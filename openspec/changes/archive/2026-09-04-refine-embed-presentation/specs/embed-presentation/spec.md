## Purpose

Defines how every reply the bot sends is laid out and rendered: the shared shape of a track card, which colour carries which meaning, how much of a value any single field may hold, where an icon may appear, and how a duration, a livestream, and missing artwork are shown. Wording belongs to `interface-copy`; this capability governs everything a reader sees that is not the words themselves.

## ADDED Requirements

### Requirement: Every track is presented in one shared shape

A reply about a single track SHALL use the same embed shape regardless of which command produced it. That shape is:

- an author line naming the outcome (that the track started, was queued, or is playing now),
- the track title as the embed title, linked to the track's source when a URI is known,
- the artist below the title,
- the track's artwork as a thumbnail,
- the remaining facts as inline fields.

A track reply MUST NOT use a full-width image, and MUST NOT state its outcome only in the embed title. Two replies about the same track MUST differ only in their author line and their fields - never in their structure or their colour, which follows the track's source.

#### Scenario: A track starts playing immediately

- **WHEN** `/play` starts a track because nothing was playing
- **THEN** the reply MUST carry an author line stating the track is now playing, the track title as the linked embed title, the artist, and the artwork as a thumbnail
- **AND** the reply MUST NOT contain a full-width image

#### Scenario: A track is queued behind another

- **WHEN** `/play` appends a track to a non-empty queue
- **THEN** the reply MUST use the same structure and colour as the started-playing reply
- **AND** it MUST differ only in its author line and in carrying the queue position as a field

#### Scenario: The current track is requested

- **WHEN** `/now-playing` reports the track that is playing
- **THEN** the reply MUST use the same structure as the other two track replies
- **AND** the track title MUST be the embed title rather than a link inside the description

#### Scenario: A track has no source URI

- **WHEN** a track reply is built for a track whose source URI is unknown
- **THEN** the title MUST still show the track title as plain text
- **AND** the reply MUST be delivered rather than omitting the title

### Requirement: Colour signals the source on a card and the state on a status line

Colour SHALL carry meaning on two axes, and a reply takes its colour from exactly one of them:

- A **track card** SHALL be coloured by the service the track was resolved from, so replies about the same service read as a group and a reader can tell where a track came from before reading the footer. A source the bot has no colour for MUST fall back to one accent colour rather than being left uncoloured.
- A **status reply** SHALL be coloured by the state it reports: one colour for a failure, one for a state change that succeeded, one for playback being held (paused), and one neutral colour for a reply that reports nothing active or is merely informational.
- A **list reply** SHALL use the neutral colour, because it reports a container rather than one event or one source.

A colour MUST NOT be chosen per command. The failure colour MUST be reserved for failures: no source colour and no state colour may equal it, so the two axes cannot be confused.

#### Scenario: Two replies about the same source

- **WHEN** a track from a given service is started and another from the same service is queued
- **THEN** both replies MUST carry the same colour

#### Scenario: Two replies about different sources

- **WHEN** cards are built for tracks from two different services the bot has colours for
- **THEN** the two replies MUST NOT carry the same colour

#### Scenario: A source the bot has no colour for

- **WHEN** a card is built for a track whose source is unknown or unnamed
- **THEN** the reply MUST carry the fallback accent colour

#### Scenario: Playback is held rather than changed

- **WHEN** playback is paused and, separately, resumed
- **THEN** the two confirmations MUST NOT carry the same colour

#### Scenario: A failure is reported

- **WHEN** any user-facing failure is reported
- **THEN** the embed MUST carry the failure colour
- **AND** no reply that is not a failure MUST use that colour

### Requirement: No single field can exceed what the platform accepts

Every value written into an embed SHALL be bounded to the platform's limit for the field it goes into - title, author name, description, field name, field value, and footer - before the embed is sent. A bound MUST be measured in characters as the platform counts them, not in bytes, so a value containing non-ASCII characters is not cut short of its allowance or left over it.

A value that is shortened MUST carry a visible marker. **No reply may fail to send because a value derived from track metadata or member input was too long.**

#### Scenario: A track title exceeds the title limit

- **WHEN** a track whose title is longer than the platform's embed title limit is played
- **THEN** the reply MUST be delivered successfully
- **AND** the title MUST be shortened with a visible marker

#### Scenario: An artist name exceeds its field limit

- **WHEN** a track reply is built for a track with an unusually long artist name
- **THEN** the reply MUST be delivered successfully with that value shortened

#### Scenario: A value is shorter than its limit

- **WHEN** a value fits within the limit for its field
- **THEN** it MUST be written unchanged and MUST NOT gain a truncation marker

#### Scenario: A value is bounded at a multi-byte character

- **WHEN** a value that must be shortened has a multi-byte character at the cut point
- **THEN** the result MUST remain valid text and MUST NOT contain a broken character

### Requirement: An icon marks a state and never decorates

The bot SHALL keep a small, fixed inventory of icons, each naming one state the reader would otherwise have to read a sentence to tell apart - a failure, a neutral notice, playing, paused, stopped, skipped, and queued. Two states MUST NOT share an icon.

An icon MAY lead exactly one element of a reply, separated from the text by a single space:

- the text of a one-line status reply,
- the author line of a track card, where it distinguishes a track that started from one that was queued,
- the line naming the track playing now in a list reply.

A reply MUST NOT carry more than one icon and MUST NOT place an icon after the text. An icon MUST NOT appear in an embed title, a field name, a field value, or a footer, and the bot MUST NOT add one per command, per source, or per embed section.

Removing every icon from a reply MUST leave text that still states the outcome unambiguously.

#### Scenario: A pause is confirmed

- **WHEN** the bot confirms that playback was paused
- **THEN** the reply MUST carry exactly one icon, leading the text
- **AND** the text alone MUST state that playback is paused

#### Scenario: Pausing and resuming are told apart

- **WHEN** the pause confirmation and the resume confirmation are compared
- **THEN** they MUST NOT carry the same icon

#### Scenario: A track card is built

- **WHEN** any reply about a single track is built
- **THEN** it MUST carry exactly one icon, leading its author line
- **AND** its title, fields and footer MUST carry none

#### Scenario: A started track and a queued track are told apart

- **WHEN** the started-playing reply and the queued reply are compared
- **THEN** their author lines MUST NOT carry the same icon

#### Scenario: The queue is listed

- **WHEN** `/queue` lists queued tracks while a track is playing
- **THEN** the line naming the track playing now MUST carry one icon
- **AND** the title and every listed waiting track MUST carry none

#### Scenario: The queue is listed with nothing playing

- **WHEN** `/queue` lists queued tracks while no track is playing
- **THEN** the reply MUST carry no icon at all

#### Scenario: Icons are stripped

- **WHEN** every icon is removed from any reply the bot can send
- **THEN** the remaining text MUST still state the outcome unambiguously

### Requirement: A livestream is shown as live rather than as a duration

When the audio source reports a track as a livestream, every reply about that track SHALL show a live marker where a duration would otherwise go, and MUST NOT show a total length, a remaining time, or a progress line for it.

#### Scenario: A livestream is playing

- **WHEN** `/now-playing` reports a track the source marks as a livestream
- **THEN** the reply MUST show a live marker instead of a duration
- **AND** it MUST NOT render a progress line

#### Scenario: A livestream is queued

- **WHEN** a livestream is queued or listed in `/queue`
- **THEN** its duration MUST be shown as the live marker rather than as a time

#### Scenario: A livestream is part of a queue total

- **WHEN** a queue total duration is computed over a queue that contains a livestream
- **THEN** the livestream MUST NOT contribute a length to that total
- **AND** the total MUST be presented so the reader can tell it is not the whole picture

### Requirement: Elapsed position is shown as a progress line

A reply reporting the track currently playing SHALL render its elapsed position as a single line combining the elapsed time, a proportional bar of fixed width built from plain text characters, and the total length. The bar MUST NOT use emoji, and MUST NOT be the only carrier of the position: the elapsed and total times MUST both be readable as text.

The bar MUST render correctly at the boundaries: a position of zero, a position at or beyond the track length, and a track whose length is unknown or zero.

#### Scenario: A track is halfway through

- **WHEN** `/now-playing` reports a track whose position is about half its length
- **THEN** the reply MUST show a bar filled about halfway, preceded by the elapsed time and followed by the total length

#### Scenario: A track has just started

- **WHEN** the reported position is zero
- **THEN** the bar MUST render as empty and the elapsed time MUST read as zero

#### Scenario: A position exceeds the reported length

- **WHEN** the reported position is greater than the reported track length
- **THEN** the bar MUST render as full and MUST NOT overflow its fixed width

#### Scenario: A length of zero is reported

- **WHEN** a track reports a length of zero and is not a livestream
- **THEN** the reply MUST omit the bar rather than divide by zero or render a misleading full bar

### Requirement: Artwork and source are shown when known and omitted cleanly when not

A track reply SHALL use the artwork the audio source reports. When no artwork is reported, the reply MAY fall back to a derived thumbnail for sources where one can be derived, and MUST otherwise be sent without a thumbnail rather than with a broken image reference.

When the audio source names itself, a track reply SHALL name that source in its footer as plain text. An unknown source MUST leave the footer without a source rather than showing an empty or placeholder value.

#### Scenario: The source reports artwork

- **WHEN** a track reply is built for a track with an artwork URL
- **THEN** that URL MUST be used as the thumbnail

#### Scenario: The source reports no artwork

- **WHEN** a track reply is built for a track with no artwork URL and no thumbnail can be derived
- **THEN** the reply MUST be sent without a thumbnail
- **AND** the reply MUST NOT contain an empty or invalid image URL

#### Scenario: The source names itself

- **WHEN** a track reply is built for a track whose source name is known
- **THEN** the footer MUST name that source

#### Scenario: The source is unknown

- **WHEN** the source name is empty
- **THEN** the footer MUST omit it entirely
