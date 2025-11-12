package music

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var PlayCmdMeta = discord.SlashCommandCreate{
	Name:        "play",
	Description: "Play music from url",
}

func PlayCmdHandler(event *events.ApplicationCommandInteractionCreate) error {
	fmt.Println("hello world")
	return nil
}
