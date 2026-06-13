package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"feb-notify/internal/bot"
	"feb-notify/internal/config"
	"feb-notify/internal/monitor"
	"feb-notify/internal/squadstats"
)

func main() {
	// Configure logger to use slog with JSON format for structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Debug enabled by default for visibility
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Discord Bot...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	// 2. Initialize dependencies
	apiClient := squadstats.NewClient()
	monitorMgr := monitor.NewManager(apiClient)

	// 3. Initialize bot
	discordBot, err := bot.New(cfg, monitorMgr)
	if err != nil {
		slog.Error("Failed to initialize bot", "err", err)
		os.Exit(1)
	}

	// 4. Start bot
	err = discordBot.Start()
	if err != nil {
		slog.Error("Failed to start bot", "err", err)
		os.Exit(1)
	}

	slog.Info("Bot is now running. Press CTRL+C to exit.")

	// 5. Wait for termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// 6. Graceful shutdown
	discordBot.Stop()
	slog.Info("Bot stopped successfully.")
}
