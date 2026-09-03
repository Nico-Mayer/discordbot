## Context

See `proposal.md` - Why. This document exists to record implementation detail while it is fresh, not to authorise building. There is no `tasks.md` yet, on purpose.

Current behaviour, in `bot/handlers.go`:

```go
b.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
    func(track lavalink.Track)       { toPlay = &track },
    func(playlist lavalink.Playlist) { toPlay = &playlist.Tracks[0] },   // 39 tracks dropped
    func(tracks []lavalink.Track)    { toPlay = &tracks[0] },            // correct: top hit
    ...
))
```

The single `toPlay *lavalink.Track` variable is the constraint. It can hold one track, so the playlist callback has nowhere to put the rest.

Verified against `disgolink v3.1.0`:

```go
type Playlist struct {
    Info       PlaylistInfo   // { Name string; SelectedTrack int }
    PluginInfo RawData
    Tracks     []Track
}
```

`Info.SelectedTrack` is the index the user actually clicked, or `-1` when the source does not name one. Nothing in the current code reads it.

## Goals / Non-Goals

**Goals:**

- One playlist URL enqueues the whole playlist, with playback starting where the user pointed.
- The queue cannot be made unbounded by a single command.
- `/queue` stays usable at playlist length.

**Non-Goals:**

- Changing search behaviour. The `func(tracks []lavalink.Track)` callback keeps taking `tracks[0]`; a search should return the top hit, not 20 results.
- Reordering, shuffling, or de-duplicating on add.

## Decisions

### The result handler returns a set, not a single track

The `toPlay *lavalink.Track` variable becomes a small struct so the playlist callback has somewhere to put everything:

```go
type loadResult struct {
    tracks   []lavalink.Track  // in play order, first element plays now
    playlist string            // playlist name, empty for a single track or search
    total    int               // tracks resolved before the cap was applied
}
```

The three success callbacks then differ only in how they fill it:

```
  track callback     -> tracks = [t],                       playlist = ""
  search callback    -> tracks = [tracks[0]],               playlist = ""     (unchanged)
  playlist callback  -> tracks = ordered(playlist),          playlist = Info.Name
```

`Queue.Add` is already variadic, so the enqueue side needs no signature change.

### Playback starts at `SelectedTrack`, and the rest follows in order

The obvious reading of "enqueue the playlist" is index 0 first. That is wrong for the common case: pasting a YouTube link that happens to sit inside a playlist sets `SelectedTrack` to that track, and the user expects *that* song to play.

```
  SelectedTrack = -1            SelectedTrack = 7
  +---------------------+       +----------------------------------+
  | play  t0            |       | play  t7                         |
  | queue t1..tN        |       | queue t8..tN, then t0..t6        |
  +---------------------+       +----------------------------------+
```

Rotating rather than truncating means nothing is silently dropped. Whether the wrap-around is right is a genuine open question, recorded below.

### The cap is applied at enqueue, and the user is told

A cap is required, otherwise one command can pin a large playlist in memory and make `/queue` meaningless. The cap applies to the resulting queue length, not the playlist length, so repeated `/play` calls cannot walk past it either.

```
  cap = 100 (proposed)

  queue has 90, playlist has 40
    -> enqueue 10, reply "added 10 of 40 tracks, queue is full"
```

Refusing loudly beats silently truncating, which is the exact failure this change exists to remove. A cap of 100 is a starting number, not a researched one.

### `/queue` becomes paginated

`refactor-bot-architecture` caps `/queue` at 20 entries with an "and N more" line, which is the correct fix for a queue that is *accidentally* long. Once a playlist can make it long by design, the cap becomes the wrong answer and pagination is the right one.

Shape: previous/next buttons on the embed, page state encoded in the component custom ID rather than held in process memory, so a restart does not orphan live message components and there is no per-message state to expire.

```
  custom_id: "queue:page:3"
      |
      v
  component handler reads page, rebuilds embed from the current queue
```

The queue may have changed between renders. That is acceptable and preferable to snapshotting: the user asked what is queued *now*. A page that no longer exists clamps to the last page.

### Embed copy

`Info.Name` carries the playlist title, so the confirmation says what was added:

```
  single track   "<title>" by <author>
  playlist       "Added 40 tracks from <playlist name>", first track shown as now playing
  capped         "Added 10 of 40 tracks from <playlist name> - queue is full"
```

Copy is German, consistent with the rest of the bot.

## Risks / Trade-offs

- **Rotating at `SelectedTrack` may surprise someone who expects the playlist from the top** → Recorded as an open question. The alternative, playing `SelectedTrack` and enqueueing only what follows it, discards the earlier tracks and reintroduces silent dropping.
- **Pagination state in the custom ID is visible and forgeable** → Harmless. The worst a crafted ID does is render a different page of a queue the user can already see, and the handler clamps out-of-range pages.
- **A cap makes `/play` partially fail, which no other command does** → Accepted. The alternative is an unbounded queue. The reply has to make the partial success obvious rather than looking like a normal confirmation.
- **`/queue` pagination is the first component interaction in the bot** → New surface: component routing, and the 15 minute interaction token window after which buttons stop working. Needs deciding whether expired buttons are disabled or left to fail.

## Open Questions

These are safe to defer because none of them changes the approach above, only its parameters:

- Should the playlist wrap around at `SelectedTrack`, or start there and drop the earlier tracks?
- What is the right queue cap? 100 is a guess. Worth checking against what the bot is actually used for.
- Should the cap be a config variable or a constant?
- Do expired pagination buttons get disabled on a best-effort basis, or left to fail with the generic error reply?
- Should a playlist that resolves to a single track render as a track confirmation rather than a playlist one?
