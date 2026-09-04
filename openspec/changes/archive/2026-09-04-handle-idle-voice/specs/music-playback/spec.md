## MODIFIED Requirements

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

## ADDED Requirements

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
