package music

import (
	"context"
	"io"
	"iter"
	"log/slog"
	"sync"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

const testGuildID = snowflake.ID(111111111111111111)
const foreignGuildID = snowflake.ID(222222222222222222)
const testChannelID = snowflake.ID(333333333333333333)
const otherUserID = snowflake.ID(444444444444444444)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// fakeLavalink stands in for the Lavalink client. Zero value: no player, no node.
type fakeLavalink struct {
	mu sync.Mutex

	player         *fakePlayer
	existingPlayer Player
	node           Node

	playerCalls         []snowflake.ID
	existingPlayerCalls []snowflake.ID
}

func (f *fakeLavalink) Player(guildID snowflake.ID) Player {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playerCalls = append(f.playerCalls, guildID)
	if f.player == nil {
		f.player = &fakePlayer{}
	}
	return f.player
}

func (f *fakeLavalink) ExistingPlayer(guildID snowflake.ID) Player {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.existingPlayerCalls = append(f.existingPlayerCalls, guildID)
	return f.existingPlayer
}

func (f *fakeLavalink) BestNode() Node {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.node
}

// fakePlayer stands in for a Lavalink player and records every update applied.
type fakePlayer struct {
	mu sync.Mutex

	track     *lavalink.Track
	paused    bool
	position  lavalink.Duration
	updateErr error

	updates    int
	lastUpdate lavalink.PlayerUpdate
}

func (f *fakePlayer) Update(_ context.Context, opts ...lavalink.PlayerUpdateOpt) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates++
	if f.updateErr != nil {
		return f.updateErr
	}

	var update lavalink.PlayerUpdate
	update.Apply(opts)
	f.lastUpdate = update

	if update.Paused != nil {
		f.paused = *update.Paused
	}
	if update.Track != nil && update.Track.Encoded != nil {
		if update.Track.Encoded.IsNull() {
			f.track = nil
		} else {
			f.track = &lavalink.Track{Encoded: update.Track.Encoded.Value()}
		}
	}
	return nil
}

func (f *fakePlayer) Track() *lavalink.Track {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.track
}

func (f *fakePlayer) Paused() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.paused
}

func (f *fakePlayer) Position() lavalink.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.position
}

func (f *fakePlayer) updateCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.updates
}

func (f *fakePlayer) lastAppliedUpdate() lavalink.PlayerUpdate {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastUpdate
}

// fakeNode stands in for a Lavalink node's track loading.
type fakeNode struct {
	mu sync.Mutex

	result *lavalink.LoadResult
	err    error
	// block, when set, makes LoadTracks wait for ctx instead of returning.
	block bool

	identifiers []string
}

func (f *fakeNode) LoadTracks(ctx context.Context, identifier string) (*lavalink.LoadResult, error) {
	f.mu.Lock()
	f.identifiers = append(f.identifiers, identifier)
	block, result, err := f.block, f.result, f.err
	f.mu.Unlock()

	if block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return result, err
}

func (f *fakeNode) loadedIdentifiers() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.identifiers...)
}

// fakeVoice stands in for the Discord gateway's voice state updates.
type fakeVoice struct {
	mu sync.Mutex

	err   error
	calls []voiceCall
}

type voiceCall struct {
	guildID   snowflake.ID
	channelID *snowflake.ID
}

func (f *fakeVoice) UpdateVoiceState(_ context.Context, guildID snowflake.ID, channelID *snowflake.ID, _ bool, _ bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, voiceCall{guildID: guildID, channelID: channelID})
	return f.err
}

func (f *fakeVoice) recorded() []voiceCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]voiceCall(nil), f.calls...)
}

var _ VoiceStates = (*fakeVoiceStates)(nil)

// fakeVoiceStates stands in for the Discord voice state cache. Zero value: the
// guild has nobody in any voice channel.
type fakeVoiceStates struct {
	mu     sync.Mutex
	states []VoiceState
}

func (f *fakeVoiceStates) VoiceStates(snowflake.ID) iter.Seq[VoiceState] {
	f.mu.Lock()
	states := append([]VoiceState(nil), f.states...)
	f.mu.Unlock()

	return func(yield func(VoiceState) bool) {
		for _, state := range states {
			if !yield(state) {
				return
			}
		}
	}
}

func (f *fakeVoiceStates) set(states ...VoiceState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states = states
}

func inChannel(userID snowflake.ID, channelID snowflake.ID) VoiceState {
	return VoiceState{UserID: userID, ChannelID: &channelID}
}

// linkedVoice keeps a fakeVoiceStates in step with the joins and leaves the
// service issues, the way the gateway cache does in production.
type linkedVoice struct {
	fakeVoice
	states *fakeVoiceStates
}

func newLinkedVoice(states *fakeVoiceStates) *linkedVoice {
	return &linkedVoice{states: states}
}

func (l *linkedVoice) UpdateVoiceState(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, mute bool, deaf bool) error {
	err := l.fakeVoice.UpdateVoiceState(ctx, guildID, channelID, mute, deaf)
	if err != nil {
		return err
	}

	l.states.mu.Lock()
	defer l.states.mu.Unlock()

	kept := l.states.states[:0]
	for _, state := range l.states.states {
		if state.UserID != selfID {
			kept = append(kept, state)
		}
	}
	if channelID != nil {
		kept = append(kept, VoiceState{UserID: selfID, ChannelID: channelID})
	}
	l.states.states = kept
	return nil
}
