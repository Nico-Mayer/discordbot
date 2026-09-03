## Why

`/play` with a playlist URL enqueues exactly one track and silently discards the rest. A 40 track playlist becomes one song with no indication that anything was dropped. This is the gap in the bot most likely to read as a bug rather than a missing feature.

It is deliberately excluded from `refactor-bot-architecture`, which preserves the current behaviour and only guards the empty-slice case. This change is scoped and designed now so the implementation detail exists before anyone starts building, but it is **not ready to implement**: it has no task breakdown yet and depends on the refactor landing first.

## What Changes

- `/play` with an identifier that resolves to a playlist enqueues every track in that playlist, not only the first.
- Playback starts from the playlist's selected track when the source names one, rather than always from index 0.
- The confirmation embed reports the playlist name and how many tracks were added.
- A queue size cap prevents a very large playlist from filling memory or making `/queue` unusable. Tracks beyond the cap are refused and the user is told.
- `/queue` gains pagination, because a playlist makes a queue longer than one embed can hold the normal case rather than an edge case. `refactor-bot-architecture` caps the list at 20 with an "and N more" line; that cap is the fallback this change replaces.
- Search results keep their current behaviour. A search returns the top hit only, which is what the user asked for. **This change is only about playlists.**

**Non-goals**

- No playlist management commands (no save, load, or share).
- No shuffle-on-add, no dedupe, no reordering.
- No per-user queue limits or permissions.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `music-playback`: the `/play` playlist scenario changes from "use the first track, discard the rest" to enqueueing the whole playlist, and the `/queue` requirement changes from a truncated single reply to a paginated one. Both requirements are created by `refactor-bot-architecture`, so this change's delta must be written against the specs as they exist once that change is archived.

## Impact

- **Code**: the playlist branch of the `/play` result handler, the queue (a cap and possibly a bulk add), the queue embed builder, and a new pagination interaction handler with component state.
- **Depends on**: `refactor-bot-architecture`. This change assumes the service and adapter split, because the playlist branch needs to be unit-testable and the pagination handler needs somewhere to live that is not an event method.
- **Dependencies**: none new expected. Pagination uses disgo's existing component and interaction handling.
- **Configuration**: likely one new variable for the queue cap, or a constant if a cap does not need tuning.
