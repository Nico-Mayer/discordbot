Discord Music Bot

Personal Discord music bot written in Go using disgo, disgolink, and Lavalink.

Setup

Requires Go 1.27+ and a running Lavalink node.

Copy .env.example to .env and configure:

TOKEN=
GUILD_ID=

NODE_NAME=main
NODE_SECURE=false

LAVALINK_HOST=localhost
LAVALINK_PORT=2333
LAVALINK_PASSWORD=youshallnotpass

IDLE_ALONE_SECONDS=60
IDLE_EMPTY_QUEUE_SECONDS=60
Run
go run .

Reset slash commands:

go run . -reset-commands
Development
mise run check

Other tasks are defined in mise.toml.

Release
mise run next-version
mise run release

Releases are versioned with svu and built/published through GoReleaser and GitHub Actions.

Preview locally:

mise run release-snapshot
Layout
main.go
internal/app
internal/config
internal/queue
internal/music
