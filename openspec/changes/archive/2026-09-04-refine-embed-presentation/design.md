## Context

See `proposal.md` - Why. The constraints that shape the approach:

- Everything user-visible is built in `internal/music/embed.go` (six builders, ~180 lines) with the German strings in `copy.go`. Handlers only choose a builder and send it, so a presentation change stays inside those two files plus their tests.
- Discord's per-field limits, confirmed against the current API docs: title 256, description 4096, field name 256, field value 1024, footer text 2048, author name 256, at most 25 fields, and at most **6000 characters across all embeds of one message**. Exceeding any of them is a `400 Bad Request`, which for `/play` means the acknowledgement is never edited and the member sees a permanent "thinking" state.
- Today only the description is bounded, in bytes, and only in `queueEmbed`. `trackEmbed` writes `track.Info.Title` into `Embed.Title` unbounded.
- `lavalink.TrackInfo` already carries `IsStream`, `SourceName`, `ArtworkURL` and `Position`; no new data has to be fetched.
- `queueEmbed` currently receives only the waiting tracks. Naming the current track needs the player's track, which lives behind `Service.NowPlaying`.
- `support-playlists` is planned and reworks the same queue embed with pagination. This change must not make that harder.

## Goals / Non-Goals

**Goals:**

- One embed shape for a track, expressed as one builder rather than three that happen to look similar.
- Field bounds that cannot be forgotten when a builder is added later.
- Keep the builders pure functions of their inputs, so every requirement in the specs is unit-testable without a Discord client.

**Non-Goals:**

- No message-component or interaction plumbing (that is `support-playlists`).
- No localisation mechanism. Strings stay Go constants in `copy.go`.
- No change to `Service`'s existing method contracts beyond one additive read method.

## Decisions

### One card builder, variant chosen by the caller

Replace `trackEmbed`, `queuedEmbed` and `nowPlayingEmbed` with a single builder taking the track plus a small variant value describing the outcome (author line, colour, and which extra fields to add). Position and queue index are passed as optional extras.

*Why:* the three builders drifted precisely because each owned its own layout - one uses `Author` + full-width `Image`, two use `Title` + `Thumbnail`, and the description repeats the title as a masked link in one of them. A single builder makes the structural requirement in `embed-presentation` true by construction instead of by review.

*Alternative considered:* keep three builders and share small helpers. Rejected: helpers constrain the parts, not the shape, which is what actually diverged.

### Bounds enforced by one pass over the finished embed, not at each call site

Add an internal `bound(embed)` step that every builder's return value passes through. It clamps title, description, author name, footer text, and each field name and value to its own limit, then, if the running total still exceeds 6000, trims the description last (it is the only field long enough to matter).

*Why:* clamping at the call site is one `clamp(...)` call per assignment and is forgotten the first time someone adds a field. A single exit pass is enforceable in a test: iterate every builder in the package with adversarial input and assert the limits hold. That test is the real deliverable here, not the clamp itself.

*Alternative considered:* clamp at each assignment. Rejected as unenforceable. *Also considered:* validating and returning an error. Rejected: a reply that cannot be sent is worse than a shortened one, and the caller has no better option than to shorten anyway.

### Truncation counts runes, and the queue budget switches to runes too

Replace the byte-based `truncate` with a rune-based `clamp`, and change `queueEmbed`'s `b.Len()+len(line)` budget check to count runes.

*Why:* German copy, CJK track titles and emoji in titles are all multi-byte. Byte counting cuts a 4096-character allowance to as little as a quarter of it, which silently drops queue lines that would have fit. Rune counting is still not exactly Discord's counting for astral-plane characters, but it errs on the safe side by at most a factor of two for those and is correct for everything else. The existing rune-boundary care in `truncate` is kept.

### Icons kept where they replace a sentence, dropped where they decorate

Inventory: `iconError`, `iconInfo`, `iconPlaying`, `iconPaused`, `iconStopped`, `iconSkipped`, `iconQueued`. Removed: `iconSuccess` (a generic tick said nothing the state icons do not say better, and `successEmbed` went with it) and `iconMusicNote`/`iconQueue` (pure decoration on a title). The five-space `iconPad` becomes a single space.

Placement is one icon per reply, leading the element it marks: the description of a status line, the author line of a card, or the "playing now" line inside `/queue`. Never a title, field or footer.

*Why:* the transport icons are not decoration - "Wiedergabe pausiert" and "Wiedergabe fortgesetzt" differ by one word, and `▶`/`⏸` tells them apart at a glance while scrolling a channel. Same for `▶` vs `➕` on a card: the difference between "this is playing now" and "this is 5th in line" is the single most-read fact in the reply. A tick on top of that adds nothing.

*Alternative considered:* the three-icon set this change first shipped (failure, tick, notice) with cards and lists carrying none. Rejected after testing it in a live guild: the replies became hard to scan, because the one fact that distinguishes them had no visual anchor.

### Colour on two axes: source for cards, state for status lines

- Card: `sourceColor(track.Info.SourceName)` - a small map of service brand colours (YouTube `0xFF0000`, Spotify `0x1DB954`, SoundCloud `0xFF5500`, Apple Music `0xFA243C`, Twitch `0x9146FF`, Vimeo `0x1AB7EA`, Bandcamp `0x629AA9`, Deezer `0xA238FF`), falling back to `colorAccent` (`0x5865F2`).
- Status line: `colorError` (`0xED4245`) for a failure, `colorSuccess` (`0x57F287`) for a state change that took effect, `colorPaused` (`0xFEE75C`) for playback being held, `colorNeutral` (`0x4E5058`) for stopped and for informational replies.
- List reply: `colorNeutral`. A listing is a container, not an event.

*Why:* a flat accent on every playback reply spent the embed stripe - the most visible pixel in the message - on nothing. Source colour makes it carry provenance, which the footer states in text but nobody reads first, and it groups a session's replies by service. On the status axis, amber for paused is the one state a reader may want to notice without reading.

*Trade-off accepted:* YouTube's brand red sits close to the failure red. The two never share a shape - a card always has a title, an author line and a thumbnail, while a failure is always a one-line description - and a test asserts no source colour equals `colorError` exactly. If it still reads wrong in practice, the fix is one map entry.

### Progress rendered as a fixed 12-cell text bar

`elapsed ━━━━●───────── total`, built from `━`, `●` and `─` at a fixed width, with the knob index derived from `position/length`. Position at or past the length renders full; `length == 0` (and not a stream) omits the bar and shows the times only.

*Why:* plain box-drawing characters are monospace-stable across clients and add no emoji. Emoji-block bars are the usual Discord approach and are exactly the "going overboard" this change avoids.

### Livestreams take a marker where a duration would go

One helper renders a track's length: the live marker when `IsStream` is set, `formatDuration` otherwise. Queue totals skip streams and the footer says the total is a lower bound when at least one stream was skipped.

*Why:* Lavalink reports a livestream's length as a sentinel that currently renders as a nonsense duration. Skipping streams silently would make the total wrong without saying so.

### Artwork falls back only where the fallback is valid

Derive the YouTube thumbnail URL only when `SourceName` identifies a YouTube source; otherwise, with no `ArtworkURL`, send the embed with no thumbnail.

*Why:* the current fallback builds `img.youtube.com/vi/<identifier>/hqdefault.jpg` for every source, so a Spotify or SoundCloud track without artwork gets a URL built from a foreign identifier - a broken image in the reply.

### The queue embed takes the current track as an explicit argument

`queueEmbed` gains a leading parameter for the currently playing track (absent when nothing plays). The `/queue` handler reads it from a new `Service.Current() (lavalink.Track, bool)`, which wraps the same player lookup as `NowPlaying` without the sentinel errors.

*Why:* the builders must stay pure functions for the tests to be worth anything, so the handler does the lookup. `Current` is additive and returns a boolean rather than `ErrNoPlayer`/`ErrNothingPlaying`, because `/queue` is not an error when nothing plays and should not be mapping sentinels to decide that.

*Alternative considered:* call `NowPlaying` from the handler and ignore both sentinels. Rejected: it makes the handler match on errors that mean "normal case".

## Risks / Trade-offs

- **The icon and colour scheme is a taste call** → the icon table, the state palette and the source map are each one block; changing any of them is a small diff. The first iteration of this change stripped icons from cards and lists and used one flat accent; live testing sent it back, and both decisions above record what replaced it and why.
- **Conflict with `support-playlists`, which rewrites the queue embed** → keep the change to `queueEmbed` confined to its signature and its footer/header lines. `support-playlists` is not yet implemented, so its delta gets rebased rather than reconciled in code.
- **Rune counting is not exactly Discord's character counting** → it is conservative in the direction that matters (never over the limit), and the bounds test asserts a margin rather than an exact count.
- **Removing `colorSuccess` and six icons breaks existing tests** → `embed_test.go`'s `allIcons` table and the per-builder assertions are updated as part of the work, not left as follow-up. The icon-stripping test survives, and gets stronger: it now asserts track cards carry no icon at all.
- **The 6000-character total is per message, not per embed** → the bot sends exactly one embed per message today, so the per-embed check is sufficient. If that ever changes, the total check moves to the send path; noted, not built.
