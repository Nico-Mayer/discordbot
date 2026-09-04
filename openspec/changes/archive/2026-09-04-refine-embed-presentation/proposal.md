## Why

The bot's replies are correct but visually inconsistent: the three track replies (`/play` started, `/play` queued, `/now-playing`) each use a different embed shape, a different colour, and a different icon, so the same track looks like three unrelated things. Two concrete defects also sit in the current builders: an over-long track title is written straight into `Embed.Title` unbounded, which can make Discord reject the very reply it belongs to, and a livestream renders a meaningless duration. This change settles one embed layout, bounds every field, and trims the icon inventory instead of growing it.

## What Changes

- **One track card.** `/play` (started), `/play` (queued) and `/now-playing` share a single embed shape: state on the author line, track title as the clickable title, artist below it, artwork as a thumbnail, and facts in inline fields. The full-width `Image` on the started-playing reply becomes a thumbnail like the other two.
- **Icons that mark a state, none that decorate.** The inventory keeps the icons that save the reader a sentence - failure, notice, playing, paused, stopped, skipped, queued - and drops the decorative ones (`🎶` beside "Warteschlange", `📋` on a title). One icon per reply, leading the element it marks: the text of a status line, the author line of a card, or the line naming the track playing now in `/queue`. Nothing decorates a title, a field or a footer, and the five-space `iconPad` hack becomes a single space.
- **Colour signals something.** Colour carries meaning on two axes instead of being one flat accent: a **track card** takes the colour of the service it was resolved from, so tracks from the same service read as a group and provenance is visible before the footer is read; a **status reply** takes the colour of the state it reports - failure, succeeded, held (paused), or nothing active. The failure colour is reserved for failures, so the two axes cannot be confused. An unknown source falls back to one accent colour.
- **Every embed field is bounded.** Titles, author lines, field values and footers are truncated to Discord's per-field limits with a visible marker, measured in runes rather than bytes. **A long track title can no longer fail an interaction.**
- **Livestreams read as livestreams.** A track with `IsStream` set shows a "Live" marker instead of a duration and no progress line.
- **`/now-playing` shows progress.** A single text progress line renders elapsed and total position, replacing the footer that shows only `position / length`.
- **`/queue` reads as a queue.** The listing names the track playing now above the waiting tracks, and the footer states the total count and total remaining duration rather than only the count.
- **Source attribution.** Where Lavalink reports `SourceName`, the track card footer names it (YouTube, Spotify), so a reader can tell where a track came from.

**Non-goals**

- No requester attribution ("added by @x"). It needs the caller's identity plumbed through the service and stored per track, which is a behaviour change, not a presentation one.
- No buttons, no pagination, no interactive controls. `/queue` pagination belongs to `support-playlists`, which already owns it.
- No change to the German wording rules, the term glossary, or the tone rules. Existing copy is reworded only where a new element (progress line, live marker, source name, queue header) needs a string.
- No custom or animated emoji, and no per-source icons.

## Capabilities

### New Capabilities

- `embed-presentation`: the visual contract for every reply - the shared track card shape, the semantic colour palette, per-field bounds, restrained icon placement, progress and live rendering, and artwork fallback.

### Modified Capabilities

- `interface-copy`: the requirement "A reply carries exactly one status icon" is narrowed to the part that is about copy - text MUST carry the outcome without an icon - while placement, count and inventory of icons move to `embed-presentation`, so only one spec owns them.
- `music-playback`: the `/now-playing` scenario changes from "elapsed alongside total length" to a rendered progress line, and gains a livestream scenario; the `/queue` scenarios gain the currently playing header and the total remaining duration in the footer.

## Impact

- **Code**: `internal/music/embed.go` (rewritten around one card builder and a bounds helper), `internal/music/copy.go` (new strings for the queue header, live marker, progress line), `internal/music/embed_test.go` (the icon-counting helpers shrink with the inventory). `handlers.go` is unchanged apart from which builder it calls.
- **Behaviour**: reply appearance only. No command signature, no service method, no error mapping, no logging changes.
- **Dependencies**: none new. `IsStream` and `SourceName` already exist on `lavalink.TrackInfo`.
- **Configuration**: none.
- **Interaction with pending changes**: `support-playlists` replaces the 20-track cap with pagination and reworks the queue embed. Its delta is written against the current queue requirement, so it must be rebased on this one if this lands first.
