## Purpose

Defines the language, voice, tone, and wording rules for every piece of text a Discord user reads, so that replies from the bot are consistently German, consistently phrased, and always leave the reader with something they can do next.

## ADDED Requirements

### Requirement: User-facing text is German

Every string a Discord user can read SHALL be German. This covers slash command descriptions, command option names and descriptions, embed titles, embed descriptions, embed field names, footers, and error replies.

Slash command names are exempt and MUST stay English (`play`, `pause`, `stop`, `skip`, `now-playing`, `queue`). They are the verbs a user types to invoke the bot, English is the established convention for them on Discord, and they are invoked rather than read.

Option names are NOT exempt. Discord renders an option name as the label beside the value the user is typing, so it is read like any other piece of copy.

Text that only an operator reads - log messages, Go error strings, code comments, command line flag help - MUST stay English and MUST NOT be translated.

#### Scenario: A command is browsed in Discord

- **WHEN** a user opens Discord's slash command picker for the bot
- **THEN** every command description, every option name, and every option description MUST be German
- **AND** the command names themselves MUST be unchanged English identifiers

#### Scenario: A reply is sent

- **WHEN** the bot answers any interaction, successfully or not
- **THEN** every word of the reply that is not a track title, an author name, or a duration MUST be German

#### Scenario: An operator reads a log line

- **WHEN** any log record is written
- **THEN** its message and attribute keys MUST be English

### Requirement: One term per concept

User-facing text SHALL use a single German term for each domain concept, listed below, and MUST NOT alternate between synonyms or leave an English term untranslated.

| Concept | Term |
| --- | --- |
| track / song | Titel |
| queue | Warteschlange |
| voice channel | Sprachkanal |
| playback | Wiedergabe |
| player | *(not shown to users; describe the state instead)* |
| artist | Interpret |
| duration | Dauer |
| position | Position |

#### Scenario: A track is referred to

- **WHEN** any reply refers to a single piece of audio
- **THEN** it MUST call it "Titel" and MUST NOT call it "Song", "Track", "Lied", or "Musikstück"

#### Scenario: The absence of a player is reported

- **WHEN** a command fails because there is no active player
- **THEN** the reply MUST describe the situation in user terms, such as that nothing is currently playing
- **AND** the reply MUST NOT contain the word "Player"

### Requirement: Labels use the reader's vocabulary

A label the reader sees SHALL name what they are providing or looking at in their own terms. It MUST NOT name the value's data type, its internal role, or the field the system stores it in.

#### Scenario: The `/play` value is labelled

- **WHEN** a user opens `/play` in Discord
- **THEN** the option label MUST describe what the user is supplying in their terms
- **AND** it MUST NOT be a technical name for the value, such as "identifier", "input", "query", "arg", or "value"

#### Scenario: A label is ambiguous on its own

- **WHEN** one input accepts more than one kind of value, such as both a link and a search phrase
- **THEN** the label MUST stay short
- **AND** the accompanying description MUST name every kind of value accepted

### Requirement: Voice and tone

User-facing text SHALL address the reader informally as *du* and never as *Sie*. It SHALL be plain, calm, and free of exclamation marks, jokes, and blame.

Text SHALL be front-loaded, so the outcome is the first thing read. A reply line SHALL stay within roughly 80 characters, and a reply SHALL be at most two sentences.

Capitalisation follows German grammar: nouns are capitalised, and headings are otherwise sentence case rather than title case.

#### Scenario: The reader is addressed

- **WHEN** a reply addresses the reader directly
- **THEN** it MUST use *du* forms
- **AND** it MUST NOT use *Sie* forms

#### Scenario: A failure is reported

- **WHEN** a reply reports a failure
- **THEN** it MUST NOT end with an exclamation mark
- **AND** it MUST NOT imply the reader did something wrong

### Requirement: An error reply names a way forward

Every user-facing error SHALL state what did not happen and, where the reader can influence the outcome, what to do next. An error MUST NOT be a dead end that only reports a failure.

A reply MUST NOT contain a vague non-explanation as its entire content. When the cause is genuinely unknown to the bot, the reply MUST still tell the reader what they can do, such as trying again.

#### Scenario: The caller is not in a voice channel

- **WHEN** a user runs `/play` while not connected to a voice channel
- **THEN** the reply MUST tell them to join a voice channel first

#### Scenario: Nothing was found

- **WHEN** an identifier resolves to no tracks
- **THEN** the reply MUST name what was searched for
- **AND** MUST suggest checking the link or searching for something else

#### Scenario: Nothing is playing

- **WHEN** a command needs an active player and there is none
- **THEN** the reply MUST say that nothing is playing
- **AND** MUST point the reader at `/play`

#### Scenario: An unrecognised failure

- **WHEN** a command fails with an error the bot does not recognise
- **THEN** the reply MUST tell the reader the action did not work and that they can try again
- **AND** MUST NOT consist solely of a phrase equivalent to "something went wrong"

### Requirement: No technical detail leaks into a user reply

A user-facing reply SHALL NOT contain raw text produced by an upstream system, a Go error string, a stack trace, an error code, an internal identifier, or a secret. The technical detail belongs in the log record for the same failure.

#### Scenario: Loading a track fails at the node

- **WHEN** the audio node returns an error while loading an identifier
- **THEN** the reply MUST be a German sentence written by the bot
- **AND** MUST NOT include the upstream error text
- **AND** the upstream error text MUST appear in the log record for that failure

#### Scenario: A guild is not the configured one

- **WHEN** an interaction arrives from a guild the bot does not serve
- **THEN** the reply MUST NOT contain any guild ID or channel ID

### Requirement: A reply carries exactly one status icon

Every reply SHALL show at most one leading status icon, placed at the start of the embed title when the embed has one, and otherwise at the start of the description. A trailing icon MUST NOT be used, and two icons MUST NOT appear in the same reply.

An icon MUST NOT be the only carrier of meaning: the text alone MUST convey the outcome, so the reply still reads correctly to a screen reader or where the icon fails to render.

#### Scenario: The queue is empty

- **WHEN** a user runs `/queue` while the queue holds no tracks
- **THEN** the reply MUST show exactly one icon
- **AND** the text MUST state that the queue is empty without relying on the icon

#### Scenario: A skip is confirmed

- **WHEN** the bot confirms a skip
- **THEN** any icon MUST come before the text, not after it

#### Scenario: Icons are stripped

- **WHEN** every icon is removed from a reply
- **THEN** the remaining text MUST still state the outcome unambiguously

### Requirement: Quoted user input is bounded

A reply that quotes the user's own input back to them SHALL bound how much it quotes, so that the reply stays within the platform's message size limits no matter what was typed. A reply MUST NOT fail to send because the value it quotes is too long.

Truncation MUST be visible, so the reader can tell that what they see is shortened rather than what they typed.

#### Scenario: A very long search value finds nothing

- **WHEN** a member runs `/play` with a value long enough that quoting it whole would breach the embed description limit
- **THEN** the reply MUST be delivered successfully
- **AND** the quoted value MUST be shortened with a visible marker

#### Scenario: A short search value finds nothing

- **WHEN** a member runs `/play` with a value short enough to quote whole
- **THEN** the reply MUST quote it unchanged and MUST NOT add a truncation marker

### Requirement: German copy is defined in one place

All German user-facing strings SHALL be defined together in a single reviewable location rather than spread across the handlers, embeds, and error definitions that use them, so the wording can be read and revised as a whole.

#### Scenario: The wording is reviewed

- **WHEN** someone wants to read every user-facing string the bot can send
- **THEN** they MUST be able to do so from one place

#### Scenario: A string is reworded

- **WHEN** a user-facing string is reworded
- **THEN** no error matching, routing, or control flow MUST change as a result
