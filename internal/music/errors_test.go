package music

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEverySentinelHasAUserMessage(t *testing.T) {
	t.Parallel()

	sentinels := []error{ErrNoPlayer, ErrNotInVoice, ErrNothingPlaying, ErrQueueEmpty, ErrNoResults, ErrForeignGuild, ErrEmptyIdentifier}
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
		{name: "no player", err: ErrNoPlayer, wantKnown: true, wantMsg: msgNothingPlaying},
		{name: "not in voice", err: ErrNotInVoice, wantKnown: true, wantMsg: msgNotInVoice},
		{name: "nothing playing", err: ErrNothingPlaying, wantKnown: true, wantMsg: msgNothingPlaying},
		{name: "queue empty", err: ErrQueueEmpty, wantKnown: true, wantMsg: msgQueueEmpty},
		{name: "no results", err: ErrNoResults, wantKnown: true, wantMsg: msgNoResults},
		{name: "foreign guild", err: ErrForeignGuild, wantKnown: true, wantMsg: msgForeignGuild},
		{name: "empty identifier", err: ErrEmptyIdentifier, wantKnown: true, wantMsg: msgEmptyInput},
		{name: "wrapped sentinel", err: fmt.Errorf("update player: %w", ErrNoPlayer), wantKnown: true, wantMsg: msgNothingPlaying},
		{name: "no results names the identifier", err: &NoResultsError{Identifier: "never gonna give you up"}, wantKnown: true, contains: "never gonna give you up"},
		{name: "load error names no upstream detail", err: &LoadError{Identifier: "x", Err: errors.New("node unreachable")}, wantKnown: true, wantMsg: msgLoadFailed},
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

func TestNoResultsErrorBoundsTheQuotedIdentifier(t *testing.T) {
	t.Parallel()

	// 6000 is the longest value Discord accepts for a string option, so this is
	// the worst case the reply has to survive.
	err := &NoResultsError{Identifier: strings.Repeat("x", 6000)}
	description := errorEmbed(err.UserMessage()).Description

	require.Less(t, len(description), limitDescription, "the reply reporting the failure must itself send")
	require.Contains(t, description, "\u2026`", "the quoted value ends with a visible truncation marker")
	require.Contains(t, description, "Pr\u00fcfe den Link", "truncating the input leaves the advice intact")
}

func TestNoResultsErrorQuotesAShortIdentifierUnchanged(t *testing.T) {
	t.Parallel()

	err := &NoResultsError{Identifier: "never gonna give you up"}
	msg := err.UserMessage()

	require.Contains(t, msg, "`never gonna give you up`")
	require.NotContains(t, msg, "\u2026", "a value that fits is not marked as shortened")
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

func TestLoadErrorKeepsTheUpstreamTextOutOfTheUserMessage(t *testing.T) {
	t.Parallel()

	const upstream = "lavalink node unreachable"
	err := &LoadError{Identifier: "x", Err: errors.New(upstream)}

	require.Contains(t, err.Error(), upstream, "the operator-facing error must carry the detail")
	for word := range strings.FieldsSeq(upstream) {
		require.NotContains(t, err.UserMessage(), word, "the reply must not quote the upstream error")
	}
}
