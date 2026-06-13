package bot

import (
	"context"
	"fmt"
	"log/slog"

	"feb-notify/internal/config"
	"feb-notify/internal/monitor"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot instance.
type Bot struct {
	Session *discordgo.Session
	config  *config.Config
	monitor *monitor.Manager
	ctx     context.Context
	cancel  context.CancelFunc
}

// New creates a new Bot instance.
func New(cfg *config.Config, monitorMgr *monitor.Manager) (*Bot, error) {
	s, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating discord session: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	b := &Bot{
		Session: s,
		config:  cfg,
		monitor: monitorMgr,
		ctx:     ctx,
		cancel:  cancel,
	}

	b.registerHandlers()
	return b, nil
}

// Start opens the connection to Discord and registers commands.
func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening discord connection: %w", err)
	}

	slog.Info("Bot connected to Discord")

	err = b.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: string(discordgo.StatusOnline),
		Activities: []*discordgo.Activity{
			{
				Name: "Jogando todo meu ódio em Tallil!",
				Type: discordgo.ActivityTypeGame,
			},
		},
	})
	if err != nil {
		slog.Warn("Failed to update bot activity", "err", err)
	}

	err = b.registerCommands()
	if err != nil {
		return fmt.Errorf("error registering commands: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the bot.
func (b *Bot) Stop() {
	slog.Info("Shutting down bot...")
	b.cancel() // Cancel all pending contexts
	b.Session.Close()
}
