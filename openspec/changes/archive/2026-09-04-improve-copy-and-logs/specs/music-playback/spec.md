## MODIFIED Requirements

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
