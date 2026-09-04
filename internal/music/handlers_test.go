package music

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"
)

// capturingLogger records the records written to it so a log line can be asserted on.
type capturingLogger struct {
	buf *bytes.Buffer
	mu  *sync.Mutex
}

func newCapturingLogger(level slog.Level) (*slog.Logger, *capturingLogger) {
	c := &capturingLogger{buf: &bytes.Buffer{}, mu: &sync.Mutex{}}
	handler := slog.NewJSONHandler(c, &slog.HandlerOptions{Level: level})
	return slog.New(handler), c
}

func (c *capturingLogger) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.Write(p)
}

// records decodes each captured line into a map.
func (c *capturingLogger) records(t *testing.T) []map[string]any {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()

	var out []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(c.buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var record map[string]any
		require.NoError(t, json.Unmarshal(line, &record))
		out = append(out, record)
	}
	return out
}

// fakeReplier records how a failure was reported back to the caller.
type fakeReplier struct {
	created []discord.MessageCreate
	updated []discord.MessageUpdate

	createErr error
	updateErr error
}

func (f *fakeReplier) CreateMessage(messageCreate discord.MessageCreate, _ ...rest.RequestOpt) error {
	f.created = append(f.created, messageCreate)
	return f.createErr
}

func (f *fakeReplier) UpdateInteractionResponse(messageUpdate discord.MessageUpdate, _ ...rest.RequestOpt) (*discord.Message, error) {
	f.updated = append(f.updated, messageUpdate)
	return nil, f.updateErr
}

func TestCheckGuild(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		guildID *snowflake.ID
		wantErr error
	}{
		{name: "configured guild", guildID: ptr(testGuildID)},
		{name: "foreign guild", guildID: ptr(foreignGuildID), wantErr: ErrForeignGuild},
		{name: "no guild at all", guildID: nil, wantErr: ErrForeignGuild},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := checkGuild(testGuildID, tt.guildID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestReplyErrorSendsAnEphemeralMessage(t *testing.T) {
	t.Parallel()

	replier := &fakeReplier{}
	require.NoError(t, replyError(replier, false, msgNothingPlaying))

	require.Len(t, replier.created, 1)
	require.Empty(t, replier.updated)
	require.Equal(t, discord.MessageFlagEphemeral, replier.created[0].Flags)
	require.Len(t, replier.created[0].Embeds, 1)
	require.Equal(t, colorError, replier.created[0].Embeds[0].Color)
	require.Contains(t, replier.created[0].Embeds[0].Description, msgNothingPlaying)
}

func TestReplyErrorEditsADeferredAcknowledgement(t *testing.T) {
	t.Parallel()

	replier := &fakeReplier{}
	require.NoError(t, replyError(replier, true, msgNoResults))

	require.Empty(t, replier.created, "a deferred interaction must not get a second message")
	require.Len(t, replier.updated, 1)
	require.NotNil(t, replier.updated[0].Embeds)
	require.Contains(t, (*replier.updated[0].Embeds)[0].Description, msgNoResults)
}

func TestReplyErrorWrapsSendFailures(t *testing.T) {
	t.Parallel()

	createFailure := errors.New("rate limited")
	require.ErrorIs(t, replyError(&fakeReplier{createErr: createFailure}, false, "x"), createFailure)

	updateFailure := errors.New("unknown webhook")
	require.ErrorIs(t, replyError(&fakeReplier{updateErr: updateFailure}, true, "x"), updateFailure)
}

func TestErrorRepliesUseTheSentinelMessageTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "no player", err: ErrNoPlayer, want: msgNothingPlaying},
		{name: "not in voice", err: ErrNotInVoice, want: msgNotInVoice},
		{name: "nothing playing", err: ErrNothingPlaying, want: msgNothingPlaying},
		{name: "queue empty", err: ErrQueueEmpty, want: msgQueueEmpty},
		{name: "no results", err: ErrNoResults, want: msgNoResults},
		{name: "foreign guild", err: ErrForeignGuild, want: msgForeignGuild},
		{name: "unknown error falls back", err: errors.New("boom"), want: GenericErrorMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, _ := UserMessage(tt.err)
			replier := &fakeReplier{}
			require.NoError(t, replyError(replier, false, msg))
			require.Contains(t, replier.created[0].Embeds[0].Description, tt.want)
		})
	}
}

func TestUnknownErrorsAreLoggedAtErrorLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		wantLevel string
	}{
		{name: "recognised sentinel", err: ErrNoPlayer, wantLevel: "INFO"},
		{name: "unrecognised error", err: errors.New("boom"), wantLevel: "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, captured := newCapturingLogger(slog.LevelDebug)
			_, known := UserMessage(tt.err)
			if known {
				logger.Info("command failed", slog.Any("error", tt.err))
			} else {
				logger.Error("command failed", slog.Any("error", tt.err))
			}

			records := captured.records(t)
			require.Len(t, records, 1)
			require.Equal(t, tt.wantLevel, records[0]["level"])
		})
	}
}

func TestLogUnknownCommand(t *testing.T) {
	t.Parallel()

	logger, captured := newCapturingLogger(slog.LevelDebug)

	require.NotPanics(t, func() {
		logUnknownCommand(context.Background(), logger, "does-not-exist")
	})

	records := captured.records(t)
	require.Len(t, records, 1)
	require.Equal(t, "WARN", records[0]["level"])
	require.Equal(t, "unknown command", records[0]["msg"])
	require.Equal(t, "does-not-exist", records[0]["command"])
}

func TestNewRouterRegistersEveryCommand(t *testing.T) {
	t.Parallel()

	s := newTestService(t, nil, nil)
	mux := NewRouter(s, discardLogger())

	for _, command := range Commands {
		name := command.(discord.SlashCommandCreate).Name
		t.Run(name, func(t *testing.T) {
			matched := mux.Match("/"+name, discord.InteractionTypeApplicationCommand, int(discord.ApplicationCommandTypeSlash))
			require.True(t, matched, "/%s is not routed", name)
		})
	}

	require.False(t, mux.Match("/does-not-exist", discord.InteractionTypeApplicationCommand, int(discord.ApplicationCommandTypeSlash)))
}
