package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"

	"github.com/disgoorg/disgolink/v3/disgolink"

	"github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/config"
)

func main() {
	cfg := config.Load()

	slog.Info("starting bot...", slog.String("disgo_version", disgo.Version))

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
		slog.Error("error building client", slog.Any("err", err))
		os.Exit(1)
	}
	b.Client = client

	bot.RegisterCommands(client, cfg.GuildID)

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
		slog.Error("failed to open gateway", slog.Any("err", err))
		os.Exit(1)
	}
	defer client.Close(context.TODO())

	node, err := b.Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Name:     cfg.NodeName,
		Address:  cfg.NodeAddress,
		Password: cfg.NodePassword,
		Secure:   cfg.NodeSecure,
	})
	if err != nil {
		slog.Error("failed to add lavalink node", slog.Any("err", err))
		os.Exit(1)
	}
	version, err := node.Version(ctx)
	if err != nil {
		slog.Error("failed to get node version", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("bot is running", slog.String("node_version", version))

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
