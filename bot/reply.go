package bot

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func reply(event *events.ApplicationCommandInteractionCreate, embed discord.Embed) error {
	return event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
}

func errorReply(event *events.ApplicationCommandInteractionCreate, msg string) error {
	return reply(event, discord.Embed{
		Description: msg,
		Color:       0xED4245, // red
	})
}

func successReply(event *events.ApplicationCommandInteractionCreate, msg string) error {
	return reply(event, discord.Embed{
		Description: msg,
		Color:       0x57F287, // green
	})
}

func infoReply(event *events.ApplicationCommandInteractionCreate, msg string) error {
	return reply(event, discord.Embed{
		Description: msg,
		Color:       0x5865F2, // blurple
	})
}
