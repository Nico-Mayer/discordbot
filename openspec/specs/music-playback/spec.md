## Purpose

Lets members of one configured Discord guild play audio in a voice channel through slash commands, with a single track queue that advances automatically and reports every result or failure back to the person who ran the command.

## Requirements

### Requirement: Play a track or search query

The `/play` command SHALL accept a `titel` string that is either a URL or a search query. A value that is not a URL MUST be treated as a YouTube Music search. The bot MUST join the voice channel the requesting member is currently in.

Because loading a track can exceed Discord's initial interaction response window, the bot MUST acknowledge the interaction before loading and then edit that acknowledgement with the result.

#### Scenario: Member is not in a voice channel

- **WHEN** a member runs `/play` while not connected to any voice channel in the guild
- **THEN** the bot MUST reply with an ephemeral error telling them to join a voice channel first
- **AND** the bot MUST NOT join a channel, load a track, or modify the queue

#### Scenario: Search query resolves to a track and nothing is playing

- **WHEN** a member in a voice channel runs `/play` with a search query and the guild has no track playing
- **THEN** the bot MUST join that member's voice channel and start playing the first matching track
- **AND** the bot MUST edit its acknowledgement into a confirmation showing the track title, author, duration, and artwork

#### Scenario: URL is played directly

- **WHEN** a member in a voice channel runs `/play` with a value starting `http://` or `https://`
- **THEN** the bot MUST load that URL as given and MUST NOT prepend a search prefix

#### Scenario: A track is already playing

- **WHEN** a member in a voice channel runs `/play` while a track is already playing in that guild
- **THEN** the bot MUST append the resolved track to the guild queue instead of interrupting playback
- **AND** the bot MUST edit its acknowledgement into a confirmation showing the track title, author, duration, and its position in the queue

#### Scenario: Identifier resolves to a playlist

- **WHEN** the identifier resolves to a playlist
- **THEN** the bot MUST use the playlist's first track and MUST NOT enqueue the remaining tracks
- **AND** if the playlist contains no tracks the bot MUST report that nothing was found rather than failing unexpectedly

Enqueueing a whole playlist is deliberately out of scope for this capability as specified; it is planned as a separate change.

#### Scenario: Nothing found

- **WHEN** the identifier resolves to no tracks
- **THEN** the bot MUST edit its acknowledgement into an error naming the identifier it searched for
- **AND** that error MUST suggest what the member can try instead

#### Scenario: Nothing found for a very long value

- **WHEN** the identifier resolves to no tracks and is long enough that quoting it whole would breach the reply size limit
- **THEN** the bot MUST shorten the quoted identifier with a visible marker
- **AND** the reply MUST be delivered successfully rather than failing the interaction

#### Scenario: Loading fails

- **WHEN** loading the identifier returns an error
- **THEN** the bot MUST edit its acknowledgement into an error stating that the track could not be loaded
- **AND** that error MUST NOT contain the upstream error text, which MUST instead be logged
- **AND** the bot MUST NOT join the voice channel or modify the queue

#### Scenario: Loading takes too long

- **WHEN** loading the identifier does not complete within the configured load timeout
- **THEN** the load MUST be abandoned
- **AND** the bot MUST edit its acknowledgement into an error rather than leaving the acknowledgement unanswered

#### Scenario: The option is named and described in German

- **WHEN** a member opens the `/play` command in Discord
- **THEN** the option MUST be presented as `titel` with a German description naming both kinds of accepted value
- **AND** the previous name `identifier` MUST no longer be accepted

### Requirement: Pause and resume playback

The `/pause` command SHALL toggle the pause state of the guild's active player and MUST report which state it moved to.

#### Scenario: Pausing a playing track

- **WHEN** a member runs `/pause` while a track is playing
- **THEN** playback MUST pause
- **AND** the bot MUST reply that playback is paused

#### Scenario: Resuming a paused track

- **WHEN** a member runs `/pause` while playback is already paused
- **THEN** playback MUST resume
- **AND** the bot MUST reply that playback has resumed

#### Scenario: No active player

- **WHEN** a member runs `/pause` and the guild has no active player
- **THEN** the bot MUST reply with an ephemeral error that no player was found
- **AND** the queue MUST be left untouched

### Requirement: Stop playback and clear the queue

The `/stop` command SHALL stop the current track, clear the guild queue, and leave the voice channel.

#### Scenario: Stopping an active player

- **WHEN** a member runs `/stop` while the guild has an active player
- **THEN** the current track MUST stop, the guild queue MUST become empty, and the bot MUST leave the voice channel
- **AND** the bot MUST reply confirming that playback stopped and the queue was cleared

#### Scenario: No active player

- **WHEN** a member runs `/stop` and the guild has no active player
- **THEN** the bot MUST reply with an ephemeral error that no player was found

### Requirement: Skip to the next queued track

The `/skip` command SHALL replace the current track with the next track in the guild queue.

#### Scenario: Queue has a next track

- **WHEN** a member runs `/skip` while the guild queue holds at least one track
- **THEN** that track MUST be removed from the front of the queue and start playing
- **AND** the bot MUST reply confirming the skip

#### Scenario: Queue is empty

- **WHEN** a member runs `/skip` while the guild queue is empty
- **THEN** the bot MUST reply with an ephemeral error that there are no further tracks
- **AND** the current track MUST keep playing

#### Scenario: No active player

- **WHEN** a member runs `/skip` and the guild has no active player
- **THEN** the bot MUST reply with an ephemeral error that no player was found

### Requirement: Show the current track

The `/now-playing` command SHALL report the track currently playing, including its elapsed and total duration.

#### Scenario: A track is playing

- **WHEN** a member runs `/now-playing` while a track is playing
- **THEN** the bot MUST reply with the track title linked to its source, the author, artwork when available, and the elapsed position alongside the total length

#### Scenario: Player exists but is idle

- **WHEN** a member runs `/now-playing` while the guild has a player with no current track
- **THEN** the bot MUST reply with an ephemeral error that nothing is playing

#### Scenario: No active player

- **WHEN** a member runs `/now-playing` and the guild has no active player
- **THEN** the bot MUST reply with an ephemeral error that no player was found

### Requirement: Show the queue

The `/queue` command SHALL list the tracks waiting to play, in play order, with a total count. The reply MUST stay within the platform's message size limits regardless of how many tracks are queued.

#### Scenario: Queue holds tracks

- **WHEN** a member runs `/queue` while the guild queue holds tracks
- **THEN** the bot MUST reply with a numbered list of the queued tracks, each with its title linked to its source and its duration
- **AND** the reply MUST state the total number of queued tracks

#### Scenario: Queue is empty

- **WHEN** a member runs `/queue` while the guild queue is empty
- **THEN** the bot MUST reply that the queue is empty
- **AND** the reply MUST carry exactly one status icon

#### Scenario: Queue is long enough to breach the reply size limit

- **WHEN** a member runs `/queue` while the queue holds more tracks than can be listed within the platform's embed description limit
- **THEN** the bot MUST list a bounded number of tracks and indicate how many further tracks are not shown
- **AND** the reply MUST still state the total number of queued tracks
- **AND** the reply MUST be delivered successfully rather than failing the interaction

### Requirement: Durations are rendered in a readable form

Track positions and lengths SHALL be rendered as `m:ss`, and as `h:mm:ss` once the duration reaches one hour. A zero duration MUST render as `0:00`.

#### Scenario: Sub-minute duration

- **WHEN** a duration of 45 seconds is rendered
- **THEN** the result MUST be `0:45`

#### Scenario: Duration under an hour

- **WHEN** a duration of 3 minutes and 7 seconds is rendered
- **THEN** the result MUST be `3:07`

#### Scenario: Duration of an hour or more

- **WHEN** a duration of 1 hour, 5 minutes, and 30 seconds is rendered
- **THEN** the result MUST be `1:05:30` and MUST NOT be `65:30`

#### Scenario: Zero duration

- **WHEN** a duration of zero is rendered
- **THEN** the result MUST be `0:00`

### Requirement: The bot serves exactly one configured guild

The bot SHALL act only on the single guild named by its configuration. There is one queue for that guild. Interactions and voice events originating from any other guild MUST be ignored, because slash commands are scoped by registration but voice events are not.

#### Scenario: Command from the configured guild

- **WHEN** a member of the configured guild runs any of the bot's commands
- **THEN** the command MUST be handled normally

#### Scenario: Command from another guild

- **WHEN** a command interaction arrives from a guild other than the configured one
- **THEN** the bot MUST refuse it with an ephemeral error and MUST NOT touch the queue or any player

#### Scenario: Voice event from another guild

- **WHEN** a voice state or voice server event arrives for a guild other than the configured one
- **THEN** the bot MUST ignore it
- **AND** the queue and the configured guild's player MUST be unaffected

#### Scenario: Bot present in an unconfigured guild

- **WHEN** the bot is a member of a guild other than the configured one
- **THEN** it MUST NOT respond to activity there
- **AND** it MUST NOT leave that guild on its own initiative

### Requirement: Queue access is safe under concurrency

Queue reads and writes arrive from concurrent gateway and player event handlers, so all queue access MUST be safe for concurrent use.

#### Scenario: Concurrent queue access

- **WHEN** tracks are added, taken, and cleared concurrently
- **THEN** no data race MUST occur and the queue MUST remain internally consistent

#### Scenario: Reading the queue yields a stable list

- **WHEN** a caller reads the queue contents while another writer modifies it
- **THEN** the reader MUST observe a self-consistent list that a later write cannot retroactively alter

#### Scenario: Queue is discarded when the bot leaves

- **WHEN** the bot is disconnected from the voice channel
- **THEN** the queue MUST be discarded

### Requirement: Queue advances automatically when a track ends

When a track finishes for a reason that permits starting the next one, the bot SHALL play the next queued track. When the queue is exhausted the bot SHALL become idle rather than leaving at once, and SHALL leave only once the queue-empty idle countdown has elapsed.

#### Scenario: Next track is available

- **WHEN** a track ends normally and the guild queue holds another track
- **THEN** that track MUST be removed from the front of the queue and start playing

#### Scenario: Queue is exhausted

- **WHEN** a track ends normally and the guild queue is empty
- **THEN** the bot MUST start the queue-empty idle countdown
- **AND** the bot MUST remain in the voice channel until that countdown elapses

#### Scenario: Track ended in a way that forbids advancing

- **WHEN** a track ends for a reason that does not permit starting the next track, such as being replaced
- **THEN** the bot MUST NOT consume a track from the queue

### Requirement: Command failures are always reported to the caller

A command that fails SHALL leave the caller with a visible error rather than an unanswered interaction. Failures MUST also be logged with the command name for diagnosis.

#### Scenario: Handler returns an unexpected error

- **WHEN** a command handler fails with an unexpected error
- **THEN** the bot MUST answer the interaction with an ephemeral error message that tells the caller they can try again
- **AND** the failure MUST be logged at `error` level with the command name

#### Scenario: Handler fails for a recognised reason

- **WHEN** a command handler fails with a recognised, user-caused reason
- **THEN** the bot MUST answer the interaction with the German message for that reason
- **AND** the failure MUST be logged below `error` level, because it needs no operator action

#### Scenario: Unknown command received

- **WHEN** an interaction arrives for a command the bot does not recognise
- **THEN** the bot MUST log the unknown command name and MUST NOT crash

### Requirement: The bot leaves a voice channel it has been left alone in

The bot SHALL treat its voice channel as unattended when that channel holds no user other than the bot itself. While the channel is unattended the bot SHALL run a countdown of `IDLE_ALONE_SECONDS`, and on elapse SHALL leave. The countdown MUST be cancelled as soon as any other user is present in the channel again.

Occupancy MUST be recomputed from the bot's current view of the guild's voice states on every voice state change, so that a dropped or reordered event self-corrects on the next one.

#### Scenario: The last other user leaves the channel

- **WHEN** the bot is connected to a voice channel and the last user other than the bot leaves it
- **THEN** the bot MUST start the unattended-channel countdown
- **AND** playback MUST continue while the countdown runs

#### Scenario: Countdown elapses with the channel still unattended

- **WHEN** the unattended-channel countdown elapses and the channel still holds no user other than the bot
- **THEN** the bot MUST stop playback, discard the queue, and leave the voice channel

#### Scenario: A user rejoins before the countdown elapses

- **WHEN** any user joins the bot's voice channel while the unattended-channel countdown is running
- **THEN** the countdown MUST be cancelled
- **AND** the bot MUST remain in the channel and keep playing

#### Scenario: Repeated departures do not extend a running countdown

- **WHEN** users join and leave the bot's voice channel repeatedly while the unattended-channel countdown is running, leaving the channel unattended each time
- **THEN** the countdown MUST keep its original deadline rather than restarting
- **AND** the bot MUST leave at that original deadline if the channel is unattended when it arrives

#### Scenario: The bot is not connected to a voice channel

- **WHEN** a voice state change occurs in the guild while the bot is not connected to any voice channel
- **THEN** no unattended-channel countdown MUST be started

#### Scenario: Another bot is the only remaining occupant

- **WHEN** the only occupant of the bot's voice channel other than itself is another bot
- **THEN** that occupant MUST count as present and no unattended-channel countdown MUST be started

### Requirement: The bot leaves a voice channel once it has nothing left to play

The bot SHALL treat itself as having nothing to play when no track is currently loaded in its player and the guild queue is empty. In that state the bot SHALL run a countdown of `IDLE_EMPTY_QUEUE_SECONDS`, and on elapse SHALL leave. The countdown MUST be cancelled as soon as a track starts playing.

A paused player that still holds a current track is a deliberate state and MUST NOT count as having nothing to play.

#### Scenario: A track is queued during the countdown

- **WHEN** a member runs `/play` while the queue-empty countdown is running
- **THEN** the track MUST start playing in the channel the bot is already connected to
- **AND** the countdown MUST be cancelled
- **AND** the bot MUST NOT have to rejoin the voice channel

#### Scenario: Countdown elapses with nothing to play

- **WHEN** the queue-empty countdown elapses while no track is loaded and the queue is still empty
- **THEN** the bot MUST stop playback, discard the queue, and leave the voice channel

#### Scenario: Playback is paused rather than finished

- **WHEN** playback is paused and the player still holds a current track
- **THEN** no queue-empty countdown MUST be started, regardless of whether the queue is empty

#### Scenario: Explicit stop is unaffected

- **WHEN** a member runs `/stop`
- **THEN** the bot MUST leave the voice channel immediately as specified for that command, without waiting for any countdown

### Requirement: Each idle countdown is configured independently

Each idle countdown SHALL be configured by its own variable, and the two SHALL be settable in any combination. A value of `0` MUST mean leave as soon as the condition becomes true, and the value `off` MUST mean never leave for that reason.

#### Scenario: A countdown is configured to zero seconds

- **WHEN** a countdown's variable is set to `0` and its condition becomes true
- **THEN** the bot MUST leave without waiting
- **AND** the other countdown's configured wait MUST be unaffected

#### Scenario: A countdown is turned off

- **WHEN** a countdown's variable is set to `off` and its condition becomes true
- **THEN** no countdown MUST be started and the bot MUST NOT leave for that reason
- **AND** the other countdown MUST still leave when its own condition persists for its configured wait

#### Scenario: Both countdowns are turned off

- **WHEN** both variables are set to `off`
- **THEN** the bot MUST never leave a voice channel on its own initiative

### Requirement: Leaving because the bot is idle happens at most once

The two idle conditions SHALL be evaluated independently and MAY both be true at once. Leaving MUST be a single idempotent action: whichever countdown elapses first performs it, and MUST cancel the other. A countdown that elapses after the bot has already left, or while the bot is shutting down, MUST do nothing.

#### Scenario: Both countdowns are running at once

- **WHEN** the last track ends and every other user leaves the channel, so both countdowns are running
- **THEN** the first countdown to elapse MUST perform the leave
- **AND** the other countdown MUST be cancelled so the bot does not attempt to leave twice

#### Scenario: A countdown elapses after the bot has already left

- **WHEN** a countdown elapses while the bot is no longer connected to a voice channel
- **THEN** it MUST do nothing

#### Scenario: The bot is disconnected by someone else

- **WHEN** the bot is disconnected from its voice channel by something other than its own countdown
- **THEN** both countdowns MUST be cancelled
- **AND** the queue MUST be discarded
