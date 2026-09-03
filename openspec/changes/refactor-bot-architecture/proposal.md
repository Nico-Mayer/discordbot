## Why

The bot works but the code has correctness and maintainability problems that will bite as it grows: `QueueManager` and `Queue` are shared across gateway goroutines with no synchronisation (a real data race), config failures only log a warning so the bot starts with an empty token, `context.TODO()` is used for every API call so nothing can be cancelled or time out, command errors are logged instead of answered so the user sees a dead interaction, and there are zero tests. Doing this now is cheap: the codebase is ~670 lines, so the full cleanup is a small, contained job rather than a rewrite later.

## What Changes

**Correctness**

- **BREAKING**: commit to being a single-guild bot. `QueueManager` and its `map[snowflake.ID]*Queue` are deleted in favour of one mutex-guarded `Queue`, and `GUILD_ID` becomes an authorization boundary: commands and voice events from any other guild are ignored. This dissolves the insert-on-read bug rather than fixing it, since that bug lived in the map.
- Replace every `context.TODO()` with a context derived from the interaction or a per-operation timeout.
- Fail fast on missing or invalid required configuration instead of warning and continuing. **BREAKING**: the bot now exits with a non-zero status on bad config instead of starting in a broken state.
- Reply to the user when a command handler fails, rather than only writing to the log.
- Fix the `/play` confirmation embed: the duration is currently rendered as a second field *name* with no value, so it shows as two broken fields.
- Fix `formatPosition` so tracks over an hour render as `1:05:30` instead of `65:30`.
- Guard the empty-slice indexing in the `/play` search and playlist result handlers.
- Shut down cleanly on SIGINT/SIGTERM: stop the gateway and Lavalink player connections before exiting, and drop the `log.Fatal` calls that currently skip every `defer`.
- Enable Lavalink session resuming, so a node restart or websocket blip no longer destroys the player and silently stops playback while the queue keeps listing tracks it cannot play. disgolink already reconnects on its own; resuming is the opt-in piece that makes the reconnect preserve state.
- Log Lavalink websocket closes with code, reason, and origin, distinguishing a retryable close from a terminal one such as failed authentication.
- Bound the `/queue` reply. It currently concatenates one line per track with no limit, so roughly 40 to 60 queued tracks breach Discord's 4096-character embed description limit and the command fails outright. Cap the list and report how many are not shown.

**Architecture**

- Move the command dispatch from the hand-rolled `map[string]func(...)` in `bot.InitHandlers` onto disgo's own `handler.Mux` router, with middleware for logging, error replies, and deferring long commands.
- Remove the two-phase construction of `Bot` (fields assigned after `New()`, `InitHandlers()` called separately). Build it once with its dependencies and no exported mutable fields.
- Inject the logger instead of calling the package-level `charm.land/log/v2` functions from every file, so handlers are testable and log output is capturable.
- Split the single `bot` package into focused packages so the pure logic (queue, formatting, config, embed construction) is separable from the Discord and Lavalink wiring.
- Replace the config singleton (`sync.Once` + package-level `instance`) with a plain `Load() (Config, error)`.
- Introduce typed sentinel errors for the recurring handler failures (no player, no voice channel, empty queue) instead of comparing strings.

**Tests**

- Unit tests for the queue (including concurrent access under `-race`), duration formatting, URL detection, embed construction, and config loading/validation.
- Table-driven tests for command handlers against small interfaces standing in for the Lavalink client and player, so handler logic is covered without a live Discord or Lavalink connection.
- Fuzz targets for the two pure string functions that take arbitrary slash command input: duration formatting and URL detection.
- A `goleak` check that the graceful shutdown path leaves no goroutines behind, since "shutdown does not hang" is otherwise untestable.
- A `-race` test run and a coverage report wired into CI, with no coverage threshold gating the build.

**Tooling and docs**

- Add `mise` tasks for `fmt`, `vet`, and `test`, plus a `check` task running all three. No third-party linter: static analysis stays with the Go toolchain's own `gofmt` and `go vet`.
- Add a CI job that runs build, `gofmt`, `go vet`, and `go test -race`. The workflow currently only builds a Docker image.
- Harden the Dockerfile: `-trimpath`, stripped binary, non-root user, build cache mounts.
- Fix the README: it documents `NODE_ADDRESS`/`NODE_PASSWORD` and a `-reset-commands` flag that the code does not implement. Wire up the flag and correct the variable names to match `.env.example`.

**Non-goals**

- No new bot features (no volume, seek, shuffle, loop, autocomplete, or the commented-out `source` option for `/play`).
- No playlist support. `/play` with a playlist URL keeps using only the first track. Planned as the separate `support-playlists` change, which also carries `/queue` pagination.
- No idle handling. Leaving when nobody is listening is the separate `handle-idle-voice` change.
- No state reconciliation when a Lavalink resume *fails*. There is no obvious hook to observe that, so it needs a spike first; recorded as a known unknown in the design.
- No change to the user-facing German copy, embed colours, or emoji, apart from the two rendering bugs above.
- No persistence layer, database, or metrics.
- No `golangci-lint` or other third-party linter, and no `Makefile`. Task running is `mise`, static analysis is `gofmt` and `go vet`.

## Capabilities

### New Capabilities

`openspec/specs/` is currently empty, so the behaviour this change fixes and hardens is being specified for the first time.

- `music-playback`: the slash command surface (`/play`, `/pause`, `/stop`, `/skip`, `/now-playing`, `/queue`), single-guild queue semantics and the guild guard, voice channel requirements, track advance on end, bounded replies, and how results and failures are reported back to the user.
- `bot-lifecycle`: configuration loading and validation, slash command registration and reset, gateway and Lavalink node startup, session resuming across node reconnects, and graceful shutdown.

### Modified Capabilities

None. There are no existing spec files to modify.

## Impact

- **Code**: every existing Go file. `main.go`, `config/config.go`, and all of `bot/` are restructured. The package layout changes, so import paths within the module change.
- **Public API**: none. This is a `package main` binary with no importers.
- **Configuration**: required environment variables are now enforced. A deployment that was silently running with a missing variable will now fail to start, which is the point, but it means the deploy environment must be checked before rollout. `GUILD_ID` gains weight: it now decides which guild the bot will act on at all, not merely where commands appear.
- **Dependencies**: `github.com/stretchr/testify` is already in `go.sum` as an indirect dependency and moves to a direct test dependency. `go.uber.org/goleak` is added as a new test-only dependency to check the shutdown path for leaked goroutines. No new runtime dependencies; neither enters the binary.
- **CI/CD**: `.github/workflows/` gains a test job. The Dockerfile changes, so image layers and the runtime user change.
- **Docs**: README corrected.
