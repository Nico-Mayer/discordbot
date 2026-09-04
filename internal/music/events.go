package music

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

// VoiceForwarder passes Discord's voice updates on to Lavalink.
// disgolink.Client satisfies it.
type VoiceForwarder interface {
	OnVoiceStateUpdate(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, sessionID string)
	OnVoiceServerUpdate(ctx context.Context, guildID snowflake.ID, token string, endpoint string)
}

// Lavalink close codes that no amount of reconnecting will fix.
var terminalCloseCodes = map[int]string{
	4004: "authentication failed",
	4011: "server not found",
	4012: "unknown protocol",
	4014: "disconnected from the channel",
	4016: "unknown encryption mode",
}

// Events handles the gateway and Lavalink events that drive playback.
type Events struct {
	service *Service
	voice   VoiceForwarder
	logger  *slog.Logger
	selfID  snowflake.ID

	// ctx is the process lifetime context. Holding a context in a struct is
	// normally wrong, and is deliberate here: disgo hands these handlers no
	// context of their own, and this is what makes them stop working the moment
	// shutdown begins.
	ctx context.Context
}

// NewEvents builds the event handlers bound to the process lifetime context.
func NewEvents(ctx context.Context, s *Service, voice VoiceForwarder, selfID snowflake.ID, logger *slog.Logger) *Events {
	return &Events{
		service: s,
		voice:   voice,
		logger:  logger,
		selfID:  selfID,
		ctx:     ctx,
	}
}

// Discord gateway events

func (e *Events) OnVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	e.handleVoiceStateUpdate(
		event.VoiceState.GuildID,
		event.VoiceState.UserID,
		event.VoiceState.ChannelID,
		event.VoiceState.SessionID,
	)
}

func (e *Events) handleVoiceStateUpdate(guildID snowflake.ID, userID snowflake.ID, channelID *snowflake.ID, sessionID string) {
	if guildID != e.service.GuildID() {
		e.logger.DebugContext(e.ctx, "ignoring voice state update from an unconfigured guild", slog.Any("guild", guildID))
		return
	}

	// Only the bot's own state belongs to Lavalink, but every user's movement
	// changes whether anyone is still listening, so the rest falls through.
	if userID == e.selfID {
		e.voice.OnVoiceStateUpdate(e.ctx, guildID, channelID, sessionID)
		if channelID == nil {
			e.service.CancelIdle()
			e.service.DiscardQueue()
			return
		}
	}

	e.service.EvaluateOccupancy(e.ctx)
}

func (e *Events) OnVoiceServerUpdate(event *events.VoiceServerUpdate) {
	if event.Endpoint == nil {
		e.logger.DebugContext(e.ctx, "voice server update carried no endpoint", slog.Any("guild", event.GuildID))
		return
	}
	e.handleVoiceServerUpdate(event.GuildID, event.Token, *event.Endpoint)
}

func (e *Events) handleVoiceServerUpdate(guildID snowflake.ID, token string, endpoint string) {
	if guildID != e.service.GuildID() {
		e.logger.DebugContext(e.ctx, "ignoring voice server update from an unconfigured guild", slog.Any("guild", guildID))
		return
	}
	e.voice.OnVoiceServerUpdate(e.ctx, guildID, token, endpoint)
}

// Lavalink player events

func (e *Events) OnPlayerPause(_ disgolink.Player, event lavalink.PlayerPauseEvent) {
	e.logger.InfoContext(e.ctx, "player paused", slog.Any("guild", event.GuildID()))
}

func (e *Events) OnPlayerResume(_ disgolink.Player, event lavalink.PlayerResumeEvent) {
	e.logger.InfoContext(e.ctx, "player resumed", slog.Any("guild", event.GuildID()))
}

func (e *Events) OnTrackStart(_ disgolink.Player, event lavalink.TrackStartEvent) {
	e.handleTrackStart(event.GuildID(), event.Track.Info.Title)
}

func (e *Events) handleTrackStart(guildID snowflake.ID, title string) {
	if guildID != e.service.GuildID() {
		e.logger.DebugContext(e.ctx, "ignoring track start from an unconfigured guild", slog.Any("guild", guildID))
		return
	}

	e.logger.InfoContext(e.ctx, "track started", slog.String("title", title))
	e.service.CancelEmptyQueue()
}

func (e *Events) OnTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	e.handleTrackEnd(player, event.GuildID(), event.Reason)
}

func (e *Events) handleTrackEnd(player Player, guildID snowflake.ID, reason lavalink.TrackEndReason) {
	if guildID != e.service.GuildID() {
		e.logger.DebugContext(e.ctx, "ignoring track end from an unconfigured guild", slog.Any("guild", guildID))
		return
	}
	if !reason.MayStartNext() {
		e.logger.DebugContext(e.ctx, "track end forbids starting the next one", slog.String("reason", string(reason)))
		return
	}

	advanced, err := e.service.Advance(e.ctx, player)
	if err != nil {
		e.logger.ErrorContext(e.ctx, "could not play the next track", slog.Any("err", err))
		return
	}
	if advanced {
		return
	}

	e.service.ArmEmptyQueue(e.ctx)
}

func (e *Events) OnTrackException(_ disgolink.Player, event lavalink.TrackExceptionEvent) {
	e.logger.ErrorContext(e.ctx, "track exception",
		slog.Any("guild", event.GuildID()),
		slog.String("track", event.Track.Info.Title),
		slog.Any("err", event.Exception),
	)
}

func (e *Events) OnTrackStuck(_ disgolink.Player, event lavalink.TrackStuckEvent) {
	e.logger.WarnContext(e.ctx, "track stuck",
		slog.Any("guild", event.GuildID()),
		slog.String("track", event.Track.Info.Title),
	)
}

func (e *Events) OnWebSocketClosed(_ disgolink.Player, event lavalink.WebSocketClosedEvent) {
	e.logWebSocketClosed(event.GuildID(), event.Code, event.Reason, event.ByRemote)
}

func (e *Events) logWebSocketClosed(guildID snowflake.ID, code int, reason string, byRemote bool) {
	attrs := []any{
		slog.Any("guild", guildID),
		slog.Int("code", code),
		slog.String("reason", reason),
		slog.Bool("by_remote", byRemote),
	}

	if cause, terminal := terminalCloseCodes[code]; terminal {
		e.logger.ErrorContext(e.ctx, "lavalink websocket closed for good: "+cause, attrs...)
		return
	}
	e.logger.WarnContext(e.ctx, "lavalink websocket closed, will be retried", attrs...)
}
