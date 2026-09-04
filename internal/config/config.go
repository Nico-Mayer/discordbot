// Package config loads and validates the bot's environment configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
)

var (
	// ErrMissing reports a required environment variable that is unset or empty.
	ErrMissing = errors.New("missing environment variable")
	// ErrInvalid reports an environment variable that is set but unusable.
	ErrInvalid = errors.New("invalid environment variable")
)

// defaultIdleTimeout is how long the bot lingers before leaving for an idle
// reason the operator has not configured.
const defaultIdleTimeout = 60 * time.Second

// IdleTimeout is how long the bot stays in a voice channel after one idle
// condition becomes true. Enabled is false for the "off" literal, which cannot
// be expressed as a duration without a sentinel value.
type IdleTimeout struct {
	After   time.Duration
	Enabled bool
}

// Config is the validated configuration of a single bot process.
type Config struct {
	Token   string
	GuildID snowflake.ID

	NodeName   string
	NodeSecure bool

	LavalinkAddress  string
	LavalinkPassword string

	IdleAlone      IdleTimeout
	IdleEmptyQueue IdleTimeout
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
	idleAlone, aloneErr := envIdleTimeout("IDLE_ALONE_SECONDS")
	idleEmpty, emptyErr := envIdleTimeout("IDLE_EMPTY_QUEUE_SECONDS")

	if err := errors.Join(tokenErr, guildErr, nameErr, secureErr, hostErr, portErr, passwordErr, aloneErr, emptyErr); err != nil {
		return Config{}, err
	}

	return Config{
		Token:            token,
		GuildID:          guildID,
		NodeName:         nodeName,
		NodeSecure:       secure,
		LavalinkAddress:  net.JoinHostPort(host, strconv.Itoa(port)),
		LavalinkPassword: password,
		IdleAlone:        idleAlone,
		IdleEmptyQueue:   idleEmpty,
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

// envIdleTimeout accepts a non-negative number of seconds or the literal "off".
// A value it cannot interpret is an error rather than a fallback to the default,
// because a silently ignored timeout looks exactly like a working one.
func envIdleTimeout(key string) (IdleTimeout, error) {
	raw := os.Getenv(key)
	switch raw {
	case "":
		return IdleTimeout{After: defaultIdleTimeout, Enabled: true}, nil
	case "off":
		return IdleTimeout{}, nil
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return IdleTimeout{}, fmt.Errorf("%s: %w: %q is neither a number of seconds nor %q", key, ErrInvalid, raw, "off")
	}
	if seconds < 0 {
		return IdleTimeout{}, fmt.Errorf("%s: %w: %d is negative", key, ErrInvalid, seconds)
	}
	return IdleTimeout{After: time.Duration(seconds) * time.Second, Enabled: true}, nil
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
