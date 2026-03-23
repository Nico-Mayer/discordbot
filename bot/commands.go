package bot

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
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
			discord.ApplicationCommandOptionString{
				Name:        "source",
				Description: "The source to search on",
				Required:    false,
				Choices: []discord.ApplicationCommandOptionChoiceString{
					{Name: "YouTube", Value: string(lavalink.SearchTypeYouTube)},
					{Name: "YouTube Music", Value: string(lavalink.SearchTypeYouTubeMusic)},
					{Name: "SoundCloud", Value: string(lavalink.SearchTypeSoundCloud)},
				},
			},
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

func RegisterCommands(client *bot.Client, guildID snowflake.ID) {
	if err := handler.SyncCommands(client, commands, []snowflake.ID{guildID}); err != nil {
		slog.Error("error registering commands", slog.Any("err", err))
	}
}
