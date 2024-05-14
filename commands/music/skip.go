package music

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	mybot "github.com/nico-mayer/discordbot/bot"
)

var SkipCommand = discord.SlashCommandCreate{
	Name:        "skip",
	Description: "Überspringt den aktuellen Song und geht zum nächsten in der Warteschlange.",
}

func SkipCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	if b.BotStatus == mybot.Resting {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Ich spiele grade keine musik du kek, versuch nicht mich zu stoppen!",
		})
	}

	if len(b.Queue) <= 1 {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Kein song mehr in der Warteschlange.",
		})
	}

	nextSong := b.Queue[1]

	b.SkipInterrupt <- true

	return event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Author: &discord.EmbedAuthor{
					Name:    event.User().Username,
					IconURL: *event.User().AvatarURL(),
				},
				Title:       "⏩ - Skipped",
				Description: fmt.Sprintf("Next Song: [`%s`](<%s>)", nextSong.Title, nextSong.URL),
				Thumbnail: &discord.EmbedResource{
					URL: nextSong.ThumbnailURL,
				},
			},
		},
	})
}
