package music

import "github.com/disgoorg/disgo/discord"

// Commands is the bot's slash command set. It is registered for the single
// configured guild, never globally.
//
// The command names stay English because they are typed to invoke the bot. The
// descriptions and the option label are read, so they are German like the rest
// of the copy.
var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "play",
		Description: descPlay,
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        optionPlayName,
				Description: optionPlayDesc,
				Required:    true,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "pause",
		Description: descPause,
	},
	discord.SlashCommandCreate{
		Name:        "stop",
		Description: descStop,
	},
	discord.SlashCommandCreate{
		Name:        "skip",
		Description: descSkip,
	},
	discord.SlashCommandCreate{
		Name:        "now-playing",
		Description: descNowPlaying,
	},
	discord.SlashCommandCreate{
		Name:        "queue",
		Description: descQueue,
	},
}
