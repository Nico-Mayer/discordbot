package config

import (
	"os"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
)

type config struct {
	ENV       string
	TOKEN     string
	GUILD_ID  snowflake.ID
	CLIENT_ID snowflake.ID
}

var instance *config
var once sync.Once

func Get() *config {
	if instance == nil {
		once.Do(func() {
			err := godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}
			instance = initConfig()
		})
	}
	return instance
}

func initConfig() *config {
	return &config{
		ENV:       getEnv("ENV", "PREV"),
		TOKEN:     getRequiredEnv("TOKEN"),
		GUILD_ID:  snowflake.MustParse(getRequiredEnv("GUILD_ID")),
		CLIENT_ID: snowflake.MustParse(getRequiredEnv("CLIENT_ID")),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Warn("env variable not set", "key", key, "fallback", fallback)
		value = fallback
	}
	return value
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal("Environment variable " + key + " is required but not set.")
	}
	return value
}
