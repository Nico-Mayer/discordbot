// Package config loads and validates the bot's environment configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
)

var (
	// ErrMissing reports a required environment variable that is unset or empty.
	ErrMissing = errors.New("missing environment variable")
	// ErrInvalid reports an environment variable that is set but unusable.
	ErrInvalid = errors.New("invalid environment variable")
)

// Config is the validated configuration of a single bot process.
type Config struct {
	Token   string
	GuildID snowflake.ID

	NodeName   string
	NodeSecure bool

	LavalinkAddress  string
	LavalinkPassword string
}

// Load reads the configuration from a .env file, if present, and the process
// environment, which takes precedence. Every validation failure is reported
// together. No error message contains the value of a secret.
func Load() (Config, error) {
	// A .env file is a local development convenience; its absence is not an error.
	_ = godotenv.Load()

	token, tokenErr := requireEnv("TOKEN")
	guildID, guildErr := envSnowflake("GUILD_ID")
	nodeName, nameErr := requireEnv("NODE_NAME")
	secure, secureErr := envBool("NODE_SECURE")
	host, hostErr := requireEnv("LAVALINK_HOST")
	port, portErr := envPort("LAVALINK_PORT")
	password, passwordErr := requireEnv("LAVALINK_PASSWORD")

	if err := errors.Join(tokenErr, guildErr, nameErr, secureErr, hostErr, portErr, passwordErr); err != nil {
		return Config{}, err
	}

	return Config{
		Token:            token,
		GuildID:          guildID,
		NodeName:         nodeName,
		NodeSecure:       secure,
		LavalinkAddress:  net.JoinHostPort(host, strconv.Itoa(port)),
		LavalinkPassword: password,
	}, nil
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s: %w", key, ErrMissing)
	}
	return value, nil
}

func envSnowflake(key string) (snowflake.ID, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("%s: %w", key, ErrMissing)
	}
	id, err := snowflake.Parse(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w: %q is not a Discord snowflake", key, ErrInvalid, raw)
	}
	return id, nil
}

func envPort(key string) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("%s: %w", key, ErrMissing)
	}
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w: %q is not a number", key, ErrInvalid, raw)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s: %w: %d is outside 1-65535", key, ErrInvalid, port)
	}
	return port, nil
}

func envBool(key string) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s: %w: %q is not a boolean", key, ErrInvalid, raw)
	}
	return value, nil
}
