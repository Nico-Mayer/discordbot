package music

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	mybot "github.com/nico-mayer/discordbot/bot"
)

var StopCommand = discord.SlashCommandCreate{
	Name:        "stop",
	Description: "Stoppt den Musikbot und löscht die Warteschlange.",
}

func StopCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	if b.BotStatus == mybot.Resting {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Ich spiele grade keine musik du kek, versuch nicht mich zu stoppen!",
		})
	}

	go func() {
		conn := b.Client.VoiceManager().GetConn(*event.GuildID())
		conn.Close(context.TODO())
	}()
	b.ClearQueue()
	b.BotStatus = mybot.Resting

	return event.CreateMessage(discord.MessageCreate{
		Content: "Die Musik wurde gestoppt und die Warteschlange wurde gelöscht.",
	})
}
