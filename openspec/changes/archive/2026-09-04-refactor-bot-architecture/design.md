## Context

See `proposal.md` - Why. What shapes the approach here:

- The whole program is ~670 lines of Go across `main.go`, `config/`, and one `bot/` package. Everything is reachable in one sitting, so a single coordinated restructure is cheaper than incremental patching.
- It is a `package main` binary with no importers, so there is no API compatibility to preserve. The only external contract is the slash command surface and the environment variables.
- The hard testability constraint is that `*events.ApplicationCommandInteractionCreate` and `*handler.CommandEvent` are concrete structs wrapping a concrete `*bot.Client`. They cannot be faked. Any design that puts logic inside a method taking one of those types is untestable, which is why there are no tests today.
- Conversely `disgolink.Client`, `disgolink.Player`, and `disgolink.Node` are already interfaces, so the Lavalink side has usable seams.
- Verified against the pinned dependencies in the module cache: `disgo v0.19.6` ships `handler.Mux` with `Use`, `Group`, `SlashCommand`, `NotFound`, and `Error(ErrorHandler)`, plus `handler/middleware` helpers; `handler.CommandEvent` carries a `Ctx context.Context`. `charm.land/log/v2` `*Logger` implements `slog.Handler` (`var _ slog.Handler = (*Logger)(nil)`).
- `mise.toml` already pins Go and openspec, so it is the natural home for task definitions. No third-party linter is in the repo and none is being added, at the user's request: static analysis is `gofmt` and `go vet`.

## Goals / Non-Goals

**Goals:**

- Make the command logic unit-testable without a live Discord or Lavalink connection, using hand-written fakes and no new runtime dependencies.
- Give the process one composition root and one lifetime context, so every network call is cancellable and shutdown is deterministic.
- Keep the German user-facing copy in one place so it can be reviewed and reused.
- Commit to being a single-guild bot, so the code stops carrying multi-guild machinery it never uses, and gain a guard where that assumption could be violated.

**Non-Goals:**

- No integration or end-to-end tests against a real Discord gateway or Lavalink node. The seams introduced here stop at the boundary of disgo's concrete event types.
- No abstraction over disgo itself. Wrapping `*bot.Client` behind an interface would mean maintaining a facade larger than the bot, for no benefit at this size.
- No coverage *threshold* gating CI. Coverage is measured and reported, because `golang-testing` makes that cheap, but chasing a percentage in a package that is mostly Discord wiring produces tests that assert on adapters.
- No `gosec` SAST and no Dependabot or Renovate, both now within reach via the newly added skills. Reasons under Go skills below.

## Decisions

### Split the logic from the Discord adapter

Handlers currently mix three concerns: reading interaction options, deciding what to do, and sending a reply. Because the event type is unfakeable, all three are untestable together.

The change splits them. A `Service` holds the queue, the Lavalink client behind narrow interfaces, and the one configured guild ID, and exposes intent-shaped methods that take plain values and return a value plus an error:

```go
func (s *Service) Pause(ctx context.Context) (Paused, error)
func (s *Service) NowPlaying() (lavalink.Track, lavalink.Duration, error)
func (s *Service) Enqueue(ctx context.Context, req PlayRequest) (PlayResult, error)
```

The guild ID is a field on the `Service`, not a parameter on every method. It is a constant for the lifetime of the process, so threading it through each call would only create the opportunity to pass the wrong one.

The handler layer becomes a thin adapter: pull options off the event, call the service, turn the return value into an embed, send it. The service is fully unit-testable; the adapter has nothing left worth testing.

*Alternative considered:* keep logic in the handlers and test through disgo's `handler` test helpers. `handler` has a `mux_test.go` and `testdata/`, but those exercise routing, not our behaviour, and they still need a client to send a response. Rejected.

*Alternative considered:* an interface per handler dependency, injected into methods that still take `*handler.CommandEvent`. This does not help, because the untestable part is the event, not the dependencies.

### Narrow interfaces for the Lavalink seam, defined by the consumer

`disgolink.Client` and `disgolink.Player` are interfaces but wide ones; implementing either by hand is a large fake that mostly returns zero values. Instead `internal/music` declares only what it uses:

```go
type Lavalink interface {
    Player(guildID snowflake.ID) Player
    ExistingPlayer(guildID snowflake.ID) Player
    BestNode() Node
}

type Player interface {
    Update(ctx context.Context, opts ...lavalink.PlayerUpdateOpt) error
    Track() *lavalink.Track
    Paused() bool
    Position() lavalink.Duration
}
```

Because `Player` returns our `Player` rather than `disgolink.Player`, a thin adapter in `internal/app` wraps the real client. That adapter is a handful of one-line methods and is the only untested code on this path.

There is one trap in it, flagged by `golang-safety`: `ExistingPlayer` returns nil when a guild has no player, and an adapter that wraps that nil in a struct pointer and returns it as our `Player` interface produces a **typed nil**, which compares `!= nil`. Every `if player == nil` check in the service would then silently stop working, and `/pause`, `/stop`, `/skip`, and `/now-playing` would panic instead of replying `ErrNoPlayer`. The adapter must check the concrete value and return an untyped nil interface. This is the single most likely way to break the refactor, so it gets its own task and its own test.

*Alternative considered:* using `disgolink.Client` directly and faking it. Rejected: the fake would be several hundred lines of stubs and would break on every disgolink release.

### Package layout

```
main.go                       flags, logger, os.Exit; nothing else
internal/app/app.go           composition root: build client, router, lavalink; run; shut down
internal/config/config.go     Load() (Config, error)
internal/queue/queue.go       Queue, mutex-guarded, one instance
internal/music/service.go     the decisions: play, pause, stop, skip, now playing, queue
internal/music/handlers.go    disgo adapters, option parsing, reply sending
internal/music/commands.go    slash command definitions
internal/music/events.go      gateway and player event handlers
internal/music/embed.go       embed builders and duration formatting
internal/music/errors.go      sentinel errors and their user-facing messages
```

`internal/` is used so nothing here can be imported by accident, which makes the package boundaries a statement of intent rather than a suggestion. Four packages, not nine: at this size a package per file costs more in indirection than it buys.

*Alternative considered:* `cmd/bot/main.go`. Rejected. The repo builds one binary, `go run .` and the Dockerfile's `go build .` both rely on the root package, and the README documents `go run .`. Adding a `cmd/` level would change all of that to buy nothing.

### Command dispatch moves to disgo's router

`bot.InitHandlers` builds a `map[string]func(...)` and `OnApplicationCommand` looks up and calls it. `handler.Mux` already does this, plus middleware, plus subcommand path matching, plus an error hook. Replacing the map with the router deletes the dispatch code and provides the two seams the specs need:

- `mux.Error(...)` becomes the single place where a handler error is turned into an ephemeral reply and a log line. This satisfies "command failures are always reported to the caller" once, instead of at every `return err`.
- `mux.NotFound(...)` covers the unknown-command scenario.
- `mux.Use(...)` carries a logging middleware, and `/play` gets `middleware.Defer` so the deferred acknowledgement is not hand-rolled in the handler.

### Errors are sentinels with a message table

The recurring failures are a small closed set. `internal/music/errors.go` defines them and maps them to the German copy:

```go
var (
    ErrNoPlayer       = errors.New("no active player")
    ErrNotInVoice     = errors.New("caller not in a voice channel")
    ErrNothingPlaying = errors.New("player has no current track")
    ErrQueueEmpty     = errors.New("queue is empty")
    ErrNoResults      = errors.New("no tracks found")
)
```

The error handler matches with `errors.Is` and falls back to a generic message plus an error-level log for anything unrecognised. Tests assert on the sentinel, not on a German string, so copy can be reworded without touching tests. Errors crossing a package boundary are wrapped with `%w`.

### Config returns an error and reports every problem at once

`config.Load()` becomes `Load() (Config, error)`, dropping the `sync.Once` singleton. A singleton that cannot fail is what let the current code start with an empty token.

Validation accumulates into `errors.Join` so a fresh deployment learns about all missing variables in one run rather than one per restart. Errors name the variable but never its value, so the token and password cannot reach the logs. `LavalinkHost` and `LavalinkPort` stop being stored as separate fields since only the joined address is used, and the current code calls `requireEnv("LAVALINK_HOST")` and `envInt("LAVALINK_PORT")` a second time to build it, producing duplicate warnings.

### The logger is `*slog.Logger` backed by charm log

Every file currently calls the package-level `charm.land/log/v2` functions, so log output cannot be captured in a test and there is no way to attach per-guild fields.

The logger is injected as `*slog.Logger`, constructed in `main.go` as `slog.New(log.New(os.Stderr))` since charm's `*Logger` implements `slog.Handler`. This keeps the pretty console output from commit `7a19e86` while typing every seam against the standard library, so tests can substitute a capturing handler with no dependency on charm.

*Alternative considered:* inject charm's `*log.Logger` directly. Rejected: it pins every package to charm for no gain over the interface the standard library already defines.

### Contexts

`main.go` creates the lifetime context with `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`. Startup steps use `context.WithTimeout` derived from it. Command handlers use `e.Ctx`, which `handler.CommandEvent` already provides. Every `context.TODO()` goes away.

Gateway event handlers (`OnVoiceStateUpdate`, `OnVoiceServerUpdate`, `OnTrackEnd`) are the one exception: disgo hands them no context, so they take theirs from a field on the handler struct. Storing a context in a struct is normally wrong, and this is a deliberate exception with a comment saying why: it is what makes those handlers stop working the moment shutdown starts.

### The guild map collapses: this is a single-guild bot

`GUILD_ID` is a single snowflake, commands register to that one guild, and the bot runs on one private server. Everything downstream was nonetheless keyed by guild, which is speculative generality:

```
BEFORE                              AFTER
QueueManager                        Queue          (one instance)
  map[snowflake.ID]*Queue             tracks []lavalink.Track
  mu sync.Mutex                       mu   sync.Mutex
  Get / GetOrCreate / Delete          Add / Next / Clear / Len / Tracks
```

`QueueManager`, its map, and its mutex are deleted. `internal/queue` becomes one mutex-guarded `Queue` whose zero value is usable. `Queue.tracks` is unexported with a `Tracks() []lavalink.Track` accessor returning a copy, because handing out the live slice under a mutex only pretends to be safe.

This also dissolves a bug rather than fixing it. The earlier plan split `Get` from `GetOrCreate` because `Get` inserted an empty queue on read, so `OnTrackEnd` and `/queue` grew the map for guilds with nothing queued. That was never a queue bug; it was a map bug, and there is no longer a map. The `Get`/`GetOrCreate` split is dropped.

*Alternative considered:* keep the map and assert a single guild. Rejected: it preserves the machinery whose only justification was multi-guild support, and leaves the same insert-on-read trap in place.

### `GUILD_ID` becomes an authorization boundary

Collapsing to one queue creates an obligation the map was accidentally discharging. Slash commands are registered per guild so they only appear in the configured server, but **voice events fire for every guild the bot is in**. A bot added to a second server would drive the single queue from there.

```
  slash commands  -> registered to GUILD_ID only  -> naturally scoped
  voice events    -> fire for ALL guilds          -> unguarded today
```

So `cfg.GuildID` stops being a registration detail and becomes a guard at three edges:

- router middleware, rejecting a command from any other guild with an ephemeral error
- `OnVoiceStateUpdate`, returning early
- `OnVoiceServerUpdate`, returning early

Foreign-guild events are ignored and logged at debug, not acted on. The bot does not leave the guild on its own; that is the operator's call, not a side effect of a refactor.

### Lavalink session resuming is enabled

disgolink already reconnects on its own: a websocket read error triggers `go n.reconnect()` with linear backoff (`try * 2s`, capped at 30s), retrying indefinitely. Reconnection is not the gap.

The gap is that a reconnect without resuming opens a **fresh** session, so Lavalink destroys every player. Playback stops, the bot stays sitting in the voice channel because the Discord voice state is independent, and our queue still lists tracks it can no longer play.

```
  websocket drops
      |
      +-> disgolink reconnects automatically
            |
            +-- resuming ON  --> Session-Id header --> session resumed
            |                     players survive, playback continues
            |
            +-- resuming OFF --> fresh session
                                  players destroyed, silence,
                                  bot still in voice, queue still full
```

Resuming is opt-in and currently off. One REST call once the node is ready turns it on:

```go
resuming, timeout := true, 60
node.Update(ctx, lavalink.SessionUpdate{Resuming: &resuming, Timeout: &timeout})
```

The 60 second timeout is how long Lavalink holds the session open after the socket drops. It is a constant, not a config variable: it needs to outlive a node restart and nothing else, and there is no reason for an operator to tune it.

`OnWebSocketClosed` also stops being a bare warning. It logs the close code, the reason, and whether the remote closed it, and distinguishes a retryable close from a terminal one such as 4004 (authentication failed), which no amount of backoff will fix.

**Known unknown, deliberately not designed here:** if a resume *fails*, our queue is stale and the bot is lying about what it can play. disgolink logs `"failed to resume session"` internally but there is no obvious hook to observe it, so the choice is a node-ready/resumed event that may not exist or polling player state. That needs a spike before it gets a mechanism. Recording it as an unknown rather than guessing.

### `/queue` output is bounded

`/queue` builds its description by unbounded string concatenation. Discord caps an embed description at 4096 characters and each line runs roughly 60 to 120 characters, so somewhere around 40 to 60 queued tracks makes the API call fail. That is reachable today by repeated `/play`, and once hit, `/queue` stays broken until the queue drains.

The list is capped at 20 entries followed by an "and N more" line, with a defensive length check so a pathological set of long titles cannot breach 4096 even under the cap. Full pagination is left to the playlist change, where it becomes genuinely necessary.

### Startup and shutdown are ordered and bounded

`main` currently calls `log.Fatal` in five places. Each one skips `defer client.Close(...)`, so a failure part-way through startup leaks whatever was already open. And after startup it blocks on a signal channel and then returns with nothing unwound.

`app.Run(ctx, cfg, logger) error` replaces this. It builds each dependency in order, registers a cleanup for each on success, and on any error unwinds the cleanups registered so far before returning. `main` prints the error and calls `os.Exit(1)`. Shutdown leaves voice channels, closes Lavalink, then closes the gateway, under a bounded shutdown context so a hung dependency cannot block the exit; a second signal aborts immediately because `signal.NotifyContext` restores the default disposition after the first.

### Go skills bound to task groups

The repo vendors 21 Go skills from `samber/cc-skills-golang` in `.claude/skills`, pinned in `skills-lock.json`. Leaving them to fire opportunistically means the ones that matter most for the trickiest groups get missed, so each is bound to the groups it applies to:

| Skill | Task groups | Why it applies here |
|---|---|---|
| `golang-how-to` | all | Orchestrator, always active. Also writes the force-trigger config, see task 1.5 |
| `golang-refactoring` | all | Behaviour-preserving transforms, breaking import cycles, small stacked PRs. Matches the migration plan directly |
| `golang-gopls` | all | Find references before each rename, diagnostics after each edit. The package moves need this |
| `golang-naming` | all | `ErrNoPlayer` vs `NoPlayerError`, `New` vs `NewService`, interface and receiver names |
| `golang-code-style` | all | Declarations, control flow, and specifically when a comment helps rather than hurts |
| `golang-project-layout` | 2-7 | Owns the `internal/` convention and the package split. See the layout tension noted below |
| `golang-testing` | 2-7 | Table-driven tests, `t.Parallel()`, fuzzing, `goleak`, coverage, test naming. Owns the whole test approach |
| `golang-safety` | 2, 4, 5, 7 | Nil derefs on `track.Info.URI`, `append` backing-array aliasing in the queue, defensive copies, usable zero values, and the typed-nil interface trap in the adapter |
| `golang-concurrency` | 2 | Mutex placement on the single queue, and the shared-slice protection the current code lacks |
| `golang-error-handling` | 3, 4, 5, 6, 7 | Sentinels, `%w`, `errors.Is`, `errors.Join` for config, `slog` for structured logging |
| `golang-context` | 5, 6, 7 | Propagation across layers, timeouts, `Background` vs `TODO`, and the context-in-a-struct exception in the event handlers |
| `golang-dependency-injection` | 5, 7 | Manual constructor injection and the testability argument behind the service split |
| `golang-design-patterns` | 5, 7 | Constructor APIs, graceful shutdown, resource lifecycle, and `init()`/global-state avoidance, which is what killed the config singleton |
| `golang-security` | 3, 8 | Secrets management and PII in logs: the token and password must not reach an error message or a log line |
| `golang-continuous-integration` | 1, 8 | Owns `.github/workflows/*.yml`, the check job, the Docker build/push job, and the quality gate wiring |
| `golang-modernize` | 1, 8 | Go 1.27 is pinned; catch anything the toolchain now does better, plus the tooling refresh |
| `golang-documentation` | 8 | README and doc comments |
| `golang-dependency-management` | 8 | `testify` indirect to direct, `go mod tidy`, and a `govulncheck` pass over the tree |
| `golang-pkg-go-dev` | 4, 5, 6 | `godig` for verifying disgo and disgolink signatures instead of guessing at them |

Two of the 21 are still deliberately not used:

- `golang-lint` is golangci-lint configuration and `.golangci.yml` authoring. That tool is out by decision above, so the skill has nothing to configure. Its `go vet` guidance is the only relevant part and does not need the skill loaded.
- `golang-performance` triggers when profiling or a benchmark has found a bottleneck. Nothing here has been profiled and no bottleneck is claimed, so loading it would invite speculative optimisation of a bot that plays one track at a time.

Three siblings referenced by the installed skills are still absent: `golang-benchmark`, `golang-troubleshooting`, and `golang-popular-libraries`. None of them is needed here. Nothing is benchmarked, nothing is being debugged after the fact, and no library choice is open.

Two scanners that the new skills bring within reach are also left out, and both are judgement calls worth stating rather than burying:

- `gosec` SAST, offered by `golang-security` and wireable by `golang-continuous-integration`. Left out to stay consistent with the decision to keep third-party static analysis out of this repo. `govulncheck` stays in at task 8.4 because it checks dependency advisories rather than linting our code.
- Dependabot or Renovate config, offered by `golang-continuous-integration`. Out of scope: dependency bumps are currently done by hand (`chore: bump deps`), and switching that to a bot is its own change with its own review burden.

### Layout tension to resolve, not rediscover

`golang-project-layout` owns `cmd/`, `internal/`, and `pkg/` conventions, and the conventional answer for a binary is `cmd/<name>/main.go`. This design keeps `main.go` at the repo root instead, for the reasons given under Package layout: `go run .`, the Dockerfile's `go build .`, and the README all depend on it, and the repo builds exactly one binary.

The apply phase should not silently flip this when the skill loads. The `internal/` half of the skill's guidance is adopted in full; the `cmd/` half is knowingly declined.

### Tests are table-driven with hand-written fakes

Standard library `testing` plus `testify/require`, which is already in `go.sum` and moves to a direct dependency. Fakes are written by hand against the narrow interfaces above; no mockgen.

- `internal/queue`: add, take, clear, usable zero value, and a concurrent test that is meaningful only under `-race`.
- `internal/music`: duration formatting including the hour boundary, URL detection, embed builders as pure functions returning `discord.Embed`, and one table-driven test per service method covering the success and each sentinel-error path.
- `internal/config`: required variables, joined errors for several missing at once, port range, invalid snowflake, `.env` precedence via `t.Setenv` and `t.TempDir`, and that no error message contains the token or password.

`golang-testing` adds four things the earlier plan did not have:

- **`t.Parallel()` on the table-driven subtests**, since they are all pure and there is no shared state to serialise. The config tests are the exception: `t.Setenv` forbids `t.Parallel` in the same test, so those stay sequential.
- **`goleak`** to assert the shutdown path leaves no goroutines behind. This is not decoration: "Shutdown is graceful" and "Shutdown does not hang" are spec requirements, and a leaked gateway or Lavalink goroutine is exactly how they fail in a way no other test catches. It costs one test-only dependency, `go.uber.org/goleak`. Nothing is added to the runtime dependency tree.
- **Fuzz targets** for the two pure string functions, duration formatting and URL detection. Both take arbitrary user input from a slash command option, both are cheap to fuzz, and the URL regex is loose enough to be worth probing.
- **Coverage reported in CI** without a threshold, per the non-goal above.

*Alternative considered:* `testify/mock` for the Lavalink seam instead of hand-written fakes. Rejected: the interfaces are four and three methods wide, so a hand-written fake is shorter than the mock setup, and it stays compile-checked rather than failing at runtime on a missing expectation.

### Tooling lives in mise, not a Makefile

The repo already pins Go and openspec in `mise.toml`, so task definitions go there rather than introducing a second mechanism alongside it. Tasks: `fmt` (`gofmt -l` failing on any output), `vet`, `test` (with `-race`), and `check` running all three. CI runs `check`; the existing Docker workflow is left alone apart from gaining a dependency on it.

No third-party linter is added. `golangci-lint` is explicitly out, at the user's request, so `gofmt` and `go vet` are the whole static analysis story.

This has one consequence worth naming: `errcheck` would have been the mechanical way to catch the ignored error returns this codebase currently has, and `errorlint` the way to catch `%v` where `%w` belongs. Both now have to be caught by review instead, so the tasks that fix them carry a grep-based check rather than a linter invocation. That is weaker, and it is a deliberate trade.

*Alternative considered:* `go vet` plus `staticcheck` alone as a middle ground. Not proposed, because the request was to keep third-party linting out rather than to swap one tool for another; it stays available as a later addition if the review burden proves annoying.

## Risks / Trade-offs

- **Fail-fast config breaks a deployment that is silently running with a missing variable** → This is the intended behaviour, but it turns a quiet degradation into a crash loop on next deploy. Mitigation: verify the deployment environment against `.env.example` before rolling out, and land the config change in its own commit so it can be reverted independently.
- **Every import path in the module moves** → Nothing outside the repo imports it, so the blast radius is this repo only. It does make the diff large and unreviewable line by line; mitigation is the commit sequence in the migration plan, so each commit is separately reviewable.
- **The service/adapter split adds a layer to a small program** → Real cost in indirection, accepted because it is the only thing that makes the logic testable at all given disgo's concrete event types.
- **Narrow consumer-defined interfaces need an adapter that drifts if disgolink changes its signatures** → Compile-time failure, not a silent one, and the adapter is small enough to fix in minutes.
- **Fail-fast on an unreachable Lavalink node means the bot will not start if Lavalink is slow to boot** → Matches the current behaviour (`log.Fatal`), so this change does not make it worse. Retry-with-backoff at startup is a reasonable follow-up but is out of scope; noted so it is not mistaken for an oversight.
- **`Tracks()` returning a copy allocates on every `/queue`** → Queues are a handful of tracks and the command is rare. Correctness first.
- **Adding `-race` to CI will surface the existing queue race if any test order hits it** → That is the point, and the queue fix lands before the tests that would catch it.
- **Without `errcheck`, newly ignored error returns will not be caught mechanically** → `go vet` catches some but not this class. Mitigation: the ignored returns in `reply.go` are fixed explicitly by task, and future ones are a review responsibility.
- **The adapter can return a typed nil and silently disable every `player == nil` check** → The most likely way to break this refactor, and it fails as a panic in production rather than a compile error. Mitigation: the adapter returns an untyped nil interface explicitly, task 7.2 covers it, and a test asserts `ExistingPlayer` on a guild with no player returns an interface value equal to nil.
- **Collapsing to one queue is irreversible without rework if the bot ever serves a second guild** → Accepted deliberately: the bot is for one private server. The guard makes a second guild fail safe and loudly rather than corrupt shared state, so the failure mode is "does not work there", not "works wrongly there".
- **Session resuming means a stale Lavalink player can survive a restart of the bot** → The session outlives our process for 60 seconds, so a fast restart could attach to a player still holding a track our queue no longer knows about. Mitigation: startup does not assume an empty node; the reconciliation unknown above covers the same ground and is the right place to solve it properly.
- **`goleak` is a new test-only dependency** → Accepted. It is the only way to test the graceful-shutdown requirements, it never enters the runtime binary, and `go mod tidy` keeps it in the test-only graph.

## Migration Plan

One branch per step, stacked with `git-spice`, so each is separately reviewable and revertable:

1. Tooling only: `mise.toml` tasks and the CI check job. No Go changes, so the new checks are visible against the current code first.
2. `internal/queue` with the mutex, the non-mutating `Get`, and its tests.
3. `internal/config` with `Load() (Config, error)` and its tests.
4. `internal/music`: the service, narrow interfaces, sentinel errors, embed builders, bug fixes, and tests.
5. `internal/app` plus the slim `main.go`: router, middleware, ordered startup, graceful shutdown, the `-reset-commands` flag.
6. Dockerfile hardening and the README corrections.

Rollback is per commit. Steps 1 through 4 are behaviour-preserving apart from the two rendering bug fixes and the queue leak, so they carry the least risk; steps 5 and 6 are where the observable startup behaviour changes and are the ones to revert if a deployment misbehaves.
