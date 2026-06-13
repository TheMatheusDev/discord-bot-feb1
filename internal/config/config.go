package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	DiscordToken string
	GuildID      string
}

// Load reads the configuration from environment variables.
// It attempts to load variables from a .env file if present.
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found or error loading it, proceeding with environment variables")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		slog.Error("DISCORD_TOKEN environment variable is required")
		return nil, os.ErrNotExist
	}

	guildID := os.Getenv("GUILD_ID")

	return &Config{
		DiscordToken: token,
		GuildID:      guildID,
	}, nil
}
