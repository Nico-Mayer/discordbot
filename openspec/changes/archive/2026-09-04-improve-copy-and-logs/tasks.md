## 1. German copy surface

- [x] 1.1 Add `internal/music/copy.go` holding every German string from design.md as constants and format helpers, with a file comment stating the German-for-users / English-for-operators rule; verify `go build ./...` succeeds and no constant contains an icon
- [x] 1.2 Point `errors.go` at the copy constants: `userMessages`, `GenericErrorMessage`, `NoResultsError.UserMessage` and `LoadError.UserMessage`, with `LoadError.UserMessage` no longer formatting the wrapped error; verify `go test ./internal/music -run TestUserMessage` passes with the assertions updated to the new copy
- [x] 1.3 Add a test that `LoadError.UserMessage` contains none of the wrapped error's text while `LoadError.Error` still does; verify it passes

## 2. One icon per reply

- [x] 2.1 Express `errorEmbed`, `successEmbed` and `infoEmbed` on top of a single `statusEmbed(icon, color, text)`; verify existing embed tests still pass
- [x] 2.2 Move the pause, resume, stop and skip confirmations in `handlers.go` to the copy constants with their icon passed to `statusEmbed`, dropping the trailing icon on skip; verify a test asserts the skip reply starts with its icon
- [x] 2.3 Fix the empty-queue reply in `queueEmbed` to render one icon instead of `ℹ️` plus `📋`; verify a test asserts the reply contains exactly one icon
- [x] 2.4 Set `trackEmbed`'s author line to `▶️ Läuft jetzt` and drop the `⏱️` from its duration field name; verify `embed_test.go` asserts the author line and the bare `Dauer` field name
- [x] 2.5 Add a test that strips icons from each reply and asserts the remaining text still names the outcome; verify it passes for the pause, resume, stop, skip and empty-queue replies

## 3. German command surface

- [x] 3.1 Translate the six slash command descriptions in `commands.go` to the design.md table, referencing the copy constants; verify a test asserts every description is non-empty and matches the constants
- [x] 3.2 Rename the `/play` option `identifier` to `titel` with the description "Link oder Suchbegriff", updating `handlers.go`'s `data.String(...)` lookup; verify `go test ./internal/music` passes and the router test still routes `/play`
- [x] 3.3 Update `README.md` where it documents the `/play` option name; verify the README no longer mentions `identifier` as an option name

## 4. Bounded quoting of user input

- [x] 4.1 Apply `embed.go`'s `truncate` to the identifier inside `NoResultsError.UserMessage`, capping it well below the embed description limit; verify a test asserts a 6000-character identifier produces a description under 4096 characters ending in the ellipsis marker
- [x] 4.2 Add a test that a short identifier is quoted unchanged with no ellipsis; verify it passes

## 5. Queue plural rule

- [x] 5.1 Make `queueEmbed`'s residual line read `… und 1 weiterer Titel` for one remaining track and `… und %d weitere Titel` otherwise; verify a table test covers 1, 2 and many remaining tracks

## 6. Log attribute keys

- [x] 6.1 Rename `err` to `error` in every `slog` call across `internal/music` and `internal/app`; verify `rg 'slog.Any\("err"' internal` returns nothing and `go test ./...` passes
- [x] 6.2 Rename `guild` to `guild_id` and `user` to `user_id` across `handlers.go`, `events.go` and `app.go`; verify `rg '"(guild|user)"' internal --glob '!*_test.go'` returns nothing
- [x] 6.3 Rename the track title attribute to `track_title` in `events.go`, replacing both `title` and `track`; verify `events_test.go` asserts the new key
- [x] 6.4 Update `app_test.go:65`, `handlers_test.go:184-186` and any other test asserting an old key; verify `go test ./...` passes

## 7. Static log messages

- [x] 7.1 Replace the concatenated `"lavalink websocket closed for good: "+cause` with a static message plus a `cause` attribute, and give the retryable case its own static message; verify a test asserts the message is identical across two different terminal close codes and that `cause` differs
- [x] 7.2 Sweep the remaining log calls for capitalisation, trailing full stops and interpolated values; verify every message in `internal` is lowercase, punctuation-free at the end, and free of `fmt.Sprintf` or `+`

## 8. Startup failures go through the logger

- [x] 8.1 Build the logger before `config.Load()` in `main.go` and report the configuration failure with `logger.Error("invalid configuration", ...)` before exiting non-zero; verify running the binary with a missing `TOKEN` prints a structured record and exits 1
- [x] 8.2 Report the `app.Run` failure the same way and remove both `fmt.Fprintln(os.Stderr, ...)` calls; verify `rg 'os.Stderr' main.go` matches only the logger construction

## 9. Whole-change verification

- [x] 9.1 Run `go build ./... && go test ./... && go vet ./...`; verify all pass
- [x] 9.2 Re-read every string in `copy.go` against the interface-copy spec: German, one glossary term per concept, *du*, no exclamation mark, no icon, an action where the reader has one; verify each entry matches the design.md table
- [ ] 9.3 Start the bot against the guild and run each of the six commands plus one failing `/play`; verify every reply is German, carries one icon, that command and option names are unchanged, and that the logs are English with the new keys
