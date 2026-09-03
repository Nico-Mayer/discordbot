package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nico-mayer/discordbot/internal/config"
)

// setValidEnv puts a complete, valid configuration in the process environment so
// each test only has to override the variable it is about.
func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TOKEN", "valid-token")
	t.Setenv("GUILD_ID", "123456789012345678")
	t.Setenv("NODE_NAME", "main")
	t.Setenv("NODE_SECURE", "")
	t.Setenv("LAVALINK_HOST", "lavalink.internal")
	t.Setenv("LAVALINK_PORT", "2333")
	t.Setenv("LAVALINK_PASSWORD", "valid-password")
}

func TestLoadHappyPath(t *testing.T) {
	setValidEnv(t)
	t.Setenv("NODE_SECURE", "true")

	cfg, err := config.Load()
	require.NoError(t, err)

	require.Equal(t, "valid-token", cfg.Token)
	require.Equal(t, "123456789012345678", cfg.GuildID.String())
	require.Equal(t, "main", cfg.NodeName)
	require.True(t, cfg.NodeSecure)
	require.Equal(t, "lavalink.internal:2333", cfg.LavalinkAddress)
	require.Equal(t, "valid-password", cfg.LavalinkPassword)
}

func TestLoadWithoutDotEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	setValidEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "valid-token", cfg.Token)
}

func TestLoadProcessEnvironmentWinsOverDotEnv(t *testing.T) {
	dir := t.TempDir()
	dotenv := "TOKEN=from-dotenv\nNODE_NAME=from-dotenv\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte(dotenv), 0o600))
	t.Chdir(dir)

	setValidEnv(t)
	t.Setenv("TOKEN", "from-process")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "from-process", cfg.Token)
}

func TestLoadReadsDotEnvWhenProcessEnvironmentIsEmpty(t *testing.T) {
	dir := t.TempDir()
	dotenv := strings.Join([]string{
		"TOKEN=dotenv-token",
		"GUILD_ID=123456789012345678",
		"NODE_NAME=dotenv-node",
		"LAVALINK_HOST=dotenv.internal",
		"LAVALINK_PORT=2333",
		"LAVALINK_PASSWORD=dotenv-password",
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte(dotenv), 0o600))
	t.Chdir(dir)

	for _, key := range []string{"TOKEN", "GUILD_ID", "NODE_NAME", "NODE_SECURE", "LAVALINK_HOST", "LAVALINK_PORT", "LAVALINK_PASSWORD"} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "dotenv-token", cfg.Token)
	require.Equal(t, "dotenv.internal:2333", cfg.LavalinkAddress)
}

func TestLoadSecureFlagDefaultsToFalse(t *testing.T) {
	setValidEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.False(t, cfg.NodeSecure)
}

func TestLoadReportsEveryMissingVariable(t *testing.T) {
	setValidEnv(t)
	t.Setenv("TOKEN", "")
	t.Setenv("NODE_NAME", "")
	t.Setenv("LAVALINK_PASSWORD", "")

	_, err := config.Load()
	require.ErrorIs(t, err, config.ErrMissing)

	msg := err.Error()
	require.Contains(t, msg, "TOKEN")
	require.Contains(t, msg, "NODE_NAME")
	require.Contains(t, msg, "LAVALINK_PASSWORD")
}

func TestLoadValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr error
		names   string
	}{
		{name: "token missing", key: "TOKEN", value: "", wantErr: config.ErrMissing, names: "TOKEN"},
		{name: "node name missing", key: "NODE_NAME", value: "", wantErr: config.ErrMissing, names: "NODE_NAME"},
		{name: "host missing", key: "LAVALINK_HOST", value: "", wantErr: config.ErrMissing, names: "LAVALINK_HOST"},
		{name: "password missing", key: "LAVALINK_PASSWORD", value: "", wantErr: config.ErrMissing, names: "LAVALINK_PASSWORD"},
		{name: "guild id missing", key: "GUILD_ID", value: "", wantErr: config.ErrMissing, names: "GUILD_ID"},
		{name: "guild id malformed", key: "GUILD_ID", value: "not-a-snowflake", wantErr: config.ErrInvalid, names: "GUILD_ID"},
		{name: "guild id negative", key: "GUILD_ID", value: "-1", wantErr: config.ErrInvalid, names: "GUILD_ID"},
		{name: "port missing", key: "LAVALINK_PORT", value: "", wantErr: config.ErrMissing, names: "LAVALINK_PORT"},
		{name: "port not numeric", key: "LAVALINK_PORT", value: "twenty", wantErr: config.ErrInvalid, names: "LAVALINK_PORT"},
		{name: "port zero", key: "LAVALINK_PORT", value: "0", wantErr: config.ErrInvalid, names: "LAVALINK_PORT"},
		{name: "port negative", key: "LAVALINK_PORT", value: "-1", wantErr: config.ErrInvalid, names: "LAVALINK_PORT"},
		{name: "port above range", key: "LAVALINK_PORT", value: "65536", wantErr: config.ErrInvalid, names: "LAVALINK_PORT"},
		{name: "secure not a boolean", key: "NODE_SECURE", value: "maybe", wantErr: config.ErrInvalid, names: "NODE_SECURE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(tt.key, tt.value)

			_, err := config.Load()
			require.ErrorIs(t, err, tt.wantErr)
			require.Contains(t, err.Error(), tt.names)
		})
	}
}

func TestLoadPortBoundariesAreAccepted(t *testing.T) {
	for _, port := range []string{"1", "65535"} {
		t.Run(port, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv("LAVALINK_PORT", port)

			cfg, err := config.Load()
			require.NoError(t, err)
			require.Equal(t, "lavalink.internal:"+port, cfg.LavalinkAddress)
		})
	}
}

func TestLoadErrorsNeverContainSecrets(t *testing.T) {
	const (
		secretToken    = "s3cr3t-discord-token-do-not-log"
		secretPassword = "s3cr3t-lavalink-password-do-not-log"
	)

	setValidEnv(t)
	t.Setenv("TOKEN", secretToken)
	t.Setenv("LAVALINK_PASSWORD", secretPassword)
	// Force every other variable to fail so the joined error is as long as it gets.
	t.Setenv("GUILD_ID", "not-a-snowflake")
	t.Setenv("NODE_NAME", "")
	t.Setenv("NODE_SECURE", "maybe")
	t.Setenv("LAVALINK_HOST", "")
	t.Setenv("LAVALINK_PORT", "70000")

	_, err := config.Load()
	require.Error(t, err)

	msg := err.Error()
	require.NotContains(t, msg, secretToken)
	require.NotContains(t, msg, secretPassword)
}
