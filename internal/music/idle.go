package music

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

// EvaluateOccupancy arms or cancels the unattended-channel countdown from the
// current voice states. The count is recomputed rather than tracked, so a voice
// state event that arrives late or not at all self-corrects on the next one.
func (s *Service) EvaluateOccupancy(ctx context.Context) {
	channelID := s.botChannel()
	if channelID == nil {
		s.cancelAlone()
		return
	}

	if s.listeners(*channelID) > 0 {
		s.cancelAlone()
		return
	}
	s.armAlone(ctx)
}

// ArmEmptyQueue starts the queue-empty countdown when there is nothing left to
// play. A paused player still holds its current track, so pausing does not start
// the countdown.
func (s *Service) ArmEmptyQueue(ctx context.Context) {
	if player := s.lavalink.ExistingPlayer(s.guildID); player != nil && player.Track() != nil {
		return
	}
	if s.queue.Len() > 0 {
		return
	}
	s.armEmptyQueue(ctx)
}

// CancelEmptyQueue stops the queue-empty countdown, for when playback resumes.
func (s *Service) CancelEmptyQueue() {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()
	stopTimer(&s.emptyTimer)
}

// CancelIdle stops both countdowns, for when the bot leaves voice by some route
// other than an idle countdown.
func (s *Service) CancelIdle() {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()
	stopTimer(&s.aloneTimer)
	stopTimer(&s.emptyTimer)
}

// Close stops both countdowns for good. A countdown still pending at shutdown
// would otherwise fire afterwards and act on an already-closed client.
func (s *Service) Close() {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()
	s.closed = true
	stopTimer(&s.aloneTimer)
	stopTimer(&s.emptyTimer)
}

// listeners counts the users in channelID other than the bot itself. Another bot
// sitting in the channel counts as a listener: telling bots apart needs the
// member cache, which needs the privileged intent this bot does not request.
func (s *Service) listeners(channelID snowflake.ID) int {
	var n int
	for state := range s.voiceStates.VoiceStates(s.guildID) {
		if state.ChannelID != nil && *state.ChannelID == channelID && state.UserID != s.applicationID {
			n++
		}
	}
	return n
}

// botChannel is the voice channel the bot is in, or nil when it is in none.
func (s *Service) botChannel() *snowflake.ID {
	for state := range s.voiceStates.VoiceStates(s.guildID) {
		if state.UserID == s.applicationID {
			return state.ChannelID
		}
	}
	return nil
}

func (s *Service) armAlone(ctx context.Context) {
	s.arm(ctx, &s.aloneTimer, s.idleAlone, "nobody is listening")
}

func (s *Service) armEmptyQueue(ctx context.Context) {
	s.arm(ctx, &s.emptyTimer, s.idleEmptyQueue, "the queue ran dry")
}

func (s *Service) cancelAlone() {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()
	stopTimer(&s.aloneTimer)
}

// arm starts a countdown, leaving an already-running one alone so that join and
// leave churn cannot keep pushing the deadline back.
func (s *Service) arm(ctx context.Context, timer **time.Timer, timeout IdleTimeout, reason string) {
	if !timeout.Enabled {
		return
	}

	s.idleMu.Lock()
	defer s.idleMu.Unlock()

	if s.closed || *timer != nil {
		return
	}

	s.logger.DebugContext(ctx, "starting idle countdown",
		slog.String("reason", reason),
		slog.Duration("after", timeout.After),
	)
	*timer = time.AfterFunc(timeout.After, func() { s.leaveIdle(ctx, reason) })
}

// leaveIdle is the one action both countdowns resolve to, so whichever elapses
// first is simply right and the other is cancelled. It re-checks its conditions
// rather than trusting the state it was armed with.
func (s *Service) leaveIdle(ctx context.Context, reason string) {
	s.idleMu.Lock()
	if s.closed || s.leaving {
		s.idleMu.Unlock()
		return
	}
	s.leaving = true
	stopTimer(&s.aloneTimer)
	stopTimer(&s.emptyTimer)
	s.idleMu.Unlock()

	defer func() {
		s.idleMu.Lock()
		s.leaving = false
		s.idleMu.Unlock()
	}()

	// The leaving flag covers two callbacks racing; this covers one that elapses
	// long after the bot already left by some other route.
	if s.botChannel() == nil {
		return
	}

	s.logger.InfoContext(ctx, "leaving the voice channel", slog.String("reason", reason))

	if player := s.lavalink.ExistingPlayer(s.guildID); player != nil {
		if err := player.Update(ctx, lavalink.WithNullTrack()); err != nil {
			s.logger.ErrorContext(ctx, "could not stop the player", slog.Any("err", err))
		}
	}
	if err := s.Leave(ctx); err != nil {
		s.logger.ErrorContext(ctx, "could not leave the voice channel", slog.Any("err", err))
	}
}

func stopTimer(timer **time.Timer) {
	if *timer == nil {
		return
	}
	(*timer).Stop()
	*timer = nil
}
