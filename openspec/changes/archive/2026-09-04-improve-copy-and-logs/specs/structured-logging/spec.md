## Purpose

Defines how the bot reports what it is doing to the operator running it: the language and shape of log messages, which level each kind of event uses, and which attribute keys carry the context needed to diagnose a failure.

## ADDED Requirements

### Requirement: Logs are English and structured

Every log record SHALL be written through the process-wide structured logger. Its message and attribute keys SHALL be English. No output MUST be written directly to standard output or standard error in place of a log record.

#### Scenario: A component reports an event

- **WHEN** any part of the bot reports an event
- **THEN** it MUST do so through the structured logger

#### Scenario: The process fails to start

- **WHEN** startup fails, including a configuration failure
- **THEN** the failure MUST be reported as a log record rather than printed directly to standard error
- **AND** the process MUST still exit with a non-zero status

### Requirement: A log message is a static event name

A log message SHALL be a short lowercase English phrase naming what happened. It MUST NOT end with a full stop, MUST NOT be capitalised, and MUST NOT have any variable value concatenated or formatted into it. Every varying value belongs in an attribute, so that all occurrences of one event share an identical message and can be grouped.

#### Scenario: A failure with a known cause is logged

- **WHEN** a failure is logged whose cause is one of a known set, such as a terminal websocket close code
- **THEN** the message MUST be the same static string for every cause
- **AND** the cause MUST be carried in its own attribute

#### Scenario: A track event is logged

- **WHEN** a track starts, ends, or fails
- **THEN** the track title MUST be an attribute and MUST NOT be part of the message

### Requirement: Attribute keys are standard and stable

Log attributes SHALL use lower snake_case keys, and the same concept SHALL always use the same key. The following keys are normative:

| Concept | Key |
| --- | --- |
| the error being reported | `error` |
| the Discord guild | `guild_id` |
| the Discord user | `user_id` |
| the Discord voice channel | `channel_id` |
| the slash command name | `command` |
| a track's title | `track_title` |
| why an action was taken | `reason` |
| what made a failure terminal | `cause` |
| a named shutdown step | `step` |

#### Scenario: An error is attached to a record

- **WHEN** any log record reports an error value
- **THEN** it MUST be attached under the key `error`

#### Scenario: A guild is attached to a record

- **WHEN** any log record identifies a guild
- **THEN** it MUST use the key `guild_id` and MUST NOT use `guild`

### Requirement: Levels reflect operator action

Log levels SHALL be chosen by what the operator has to do about the record.

- `error`: the bot could not do its job and an operator needs to look.
- `warn`: something degraded but the bot carried on, including a failure that will be retried.
- `info`: a lifecycle milestone or a user-visible action.
- `debug`: routine detail, including events filtered out because they belong to another guild.

A command that fails for an expected, user-caused reason MUST NOT be logged at `error`, because it needs no operator action.

#### Scenario: A user runs a command with nothing playing

- **WHEN** a command fails with a recognised user-facing reason
- **THEN** the failure MUST be logged below `error` level

#### Scenario: A command fails unexpectedly

- **WHEN** a command fails with an error the bot does not recognise
- **THEN** the failure MUST be logged at `error` level with the command name attached

#### Scenario: A connection drops and will be retried

- **WHEN** the connection to the audio node closes with a code that permits reconnecting
- **THEN** the event MUST be logged at `warn` level

#### Scenario: A connection drops for good

- **WHEN** the connection to the audio node closes with a code that no reconnect can fix
- **THEN** the event MUST be logged at `error` level

#### Scenario: An event from another guild arrives

- **WHEN** a voice or track event arrives for a guild the bot does not serve
- **THEN** it MUST be logged at `debug` level

### Requirement: Logs carry the context needed to diagnose

Every record reporting a command SHALL carry the command name. Every record reporting a failure SHALL carry the underlying error. Every record reporting a guild-scoped event SHALL carry the guild.

A log record MUST NOT contain the bot token or the audio node password, in the message or in any attribute.

#### Scenario: A command is handled

- **WHEN** a slash command is handled
- **THEN** the record MUST carry the command name, the guild, and the invoking user

#### Scenario: The bot leaves because it is idle

- **WHEN** the bot leaves a voice channel because an idle countdown elapsed
- **THEN** the record MUST carry which idle condition caused it

#### Scenario: A secret is in scope

- **WHEN** any record is written while the configuration is in scope
- **THEN** it MUST NOT contain the token or the node password
