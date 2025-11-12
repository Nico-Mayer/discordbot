package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/cmd"
	"github.com/nico-mayer/discordbot/config"
)

func main() {
	slog.Info("starting...")
	slog.Info("disgo version", slog.String("version", disgo.Version))

	bot := bot.Get()
	if err := bot.Start(); err != nil {
		slog.Error("failed to start bot", slog.String("error", err.Error()))
		return
	}
	defer bot.Client.Close(context.TODO())
	cmd.RegisterSlashCommands(bot.Client, config.Get().GUILD_ID)

	slog.Info("bot is now running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
