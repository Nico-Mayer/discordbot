package commands

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nico-mayer/discordbot/commands/general"
	"github.com/nico-mayer/discordbot/commands/lol"
	"github.com/nico-mayer/discordbot/commands/music"
	"github.com/nico-mayer/discordbot/commands/nasen"
)

var commands = []discord.ApplicationCommandCreate{
	general.HelpCommand,
	general.UserCommand,
	nasen.ClownsnaseCommand,
	nasen.ClownfiestaCommand,
	nasen.NasenCommand,
	nasen.LeaderboardCommand,
	music.PlayCommand,
	music.StopCommand,
	music.SkipCommand,
	lol.AddRiotAccountCommand,
	lol.LiveGameCommand,
}

func RegisterSlashCommands(client bot.Client, guildID snowflake.ID) {
	if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), guildID, commands); err != nil {
		slog.Error("error while registering commands", slog.Any("err", err))
	}
}
