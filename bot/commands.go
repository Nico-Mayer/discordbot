package bot

import (
	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

var commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "play",
		Description: "Plays a song",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "identifier",
				Description: "The song link or search query",
				Required:    true,
			},
			// discord.ApplicationCommandOptionString{
			// 	Name:        "source",
			// 	Description: "The source to search on",
			// 	Required:    false,
			// 	Choices: []discord.ApplicationCommandOptionChoiceString{
			// 		{Name: "YouTube", Value: string(lavalink.SearchTypeYouTube)},
			// 		{Name: "YouTube Music", Value: string(lavalink.SearchTypeYouTubeMusic)},
			// 		{Name: "SoundCloud", Value: string(lavalink.SearchTypeSoundCloud)},
			// 	},
			// },
		},
	},
	discord.SlashCommandCreate{
		Name:        "pause",
		Description: "Pauses or resumes the current song",
	},
	discord.SlashCommandCreate{
		Name:        "stop",
		Description: "Stops the player and clears the queue",
	},
	discord.SlashCommandCreate{
		Name:        "skip",
		Description: "Skips the current song",
	},
	discord.SlashCommandCreate{
		Name:        "now-playing",
		Description: "Shows the current playing song",
	},
	discord.SlashCommandCreate{
		Name:        "queue",
		Description: "Shows the current queue",
	},
}

func RegisterCommands(client *bot.Client, guildID snowflake.ID, reset bool) {
	if reset {
		log.Info("Resetting all commands")
		if _, err := client.Rest.SetGuildCommands(client.ApplicationID, guildID, nil); err != nil {
			log.Error("Failed to clear guild commands", "err", err)
		}
		if _, err := client.Rest.SetGlobalCommands(client.ApplicationID, nil); err != nil {
			log.Error("Failed to clear global commands", "err", err)
		}
	}
	if err := handler.SyncCommands(client, commands, []snowflake.ID{guildID}); err != nil {
		log.Error("Failed to register commands", "err", err)
	}
}
