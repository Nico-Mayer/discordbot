# Discord Music Bot

This is a minimal Discord music bot written in Go using [disgo](https://github.com/disgoorg/disgo) and [Lavalink](https://github.com/lavalink-devs/Lavalink) via [disgolink](https://github.com/disgoorg/disgolink). (For private use only)

It serves exactly one guild: the one named by `GUILD_ID`. Commands and voice events from any other guild are ignored.

### Prerequisites

- Go 1.27+ (pinned in `mise.toml`)
- A running [Lavalink](https://github.com/lavalink-devs/Lavalink) node

### Configuration

Copy `.env.example` to `.env` and fill in the values:

```env
TOKEN=your-bot-token
GUILD_ID=your-guild-id

NODE_NAME=main
NODE_SECURE=false

LAVALINK_HOST=localhost
LAVALINK_PORT=2333
LAVALINK_PASSWORD=youshallnotpass
```

Every variable except `NODE_SECURE` is required. `NODE_SECURE` defaults to `false`. The bot reports every configuration problem at once and exits with a non-zero status rather than starting in a broken state.

Values in the process environment take precedence over `.env`, and a missing `.env` file is not an error.

### Run

```bash
go run .
```

Reset slash commands before starting:

```bash
go run . -reset-commands
```

This clears all previously registered guild and global commands, then registers the current set.

### Development

Task definitions live in `mise.toml`:

```bash
mise run build      # go build ./...
mise run fmt        # fail if anything is not gofmt'd
mise run vet        # go vet ./...
mise run test       # go test -race ./...
mise run coverage   # report total statement coverage
mise run check      # all of the above, in order, stopping at the first failure
```

`mise run check` is what CI runs. There is no third-party linter: static analysis is `gofmt` and `go vet`.

### Layout

```
main.go              flags, logger, signal context, exit code
internal/app         composition root: build, run, shut down
internal/config      environment loading and validation
internal/queue       the guild's track queue, mutex-guarded
internal/music       playback logic, commands, events, embeds
```
