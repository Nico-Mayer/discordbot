package bot

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

var urlPattern = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")

func (b *Bot) play(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	identifier := data.String("identifier")
	if source, ok := data.OptString("source"); ok {
		identifier = lavalink.SearchType(source).Apply(identifier)
	} else if !urlPattern.MatchString(identifier) {
		identifier = lavalink.SearchTypeYouTubeMusic.Apply(identifier)
	}

	voiceState, ok := b.Client.Caches.VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return errorReply(event, "Du musst in einem Sprachkanal sein!")
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track
	b.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			toPlay = &track
		},
		func(playlist lavalink.Playlist) {
			toPlay = &playlist.Tracks[0]
		},
		func(tracks []lavalink.Track) {
			toPlay = &tracks[0]
		},
		func() {
			b.deferredEmbedReply(event, 0xED4245, fmt.Sprintf("❌ Nichts gefunden für: `%s`", identifier))
		},
		func(err error) {
			b.deferredEmbedReply(event, 0xED4245, fmt.Sprintf("❌ Fehler beim Laden: `%s`", err))
		},
	))
	if toPlay == nil {
		return nil
	}

	if err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), voiceState.ChannelID, false, false); err != nil {
		return err
	}

	queue := b.Queues.Get(*event.GuildID())
	player := b.Lavalink.Player(*event.GuildID())

	if player.Track() != nil {
		queue.Add(*toPlay)
		b.deferredQueuedReply(event, toPlay, len(queue.Tracks))
		return nil
	}

	b.deferredTrackReply(event, toPlay)
	return player.Update(context.TODO(), lavalink.WithTrack(*toPlay))
}

func (b *Bot) pause(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return errorReply(event, "Kein Player gefunden")
	}

	if err := player.Update(context.TODO(), lavalink.WithPaused(!player.Paused())); err != nil {
		return errorReply(event, fmt.Sprintf("Fehler beim Pausieren: `%s`", err))
	}

	if player.Paused() {
		return infoReply(event, "⏸️ Wiedergabe pausiert")
	}
	return infoReply(event, "▶️ Wiedergabe fortgesetzt")
}

func (b *Bot) stop(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return errorReply(event, "Kein Player gefunden")
	}

	b.Queues.Get(*event.GuildID()).Clear()

	if err := player.Update(context.TODO(), lavalink.WithNullTrack()); err != nil {
		return errorReply(event, fmt.Sprintf("Fehler beim Stoppen: `%s`", err))
	}

	if err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), nil, false, false); err != nil {
		return errorReply(event, fmt.Sprintf("Fehler beim Trennen: `%s`", err))
	}

	return successReply(event, "Wiedergabe gestoppt und Warteschlange geleert")
}

func (b *Bot) skip(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	queue := b.Queues.Get(*event.GuildID())
	if player == nil {
		return errorReply(event, "Kein Player gefunden")
	}

	track, ok := queue.Next()
	if !ok {
		return errorReply(event, "Keine weiteren Titel in der Warteschlange")
	}

	if err := player.Update(context.TODO(), lavalink.WithTrack(track)); err != nil {
		return errorReply(event, fmt.Sprintf("Fehler beim Überspringen: `%s`", err))
	}

	return successReply(event, "Titel übersprungen ⏭️")
}

func (b *Bot) nowPlaying(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return errorReply(event, "Kein Player gefunden")
	}

	track := player.Track()
	if track == nil {
		return errorReply(event, "Es wird gerade nichts abgespielt")
	}

	embed := discord.Embed{
		Title:       "🎶 Läuft gerade",
		Description: fmt.Sprintf("[%s](<%s>)\nvon **%s**", track.Info.Title, *track.Info.URI, track.Info.Author),
		Color:       0x5865F2,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%s / %s", formatPosition(player.Position()), formatPosition(track.Info.Length)),
		},
	}
	if track.Info.ArtworkURL != nil {
		embed.Thumbnail = &discord.EmbedResource{URL: *track.Info.ArtworkURL}
	}
	return reply(event, embed)
}

func (b *Bot) queue(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	queue := b.Queues.Get(*event.GuildID())
	if len(queue.Tracks) == 0 {
		return infoReply(event, "📋 Die Warteschlange ist leer")
	}

	var tracks string
	for i, track := range queue.Tracks {
		tracks += fmt.Sprintf("`%d.` [%s](<%s>) · %s\n", i+1, track.Info.Title, *track.Info.URI, formatPosition(track.Info.Length))
	}

	return reply(event, discord.Embed{
		Title:       "📋 Warteschlange",
		Description: tracks,
		Color:       0x5865F2,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%d Titel in der Warteschlange", len(queue.Tracks)),
		},
	})
}

func formatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}
