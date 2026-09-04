## MODIFIED Requirements

### Requirement: Play a track or search query

The `/play` command SHALL accept a `titel` string that is either a URL or a search query.

The value SHALL be normalised before it is classified: surrounding whitespace is removed, a single pair of wrapping angle brackets (`<` and `>`) is removed, and the URL scheme is recognised regardless of its case. Classification and loading SHALL both use the normalised value, so a value supplied as a link is loaded as a link.

A value that is empty after normalisation MUST be rejected without contacting the audio node. A value that is not a URL MUST be treated as a YouTube Music search. The bot MUST join the voice channel the requesting member is currently in.

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

#### Scenario: URL is surrounded by whitespace

- **WHEN** a member runs `/play` with a URL that has a leading or trailing space, tab, or newline
- **THEN** the bot MUST treat it as a URL
- **AND** the value loaded MUST have the surrounding whitespace removed

#### Scenario: URL is wrapped in angle brackets

- **WHEN** a member runs `/play` with a URL wrapped as `<https://example.com/track>`, as produced by copying a link out of one of the bot's own replies
- **THEN** the bot MUST treat it as a URL
- **AND** the value loaded MUST NOT include the angle brackets

#### Scenario: URL scheme is not lowercase

- **WHEN** a member runs `/play` with a value whose scheme is written `HTTPS://` or `Http://`
- **THEN** the bot MUST treat it as a URL rather than as a search query

#### Scenario: A search phrase that merely contains a URL

- **WHEN** a member runs `/play` with a value that contains a URL somewhere other than at its start, such as `listen to https://example.com/track now`
- **THEN** the bot MUST treat the whole value as a search query
- **AND** the bot MUST NOT extract the URL from within it

#### Scenario: Value is empty after normalisation

- **WHEN** a member runs `/play` with a value that is empty or contains only whitespace
- **THEN** the bot MUST reply with an error telling them to supply a link or a search term
- **AND** the bot MUST NOT contact the audio node, join a channel, or modify the queue

#### Scenario: Value exceeds the option's maximum length

- **WHEN** a member supplies a `/play` value longer than the option's configured maximum
- **THEN** Discord MUST reject it before the interaction reaches the bot

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

#### Scenario: Nothing found quotes what was searched for

- **WHEN** a normalised value resolves to no tracks
- **THEN** the error MUST quote the normalised value rather than the raw one, so it matches what the bot actually searched for

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
