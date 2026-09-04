## 1. Bounds and primitives

- [x] 1.1 Replace the byte-based `truncate` in `internal/music/embed.go` with a rune-based `clamp(s string, limit int) string` that keeps the visible ellipsis and never cuts inside a rune; verify with table tests covering an ASCII value under and over the limit, a German value, a CJK value, and a limit landing mid-rune
- [x] 1.2 Add the Discord field limit constants (title 256, description 4096, field name 256, field value 1024, footer 2048, author name 256, total 6000) and a `bound(discord.Embed) discord.Embed` pass that clamps every field and trims the description last if the total is still over 6000; verify with a unit test feeding an embed that is over every limit at once
- [x] 1.3 Route every builder's return value through `bound`, then add the enforcement test: for each builder in the package, build it with adversarially long track metadata and assert every limit holds; verify the test fails if `bound` is removed from any one builder
- [x] 1.4 Switch `queueEmbed`'s description budget check from `len` to a rune count; verify with a queue of multi-byte titles that previously stopped listing early and now lists the full 20

## 2. Palette and icon inventory

- [x] 2.1 Reduce the colour constants to `colorError`, `colorAccent`, `colorNeutral` and remove `colorSuccess`; verify no reference to the removed constant remains (`rg colorSuccess internal/` is empty) and the package builds
- [x] 2.2 Reduce the icon constants to `iconError`, `iconSuccess`, `iconInfo`, remove the transport, queue and music-note icons, and replace `iconPad` with a single space; verify `allIcons` in `embed_test.go` is updated and the icon-count tests pass
- [x] 2.3 Point the four confirmation builders (paused, resumed, stopped, skipped) at `iconSuccess` with `colorAccent`, and the informational ones at `iconInfo` with `colorNeutral`; verify each confirmation still carries exactly one icon and its text states the outcome with every icon stripped

## 3. Track rendering helpers

- [x] 3.1 Add a track-length renderer that returns the live marker when `IsStream` is set and `formatDuration` otherwise, with the German marker string added to `copy.go`; verify with unit tests for a stream, a normal track, and a zero length
- [x] 3.2 Fix `artworkURL` to derive the YouTube thumbnail only when `SourceName` names a YouTube source and to report "no thumbnail" otherwise; verify with tests for a track with artwork, a YouTube track without artwork, and a Spotify track without artwork (which must produce no thumbnail rather than a YouTube URL)
- [x] 3.3 Add the fixed-width progress line renderer (elapsed, 12-cell bar, total); verify with tests at position zero, mid-track, exactly at the length, past the length, and with a zero length (bar omitted, times still shown)

## 4. The shared track card

- [x] 4.1 Add the single track card builder with a variant value carrying the author line, colour, and extra fields, replacing `trackEmbed`, `queuedEmbed` and `nowPlayingEmbed`; verify all three call sites in `handlers.go` compile and the existing handler tests pass
- [x] 4.2 Assert the shared structure in tests: for all three variants, the title is the track title, the URL is set when a URI exists, the artwork is a thumbnail and never a full-width image, and no field or author line carries a status icon
- [x] 4.3 Add the source name to the card footer when `SourceName` is non-empty and omit the footer cleanly when it is not; verify with tests for both cases
- [x] 4.4 Add the track-without-URI and track-without-artwork cases for the card; verify the title renders as plain text and the embed sends without a thumbnail

## 5. Queue presentation

- [x] 5.1 Add `Service.Current() (lavalink.Track, bool)` wrapping the player lookup without sentinel errors; verify with service tests for no player, a player with no track, and a playing track
- [x] 5.2 Change `queueEmbed` to take the current track as an optional leading argument and render it above the waiting tracks, and update the `/queue` handler to supply it from `Service.Current()`; verify with tests for a queue with a track playing and a queue with nothing playing (header omitted, not blank)
- [x] 5.3 Extend the queue footer to state the total waiting count and their total duration, skipping streams and marking the total as a lower bound when a stream was skipped; verify with tests for an all-normal queue, a queue containing one stream, and an empty queue
- [x] 5.4 Confirm the empty-queue reply still states the outcome in words with one icon; verify the icon-stripping test covers it

## 6. Verification

- [x] 6.1 Run `go test ./... -race` and `go vet ./...`; verify both pass with no skipped tests
- [x] 6.2 Run the project linter and the modernize check as configured in `mise.toml`; verify no new findings
- [x] 6.3 Review the diff for comments that only restate the code and delete them; verify each remaining comment explains a constraint (a Discord limit, the rune-counting rationale, the artwork fallback condition) rather than the mechanics
- [x] 6.4 Run `openspec validate refine-embed-presentation --strict`; verify it reports the change as valid
- [x] 6.5 Send each of the six commands against a live guild, including one livestream URL, one track with a very long title, and one non-YouTube track without artwork; verify every reply renders and none fails with a Bad Request

## 7. Icons and colour after live testing

- [x] 7.1 Restore the state icons (playing, paused, stopped, skipped, queued) alongside failure and notice, drop `iconSuccess` and `successEmbed`, and point each confirmation at the icon and colour naming the state it moved to; verify pause and resume share neither icon nor colour
- [x] 7.2 Lead each track card's author line with one icon - playing for started and now-playing, queued for appended - and mark the "playing now" line inside `/queue`; verify every reply carries at most one icon and none appears in a title, field or footer
- [x] 7.3 Colour a track card by its source from a brand-colour map with an accent fallback, and keep status replies on the state palette (failure, succeeded, paused, neutral); verify two tracks from one source share a colour, two sources differ, an unknown source falls back, and no source or state colour equals the failure colour
- [x] 7.4 Re-run `mise run check` and `go fix -diff ./internal/music/...`; verify the suite passes and no modernize suggestion lands in a changed file
