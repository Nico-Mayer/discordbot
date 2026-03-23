package bot

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

// Discord gateway events

func (b *Bot) OnApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	handler, ok := b.handlers[data.CommandName()]
	if !ok {
		log.Info("Unknown command", "command", data.CommandName())
		return
	}
	if err := handler(event, data); err != nil {
		log.Error("Command failed", "command", data.CommandName(), "err", err)
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
	log.Info("Player paused", "guild", event.GuildID())
}

func (b *Bot) OnPlayerResume(_ disgolink.Player, event lavalink.PlayerResumeEvent) {
	log.Info("Player resumed", "guild", event.GuildID())
}

func (b *Bot) OnTrackStart(_ disgolink.Player, event lavalink.TrackStartEvent) {
	log.Info("Track started", "title", event.Track.Info.Title)
}

func (b *Bot) OnTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	if !event.Reason.MayStartNext() {
		return
	}
	queue := b.Queues.Get(event.GuildID())
	nextTrack, ok := queue.Next()
	if !ok {
		_ = b.Client.UpdateVoiceState(context.TODO(), event.GuildID(), nil, false, false)
		return
	}
	if err := player.Update(context.TODO(), lavalink.WithTrack(nextTrack)); err != nil {
		log.Error("Failed to play next track", "err", err)
	}
}

func (b *Bot) OnTrackException(_ disgolink.Player, event lavalink.TrackExceptionEvent) {
	log.Error("Track exception", "guild", event.GuildID(), "track", event.Track.Info.Title)
}

func (b *Bot) OnTrackStuck(_ disgolink.Player, event lavalink.TrackStuckEvent) {
	log.Warn("Track stuck", "guild", event.GuildID(), "track", event.Track.Info.Title)
}

func (b *Bot) OnWebSocketClosed(_ disgolink.Player, event lavalink.WebSocketClosedEvent) {
	log.Warn("Lavalink websocket closed", "guild", event.GuildID(), "code", event.Code, "reason", event.Reason)
}
