package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func newCapturingLogger(t *testing.T) (*slog.Logger, func() []map[string]any) {
	t.Helper()

	var mu sync.Mutex
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return buf.Write(p)
	}), &slog.HandlerOptions{Level: slog.LevelDebug}))

	return logger, func() []map[string]any {
		mu.Lock()
		defer mu.Unlock()

		var out []map[string]any
		for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var record map[string]any
			require.NoError(t, json.Unmarshal(line, &record))
			out = append(out, record)
		}
		return out
	}
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

func TestLogSelfWarnsWhenTheBotUserCannotBeFetched(t *testing.T) {
	logger, records := newCapturingLogger(t)

	logSelf(context.Background(), logger, func() (*discord.User, error) {
		return nil, errors.New("401 unauthorized")
	})

	got := records()
	require.Len(t, got, 1)
	require.Equal(t, "WARN", got[0]["level"])
	require.Equal(t, "could not fetch the bot user", got[0]["msg"])
	require.Contains(t, got[0]["err"], "401 unauthorized")
}

func TestLogSelfReportsTheIdentity(t *testing.T) {
	logger, records := newCapturingLogger(t)

	logSelf(context.Background(), logger, func() (*discord.User, error) {
		return &discord.User{Username: "musicbot"}, nil
	})

	got := records()
	require.Len(t, got, 1)
	require.Equal(t, "INFO", got[0]["level"])
	require.Equal(t, "musicbot", got[0]["username"])
}

func TestReleaseAllRunsCleanupsInOrder(t *testing.T) {
	logger, _ := newCapturingLogger(t)

	var order []string
	step := func(name string) cleanup {
		return cleanup{name: name, run: func(context.Context) { order = append(order, name) }}
	}
	cleanups := []cleanup{step("leave voice channel"), step("close lavalink"), step("close gateway")}

	releaseAll(context.Background(), cleanups, time.Second, logger)

	require.Equal(t, []string{"leave voice channel", "close lavalink", "close gateway"}, order)
}

func TestReleaseAllLeavesNoGoroutinesBehind(t *testing.T) {
	logger, _ := newCapturingLogger(t)

	var ran int
	cleanups := []cleanup{
		{name: "a", run: func(context.Context) { ran++ }},
		{name: "b", run: func(context.Context) { ran++ }},
	}

	releaseAll(context.Background(), cleanups, time.Second, logger)
	require.Equal(t, 2, ran)
	// goleak.VerifyTestMain in TestMain fails the run if a cleanup goroutine is
	// still alive, which is how a missing cleanup step shows up.
}

func TestReleaseAllReturnsWhenACleanupBlocks(t *testing.T) {
	logger, records := newCapturingLogger(t)

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	cleanups := []cleanup{
		{name: "blocking close", run: func(context.Context) { <-release }},
		{name: "close gateway", run: func(context.Context) {}},
	}

	const timeout = 50 * time.Millisecond
	done := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(done)
		releaseAll(context.Background(), cleanups, timeout, logger)
	}()

	select {
	case <-done:
		require.Less(t, time.Since(start), 5*time.Second, "shutdown must not wait on a stuck cleanup")
	case <-time.After(5 * time.Second):
		t.Fatal("releaseAll hung on a blocking cleanup")
	}

	// Once the shutdown budget is spent, later steps get an already-cancelled
	// context too, so only the blocking step is asserted on.
	var warned []any
	for _, record := range records() {
		if record["msg"] == "shutdown step did not finish in time" {
			warned = append(warned, record["step"])
		}
	}
	require.Contains(t, warned, "blocking close", "a stuck shutdown step must be reported")
}
