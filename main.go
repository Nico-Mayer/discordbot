package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"

	"github.com/disgoorg/disgolink/v3/disgolink"

	"github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/config"
)

func main() {
	flag.Parse()

	log.SetReportTimestamp(true)

	cfg := config.Load()

	log.Info("Starting bot", "disgo_version", disgo.Version)

	b := bot.New()

	client, err := disgo.New(cfg.Token,
		disgobot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates),
		),
		disgobot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
		),
		disgobot.WithEventListenerFunc(b.OnApplicationCommand),
		disgobot.WithEventListenerFunc(b.OnVoiceStateUpdate),
		disgobot.WithEventListenerFunc(b.OnVoiceServerUpdate),
	)
	if err != nil {
		log.Fatal("Failed to build client", "err", err)
	}
	b.Client = client

	bot.RegisterCommands(client, cfg.GuildID, false)

	b.Lavalink = disgolink.New(client.ApplicationID,
		disgolink.WithListenerFunc(b.OnPlayerPause),
		disgolink.WithListenerFunc(b.OnPlayerResume),
		disgolink.WithListenerFunc(b.OnTrackStart),
		disgolink.WithListenerFunc(b.OnTrackEnd),
		disgolink.WithListenerFunc(b.OnTrackException),
		disgolink.WithListenerFunc(b.OnTrackStuck),
		disgolink.WithListenerFunc(b.OnWebSocketClosed),
	)
	b.InitHandlers()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.OpenGateway(ctx); err != nil {
		log.Fatal("Failed to open gateway", "err", err)
	}
	defer client.Close(context.TODO())

	self, err := client.Rest.GetUser(client.ApplicationID)
	if err != nil {
		log.Warn("Could not fetch bot user", "err", err)
	} else {
		log.Info("Connected as", "username", self.Username, "id", self.ID)
	}

	node, err := b.Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Name:     cfg.NodeName,
		Address:  cfg.NodeAddress,
		Password: cfg.NodePassword,
		Secure:   cfg.NodeSecure,
	})
	if err != nil {
		log.Fatal("Failed to add lavalink node", "err", err)
	}
	version, err := node.Version(ctx)
	if err != nil {
		log.Fatal("Failed to get node version", "err", err)
	}

	log.Info("Bot is running", "lavalink_node_version", version)

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
