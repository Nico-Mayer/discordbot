## Purpose

Covers how the bot process starts, validates its configuration, registers its slash commands, connects to Discord and Lavalink, and shuts down without leaving connections or voice sessions dangling.

## Requirements

### Requirement: Configuration is loaded from the environment

The bot SHALL read its configuration from environment variables, and SHALL load a `.env` file first when one is present so local development can use it. A missing `.env` file MUST NOT be an error.

The required variables are the Discord bot token, the guild ID, the Lavalink node name, the Lavalink host, the Lavalink port, and the Lavalink password. Whether the node connection is secure is optional and MUST default to insecure.

The two idle countdowns are optional. `IDLE_ALONE_SECONDS` and `IDLE_EMPTY_QUEUE_SECONDS` each accept a non-negative whole number of seconds or the literal `off`, and each MUST default to 60 seconds when unset or empty.

#### Scenario: All variables present

- **WHEN** the process starts with every required variable set
- **THEN** configuration loading MUST succeed

#### Scenario: No .env file present

- **WHEN** the process starts with no `.env` file but every required variable set in the environment
- **THEN** configuration loading MUST succeed

#### Scenario: Values in the environment win

- **WHEN** a variable is set both in `.env` and in the process environment
- **THEN** the process environment value MUST be used

#### Scenario: Optional secure flag omitted

- **WHEN** the secure-node variable is not set
- **THEN** the node MUST be treated as insecure

#### Scenario: Idle variables omitted

- **WHEN** neither idle variable is set
- **THEN** both countdowns MUST be configured to 60 seconds

#### Scenario: Idle variable set to a number of seconds

- **WHEN** an idle variable is set to `0` or to a positive whole number
- **THEN** that countdown MUST be configured to wait that many seconds

#### Scenario: Idle variable turned off

- **WHEN** an idle variable is set to `off`
- **THEN** that countdown MUST be configured never to leave for its reason

#### Scenario: Idle variables are independent

- **WHEN** one idle variable is set and the other is not
- **THEN** the set one MUST take its configured value and the unset one MUST take the 60 second default

### Requirement: Invalid configuration stops startup

Configuration errors SHALL be reported together and MUST prevent the bot from starting. The process MUST exit with a non-zero status. A value that cannot be interpreted MUST be an error rather than silently falling back to a default.

#### Scenario: Required variable missing

- **WHEN** a required variable is empty or unset
- **THEN** startup MUST fail with an error naming the missing variable
- **AND** the process MUST exit with a non-zero status without connecting to Discord

#### Scenario: Several variables missing

- **WHEN** more than one required variable is missing
- **THEN** the reported error MUST name every missing variable, not only the first

#### Scenario: Port is not a valid port number

- **WHEN** the Lavalink port is not an integer, or is outside the range 1 to 65535
- **THEN** startup MUST fail with an error naming the offending variable and its value

#### Scenario: Guild ID is not a valid ID

- **WHEN** the guild ID is not a valid Discord snowflake
- **THEN** startup MUST fail with an error naming the offending variable

#### Scenario: Idle variable is neither a non-negative number nor off

- **WHEN** an idle variable is set to a value that is not `off` and not a non-negative whole number, such as `-1`, `30s`, or `never`
- **THEN** startup MUST fail with an error naming the offending variable and its value
- **AND** the bot MUST NOT fall back to the default for that variable

#### Scenario: Secrets are not leaked in errors

- **WHEN** a configuration error is reported
- **THEN** the message MUST NOT include the value of the token or the Lavalink password

### Requirement: Slash commands are registered at startup

On startup the bot SHALL register its slash command set for the single configured guild, so that the commands available to users match the commands the running build implements. Commands MUST NOT be registered globally.

#### Scenario: Commands are synchronised

- **WHEN** the bot starts
- **THEN** the guild's slash commands MUST be brought in line with the command set of the running build

#### Scenario: Registration fails

- **WHEN** command registration fails
- **THEN** the failure MUST be logged
- **AND** the bot MUST continue starting, because previously registered commands may still work

### Requirement: Commands can be reset from the command line

The bot SHALL accept a documented command line flag that clears all previously registered guild and global commands before registering the current set. Without the flag, no clearing MUST take place.

#### Scenario: Reset flag supplied

- **WHEN** the bot is started with the reset flag
- **THEN** existing guild and global commands MUST be cleared before the current command set is registered

#### Scenario: Reset flag omitted

- **WHEN** the bot is started without the reset flag
- **THEN** no commands MUST be cleared

#### Scenario: Unknown flag supplied

- **WHEN** the bot is started with a flag it does not recognise
- **THEN** the process MUST exit with a non-zero status and a usage message

### Requirement: Startup connects to Discord and Lavalink

The bot SHALL open the Discord gateway and register at least one Lavalink node before reporting itself ready, and each connection attempt MUST be bounded by a timeout so a hung dependency cannot stall startup indefinitely.

#### Scenario: Successful startup

- **WHEN** the gateway opens and the Lavalink node is reachable
- **THEN** the bot MUST log that it is running, including the identity it connected as and the Lavalink node version

#### Scenario: Gateway cannot be opened

- **WHEN** the Discord gateway fails to open or does not open within the startup timeout
- **THEN** startup MUST fail and the process MUST exit with a non-zero status
- **AND** any connection already opened MUST be closed before exit

#### Scenario: Lavalink node is unreachable

- **WHEN** the Lavalink node cannot be added or does not respond within the startup timeout
- **THEN** startup MUST fail and the process MUST exit with a non-zero status
- **AND** the Discord gateway connection MUST be closed before exit

#### Scenario: Bot identity cannot be fetched

- **WHEN** the bot's own user cannot be fetched
- **THEN** a warning MUST be logged and startup MUST continue, because the identity is informational only

### Requirement: The Lavalink session survives a reconnect

The bot SHALL enable session resuming on its Lavalink node once the node is ready, so that a dropped node connection does not destroy the active player and silently stop playback. Reconnection itself is handled by the Lavalink client library and MUST NOT be reimplemented.

#### Scenario: Resuming is enabled at startup

- **WHEN** the Lavalink node becomes ready
- **THEN** the bot MUST enable session resuming with a timeout long enough to outlive a node restart

#### Scenario: Node connection drops and is resumed

- **WHEN** the node connection drops and is re-established within the resume timeout
- **THEN** the active player MUST survive and playback MUST continue
- **AND** the bot MUST NOT be required to re-join the voice channel

#### Scenario: Enabling resuming fails

- **WHEN** the call to enable resuming fails
- **THEN** the failure MUST be logged as a warning and startup MUST continue, because the bot is still usable without it

#### Scenario: Node connection closes

- **WHEN** the node websocket closes
- **THEN** the bot MUST log the close code, the reason, and whether the remote initiated it
- **AND** a close code indicating a terminal condition such as failed authentication MUST be logged distinguishably from one that will be retried

### Requirement: Shutdown is graceful

On receiving an interrupt or termination signal the bot SHALL shut down in an orderly way: stop accepting new work, cancel any pending idle countdowns, leave any voice channels it is connected to, close the Lavalink and gateway connections, and exit with a zero status.

#### Scenario: Termination signal received

- **WHEN** the process receives SIGINT or SIGTERM
- **THEN** the bot MUST leave the voice channels it is connected to, close its Lavalink and gateway connections, and exit with status zero

#### Scenario: Idle countdown pending at shutdown

- **WHEN** shutdown begins while either idle countdown is running
- **THEN** that countdown MUST be cancelled
- **AND** no work started by it MUST outlive the process
- **AND** it MUST NOT act on a connection that shutdown has already closed

#### Scenario: Shutdown does not hang

- **WHEN** shutdown steps do not complete within the shutdown timeout
- **THEN** the process MUST exit anyway rather than hanging indefinitely

#### Scenario: Second signal during shutdown

- **WHEN** a second termination signal arrives while shutdown is in progress
- **THEN** the process MUST exit immediately

#### Scenario: Startup failure still releases resources

- **WHEN** startup fails after some connections have already been established
- **THEN** those connections MUST be closed before the process exits
