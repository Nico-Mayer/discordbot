package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/charmbracelet/log"
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
			log.Debug("No .env file found, using environment variables")
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
		log.Warn("Missing environment variable", "key", key)
	}
	return val
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}
