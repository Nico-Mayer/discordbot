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

// deferredEmbedReply updates a deferred interaction response with an embed.
func (b *Bot) deferredEmbedReply(event *events.ApplicationCommandInteractionCreate, color int, msg string) {
	embeds := []discord.Embed{{Description: msg, Color: color}}
	_, _ = b.Client.Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds: &embeds,
	})
}

func (b *Bot) play(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	identifier := data.String("identifier")
	if source, ok := data.OptString("source"); ok {
		identifier = lavalink.SearchType(source).Apply(identifier)
	} else if !urlPattern.MatchString(identifier) {
		identifier = lavalink.SearchTypeYouTube.Apply(identifier)
	}

	voiceState, ok := b.Client.Caches.VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return errorReply(event, "You need to be in a voice channel to use this command")
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track
	b.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			b.deferredEmbedReply(event, 0x57F287, fmt.Sprintf("Loaded track: [%s](<%s>)", track.Info.Title, *track.Info.URI))
			toPlay = &track
		},
		func(playlist lavalink.Playlist) {
			b.deferredEmbedReply(event, 0x57F287, fmt.Sprintf("Loaded playlist: **%s** with **%d** tracks", playlist.Info.Name, len(playlist.Tracks)))
			toPlay = &playlist.Tracks[0]
			b.Queues.Get(*event.GuildID()).Add(playlist.Tracks[1:]...)
		},
		func(tracks []lavalink.Track) {
			b.deferredEmbedReply(event, 0x57F287, fmt.Sprintf("Loaded search result: [%s](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI))
			toPlay = &tracks[0]
		},
		func() {
			b.deferredEmbedReply(event, 0xED4245, fmt.Sprintf("Nothing found for: `%s`", identifier))
		},
		func(err error) {
			b.deferredEmbedReply(event, 0xED4245, fmt.Sprintf("Error loading track: `%s`", err))
		},
	))
	if toPlay == nil {
		return nil
	}

	if err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), voiceState.ChannelID, false, false); err != nil {
		return err
	}

	return b.Lavalink.Player(*event.GuildID()).Update(context.TODO(), lavalink.WithTrack(*toPlay))
}

func (b *Bot) pause(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return errorReply(event, "No player found")
	}

	if err := player.Update(context.TODO(), lavalink.WithPaused(!player.Paused())); err != nil {
		return errorReply(event, fmt.Sprintf("Error while pausing: `%s`", err))
	}

	status := "resumed ▶️"
	if player.Paused() {
		status = "paused ⏸️"
	}
	return infoReply(event, fmt.Sprintf("Player is now %s", status))
}

func (b *Bot) stop(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return errorReply(event, "No player found")
	}

	b.Queues.Get(*event.GuildID()).Clear()

	if err := player.Update(context.TODO(), lavalink.WithNullTrack()); err != nil {
		return errorReply(event, fmt.Sprintf("Error while stopping: `%s`", err))
	}

	return successReply(event, "Player stopped ⏹️")
}

func (b *Bot) skip(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	queue := b.Queues.Get(*event.GuildID())
	if player == nil {
		return errorReply(event, "No player found")
	}

	track, ok := queue.Next()
	if !ok {
		return errorReply(event, "No more tracks in queue")
	}

	if err := player.Update(context.TODO(), lavalink.WithTrack(track)); err != nil {
		return errorReply(event, fmt.Sprintf("Error while skipping: `%s`", err))
	}

	return successReply(event, "Skipped track ⏭️")
}

func (b *Bot) nowPlaying(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return errorReply(event, "No player found")
	}

	track := player.Track()
	if track == nil {
		return errorReply(event, "No track playing")
	}

	return reply(event, discord.Embed{
		Title:       "Now Playing 🎶",
		Description: fmt.Sprintf("[%s](<%s>)", track.Info.Title, *track.Info.URI),
		Color:       0x5865F2,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%s / %s", formatPosition(player.Position()), formatPosition(track.Info.Length)),
		},
	})
}

func (b *Bot) queue(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	queue := b.Queues.Get(*event.GuildID())
	if len(queue.Tracks) == 0 {
		return infoReply(event, "Queue is empty")
	}

	var tracks string
	for i, track := range queue.Tracks {
		tracks += fmt.Sprintf("`%d.` [%s](<%s>)\n", i+1, track.Info.Title, *track.Info.URI)
	}

	return reply(event, discord.Embed{
		Title:       "Queue 📋",
		Description: tracks,
		Color:       0x5865F2,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%d tracks", len(queue.Tracks)),
		},
	})
}

func formatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}
