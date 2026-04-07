package bot

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func reply(event *events.ApplicationCommandInteractionCreate, embed discord.Embed) error {
	return event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
}

func ephemeralReply(event *events.ApplicationCommandInteractionCreate, embed discord.Embed) error {
	return event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
}

func errorReply(event *events.ApplicationCommandInteractionCreate, msg string) error {
	return ephemeralReply(event, discord.Embed{
		Description: "❌\u00A0\u00A0\u00A0\u00A0\u00A0" + msg,
		Color:       0xED4245,
	})
}

func successReply(event *events.ApplicationCommandInteractionCreate, msg string) error {
	return reply(event, discord.Embed{
		Description: "✅\u00A0\u00A0\u00A0\u00A0\u00A0" + msg,
		Color:       0x57F287,
	})
}

func infoReply(event *events.ApplicationCommandInteractionCreate, msg string) error {
	return reply(event, discord.Embed{
		Description: "ℹ️\u00A0\u00A0\u00A0\u00A0\u00A0" + msg,
		Color:       0x5865F2,
	})
}

func (b *Bot) deferredTrackReply(event *events.ApplicationCommandInteractionCreate, track *lavalink.Track) {
	embed := discord.Embed{
		Title:       track.Info.Title,
		Description: fmt.Sprintf("von **%s**", track.Info.Author),
		URL:         *track.Info.URI,
		Color:       0x57F287,
		Image:       &discord.EmbedResource{URL: getArtworkURL(track)},
		Fields: []discord.EmbedField{
			{
				Name:   "⏱️ Dauer",
				Inline: new(true),
			},
			{
				Name:   fmt.Sprintf("%s min", formatPosition(track.Info.Length)),
				Inline: new(true),
			},
		},
	}
	embeds := []discord.Embed{embed}
	_, _ = b.Client.Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds: &embeds,
	})
}

func (b *Bot) deferredEmbedReply(event *events.ApplicationCommandInteractionCreate, color int, msg string) {
	embeds := []discord.Embed{{Description: msg, Color: color}}
	_, _ = b.Client.Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds: &embeds,
	})
}

func (b *Bot) deferredQueuedReply(event *events.ApplicationCommandInteractionCreate, track *lavalink.Track, position int) {
	embed := discord.Embed{
		Title:       "📋 Zur Warteschlange hinzugefügt",
		Description: fmt.Sprintf("[%s](<%s>)\nvon **%s**", track.Info.Title, *track.Info.URI, track.Info.Author),
		Color:       0x5865F2,
		Fields: []discord.EmbedField{
			{Name: "Position", Value: fmt.Sprintf("#%d", position), Inline: new(true)},
			{Name: "Dauer", Value: formatPosition(track.Info.Length), Inline: new(true)},
		},
	}
	embed.Thumbnail = &discord.EmbedResource{URL: getArtworkURL(track)}
	embeds := []discord.Embed{embed}
	_, _ = b.Client.Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds: &embeds,
	})
}

func getArtworkURL(track *lavalink.Track) string {
	if track.Info.ArtworkURL != nil {
		return *track.Info.ArtworkURL
	}
	return fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", track.Info.Identifier)
}
