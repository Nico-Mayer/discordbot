## MODIFIED Requirements

### Requirement: Invalid configuration stops startup

Configuration errors SHALL be reported together and MUST prevent the bot from starting. The process MUST exit with a non-zero status. A value that cannot be interpreted MUST be an error rather than silently falling back to a default.

The failure SHALL be reported as a structured log record, in the same format as every other record the bot writes, rather than printed directly to standard error.

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

#### Scenario: The failure is written as a log record

- **WHEN** configuration loading fails, or startup fails for any other reason
- **THEN** the failure MUST be emitted through the structured logger at `error` level
- **AND** it MUST NOT be printed to standard error outside the logger
