package music

import (
	"context"
	"testing"
	"time"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"
)

// leaveWait is generous enough that a scheduled countdown is observed even on a
// loaded machine, without making a negative assertion slow.
const leaveWait = 2 * time.Second

func enabled(after time.Duration) IdleTimeout {
	return IdleTimeout{After: after, Enabled: true}
}

type idleFixture struct {
	service *Service
	states  *fakeVoiceStates
	voice   *linkedVoice
	lava    *fakeLavalink
}

func newIdleFixture(t *testing.T, alone IdleTimeout, empty IdleTimeout) *idleFixture {
	t.Helper()

	states := &fakeVoiceStates{}
	voice := newLinkedVoice(states)
	lava := &fakeLavalink{}

	service := NewService(ServiceConfig{
		GuildID:        testGuildID,
		ApplicationID:  selfID,
		Lavalink:       lava,
		Voice:          voice,
		VoiceStates:    states,
		Logger:         discardLogger(),
		IdleAlone:      alone,
		IdleEmptyQueue: empty,
	})
	t.Cleanup(service.Close)

	return &idleFixture{service: service, states: states, voice: voice, lava: lava}
}

// leaves counts the voice updates that disconnected the bot.
func (f *idleFixture) leaves() int {
	var n int
	for _, call := range f.voice.recorded() {
		if call.channelID == nil {
			n++
		}
	}
	return n
}

func (f *idleFixture) requireLeft(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool { return f.leaves() == 1 }, leaveWait, time.Millisecond)
}

func (f *idleFixture) requireStayed(t *testing.T) {
	t.Helper()
	require.Never(t, func() bool { return f.leaves() > 0 }, 100*time.Millisecond, time.Millisecond)
}

func TestServiceListeners(t *testing.T) {
	t.Parallel()

	otherChannelID := snowflake.ID(555555555555555555)

	tests := []struct {
		name          string
		states        []VoiceState
		wantListeners int
		wantConnected bool
	}{
		{name: "nobody in the guild"},
		{
			name:          "only the bot",
			states:        []VoiceState{inChannel(selfID, testChannelID)},
			wantConnected: true,
		},
		{
			name:          "the bot and a user",
			states:        []VoiceState{inChannel(selfID, testChannelID), inChannel(otherUserID, testChannelID)},
			wantListeners: 1,
			wantConnected: true,
		},
		{
			name:          "a user in a different channel",
			states:        []VoiceState{inChannel(selfID, testChannelID), inChannel(otherUserID, otherChannelID)},
			wantConnected: true,
		},
		{
			name:          "a user without the bot",
			states:        []VoiceState{inChannel(otherUserID, testChannelID)},
			wantListeners: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
			f.states.set(test.states...)

			require.Equal(t, test.wantListeners, f.service.listeners(testChannelID))
			if test.wantConnected {
				require.NotNil(t, f.service.botChannel())
				require.Equal(t, testChannelID, *f.service.botChannel())
				return
			}
			require.Nil(t, f.service.botChannel())
		})
	}
}

func TestServiceLeavesWhenAlone(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(0), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())

	f.requireLeft(t)
}

func TestServiceStaysWhileSomeoneIsListening(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(0), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID), inChannel(otherUserID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())

	f.requireStayed(t)
}

func TestServiceCancelsAloneCountdownWhenAUserRejoins(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())
	require.NotNil(t, f.service.aloneTimer)

	f.states.set(inChannel(selfID, testChannelID), inChannel(otherUserID, testChannelID))
	f.service.EvaluateOccupancy(context.Background())

	require.Nil(t, f.service.aloneTimer)
	f.requireStayed(t)
}

func TestServiceDoesNotRestartARunningCountdown(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())
	armed := f.service.aloneTimer
	require.NotNil(t, armed)

	for range 5 {
		f.service.EvaluateOccupancy(context.Background())
	}

	require.Same(t, armed, f.service.aloneTimer)
}

func TestServiceDoesNotArmWhileNotConnected(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(0), enabled(0))

	f.service.EvaluateOccupancy(context.Background())

	require.Nil(t, f.service.aloneTimer)
	f.requireStayed(t)
}

func TestServiceNeverLeavesWhenTheTimeoutIsOff(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, IdleTimeout{}, IdleTimeout{})
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())
	f.service.ArmEmptyQueue(context.Background())

	require.Nil(t, f.service.aloneTimer)
	require.Nil(t, f.service.emptyTimer)
	f.requireStayed(t)
}

func TestServiceArmEmptyQueue(t *testing.T) {
	t.Parallel()

	playing := &fakePlayer{track: &lavalink.Track{Encoded: "encoded"}}

	tests := []struct {
		name     string
		player   Player
		queued   []lavalink.Track
		wantArms bool
	}{
		{name: "no player and an empty queue", wantArms: true},
		{name: "a player with no track", player: &fakePlayer{}, wantArms: true},
		{name: "a player still holding a track", player: playing},
		{name: "a paused player still holding a track", player: &fakePlayer{track: playing.track, paused: true}},
		{name: "the queue still holds tracks", queued: []lavalink.Track{testTrack("next")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
			f.lava.existingPlayer = test.player
			for _, track := range test.queued {
				f.service.queue.Add(track)
			}

			f.service.ArmEmptyQueue(context.Background())

			if test.wantArms {
				require.NotNil(t, f.service.emptyTimer)
				return
			}
			require.Nil(t, f.service.emptyTimer)
		})
	}
}

func TestServiceLeavesWhenTheQueueStaysEmpty(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(0))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.ArmEmptyQueue(context.Background())

	f.requireLeft(t)
}

func TestServiceCancelEmptyQueueStopsTheCountdown(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.ArmEmptyQueue(context.Background())
	require.NotNil(t, f.service.emptyTimer)

	f.service.CancelEmptyQueue()

	require.Nil(t, f.service.emptyTimer)
	f.requireStayed(t)
}

func TestServiceLeaveIdleIsIdempotent(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.leaveIdle(context.Background(), "test")
	f.service.leaveIdle(context.Background(), "test")

	require.Equal(t, 1, f.leaves())
}

func TestServiceLeaveIdleDoesNothingWhenAlreadyGone(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))

	f.service.leaveIdle(context.Background(), "test")

	require.Zero(t, f.leaves())
}

func TestServiceLeaveIdleStopsThePlayerAndDiscardsTheQueue(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))
	player := &fakePlayer{track: &lavalink.Track{Encoded: "encoded"}}
	f.lava.existingPlayer = player
	f.service.queue.Add(testTrack("queued"))

	f.service.leaveIdle(context.Background(), "test")

	require.Nil(t, player.Track())
	require.Empty(t, f.service.Queue())
	require.Equal(t, 1, f.leaves())
}

func TestServiceBothCountdownsLeaveOnlyOnce(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(0), enabled(0))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())
	f.service.ArmEmptyQueue(context.Background())

	f.requireLeft(t)
	require.Never(t, func() bool { return f.leaves() > 1 }, 100*time.Millisecond, time.Millisecond)
}

func TestServiceCancelIdleStopsBothCountdowns(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())
	f.service.ArmEmptyQueue(context.Background())
	require.NotNil(t, f.service.aloneTimer)
	require.NotNil(t, f.service.emptyTimer)

	f.service.CancelIdle()

	require.Nil(t, f.service.aloneTimer)
	require.Nil(t, f.service.emptyTimer)
}

func TestServiceCloseStopsPendingCountdowns(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(50*time.Millisecond), enabled(50*time.Millisecond))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.EvaluateOccupancy(context.Background())
	f.service.ArmEmptyQueue(context.Background())

	f.service.Close()
	f.service.Close()

	f.requireStayed(t)
}

func TestServiceDoesNotArmAfterClose(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(0), enabled(0))
	f.states.set(inChannel(selfID, testChannelID))

	f.service.Close()
	f.service.EvaluateOccupancy(context.Background())
	f.service.ArmEmptyQueue(context.Background())

	require.Nil(t, f.service.aloneTimer)
	require.Nil(t, f.service.emptyTimer)
	f.requireStayed(t)
}

func TestServiceQueueingDuringTheCountdownKeepsTheBotInTheChannel(t *testing.T) {
	t.Parallel()

	f := newIdleFixture(t, enabled(time.Minute), enabled(time.Minute))
	f.states.set(inChannel(selfID, testChannelID), inChannel(otherUserID, testChannelID))
	f.lava.node = &fakeNode{result: &lavalink.LoadResult{Data: encodedTrack("next")}}
	e := newTestEvents(t, f.service, &fakeForwarder{}, nil)

	f.service.ArmEmptyQueue(context.Background())
	require.NotNil(t, f.service.emptyTimer)

	result, err := f.service.Enqueue(context.Background(), PlayRequest{
		Identifier:     "next",
		VoiceChannelID: new(testChannelID),
	})
	require.NoError(t, err)
	require.False(t, result.Queued, "the track must start rather than wait behind a finished one")

	e.handleTrackStart(testGuildID, "next")

	require.Nil(t, f.service.emptyTimer)
	require.Zero(t, f.leaves(), "the bot must not have left the channel")
}
