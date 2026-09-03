package music

import "github.com/disgoorg/disgo/discord"

// Commands is the bot's slash command set. It is registered for the single
// configured guild, never globally.
var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "play",
		Description: "Plays a song",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "identifier",
				Description: "The song link or search query",
				Required:    true,
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
