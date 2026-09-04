package music

import (
	"errors"
	"fmt"
)

// The recurring command failures. Handlers match with errors.Is and never on the
// user-facing text, so the German copy can be reworded without touching tests.
var (
	ErrNoPlayer       = errors.New("no active player")
	ErrNotInVoice     = errors.New("caller not in a voice channel")
	ErrNothingPlaying = errors.New("player has no current track")
	ErrQueueEmpty     = errors.New("queue is empty")
	ErrNoResults      = errors.New("no tracks found")
	ErrForeignGuild   = errors.New("interaction is not from the configured guild")
)

// GenericErrorMessage is shown for any failure that is not one of the sentinels.
const GenericErrorMessage = msgGeneric

// ErrNoPlayer and ErrNothingPlaying stay separate errors because the code
// distinguishes them, but a caller who hits either one is in the same situation
// and takes the same action, so they share a message.
var userMessages = []struct {
	err     error
	message string
}{
	{ErrNoPlayer, msgNothingPlaying},
	{ErrNotInVoice, msgNotInVoice},
	{ErrNothingPlaying, msgNothingPlaying},
	{ErrQueueEmpty, msgQueueEmpty},
	{ErrNoResults, msgNoResults},
	{ErrForeignGuild, msgForeignGuild},
}

// NoResultsError reports that an identifier resolved to nothing, so the reply can
// name what was searched for.
type NoResultsError struct {
	Identifier string
}

func (e *NoResultsError) Error() string {
	return fmt.Sprintf("no tracks found for %q", e.Identifier)
}

func (e *NoResultsError) Unwrap() error { return ErrNoResults }

// UserMessage bounds the quoted identifier, because /play sets no maximum length
// and an unbounded quote would push the embed description past what Discord
// accepts - failing the very reply that reports the failure.
func (e *NoResultsError) UserMessage() string {
	return msgNoResultsFor(truncate(e.Identifier, quotedInputLimit))
}

// LoadError reports that loading an identifier failed at the Lavalink node.
type LoadError struct {
	Identifier string
	Err        error
}

func (e *LoadError) Error() string {
	return fmt.Sprintf("load %q: %v", e.Identifier, e.Err)
}

func (e *LoadError) Unwrap() error { return e.Err }

// UserMessage deliberately drops e.Err: the upstream text is English, is written
// by Lavalink rather than by the bot, and belongs in the log record instead.
func (e *LoadError) UserMessage() string {
	return msgLoadFailed
}

// UserMessage translates err into the German message shown to the caller and
// reports whether the error was recognised. Unrecognised errors get
// GenericErrorMessage and are the caller's cue to log at error level.
func UserMessage(err error) (string, bool) {
	var withMessage interface{ UserMessage() string }
	if errors.As(err, &withMessage) {
		return withMessage.UserMessage(), true
	}
	for _, m := range userMessages {
		if errors.Is(err, m.err) {
			return m.message, true
		}
	}
	return GenericErrorMessage, false
}
