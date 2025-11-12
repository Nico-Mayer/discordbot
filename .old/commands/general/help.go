package general

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	mybot "github.com/nico-mayer/discordbot/bot"
)

var HelpCommand = discord.SlashCommandCreate{
	Name:        "help",
	Description: "Zeige alle verfügbaren Bot-Befehle an.",
}

func HelpCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	var slashCommands []discord.SlashCommand

	if err := event.DeferCreateMessage(true); err != nil {
		return err
	}

	commands, err := event.Client().Rest().GetGuildCommands(event.ApplicationID(), *event.GuildID(), false)
	if err != nil {
		return err
	}

	for _, command := range commands {
		if slashCommand, ok := command.(discord.SlashCommand); ok {
			slashCommands = append(slashCommands, slashCommand)
		}
	}

	_, err = event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title:       "ℹ️ - Help",
				Description: generateList(slashCommands),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil

}

func generateList(slashCommands []discord.SlashCommand) (desc string) {
	for _, command := range slashCommands {
		line := fmt.Sprintf("- `/%s` - (%s)\n", command.Name(), command.Description)
		desc = desc + line
	}
	return desc
}
