## Context

See proposal.md - Why. The relevant current state:

- German strings live in four files: `internal/music/errors.go` (sentinel messages), `embed.go` (titles, footers, field names), `handlers.go` (pause/stop/skip confirmations) and `commands.go` (still English).
- `errorEmbed`, `successEmbed` and `infoEmbed` each hard-code their own icon and prepend it to the description. Call sites that want a different icon add a second one, which is how `queueEmbed`'s empty case ends up rendering `ℹ️` and `📋` side by side and how `/skip` ends up with a trailing icon.
- Handlers already match with `errors.Is` and never on message text (`internal/music/errors.go:8-16`), so rewording is safe for control flow.
- Tests already have a `capturingLogger` that decodes records into maps (`internal/music/handlers_test.go:19-50`), so log keys are cheap to assert.
- Copy is asserted verbatim in `errors_test.go`, `handlers_test.go` and `embed_test.go`. Those assertions move with the strings.

## Goals / Non-Goals

**Goals:**

- One reviewable German copy surface and one written rule for which language goes where.
- Every reply readable on its own, without an icon and without upstream English text.
- Log records that group: static messages, standard keys.

**Non-Goals:**

- Runtime language selection, locale detection, or a second language. The bot serves one German guild.
- An i18n dependency or message catalogue format.
- Changing which commands exist, what they do, or the embed colour scheme.
- Restyling the logs' output format (`charm.land/log/v2` stays as configured in `main.go`).

## Decisions

### German copy lives in `internal/music/copy.go` as constants, not in an i18n catalogue

A new `internal/music/copy.go` holds every German string as a constant or small formatting helper, with a file comment stating the German-for-users / English-for-operators rule. `errors.go`, `embed.go`, `handlers.go` and `commands.go` reference those constants instead of literals.

*Alternatives considered.* `golang.org/x/text/message` or `go-i18n`: both solve translation across locales, which is not the problem here - there is one locale, and only one string needs a plural rule. They add a dependency, a catalogue file, and an initialisation step to save nothing. Rejected. Leaving the literals where they are: rejected, the specs require one reviewable place, and it is exactly the spread that let the two-icon bug survive.

### Icon becomes a parameter, so a reply cannot accumulate two

`errorEmbed`, `successEmbed` and `infoEmbed` are re-expressed on top of one `statusEmbed(icon, color, text string)`. A call site that needs a different icon passes it rather than embedding a second one in the text. No user-facing string in `copy.go` contains an icon.

That makes the one-icon rule structural rather than a convention to remember, and it is what fixes the empty-queue reply.

### Sentinel copy collapses "no player" and "nothing playing" into one message

`ErrNoPlayer` and `ErrNothingPlaying` stay separate errors, because the code distinguishes them, but they map to the same user message. The distinction is internal: a user who sees either one is in the same situation and takes the same action.

### The proposed copy

Command descriptions:

| Command | Description |
| --- | --- |
| `/play` | Spielt einen Titel ab oder stellt ihn in die Warteschlange |
| `/play` option, renamed `identifier` → `titel` | Link oder Suchbegriff |
| `/pause` | Pausiert die Wiedergabe oder setzt sie fort |
| `/stop` | Stoppt die Wiedergabe und leert die Warteschlange |
| `/skip` | Springt zum nächsten Titel |
| `/now-playing` | Zeigt den Titel, der gerade läuft |
| `/queue` | Zeigt die Warteschlange |

Errors:

| Error | Message |
| --- | --- |
| `ErrNoPlayer`, `ErrNothingPlaying` | Gerade läuft nichts. Mit `/play` startest du einen Titel. |
| `ErrNotInVoice` | Tritt zuerst einem Sprachkanal bei. |
| `ErrQueueEmpty` | Die Warteschlange ist leer. Mit `/play` fügst du Titel hinzu. |
| `ErrNoResults` | Nichts gefunden. Prüfe den Link oder versuche einen anderen Suchbegriff. |
| `NoResultsError` | Nichts gefunden für `%s`. Prüfe den Link oder versuche einen anderen Suchbegriff. |
| `ErrForeignGuild` | Dieser Bot ist für diesen Server nicht freigeschaltet. |
| `LoadError` | Der Titel konnte nicht geladen werden. Versuche es noch einmal. |
| generic | Das hat nicht geklappt. Versuche es noch einmal. |

Confirmations and embeds:

| Reply | Icon | Text |
| --- | --- | --- |
| `/pause` paused | ⏸️ | Wiedergabe pausiert |
| `/pause` resumed | ▶️ | Wiedergabe fortgesetzt |
| `/stop` | ⏹️ | Wiedergabe gestoppt, Warteschlange geleert |
| `/skip` | ⏭️ | Titel übersprungen |
| `/queue` empty | 📋 | Die Warteschlange ist leer |
| `/play` started (author line) | ▶️ | Läuft jetzt |
| `/play` queued (title) | 📋 | Zur Warteschlange hinzugefügt |
| `/now-playing` (title) | 🎶 | Läuft gerade |
| queue list (title) | 📋 | Warteschlange |
| field names | - | Dauer, Position |
| queue footer | - | %d Titel in der Warteschlange |
| queue residual, n = 1 | - | … und 1 weiterer Titel |
| queue residual, n > 1 | - | … und %d weitere Titel |

Wording rules applied throughout: informal *du*; the full `-e` imperative (`Prüfe`, `Versuche`) rather than the clipped colloquial form; `noch einmal` rather than `nochmal`; no exclamation marks; the outcome first and the recovery action second.

### `trackEmbed` gains an author line rather than a restructured body

`trackEmbed` currently puts the track title in the embed title and says nothing about what just happened, so stripping its icons leaves a reply that does not state its own outcome. Setting the embed's author line to `▶️ Läuft jetzt` adds the missing statement above the title and leaves the existing title, link, artwork and field layout untouched.

*Alternative considered.* Moving the track into the description with a `▶️ Läuft jetzt` title, matching `queuedEmbed`. More consistent between the two `/play` outcomes, but it visibly restyles the most common reply the bot sends, which is more than a copy change needs to do. The author line is reversible in one line if the consistency turns out to matter more.

### Quoted input is bounded where it is formatted, not where it is displayed

`NoResultsError.UserMessage` is the only copy that quotes the member's own input back. The `/play` option sets no `MaxLength`, so Discord accepts up to 6000 characters, and `errorEmbed` - unlike `queueEmbed` - never truncates. A long enough value therefore produces an embed description over the 4096-character limit, and the reply reporting the failure fails to send.

The existing `truncate` helper in `embed.go` already cuts on a rune boundary and appends an ellipsis. `NoResultsError.UserMessage` applies it to the identifier before formatting, capping the quoted value well below the limit rather than capping the assembled description. Bounding the input keeps the German sentence around it intact; bounding the finished description would cut the recovery advice off the end, which is the part the member needs.

Setting `MaxLength` on the option as well is left to the separate input-handling change: it is a second line of defence, and it does not remove the need to bound what is quoted.

### Log changes are a rename pass plus two fixes

Keys are renamed (`err`→`error`, `guild`→`guild_id`, `user`→`user_id`, `title`/`track`→`track_title`), the concatenated websocket message becomes static with a `cause` attribute, and `main.go` reports its two fatal paths through the logger. The existing message style - `could not <verb phrase>` for failures, `<subject> <verb>ed` for events - is already consistent and is kept, so the pass stays mechanical.

Building the logger before `config.Load()` in `main.go` is what lets the configuration failure be logged. The logger needs no configuration itself, so the reorder is free.

### The `/play` option is renamed, the command names are not

A command name is invoked; an option name is read. Discord prints the option name as the label beside the value being typed, which makes `identifier` the one piece of system vocabulary a member cannot avoid reading to use the bot's main command - and it names the value's data type rather than what the member is supplying. It becomes `titel`, matching the glossary term the rest of the copy uses for a track.

`titel` alone does not say that a link is accepted too, so the description carries that: "Link oder Suchbegriff". Short label, description resolving the ambiguity.

*Alternatives considered.* `song`: an English loanword, and the glossary already settles on Titel. `suche`: names only half of what the option takes. `link-oder-suche`: accurate, but a label should not carry what a description can. `wunsch`: natural for a music bot but coy about what to actually type.

Command names stay English. `/play`, `/queue` and `/now-playing` are the established Discord convention, members type them from muscle memory, and translating them buys nothing a description cannot say.

## Risks / Trade-offs

- The option rename breaks a member who types `identifier:` by hand → Discord re-syncs the command set on startup, so the picker updates itself; only spelled-out invocations are affected, and the task is separable from the rest of the change.
- Verbatim copy assertions in three test files must all move at once, or the suite fails midway through the change → the copy constants land first, and the tests are updated to assert against those constants rather than against repeated literals, so a future rewording touches one place.
- Renaming log keys breaks any external log query or alert that matched `err` or `guild` → this is a personal bot with no log pipeline; the keys are renamed once, now, rather than growing a compatibility layer.
- A German string table has no test that can prove the wording is good → the checkable parts are specified instead: no English glossary term in a user reply, no upstream error text in `LoadError`'s message, exactly one icon per reply.

## Migration Plan

Deploy is a restart. On startup the command set is re-synced, which is what publishes the German descriptions and the renamed option. Rollback is a redeploy of the previous image; the previous build re-syncs the old command set the same way. Nothing is persisted, so there is no data to migrate.

Input handling for the `/play` value - whitespace, wrapping characters, scheme case, blank values - is deliberately not touched here; it is the `normalize-play-identifier` change.
