package config

import (
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
)

var (
	once     sync.Once
	instance Config
)

type Config struct {
	Token   string
	GuildID snowflake.ID

	NodeName     string
	NodeAddress  string
	NodePassword string
	NodeSecure   bool
}

func Load() Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			slog.Debug("no .env file found, using environment variables")
		}

		instance = Config{
			Token:        requireEnv("TOKEN"),
			GuildID:      snowflake.GetEnv("GUILD_ID"),
			NodeName:     requireEnv("NODE_NAME"),
			NodeAddress:  requireEnv("NODE_ADDRESS"),
			NodePassword: requireEnv("NODE_PASSWORD"),
			NodeSecure:   envBool("NODE_SECURE"),
		}
	})
	return instance
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		slog.Warn("missing environment variable", slog.String("key", key))
	}
	return val
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}
