package music

import "fmt"

// Every string a Discord user reads is German and defined here, so the wording
// can be reviewed as a whole. Everything an operator reads - log messages, Go
// error strings, comments - stays English.
//
// No string here carries a status icon. Icons are passed to statusEmbed, which
// is what keeps a reply from accumulating two.

// Slash command descriptions. The command names themselves stay English.
const (
	descPlay       = "Spielt einen Titel ab oder in die Warteschlange"
	descPause      = "Pausiert die Wiedergabe oder setzt sie fort"
	descStop       = "Stoppt die Wiedergabe und leert die Warteschlange"
	descSkip       = "Springt zum nächsten Titel"
	descNowPlaying = "Zeigt den Titel, der gerade läuft"
	descQueue      = "Zeigt die Warteschlange"

	// The option label is read by the member, so it is German too. Only the
	// command names stay English.
	optionPlayName = "titel"
	optionPlayDesc = "Link oder Suchbegriff"
)

// Error messages. Each states what did not happen and, where the reader can do
// something about it, what to try next.
const (
	msgNothingPlaying = "Gerade läuft nichts. Mit `/play` startest du einen Titel."
	msgNotInVoice     = "Tritt zuerst einem Sprachkanal bei."
	msgQueueEmpty     = "Die Warteschlange ist leer. Mit `/play` fügst du Titel hinzu."
	msgNoResults      = "Nichts gefunden. Prüfe den Link oder versuche einen anderen Suchbegriff."
	msgForeignGuild   = "Dieser Bot ist für diesen Server nicht freigeschaltet."
	msgLoadFailed     = "Der Titel konnte nicht geladen werden. Versuche es noch einmal."
	msgGeneric        = "Das hat nicht geklappt. Versuche es noch einmal."
)

// Confirmations.
const (
	replyPaused     = "Wiedergabe pausiert"
	replyResumed    = "Wiedergabe fortgesetzt"
	replyStopped    = "Wiedergabe gestoppt, Warteschlange geleert"
	replySkipped    = "Titel übersprungen"
	replyQueueEmpty = "Die Warteschlange ist leer"
)

// Embed titles, author lines and field names.
const (
	titleNowPlaying  = "Läuft gerade"
	titleQueued      = "Zur Warteschlange hinzugefügt"
	titleQueue       = "Warteschlange"
	authorNowPlaying = "Läuft jetzt"
	fieldDuration    = "Dauer"
	fieldPosition    = "Position"
)

// quotedInputLimit caps how much of the member's own input a reply quotes back.
// It is far below the embed description limit, so the surrounding sentence can
// never push the reply past what Discord accepts.
const quotedInputLimit = 200

// msgNoResultsFor names what was searched for, bounded so that a value long
// enough to breach the description limit cannot stop the reply from sending.
func msgNoResultsFor(identifier string) string {
	return fmt.Sprintf("Nichts gefunden für `%s`. Prüfe den Link oder versuche einen anderen Suchbegriff.", identifier)
}

// footerQueueCount is the total shown under the queue listing.
func footerQueueCount(n int) string {
	return fmt.Sprintf("%d Titel in der Warteschlange", n)
}

// lineQueueResidual summarises the queued tracks the listing left out.
func lineQueueResidual(n int) string {
	if n == 1 {
		return "… und 1 weiterer Titel"
	}
	return fmt.Sprintf("… und %d weitere Titel", n)
}
