package music

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"
)

func encodedTrack(title string) lavalink.Track {
	track := testTrack(title)
	track.Encoded = "encoded-" + title
	return track
}

func newTestService(t *testing.T, lava *fakeLavalink, voice *fakeVoice) *Service {
	t.Helper()
	if lava == nil {
		lava = &fakeLavalink{}
	}
	if voice == nil {
		voice = &fakeVoice{}
	}
	return NewService(testGuildID, lava, voice, discardLogger())
}

func TestNewServiceSmoke(t *testing.T) {
	t.Parallel()

	s := newTestService(t, nil, nil)

	require.Equal(t, testGuildID, s.GuildID())
	require.Empty(t, s.Queue())
	require.Equal(t, defaultLoadTimeout, s.loadTimeout)
}

func TestServicePause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		player      *fakePlayer
		wantPaused  bool
		wantErr     error
		wantUpdates int
	}{
		{
			name:        "pauses a playing track",
			player:      &fakePlayer{paused: false},
			wantPaused:  true,
			wantUpdates: 1,
		},
		{
			name:        "resumes a paused track",
			player:      &fakePlayer{paused: true},
			wantPaused:  false,
			wantUpdates: 1,
		},
		{
			name:    "no player",
			wantErr: ErrNoPlayer,
		},
		{
			name:        "update fails",
			player:      &fakePlayer{updateErr: errors.New("node down")},
			wantErr:     errors.New("node down"),
			wantUpdates: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lava := &fakeLavalink{}
			if tt.player != nil {
				lava.existingPlayer = tt.player
			}
			s := newTestService(t, lava, nil)
			s.queue.Add(encodedTrack("queued"))

			paused, err := s.Pause(context.Background())

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrNoPlayer) {
					require.ErrorIs(t, err, ErrNoPlayer)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantPaused, paused)
				require.Equal(t, tt.wantPaused, tt.player.Paused())
			}

			require.Equal(t, 1, s.queue.Len(), "pause must leave the queue untouched")
			if tt.player != nil {
				require.Equal(t, tt.wantUpdates, tt.player.updateCount())
			}
		})
	}
}

func TestServiceStop(t *testing.T) {
	t.Parallel()

	t.Run("stops, leaves voice and clears the queue", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{track: ptr(encodedTrack("current"))}
		lava := &fakeLavalink{existingPlayer: player}
		voice := &fakeVoice{}
		s := newTestService(t, lava, voice)
		s.queue.Add(encodedTrack("a"), encodedTrack("b"))

		require.NoError(t, s.Stop(context.Background()))

		require.Zero(t, s.queue.Len())
		require.Nil(t, player.Track())
		require.Len(t, voice.recorded(), 1)
		require.Equal(t, testGuildID, voice.recorded()[0].guildID)
		require.Nil(t, voice.recorded()[0].channelID, "a nil channel leaves the voice channel")
	})

	t.Run("no player", func(t *testing.T) {
		t.Parallel()

		voice := &fakeVoice{}
		s := newTestService(t, &fakeLavalink{}, voice)
		s.queue.Add(encodedTrack("a"))

		require.ErrorIs(t, s.Stop(context.Background()), ErrNoPlayer)
		require.Equal(t, 1, s.queue.Len())
		require.Empty(t, voice.recorded())
	})

	t.Run("failing to null the track leaves no partial state", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{track: ptr(encodedTrack("current")), updateErr: errors.New("node down")}
		voice := &fakeVoice{}
		s := newTestService(t, &fakeLavalink{existingPlayer: player}, voice)
		s.queue.Add(encodedTrack("a"))

		err := s.Stop(context.Background())
		require.Error(t, err)
		require.Equal(t, 1, s.queue.Len(), "the queue must survive a failed stop")
		require.Empty(t, voice.recorded(), "the bot must stay in the channel")
	})

	t.Run("failing to leave voice leaves the queue intact", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{track: ptr(encodedTrack("current"))}
		voice := &fakeVoice{err: errors.New("gateway down")}
		s := newTestService(t, &fakeLavalink{existingPlayer: player}, voice)
		s.queue.Add(encodedTrack("a"))

		err := s.Stop(context.Background())
		require.Error(t, err)
		require.Equal(t, 1, s.queue.Len(), "the queue must survive a failed stop")
	})
}

func TestServiceSkip(t *testing.T) {
	t.Parallel()

	t.Run("plays the next queued track", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{track: ptr(encodedTrack("current"))}
		s := newTestService(t, &fakeLavalink{existingPlayer: player}, nil)
		s.queue.Add(encodedTrack("next"), encodedTrack("after"))

		got, err := s.Skip(context.Background())
		require.NoError(t, err)
		require.Equal(t, "next", got.Info.Title)
		require.Equal(t, "encoded-next", player.Track().Encoded)
		require.Equal(t, 1, s.queue.Len())
	})

	t.Run("empty queue keeps the current track playing", func(t *testing.T) {
		t.Parallel()

		current := encodedTrack("current")
		player := &fakePlayer{track: &current}
		s := newTestService(t, &fakeLavalink{existingPlayer: player}, nil)

		_, err := s.Skip(context.Background())
		require.ErrorIs(t, err, ErrQueueEmpty)
		require.Equal(t, "encoded-current", player.Track().Encoded)
		require.Zero(t, player.updateCount(), "the player must not be touched")
	})

	t.Run("no player keeps the queue intact", func(t *testing.T) {
		t.Parallel()

		s := newTestService(t, &fakeLavalink{}, nil)
		s.queue.Add(encodedTrack("next"))

		_, err := s.Skip(context.Background())
		require.ErrorIs(t, err, ErrNoPlayer)
		require.Equal(t, 1, s.queue.Len(), "the queue must not be consumed without a player")
	})

	t.Run("update failure", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{track: ptr(encodedTrack("current")), updateErr: errors.New("node down")}
		s := newTestService(t, &fakeLavalink{existingPlayer: player}, nil)
		s.queue.Add(encodedTrack("next"))

		_, err := s.Skip(context.Background())
		require.ErrorContains(t, err, "node down")
	})
}

func TestServiceNowPlaying(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		player  *fakePlayer
		wantErr error
		want    string
	}{
		{
			name:   "a track is playing",
			player: &fakePlayer{track: ptr(encodedTrack("current")), position: 30 * lavalink.Second},
			want:   "current",
		},
		{
			name:    "player exists but is idle",
			player:  &fakePlayer{},
			wantErr: ErrNothingPlaying,
		},
		{
			name:    "no player",
			wantErr: ErrNoPlayer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lava := &fakeLavalink{}
			if tt.player != nil {
				lava.existingPlayer = tt.player
			}
			s := newTestService(t, lava, nil)

			track, position, err := s.NowPlaying()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, position)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, track.Info.Title)
			require.Equal(t, 30*lavalink.Second, position)
		})
	}
}

func TestServiceQueue(t *testing.T) {
	t.Parallel()

	t.Run("empty queue reports empty", func(t *testing.T) {
		t.Parallel()

		s := newTestService(t, nil, nil)
		require.Empty(t, s.Queue())
	})

	t.Run("returns a copy in play order", func(t *testing.T) {
		t.Parallel()

		s := newTestService(t, nil, nil)
		s.queue.Add(encodedTrack("a"), encodedTrack("b"))

		got := s.Queue()
		require.Equal(t, []string{"a", "b"}, []string{got[0].Info.Title, got[1].Info.Title})

		got[0] = encodedTrack("hijacked")
		require.Equal(t, "a", s.Queue()[0].Info.Title, "the caller must not be able to mutate the queue")
	})
}

func TestServiceEnqueue(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(333333333333333333)

	trackResult := &lavalink.LoadResult{LoadType: lavalink.LoadTypeTrack, Data: encodedTrack("loaded")}
	playlistResult := func(tracks ...lavalink.Track) *lavalink.LoadResult {
		return &lavalink.LoadResult{LoadType: lavalink.LoadTypePlaylist, Data: lavalink.Playlist{Tracks: tracks}}
	}
	searchResult := func(tracks ...lavalink.Track) *lavalink.LoadResult {
		return &lavalink.LoadResult{LoadType: lavalink.LoadTypeSearch, Data: lavalink.Search(tracks)}
	}

	tests := []struct {
		name         string
		identifier   string
		voiceChannel *snowflake.ID
		node         *fakeNode
		currentTrack *lavalink.Track
		voiceErr     error
		playerErr    error

		wantErr        error
		wantErrMsg     string
		wantTitle      string
		wantQueued     bool
		wantPosition   int
		wantQueueLen   int
		wantIdentifier string
		// wantJoin records that a join was attempted, successfully or not.
		wantJoin bool
	}{
		{
			name:         "not in a voice channel",
			identifier:   "song",
			voiceChannel: nil,
			node:         &fakeNode{result: trackResult},
			wantErr:      ErrNotInVoice,
		},
		{
			name:           "search query plays now",
			identifier:     "never gonna give you up",
			voiceChannel:   &channelID,
			node:           &fakeNode{result: searchResult(encodedTrack("first"), encodedTrack("second"))},
			wantTitle:      "first",
			wantIdentifier: lavalink.SearchTypeYouTubeMusic.Apply("never gonna give you up"),
			wantJoin:       true,
		},
		{
			name:           "url is loaded as given",
			identifier:     "https://example.com/track",
			voiceChannel:   &channelID,
			node:           &fakeNode{result: trackResult},
			wantTitle:      "loaded",
			wantIdentifier: "https://example.com/track",
			wantJoin:       true,
		},
		{
			name:           "playlist uses its first track only",
			identifier:     "https://example.com/playlist",
			voiceChannel:   &channelID,
			node:           &fakeNode{result: playlistResult(encodedTrack("one"), encodedTrack("two"), encodedTrack("three"))},
			wantTitle:      "one",
			wantIdentifier: "https://example.com/playlist",
			wantJoin:       true,
		},
		{
			name:         "empty playlist reports nothing found",
			identifier:   "https://example.com/playlist",
			voiceChannel: &channelID,
			node:         &fakeNode{result: playlistResult()},
			wantErr:      ErrNoResults,
			wantErrMsg:   "https://example.com/playlist",
		},
		{
			name:         "empty search reports nothing found",
			identifier:   "obscure",
			voiceChannel: &channelID,
			node:         &fakeNode{result: searchResult()},
			wantErr:      ErrNoResults,
		},
		{
			name:         "empty load type reports nothing found",
			identifier:   "obscure",
			voiceChannel: &channelID,
			node:         &fakeNode{result: &lavalink.LoadResult{LoadType: lavalink.LoadTypeEmpty, Data: lavalink.Empty{}}},
			wantErr:      ErrNoResults,
		},
		{
			name:         "lavalink exception",
			identifier:   "song",
			voiceChannel: &channelID,
			node: &fakeNode{result: &lavalink.LoadResult{
				LoadType: lavalink.LoadTypeError,
				Data:     lavalink.Exception{Message: "track unavailable", Severity: lavalink.SeverityCommon},
			}},
			wantErrMsg: "track unavailable",
		},
		{
			name:         "load error",
			identifier:   "song",
			voiceChannel: &channelID,
			node:         &fakeNode{err: errors.New("node unreachable")},
			wantErrMsg:   "node unreachable",
		},
		{
			name:         "a track is already playing so the new one is queued",
			identifier:   "song",
			voiceChannel: &channelID,
			node:         &fakeNode{result: trackResult},
			currentTrack: ptr(encodedTrack("current")),
			wantTitle:    "loaded",
			wantQueued:   true,
			wantPosition: 1,
			wantQueueLen: 1,
			wantJoin:     true,
		},
		{
			name:         "joining voice fails",
			identifier:   "song",
			voiceChannel: &channelID,
			node:         &fakeNode{result: trackResult},
			voiceErr:     errors.New("gateway down"),
			wantErrMsg:   "join voice channel",
			wantJoin:     true,
		},
		{
			name:         "starting the track fails",
			identifier:   "song",
			voiceChannel: &channelID,
			node:         &fakeNode{result: trackResult},
			playerErr:    errors.New("node down"),
			wantErrMsg:   "start track",
			wantJoin:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lava := &fakeLavalink{
				node:   tt.node,
				player: &fakePlayer{track: tt.currentTrack, updateErr: tt.playerErr},
			}
			voice := &fakeVoice{err: tt.voiceErr}
			s := newTestService(t, lava, voice)

			result, err := s.Enqueue(context.Background(), PlayRequest{
				Identifier:     tt.identifier,
				VoiceChannelID: tt.voiceChannel,
			})

			if tt.wantErr != nil || tt.wantErrMsg != "" {
				require.Error(t, err)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
				}
				if tt.wantErrMsg != "" {
					require.ErrorContains(t, err, tt.wantErrMsg)
				}
				require.Zero(t, s.queue.Len(), "a failure must not modify the queue")
				if !tt.wantJoin {
					require.Empty(t, voice.recorded(), "a failure before joining must not join")
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantTitle, result.Track.Info.Title)
			require.Equal(t, tt.wantQueued, result.Queued)
			require.Equal(t, tt.wantPosition, result.Position)
			require.Equal(t, tt.wantQueueLen, s.queue.Len())

			require.Len(t, voice.recorded(), 1)
			require.Equal(t, &channelID, voice.recorded()[0].channelID)

			if tt.wantIdentifier != "" {
				require.Equal(t, []string{tt.wantIdentifier}, tt.node.loadedIdentifiers())
			}
		})
	}
}

func TestServiceEnqueueQueuePositionsCountUp(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(333333333333333333)
	lava := &fakeLavalink{
		node:   &fakeNode{result: &lavalink.LoadResult{LoadType: lavalink.LoadTypeTrack, Data: encodedTrack("loaded")}},
		player: &fakePlayer{track: ptr(encodedTrack("current"))},
	}
	s := newTestService(t, lava, &fakeVoice{})

	for want := 1; want <= 3; want++ {
		result, err := s.Enqueue(context.Background(), PlayRequest{Identifier: "song", VoiceChannelID: &channelID})
		require.NoError(t, err)
		require.True(t, result.Queued)
		require.Equal(t, want, result.Position)
	}
}

func TestServiceEnqueueWithoutANode(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(333333333333333333)
	s := newTestService(t, &fakeLavalink{}, nil)

	_, err := s.Enqueue(context.Background(), PlayRequest{Identifier: "song", VoiceChannelID: &channelID})
	require.ErrorIs(t, err, ErrNoNode)
}

func TestServiceEnqueueRespectsACancelledContext(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(333333333333333333)
	lava := &fakeLavalink{node: &fakeNode{block: true}}
	s := newTestService(t, lava, &fakeVoice{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() {
		_, err := s.Enqueue(ctx, PlayRequest{Identifier: "song", VoiceChannelID: &channelID})
		done <- err
	}()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("Enqueue hung on an already-cancelled context")
	}
}

func TestServiceEnqueueBoundsTheLoadWithATimeout(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(333333333333333333)
	lava := &fakeLavalink{node: &fakeNode{block: true}}
	s := newTestService(t, lava, &fakeVoice{})
	s.loadTimeout = 10 * time.Millisecond

	done := make(chan error, 1)
	go func() {
		_, err := s.Enqueue(context.Background(), PlayRequest{Identifier: "song", VoiceChannelID: &channelID})
		done <- err
	}()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(5 * time.Second):
		t.Fatal("Enqueue ignored its load timeout")
	}
}

func TestServiceAdvance(t *testing.T) {
	t.Parallel()

	t.Run("plays the next queued track", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{}
		s := newTestService(t, nil, nil)
		s.queue.Add(encodedTrack("next"))

		advanced, err := s.Advance(context.Background(), player)
		require.NoError(t, err)
		require.True(t, advanced)
		require.Equal(t, "encoded-next", player.Track().Encoded)
		require.Zero(t, s.queue.Len())
	})

	t.Run("reports false on an exhausted queue", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{}
		s := newTestService(t, nil, nil)

		advanced, err := s.Advance(context.Background(), player)
		require.NoError(t, err)
		require.False(t, advanced)
		require.Zero(t, player.updateCount())
	})

	t.Run("wraps an update failure", func(t *testing.T) {
		t.Parallel()

		player := &fakePlayer{updateErr: errors.New("node down")}
		s := newTestService(t, nil, nil)
		s.queue.Add(encodedTrack("next"))

		_, err := s.Advance(context.Background(), player)
		require.ErrorContains(t, err, "node down")
	})
}

func TestServiceLeave(t *testing.T) {
	t.Parallel()

	voice := &fakeVoice{}
	s := newTestService(t, nil, voice)
	s.queue.Add(encodedTrack("a"))

	require.NoError(t, s.Leave(context.Background()))
	require.Zero(t, s.queue.Len())
	require.Len(t, voice.recorded(), 1)
	require.Nil(t, voice.recorded()[0].channelID)
}

func TestServiceDiscardQueue(t *testing.T) {
	t.Parallel()

	s := newTestService(t, nil, nil)
	s.queue.Add(encodedTrack("a"), encodedTrack("b"))

	s.DiscardQueue()
	require.Zero(t, s.queue.Len())
}

func ptr[T any](v T) *T { return &v }
