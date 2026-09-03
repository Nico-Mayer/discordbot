package music

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEverySentinelHasAUserMessage(t *testing.T) {
	t.Parallel()

	sentinels := []error{ErrNoPlayer, ErrNotInVoice, ErrNothingPlaying, ErrQueueEmpty, ErrNoResults, ErrForeignGuild}
	require.Len(t, userMessages, len(sentinels))

	for _, sentinel := range sentinels {
		t.Run(sentinel.Error(), func(t *testing.T) {
			t.Parallel()

			msg, known := UserMessage(sentinel)
			require.True(t, known)
			require.NotEmpty(t, msg)
			require.NotEqual(t, GenericErrorMessage, msg)
		})
	}
}

func TestUserMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		wantKnown bool
		wantMsg   string
		contains  string
	}{
		{name: "no player", err: ErrNoPlayer, wantKnown: true, wantMsg: "Kein Player gefunden"},
		{name: "not in voice", err: ErrNotInVoice, wantKnown: true, wantMsg: "Du musst in einem Sprachkanal sein!"},
		{name: "nothing playing", err: ErrNothingPlaying, wantKnown: true, wantMsg: "Es wird gerade nichts abgespielt"},
		{name: "queue empty", err: ErrQueueEmpty, wantKnown: true, wantMsg: "Keine weiteren Titel in der Warteschlange"},
		{name: "no results", err: ErrNoResults, wantKnown: true, wantMsg: "Nichts gefunden"},
		{name: "foreign guild", err: ErrForeignGuild, wantKnown: true, wantMsg: "Dieser Bot ist für diesen Server nicht freigeschaltet"},
		{name: "wrapped sentinel", err: fmt.Errorf("update player: %w", ErrNoPlayer), wantKnown: true, wantMsg: "Kein Player gefunden"},
		{name: "no results names the identifier", err: &NoResultsError{Identifier: "never gonna give you up"}, wantKnown: true, contains: "never gonna give you up"},
		{name: "load error describes the failure", err: &LoadError{Identifier: "x", Err: errors.New("node unreachable")}, wantKnown: true, contains: "node unreachable"},
		{name: "unknown error", err: errors.New("boom"), wantKnown: false, wantMsg: GenericErrorMessage},
		{name: "nil error", err: nil, wantKnown: false, wantMsg: GenericErrorMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, known := UserMessage(tt.err)
			require.Equal(t, tt.wantKnown, known)
			if tt.wantMsg != "" {
				require.Equal(t, tt.wantMsg, msg)
			}
			if tt.contains != "" {
				require.Contains(t, msg, tt.contains)
			}
		})
	}
}

func TestNoResultsErrorUnwrapsToSentinel(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("enqueue: %w", &NoResultsError{Identifier: "x"})
	require.ErrorIs(t, err, ErrNoResults)
}

func TestLoadErrorUnwrapsToCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("timeout")
	err := fmt.Errorf("enqueue: %w", &LoadError{Identifier: "x", Err: cause})
	require.ErrorIs(t, err, cause)
}
