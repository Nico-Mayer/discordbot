package bot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

// Discord gateway events

func (b *Bot) OnApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	handler, ok := b.handlers[data.CommandName()]
	if !ok {
		slog.Info("unknown command", slog.String("command", data.CommandName()))
		return
	}
	if err := handler(event, data); err != nil {
		slog.Error("error handling command", slog.Any("err", err))
	}
}

func (b *Bot) OnVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.UserID != b.Client.ApplicationID {
		return
	}
	b.Lavalink.OnVoiceStateUpdate(context.TODO(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
	if event.VoiceState.ChannelID == nil {
		b.Queues.Delete(event.VoiceState.GuildID)
	}
}

func (b *Bot) OnVoiceServerUpdate(event *events.VoiceServerUpdate) {
	b.Lavalink.OnVoiceServerUpdate(context.TODO(), event.GuildID, event.Token, *event.Endpoint)
}

// Lavalink player events

func (b *Bot) OnPlayerPause(_ disgolink.Player, event lavalink.PlayerPauseEvent) {
	slog.Info("player paused", slog.Any("guild", event.GuildID()))
}

func (b *Bot) OnPlayerResume(_ disgolink.Player, event lavalink.PlayerResumeEvent) {
	slog.Info("player resumed", slog.Any("guild", event.GuildID()))
}

func (b *Bot) OnTrackStart(_ disgolink.Player, event lavalink.TrackStartEvent) {
	slog.Info("track started", slog.String("title", event.Track.Info.Title))
}

func (b *Bot) OnTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	if !event.Reason.MayStartNext() {
		return
	}
	queue := b.Queues.Get(event.GuildID())
	nextTrack, ok := queue.Next()
	if !ok {
		return
	}
	if err := player.Update(context.TODO(), lavalink.WithTrack(nextTrack)); err != nil {
		slog.Error("failed to play next track", slog.Any("err", err))
	}
}

func (b *Bot) OnTrackException(_ disgolink.Player, event lavalink.TrackExceptionEvent) {
	slog.Error("track exception", slog.Any("event", event))
}

func (b *Bot) OnTrackStuck(_ disgolink.Player, event lavalink.TrackStuckEvent) {
	slog.Warn("track stuck", slog.Any("event", event))
}

func (b *Bot) OnWebSocketClosed(_ disgolink.Player, event lavalink.WebSocketClosedEvent) {
	slog.Warn("websocket closed", slog.Any("event", event))
}
