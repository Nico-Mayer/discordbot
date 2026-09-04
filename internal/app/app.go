// Package app is the composition root: it builds every dependency, runs the bot
// and unwinds what it built.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/nico-mayer/discordbot/internal/config"
	"github.com/nico-mayer/discordbot/internal/music"
)

const (
	// startupTimeout bounds each startup step so a hung dependency cannot stall
	// the process indefinitely.
	startupTimeout = 10 * time.Second
	// shutdownTimeout bounds the whole unwind, so a blocking close still exits.
	shutdownTimeout = 10 * time.Second
	// resumeTimeoutSeconds is how long Lavalink holds a session open after the
	// websocket drops. It only has to outlive a node restart.
	resumeTimeoutSeconds = 60
)

// Options carries the command line switches through to startup.
type Options struct {
	// ResetCommands clears every registered guild and global command before the
	// current set is registered.
	ResetCommands bool
}

// cleanup releases one successfully built dependency.
type cleanup struct {
	name string
	run  func(ctx context.Context)
}

// Run builds the bot, serves until ctx is done, and closes everything it opened.
// A failure part-way through startup unwinds the dependencies built so far.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, opts Options) error {
	// cleanups are held in shutdown order rather than as a reverse-order stack:
	// leaving the voice channel needs the gateway that was opened after the
	// Lavalink client was built, so the required order inverts registration.
	var cleanups []cleanup
	register := func(c cleanup) { cleanups = append(cleanups, c) }
	registerFirst := func(c cleanup) { cleanups = slices.Insert(cleanups, 0, c) }

	// unwind releases whatever has been built so far.
	unwind := func() { releaseAll(ctx, cleanups, shutdownTimeout, logger) }

	logger.InfoContext(ctx, "starting bot", slog.String("disgo_version", disgo.Version))

	client, err := disgo.New(cfg.Token,
		disgobot.WithLogger(logger),
		disgobot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates),
		),
		disgobot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
		),
	)
	if err != nil {
		return fmt.Errorf("build discord client: %w", err)
	}

	lava := disgolink.New(client.ApplicationID, disgolink.WithLogger(logger))
	service := music.NewService(music.ServiceConfig{
		GuildID:        cfg.GuildID,
		ApplicationID:  client.ApplicationID,
		Lavalink:       newLavalinkAdapter(lava),
		Voice:          client,
		VoiceStates:    newVoiceStateAdapter(client.Caches),
		Logger:         logger,
		IdleAlone:      music.IdleTimeout(cfg.IdleAlone),
		IdleEmptyQueue: music.IdleTimeout(cfg.IdleEmptyQueue),
	})
	events := music.NewEvents(ctx, service, lava, client.ApplicationID, logger)

	lava.AddListeners(
		disgolink.NewListenerFunc(events.OnPlayerPause),
		disgolink.NewListenerFunc(events.OnPlayerResume),
		disgolink.NewListenerFunc(events.OnTrackStart),
		disgolink.NewListenerFunc(events.OnTrackEnd),
		disgolink.NewListenerFunc(events.OnTrackException),
		disgolink.NewListenerFunc(events.OnTrackStuck),
		disgolink.NewListenerFunc(events.OnWebSocketClosed),
	)

	router := music.NewRouter(service, logger)
	// Every interaction handler inherits the process lifetime context, so a
	// signal cancels in-flight commands too.
	router.DefaultContext(func() context.Context { return ctx })

	client.AddEventListeners(
		router,
		disgobot.NewListenerFunc(events.OnVoiceStateUpdate),
		disgobot.NewListenerFunc(events.OnVoiceServerUpdate),
	)

	register(cleanup{
		name: "close lavalink",
		run:  func(context.Context) { lava.Close() },
	})

	if err := openGateway(ctx, client); err != nil {
		unwind()
		return err
	}
	register(cleanup{name: "close gateway", run: client.Close})

	// Leaving the voice channel needs the gateway, and must happen before
	// anything closes, so the bot does not linger in the channel after exit.
	registerFirst(cleanup{
		name: "leave voice channel",
		run: func(ctx context.Context) {
			if err := service.Leave(ctx); err != nil {
				logger.ErrorContext(ctx, "could not leave the voice channel", slog.Any("err", err))
			}
		},
	})

	// Registered after the leave step so it lands ahead of it: a countdown that
	// elapses mid-shutdown would otherwise act on a client already being closed.
	registerFirst(cleanup{
		name: "stop idle countdowns",
		run:  func(context.Context) { service.Close() },
	})

	logSelf(ctx, logger, func() (*discord.User, error) {
		return client.Rest.GetUser(client.ApplicationID)
	})
	registerCommands(ctx, client, cfg.GuildID, opts.ResetCommands, logger)

	node, err := addNode(ctx, lava, cfg)
	if err != nil {
		unwind()
		return err
	}

	version, err := nodeVersion(ctx, node)
	if err != nil {
		unwind()
		return err
	}

	enableResuming(ctx, node, logger)

	logger.InfoContext(ctx, "bot is running", slog.String("lavalink_node_version", version))

	<-ctx.Done()
	logger.Info("shutdown signal received")
	unwind()
	return nil
}

func openGateway(ctx context.Context, client *disgobot.Client) error {
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	if err := client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("open discord gateway: %w", err)
	}
	return nil
}

func addNode(ctx context.Context, lava disgolink.Client, cfg config.Config) (disgolink.Node, error) {
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	node, err := lava.AddNode(ctx, disgolink.NodeConfig{
		Name:     cfg.NodeName,
		Address:  cfg.LavalinkAddress,
		Password: cfg.LavalinkPassword,
		Secure:   cfg.NodeSecure,
	})
	if err != nil {
		return nil, fmt.Errorf("add lavalink node %q: %w", cfg.NodeName, err)
	}
	return node, nil
}

func nodeVersion(ctx context.Context, node disgolink.Node) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	version, err := node.Version(ctx)
	if err != nil {
		return "", fmt.Errorf("get lavalink node version: %w", err)
	}
	return version, nil
}

// enableResuming makes a node reconnect keep its players instead of opening a
// fresh session and destroying them. The bot is still usable without it, so a
// failure is only a warning.
func enableResuming(ctx context.Context, node disgolink.Node, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	resuming, timeout := true, resumeTimeoutSeconds
	if err := node.Update(ctx, lavalink.SessionUpdate{Resuming: &resuming, Timeout: &timeout}); err != nil {
		logger.WarnContext(ctx, "could not enable lavalink session resuming", slog.Any("err", err))
		return
	}
	logger.InfoContext(ctx, "lavalink session resuming enabled", slog.Int("timeout_seconds", resumeTimeoutSeconds))
}

// logSelf reports the identity the bot connected as. The identity is
// informational, so a failure to fetch it must not stop startup.
func logSelf(ctx context.Context, logger *slog.Logger, fetch func() (*discord.User, error)) {
	self, err := fetch()
	if err != nil {
		logger.WarnContext(ctx, "could not fetch the bot user", slog.Any("err", err))
		return
	}
	logger.InfoContext(ctx, "connected", slog.String("username", self.Username), slog.Any("id", self.ID))
}

// releaseAll runs the cleanups in shutdown order under a bounded context derived
// from ctx but not cancelled by it, so a signal does not abort the unwind.
func releaseAll(ctx context.Context, cleanups []cleanup, timeout time.Duration, logger *slog.Logger) {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	for _, c := range cleanups {
		logger.InfoContext(shutdownCtx, "shutting down", slog.String("step", c.name))
		runCleanup(shutdownCtx, c, logger)
	}
}

// runCleanup waits for one cleanup, giving up when the shutdown context expires
// so a dependency that never returns cannot stop the process from exiting.
func runCleanup(ctx context.Context, c cleanup, logger *slog.Logger) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.run(ctx)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		logger.Warn("shutdown step did not finish in time", slog.String("step", c.name))
	}
}

// registerCommands brings the guild's command set in line with this build.
// Previously registered commands may still work, so a failure is only logged.
func registerCommands(ctx context.Context, client *disgobot.Client, guildID snowflake.ID, reset bool, logger *slog.Logger) {
	if reset {
		logger.InfoContext(ctx, "resetting all commands")
		if _, err := client.Rest.SetGuildCommands(client.ApplicationID, guildID, nil); err != nil {
			logger.ErrorContext(ctx, "could not clear guild commands", slog.Any("err", err))
		}
		if _, err := client.Rest.SetGlobalCommands(client.ApplicationID, nil); err != nil {
			logger.ErrorContext(ctx, "could not clear global commands", slog.Any("err", err))
		}
	}

	if err := handler.SyncCommands(client, music.Commands, []snowflake.ID{guildID}); err != nil {
		logger.ErrorContext(ctx, "could not register commands", slog.Any("err", err))
	}
}
