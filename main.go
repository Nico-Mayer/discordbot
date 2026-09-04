package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"charm.land/log/v2"

	"github.com/nico-mayer/discordbot/internal/app"
	"github.com/nico-mayer/discordbot/internal/config"
)

func main() {
	resetCommands := flag.Bool("reset-commands", false, "clear all registered guild and global commands before registering the current set")
	flag.Parse()

	// The logger is built before anything can fail, so a configuration failure is
	// reported as a record like every other event rather than printed raw.
	handler := log.New(os.Stderr)
	handler.SetReportTimestamp(true)
	logger := slog.New(handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logger.ErrorContext(ctx, "invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	if err := app.Run(ctx, cfg, logger, app.Options{ResetCommands: *resetCommands}); err != nil {
		logger.ErrorContext(ctx, "bot failed", slog.Any("error", err))
		os.Exit(1)
	}
}
