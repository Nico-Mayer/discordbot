package music

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"
)

const selfID = snowflake.ID(999999999999999999)

// fakeForwarder records the voice updates handed on to Lavalink.
type fakeForwarder struct {
	mu sync.Mutex

	stateUpdates  []voiceCall
	serverUpdates []snowflake.ID
}

func (f *fakeForwarder) OnVoiceStateUpdate(_ context.Context, guildID snowflake.ID, channelID *snowflake.ID, _ string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stateUpdates = append(f.stateUpdates, voiceCall{guildID: guildID, channelID: channelID})
}

func (f *fakeForwarder) OnVoiceServerUpdate(_ context.Context, guildID snowflake.ID, _ string, _ string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.serverUpdates = append(f.serverUpdates, guildID)
}

func (f *fakeForwarder) recordedStateUpdates() []voiceCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]voiceCall(nil), f.stateUpdates...)
}

func (f *fakeForwarder) recordedServerUpdates() []snowflake.ID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]snowflake.ID(nil), f.serverUpdates...)
}

func newTestEvents(t *testing.T, s *Service, forwarder *fakeForwarder, logger *slog.Logger) *Events {
	t.Helper()
	if logger == nil {
		logger = discardLogger()
	}
	return NewEvents(context.Background(), s, forwarder, selfID, logger)
}

func TestOnVoiceStateUpdateGuildGuard(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(444444444444444444)

	tests := []struct {
		name      string
		guildID   snowflake.ID
		wantForwa bool
	}{
		{name: "configured guild is forwarded", guildID: testGuildID, wantForwa: true},
		{name: "foreign guild is ignored", guildID: foreignGuildID, wantForwa: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			forwarder := &fakeForwarder{}
			s := newTestService(t, nil, nil)
			s.queue.Add(encodedTrack("a"))
			e := newTestEvents(t, s, forwarder, nil)

			e.handleVoiceStateUpdate(tt.guildID, selfID, &channelID, "session")

			if tt.wantForwa {
				require.Len(t, forwarder.recordedStateUpdates(), 1)
				require.Equal(t, testGuildID, forwarder.recordedStateUpdates()[0].guildID)
			} else {
				require.Empty(t, forwarder.recordedStateUpdates())
			}
			require.Equal(t, 1, s.queue.Len())
		})
	}
}

func TestOnVoiceStateUpdateIgnoresOtherMembers(t *testing.T) {
	t.Parallel()

	channelID := snowflake.ID(444444444444444444)
	forwarder := &fakeForwarder{}
	s := newTestService(t, nil, nil)
	e := newTestEvents(t, s, forwarder, nil)

	e.handleVoiceStateUpdate(testGuildID, snowflake.ID(123), &channelID, "session")
	require.Empty(t, forwarder.recordedStateUpdates())
}

func TestOnVoiceStateUpdateDiscardsTheQueueOnDisconnect(t *testing.T) {
	t.Parallel()

	forwarder := &fakeForwarder{}
	s := newTestService(t, nil, nil)
	s.queue.Add(encodedTrack("a"), encodedTrack("b"))
	e := newTestEvents(t, s, forwarder, nil)

	e.handleVoiceStateUpdate(testGuildID, selfID, nil, "session")

	require.Len(t, forwarder.recordedStateUpdates(), 1)
	require.Zero(t, s.queue.Len(), "the queue is discarded when the bot leaves voice")
}

func TestOnVoiceStateUpdateFromAForeignGuildLogsAtDebug(t *testing.T) {
	t.Parallel()

	logger, captured := newCapturingLogger(slog.LevelDebug)
	e := newTestEvents(t, newTestService(t, nil, nil), &fakeForwarder{}, logger)

	e.handleVoiceStateUpdate(foreignGuildID, selfID, nil, "session")

	records := captured.records(t)
	require.Len(t, records, 1)
	require.Equal(t, "DEBUG", records[0]["level"])
	require.Contains(t, records[0]["msg"], "unconfigured guild")
}

func TestOnVoiceServerUpdateGuildGuard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		guildID     snowflake.ID
		wantForward bool
	}{
		{name: "configured guild is forwarded", guildID: testGuildID, wantForward: true},
		{name: "foreign guild is ignored", guildID: foreignGuildID, wantForward: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			forwarder := &fakeForwarder{}
			logger, captured := newCapturingLogger(slog.LevelDebug)
			e := newTestEvents(t, newTestService(t, nil, nil), forwarder, logger)

			e.handleVoiceServerUpdate(tt.guildID, "token", "eu-central.discord.media")

			if tt.wantForward {
				require.Equal(t, []snowflake.ID{testGuildID}, forwarder.recordedServerUpdates())
				return
			}
			require.Empty(t, forwarder.recordedServerUpdates())
			records := captured.records(t)
			require.Len(t, records, 1)
			require.Equal(t, "DEBUG", records[0]["level"])
		})
	}
}

func TestForeignGuildVoiceEventsLeaveTheConfiguredStateUntouched(t *testing.T) {
	t.Parallel()

	player := &fakePlayer{track: ptr(encodedTrack("current"))}
	voice := &fakeVoice{}
	forwarder := &fakeForwarder{}
	s := newTestService(t, &fakeLavalink{existingPlayer: player}, voice)
	s.queue.Add(encodedTrack("a"), encodedTrack("b"))
	e := newTestEvents(t, s, forwarder, nil)

	channelID := snowflake.ID(444444444444444444)
	e.handleVoiceStateUpdate(foreignGuildID, selfID, &channelID, "session")
	e.handleVoiceStateUpdate(foreignGuildID, selfID, nil, "session")
	e.handleVoiceServerUpdate(foreignGuildID, "token", "endpoint")
	e.handleTrackEnd(player, foreignGuildID, lavalink.TrackEndReasonFinished)

	require.Equal(t, 2, s.queue.Len(), "the configured guild's queue must be untouched")
	require.Equal(t, "encoded-current", player.Track().Encoded, "the configured guild's player must be untouched")
	require.Zero(t, player.updateCount())
	require.Empty(t, voice.recorded(), "the bot must not leave any guild on its own")
	require.Empty(t, forwarder.recordedStateUpdates())
	require.Empty(t, forwarder.recordedServerUpdates())
}

func TestOnTrackEnd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		reason       lavalink.TrackEndReason
		queued       []lavalink.Track
		wantPlaying  string
		wantQueueLen int
		wantLeave    bool
		wantUpdates  int
	}{
		{
			name:         "advancing reason plays the next track",
			reason:       lavalink.TrackEndReasonFinished,
			queued:       []lavalink.Track{encodedTrack("next"), encodedTrack("after")},
			wantPlaying:  "encoded-next",
			wantQueueLen: 1,
			wantUpdates:  1,
		},
		{
			name:         "load failure also advances",
			reason:       lavalink.TrackEndReasonLoadFailed,
			queued:       []lavalink.Track{encodedTrack("next")},
			wantPlaying:  "encoded-next",
			wantQueueLen: 0,
			wantUpdates:  1,
		},
		{
			name:         "replaced does not consume the queue",
			reason:       lavalink.TrackEndReasonReplaced,
			queued:       []lavalink.Track{encodedTrack("next")},
			wantPlaying:  "encoded-current",
			wantQueueLen: 1,
			wantUpdates:  0,
		},
		{
			name:         "stopped does not consume the queue",
			reason:       lavalink.TrackEndReasonStopped,
			queued:       []lavalink.Track{encodedTrack("next")},
			wantPlaying:  "encoded-current",
			wantQueueLen: 1,
			wantUpdates:  0,
		},
		{
			name:        "exhausted queue leaves the voice channel",
			reason:      lavalink.TrackEndReasonFinished,
			wantPlaying: "encoded-current",
			wantLeave:   true,
			wantUpdates: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			player := &fakePlayer{track: ptr(encodedTrack("current"))}
			voice := &fakeVoice{}
			s := newTestService(t, &fakeLavalink{existingPlayer: player}, voice)
			s.queue.Add(tt.queued...)
			e := newTestEvents(t, s, &fakeForwarder{}, nil)

			e.handleTrackEnd(player, testGuildID, tt.reason)

			require.Equal(t, tt.wantPlaying, player.Track().Encoded)
			require.Equal(t, tt.wantQueueLen, s.queue.Len())
			require.Equal(t, tt.wantUpdates, player.updateCount())

			if tt.wantLeave {
				require.Len(t, voice.recorded(), 1)
				require.Nil(t, voice.recorded()[0].channelID)
			} else {
				require.Empty(t, voice.recorded())
			}
		})
	}
}

func TestOnTrackEndReasonsThatForbidAdvancing(t *testing.T) {
	t.Parallel()

	for _, reason := range []lavalink.TrackEndReason{lavalink.TrackEndReasonReplaced, lavalink.TrackEndReasonCleanup} {
		t.Run(string(reason), func(t *testing.T) {
			t.Parallel()
			require.False(t, reason.MayStartNext())

			player := &fakePlayer{track: ptr(encodedTrack("current"))}
			s := newTestService(t, &fakeLavalink{existingPlayer: player}, &fakeVoice{})
			s.queue.Add(encodedTrack("next"))
			e := newTestEvents(t, s, &fakeForwarder{}, nil)

			e.handleTrackEnd(player, testGuildID, reason)
			require.Equal(t, 1, s.queue.Len())
		})
	}
}

func TestWebSocketCloseLogging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      int
		wantLevel string
	}{
		{name: "authentication failure is terminal", code: 4004, wantLevel: "ERROR"},
		{name: "server not found is terminal", code: 4011, wantLevel: "ERROR"},
		{name: "session timeout is retried", code: 4009, wantLevel: "WARN"},
		{name: "normal closure is retried", code: 1000, wantLevel: "WARN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, captured := newCapturingLogger(slog.LevelDebug)
			e := newTestEvents(t, newTestService(t, nil, nil), &fakeForwarder{}, logger)

			e.logWebSocketClosed(testGuildID, tt.code, "because", true)

			records := captured.records(t)
			require.Len(t, records, 1)
			require.Equal(t, tt.wantLevel, records[0]["level"])
			require.EqualValues(t, tt.code, records[0]["code"])
			require.Equal(t, "because", records[0]["reason"])
			require.Equal(t, true, records[0]["by_remote"])
		})
	}
}
