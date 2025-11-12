package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/commands"
	"github.com/nico-mayer/discordbot/commands/general"
	"github.com/nico-mayer/discordbot/commands/lol"
	"github.com/nico-mayer/discordbot/commands/music"
	"github.com/nico-mayer/discordbot/commands/nasen"
)

func main() {
	slog.Info("starting bot...")
	slog.Info("disgo version", slog.String("version", disgo.Version))
	osSignals := make(chan os.Signal, 1)

	// Setup bot
	bot := mybot.NewBot()

	// Populate slash handler slice
	bot.Handlers = map[string]func(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error{
		general.HelpCommand.Name:       general.HelpCommandHandler,
		general.UserCommand.Name:       general.UserCommandHandler,
		nasen.ClownsnaseCommand.Name:   nasen.ClownsnaseCommandHandler,
		nasen.ClownfiestaCommand.Name:  nasen.ClownfiestaCommandHandler,
		nasen.NasenCommand.Name:        nasen.NasenCommandHandler,
		nasen.LeaderboardCommand.Name:  nasen.LeaderboardCommandHandler,
		music.PlayCommand.Name:         music.PlayCommandHandler,
		music.StopCommand.Name:         music.StopCommandHandler,
		music.SkipCommand.Name:         music.SkipCommandHandler,
		lol.AddRiotAccountCommand.Name: lol.AddRiotAccountCommandHandler,
		lol.LiveGameCommand.Name:       lol.LiveGameCommandHandler,
	}

	bot.SetupBot()

	// Register slash commands
	commands.RegisterSlashCommands(bot.Client, snowflake.GetEnv("GUILD_ID"))

	// Open Gateway
	if err := bot.Client.OpenGateway(context.TODO()); err != nil {
		slog.Error("error while connecting to gateway", slog.Any("err", err))
	}
	slog.Info("bot is now running. Press CTRL-C to exit.")

	err := bot.SetStatus(os.Getenv("BOT_STATUS"))
	if err != nil {
		slog.Error("setting bot status", err)
	}

	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-osSignals
}
