package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"charm.land/log/v2"
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

	NodeName   string
	NodeSecure bool

	LavalinkHost     string
	LavalinkPort     int
	LavalinkPassword string
	LavalinkAddress  string
}

func Load() Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Debug("No .env file found, using environment variables")
		}

		instance = Config{
			Token:            requireEnv("TOKEN"),
			GuildID:          snowflake.GetEnv("GUILD_ID"),
			NodeName:         requireEnv("NODE_NAME"),
			NodeSecure:       envBool("NODE_SECURE"),
			LavalinkHost:     requireEnv("LAVALINK_HOST"),
			LavalinkPort:     envInt("LAVALINK_PORT"),
			LavalinkPassword: requireEnv("LAVALINK_PASSWORD"),
			LavalinkAddress:  fmt.Sprintf("%s:%d", requireEnv("LAVALINK_HOST"), envInt("LAVALINK_PORT")),
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

func envInt(key string) int {
	v, _ := strconv.Atoi(os.Getenv(key))
	return v
}
