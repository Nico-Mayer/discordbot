package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo"
	"github.com/nico-mayer/discordbot/bot"
)

func main() {
	log.Info("starting...")
	log.Info("disgo version", "version", disgo.Version)

	bot := bot.Get()
	if err := bot.Start(); err != nil {
		log.Error("failed to start bot", "error", err.Error())
		return
	}
	defer bot.Client.Close(context.TODO())

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
