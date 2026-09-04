## Why

The bot never leaves a voice channel on its own. If everyone disconnects mid-song it keeps playing to an empty channel indefinitely, holding a voice session and burning Lavalink CPU, and the only way out is someone rejoining to run `/stop`. Separately, when the queue runs dry the bot leaves *instantly*, so queueing one more track means it has already gone.

Both are the same missing concept: the bot has no notion of being idle.

## What Changes

- Introduce two independent idle conditions, each with its own countdown and its own configuration:
  - **nobody listening** (`IDLE_ALONE_SECONDS`): the bot's voice channel contains no users other than the bot itself
  - **queue dry** (`IDLE_EMPTY_QUEUE_SECONDS`): nothing is playing and the queue is empty
- When a condition begins, start its countdown. When either countdown expires, stop the player, clear the queue, and leave the voice channel. Leaving is the same action in both cases.
- Cancel a countdown when its condition clears: a user rejoins the channel, or a track starts playing.
- Both variables default to `60`. Each accepts a non-negative number of seconds, where `0` means leave immediately, or the literal `off` to never leave for that reason. The two are set independently, so any combination of patience is expressible.
- **BREAKING**: the queue-dry path changes observable behaviour at the default. The bot currently leaves the moment the last track ends; it will now linger for `IDLE_EMPTY_QUEUE_SECONDS`, so a track queued within that window continues playback without the bot having to rejoin. `IDLE_EMPTY_QUEUE_SECONDS=0` reproduces the current behaviour exactly.

**Non-goals**

- No pausing while alone. The bot either stays and plays or leaves; a paused-but-present state adds a mode without solving anything.
- No announcement message when leaving. The channel is empty in the nobody-listening case, and a "left due to inactivity" message in the queue-dry case is noise.
- No per-user or per-role exemptions, and no "stay forever" command. `off` covers the permanent case at the configuration level.
- No distinction between other bots and humans, which is not achievable with the current intents. See the design.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `music-playback`: adds idle-leaving behaviour and changes the queue-exhausted scenario, which currently requires leaving the voice channel immediately. That requirement is created by `refactor-bot-architecture`, so this delta must be written against the spec as it exists once that change is archived.
- `bot-lifecycle`: adds the `IDLE_ALONE_SECONDS` and `IDLE_EMPTY_QUEUE_SECONDS` variables to configuration loading and validation, including the `off` literal, and requires both idle timers to be stopped during shutdown.

## Impact

- **Code**: the voice state event handler gains occupancy counting, the service gains two timers and their arm/cancel logic, the track-end handler stops leaving directly and defers to its timer, and shutdown must stop both.
- **Depends on**: `refactor-bot-architecture`. The timer belongs on the service, which does not exist yet, and the cleanup registration that change introduces is where stopping the timers hangs.
- **Configuration**: two new optional variables, plus their entries in `.env.example` and the README. An existing deployment that sets neither gets the 60 second default for both, which is a behaviour change on upgrade for the queue-dry case.
- **Dependencies**: none. `time.AfterFunc` and the existing voice state cache are sufficient. No new gateway intents and no new cache flags.
