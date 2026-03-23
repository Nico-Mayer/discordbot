# Discord Music Bot

This is a minimal Discord music bot written in Go using [disgo](https://github.com/disgoorg/disgo) and [Lavalink](https://github.com/lavalink-devs/Lavalink) via [disgolink](https://github.com/disgoorg/disgolink). (For private use only)

### Prerequisites

- Go 1.26+
- A running [Lavalink](https://github.com/lavalink-devs/Lavalink) node

### Configuration

Copy `.env.example` to `.env` and fill in the values:

```env
TOKEN=your-bot-token
GUILD_ID=your-guild-id

NODE_NAME=main
NODE_ADDRESS=localhost:2333
NODE_PASSWORD=youshallnotpass
NODE_SECURE=false
```

### Run

```bash
go run .
```

Reset slash commands before starting:

```bash
go run . -reset-commands
```
