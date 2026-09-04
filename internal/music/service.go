// Package music holds the bot's playback logic, its Discord command surface and
// the embeds it replies with.
package music

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/nico-mayer/discordbot/internal/queue"
)

// defaultLoadTimeout bounds a track load so a slow node cannot leave an
// interaction unanswered past Discord's window.
const defaultLoadTimeout = 10 * time.Second

// ErrNoNode reports that no Lavalink node is currently usable.
var ErrNoNode = errors.New("no lavalink node available")

// Service decides what playback commands do. It holds the single guild's queue
// and reaches Lavalink and the gateway only through the seams above, so every
// method is unit-testable.
type Service struct {
	guildID       snowflake.ID
	applicationID snowflake.ID
	lavalink      Lavalink
	voice         Voice
	voiceStates   VoiceStates
	queue         *queue.Queue
	logger        *slog.Logger
	loadTimeout   time.Duration

	idleAlone      IdleTimeout
	idleEmptyQueue IdleTimeout

	idleMu     sync.Mutex
	aloneTimer *time.Timer
	emptyTimer *time.Timer
	leaving    bool
	closed     bool
}

// IdleTimeout is how long the service waits before leaving for one idle reason.
// Enabled is false when the operator turned that reason off.
type IdleTimeout struct {
	After   time.Duration
	Enabled bool
}

// ServiceConfig is everything NewService needs. It is a struct because the
// seams, the two idle timeouts and the two IDs would otherwise be eight
// positional arguments of largely interchangeable types.
type ServiceConfig struct {
	GuildID       snowflake.ID
	ApplicationID snowflake.ID
	Lavalink      Lavalink
	Voice         Voice
	VoiceStates   VoiceStates
	Logger        *slog.Logger

	IdleAlone      IdleTimeout
	IdleEmptyQueue IdleTimeout
}

// NewService builds the service for the one guild the bot serves.
func NewService(cfg ServiceConfig) *Service {
	return &Service{
		guildID:        cfg.GuildID,
		applicationID:  cfg.ApplicationID,
		lavalink:       cfg.Lavalink,
		voice:          cfg.Voice,
		voiceStates:    cfg.VoiceStates,
		queue:          &queue.Queue{},
		logger:         cfg.Logger,
		loadTimeout:    defaultLoadTimeout,
		idleAlone:      cfg.IdleAlone,
		idleEmptyQueue: cfg.IdleEmptyQueue,
	}
}

// GuildID is the only guild the service acts on.
func (s *Service) GuildID() snowflake.ID { return s.guildID }

// PlayRequest is a /play invocation reduced to plain values. VoiceChannelID is
// nil when the caller is not connected to a voice channel.
type PlayRequest struct {
	Identifier     string
	VoiceChannelID *snowflake.ID
}

// PlayResult reports what Enqueue did with the resolved track.
type PlayResult struct {
	Track lavalink.Track
	// Queued is false when the track started playing immediately.
	Queued bool
	// Position is the track's 1-based place in the queue when Queued is true.
	Position int
}

// Enqueue resolves an identifier, joins the caller's voice channel and either
// starts the track or appends it to the queue.
func (s *Service) Enqueue(ctx context.Context, req PlayRequest) (PlayResult, error) {
	if req.VoiceChannelID == nil {
		return PlayResult{}, ErrNotInVoice
	}

	track, err := s.loadTrack(ctx, req.Identifier)
	if err != nil {
		return PlayResult{}, err
	}

	if err := s.voice.UpdateVoiceState(ctx, s.guildID, req.VoiceChannelID, false, false); err != nil {
		return PlayResult{}, fmt.Errorf("join voice channel: %w", err)
	}

	player := s.lavalink.Player(s.guildID)
	if player.Track() != nil {
		s.queue.Add(track)
		return PlayResult{Track: track, Queued: true, Position: s.queue.Len()}, nil
	}

	if err := player.Update(ctx, lavalink.WithTrack(track)); err != nil {
		return PlayResult{}, fmt.Errorf("start track: %w", err)
	}
	return PlayResult{Track: track}, nil
}

// Pause toggles playback and reports the state it moved to.
func (s *Service) Pause(ctx context.Context) (bool, error) {
	player := s.lavalink.ExistingPlayer(s.guildID)
	if player == nil {
		return false, ErrNoPlayer
	}

	paused := !player.Paused()
	if err := player.Update(ctx, lavalink.WithPaused(paused)); err != nil {
		return false, fmt.Errorf("set paused to %t: %w", paused, err)
	}
	return paused, nil
}

// Stop stops the current track, leaves the voice channel and clears the queue.
// The queue survives a failure, so a failed stop leaves nothing half-applied.
func (s *Service) Stop(ctx context.Context) error {
	player := s.lavalink.ExistingPlayer(s.guildID)
	if player == nil {
		return ErrNoPlayer
	}

	if err := player.Update(ctx, lavalink.WithNullTrack()); err != nil {
		return fmt.Errorf("stop player: %w", err)
	}
	if err := s.voice.UpdateVoiceState(ctx, s.guildID, nil, false, false); err != nil {
		return fmt.Errorf("leave voice channel: %w", err)
	}

	s.queue.Clear()
	return nil
}

// Skip replaces the current track with the next queued one. The queue keeps its
// tracks when there is no player, and the current track keeps playing when the
// queue is empty.
func (s *Service) Skip(ctx context.Context) (lavalink.Track, error) {
	player := s.lavalink.ExistingPlayer(s.guildID)
	if player == nil {
		return lavalink.Track{}, ErrNoPlayer
	}

	next, ok := s.queue.Next()
	if !ok {
		return lavalink.Track{}, ErrQueueEmpty
	}

	if err := player.Update(ctx, lavalink.WithTrack(next)); err != nil {
		return lavalink.Track{}, fmt.Errorf("skip to next track: %w", err)
	}
	return next, nil
}

// NowPlaying returns the current track and how far into it playback is.
func (s *Service) NowPlaying() (lavalink.Track, lavalink.Duration, error) {
	player := s.lavalink.ExistingPlayer(s.guildID)
	if player == nil {
		return lavalink.Track{}, 0, ErrNoPlayer
	}

	track := player.Track()
	if track == nil {
		return lavalink.Track{}, 0, ErrNothingPlaying
	}
	return *track, player.Position(), nil
}

// Current returns the track playing right now. Unlike NowPlaying it reports
// absence as a boolean, because a caller that only wants to name the current
// track - /queue does - is not in an error case when there is none.
func (s *Service) Current() (lavalink.Track, bool) {
	player := s.lavalink.ExistingPlayer(s.guildID)
	if player == nil {
		return lavalink.Track{}, false
	}

	track := player.Track()
	if track == nil {
		return lavalink.Track{}, false
	}
	return *track, true
}

// Queue returns the queued tracks in play order as a copy. The reply is bounded
// by queueEmbed, which needs the full count for its footer.
func (s *Service) Queue() []lavalink.Track {
	return s.queue.Tracks()
}

// Advance plays the next queued track, reporting false when the queue is empty.
func (s *Service) Advance(ctx context.Context, player Player) (bool, error) {
	next, ok := s.queue.Next()
	if !ok {
		return false, nil
	}
	if err := player.Update(ctx, lavalink.WithTrack(next)); err != nil {
		return false, fmt.Errorf("play next track: %w", err)
	}
	return true, nil
}

// Leave disconnects from the guild's voice channel and discards the queue.
func (s *Service) Leave(ctx context.Context) error {
	s.queue.Clear()
	if err := s.voice.UpdateVoiceState(ctx, s.guildID, nil, false, false); err != nil {
		return fmt.Errorf("leave voice channel: %w", err)
	}
	return nil
}

// DiscardQueue drops every queued track, for when the bot is disconnected from
// voice by something other than a command.
func (s *Service) DiscardQueue() {
	s.queue.Clear()
}

func (s *Service) loadTrack(ctx context.Context, identifier string) (lavalink.Track, error) {
	node := s.lavalink.BestNode()
	if node == nil {
		return lavalink.Track{}, ErrNoNode
	}

	ctx, cancel := context.WithTimeout(ctx, s.loadTimeout)
	defer cancel()

	result, err := node.LoadTracks(ctx, resolveIdentifier(identifier))
	if err != nil {
		return lavalink.Track{}, &LoadError{Identifier: identifier, Err: err}
	}
	if result == nil {
		return lavalink.Track{}, &NoResultsError{Identifier: identifier}
	}

	switch data := result.Data.(type) {
	case lavalink.Track:
		return data, nil
	case lavalink.Playlist:
		// Playlists deliberately contribute only their first track; enqueueing
		// the rest is the separate support-playlists change.
		if len(data.Tracks) == 0 {
			return lavalink.Track{}, &NoResultsError{Identifier: identifier}
		}
		return data.Tracks[0], nil
	case lavalink.Search:
		if len(data) == 0 {
			return lavalink.Track{}, &NoResultsError{Identifier: identifier}
		}
		return data[0], nil
	case lavalink.Empty:
		return lavalink.Track{}, &NoResultsError{Identifier: identifier}
	case lavalink.Exception:
		return lavalink.Track{}, &LoadError{Identifier: identifier, Err: data}
	default:
		return lavalink.Track{}, fmt.Errorf("unexpected load type %q for %q: %w", result.LoadType, identifier, ErrNoResults)
	}
}
