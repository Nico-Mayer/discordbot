## MODIFIED Requirements

### Requirement: Show the current track

The `/now-playing` command SHALL report the track currently playing, including how far into the track playback has reached and how long the track is.

#### Scenario: A track is playing

- **WHEN** a member runs `/now-playing` while a track is playing
- **THEN** the bot MUST reply with the track title linked to its source, the artist, artwork when available, and a progress line combining the elapsed position, a proportional bar, and the total length

#### Scenario: A livestream is playing

- **WHEN** a member runs `/now-playing` while the current track is a livestream
- **THEN** the bot MUST reply with the track title, the artist, and a live marker in place of a duration
- **AND** the reply MUST NOT contain a progress line

#### Scenario: Player exists but is idle

- **WHEN** a member runs `/now-playing` while the guild has a player with no current track
- **THEN** the bot MUST reply with an ephemeral error that nothing is playing

#### Scenario: No active player

- **WHEN** a member runs `/now-playing` and the guild has no active player
- **THEN** the bot MUST reply with an ephemeral error that no player was found

### Requirement: Show the queue

The `/queue` command SHALL list the tracks waiting to play, in play order, above them name the track playing now, and state the total number of waiting tracks together with how long they will take to play. The reply MUST stay within the platform's message size limits regardless of how many tracks are queued.

#### Scenario: Queue holds tracks

- **WHEN** a member runs `/queue` while the guild queue holds tracks
- **THEN** the bot MUST reply with a numbered list of the waiting tracks, each with its title linked to its source and its duration
- **AND** the reply MUST name the track playing now, distinguished from the waiting tracks
- **AND** the reply MUST state the total number of waiting tracks and their total duration

#### Scenario: Queue holds tracks but nothing is playing

- **WHEN** a member runs `/queue` while tracks are queued and no track is currently playing
- **THEN** the bot MUST list the waiting tracks and MUST omit the currently playing line rather than showing it empty

#### Scenario: Queue is empty

- **WHEN** a member runs `/queue` while the guild queue is empty
- **THEN** the bot MUST reply that the queue is empty
- **AND** the text MUST state this without relying on an icon

#### Scenario: Queue is long enough to breach the reply size limit

- **WHEN** a member runs `/queue` while the queue holds more tracks than can be listed within the platform's embed description limit
- **THEN** the bot MUST list a bounded number of tracks and indicate how many further tracks are not shown
- **AND** the reply MUST still state the total number of waiting tracks and their total duration
- **AND** the reply MUST be delivered successfully rather than failing the interaction
