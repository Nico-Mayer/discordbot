## Why

The bot serves one German Discord guild, but its user-facing text is inconsistent: slash command descriptions are English, the German embed copy mixes styles (trailing vs. leading icons, an exclamation mark, "Player" left untranslated), several errors are dead ends that state a failure without a way forward, and one error pastes a raw English Lavalink message into a German reply. Nothing records which language belongs where, so every new string re-decides it.

The English logs have the mirror problem: attribute keys drift (`err`, `guild`, `title`), one message interpolates a variable into the message text so identical failures cannot be grouped, and the two fatal startup paths in `main.go` bypass the structured logger entirely and print to stderr.

## What Changes

- Add a written copy standard: German for everything a Discord user reads, English for everything an operator reads. One glossary (Titel, Warteschlange, Sprachkanal, Wiedergabe), informal *du*, sentence case, no exclamation marks.
- Rewrite every user-facing string in German against that standard, so each failure states what happened and what to do next.
- Translate the six slash command descriptions to German. Command names stay English (`/play`, `/pause`, …) because they are the invocation verbs users type, not prose they read.
- Rename the `/play` option `identifier` to `titel` with a German description. `identifier` is system vocabulary for the value's data type; Discord shows it as the label the user types against, so it is read rather than referenced.
- Replace `GenericErrorMessage` ("Etwas ist schiefgelaufen") and stop showing raw Lavalink error text to users; the technical detail moves to the log.
- Fix the double icon on the empty-queue reply, which renders `ℹ️` and `📋` together today.
- Bound the search text quoted back in a "nothing found" reply. The `/play` option sets no maximum length, so a long value is echoed whole into an embed description that never truncates; past Discord's 4096-character limit the failure reply itself fails to send.
- Collect the German strings in one place so the wording can be reviewed as a whole rather than hunted across four files.
- Standardise log messages: lowercase, static, no variables interpolated into the message, and snake_case attribute keys (`guild_id`, `user_id`, `command`, `track_title`, `error`, `reason`, `cause`, `step`).
- Route the two fatal startup failures in `main.go` through the structured logger instead of `fmt.Fprintln(os.Stderr, …)`.

## Capabilities

### New Capabilities

- `interface-copy`: the language rule (German for users, English for operators), the voice and tone standard, the glossary, and the required shape of an error, success, and empty-state reply.
- `structured-logging`: the log language, message style, level policy, required attribute keys, and the rule that no secret or raw upstream error text reaches a user reply.

### Modified Capabilities

- `music-playback`: the `/play` option is renamed `identifier` → `titel`; error replies must name a recovery action instead of only stating the failure and must bound any user input they quote back; the empty-queue reply carries exactly one icon.
- `bot-lifecycle`: configuration and startup failures are reported through the structured logger rather than written directly to stderr.

## Impact

- `internal/music/`: `commands.go`, `errors.go`, `embed.go`, `handlers.go`, `events.go`, `idle.go`, plus a new file holding the German copy.
- `internal/app/app.go`, `main.go`: log attribute keys and the two fatal exit paths.
- Tests: `errors_test.go`, `embed_test.go`, `handlers_test.go` assert on some copy; `service_test.go` and `app_test.go` do not. Handlers already match errors with `errors.Is`, so rewording sentinel copy does not touch matching logic.
- `README.md`: documents the `/play` option name.
- No dependency, API, or persistence change, and no command name change. The only observable behaviour changes are the option rename and the bound on quoted input.
