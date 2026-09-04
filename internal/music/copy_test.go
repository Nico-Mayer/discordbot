package music

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// everyUserFacingString is the whole copy surface, so the rules below are
// checked against all of it rather than against whatever a test happens to call.
func everyUserFacingString() map[string]string {
	return map[string]string{
		"descPlay":           descPlay,
		"descPause":          descPause,
		"descStop":           descStop,
		"descSkip":           descSkip,
		"descNowPlaying":     descNowPlaying,
		"descQueue":          descQueue,
		"optionPlayName":     optionPlayName,
		"optionPlayDesc":     optionPlayDesc,
		"msgNothingPlaying":  msgNothingPlaying,
		"msgNotInVoice":      msgNotInVoice,
		"msgQueueEmpty":      msgQueueEmpty,
		"msgNoResults":       msgNoResults,
		"msgEmptyInput":      msgEmptyInput,
		"msgForeignGuild":    msgForeignGuild,
		"msgLoadFailed":      msgLoadFailed,
		"msgGeneric":         msgGeneric,
		"replyPaused":        replyPaused,
		"replyResumed":       replyResumed,
		"replyStopped":       replyStopped,
		"replySkipped":       replySkipped,
		"replyQueueEmpty":    replyQueueEmpty,
		"titleNowPlaying":    titleNowPlaying,
		"titleQueued":        titleQueued,
		"titleQueue":         titleQueue,
		"authorNowPlaying":   authorNowPlaying,
		"headingUpNext":      headingUpNext,
		"markerLive":         markerLive,
		"fieldDuration":      fieldDuration,
		"fieldPosition":      fieldPosition,
		"msgNoResultsFor":    msgNoResultsFor("etwas"),
		"footerQueueTotal":   footerQueueSummary(3, "9:21", false),
		"footerQueuePartial": footerQueueSummary(3, "9:21", true),
		"residualOne":        lineQueueResidual(1),
		"residualMany":       lineQueueResidual(7),
	}
}

func TestCopyCarriesNoIcon(t *testing.T) {
	t.Parallel()

	for name, text := range everyUserFacingString() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Zero(t, countIcons(text), "the icon is passed to statusEmbed, never baked into the copy")
		})
	}
}

// untranslated catches the English terms the glossary settles, plus "Player",
// which names an internal concept the reader has no way to act on.
var untranslated = regexp.MustCompile(`(?i)\b(songs?|tracks?|player|queue|lieder?|musikstücke?)\b`)

func TestCopyUsesOneGermanTermPerConcept(t *testing.T) {
	t.Parallel()

	for name, text := range everyUserFacingString() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.NotRegexp(t, untranslated, text, "the glossary settles on Titel, Warteschlange, Sprachkanal, Wiedergabe")
		})
	}
}

var formalAddress = regexp.MustCompile(`\b(Sie|Ihnen|Ihre[nmrs]?)\b`)

func TestCopyAddressesTheReaderInformally(t *testing.T) {
	t.Parallel()

	for name, text := range everyUserFacingString() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.NotRegexp(t, formalAddress, text, "the bot says du, never Sie")
			require.NotContains(t, text, "!", "a reply never exclaims, least of all at a failure")
		})
	}
}

// TestErrorCopyNamesAWayForward covers the errors the reader can act on. A
// foreign guild is deliberately absent: there is nothing the reader can do.
func TestErrorCopyNamesAWayForward(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "nothing playing points at /play", text: msgNothingPlaying, want: "/play"},
		{name: "not in voice names the channel", text: msgNotInVoice, want: "Sprachkanal"},
		{name: "queue empty points at /play", text: msgQueueEmpty, want: "/play"},
		{name: "no results suggests another search", text: msgNoResults, want: "Suchbegriff"},
		{name: "an empty value says what to supply", text: msgEmptyInput, want: "Suchbegriff"},
		{name: "load failure invites a retry", text: msgLoadFailed, want: "noch einmal"},
		{name: "the generic failure invites a retry", text: msgGeneric, want: "noch einmal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Contains(t, tt.text, tt.want)
		})
	}
}

func TestGenericMessageIsNotAVagueNonExplanation(t *testing.T) {
	t.Parallel()

	require.NotContains(t, strings.ToLower(msgGeneric), "schiefgelaufen")
	require.Contains(t, msgGeneric, "noch einmal", "the reader is told what they can do")
}
