_Applies to every group: `golang-how-to` (orchestrator), `golang-refactoring` (behaviour-preserving transforms, stacked PRs), `golang-gopls` (find references before a rename, diagnostics after an edit), `golang-naming`, `golang-code-style`. Groups 2-7 also carry `golang-project-layout` (the `internal/` convention; the `cmd/` half is knowingly declined, see design.md) and `golang-testing`. `golang-lint` and `golang-performance` are deliberately not used - see design.md._

## 1. Tooling and CI

_Skills: `golang-continuous-integration` (owns `.github/workflows/*.yml` and the quality gate), `golang-modernize` (Go 1.27 toolchain refresh)._

- [x] 1.1 Add `mise.toml` tasks `fmt` (`gofmt -l .`, failing on any output), `vet` (`go vet ./...`), and `test` (`go test -race ./...`) and verify each runs standalone via `mise run <task>`
- [x] 1.2 Add a `mise.toml` `check` task running build, `fmt`, `vet`, and `test` in that order, and verify `mise run check` executes every step and stops at the first failure
- [x] 1.3 Add `.github/workflows/ci.yml` running `mise run check` on push and pull request to `main`, and verify the workflow is picked up by `gh workflow list` and its first run completes green
- [x] 1.4 Make the existing Docker workflow depend on the new check job so an image is never pushed from a build that fails formatting, vet, or tests, and verify the `needs:` reference resolves
- [x] 1.5 Add a coverage step to the check task reporting total coverage via `go test -coverprofile`, with no threshold gate; verify the percentage is printed and that a low number does not fail the build
- [ ] 1.6 Run `/golang-how-to configure` to write force-trigger rules for the vendored Go skills into `.claude/CLAUDE.md`, so the skills listed per group below load during apply instead of firing opportunistically; verify `.claude/CLAUDE.md` names the skills and that `golang-lint` and `golang-performance` are excluded per the design

## 2. Queue package

_Skills: `golang-concurrency` (mutex placement on a single shared queue), `golang-safety` (`append` backing-array aliasing, defensive copies, usable zero value), `golang-testing` (table-driven, `t.Parallel()`)._

- [x] 2.1 Create `internal/queue` with a single `Queue` holding an unexported track slice guarded by a `sync.Mutex`, exposing `Add`, `Next`, `Clear`, `Len`, and `Tracks()` returning a copy; verify `go build ./...` succeeds
- [x] 2.2 Delete `QueueManager` and its `map[snowflake.ID]*Queue` entirely, including the `Get`/`GetOrCreate`/`Delete` surface; verify `rg -i 'queuemanager|GetOrCreate' .` returns nothing outside the openspec directory
- [x] 2.3 Confirm the zero value of `Queue` is usable with no constructor, so a plain `queue.Queue{}` field needs no initialisation; verify a test exercises `Add` and `Next` on a zero-value queue
- [x] 2.4 Write `internal/queue/queue_test.go` covering add, take in FIFO order, take from empty, clear, `Len`, and that `Tracks()` returns a copy the caller cannot use to mutate the queue; verify `go test ./internal/queue/` passes
- [x] 2.5 Add a concurrency test driving `Add`, `Next`, and `Clear` from multiple goroutines; verify `go test -race ./internal/queue/` passes and fails if the mutex is removed
- [x] 2.6 Confirm `Next()` advancing with `q.tracks = q.tracks[1:]` cannot let a caller observe a stale backing array; verify a test appends after taking and asserts the taken track is unchanged
- [x] 2.7 Mark every pure subtest in this package `t.Parallel()`; verify `go test -race ./internal/queue/` still passes

## 3. Config package

_Skills: `golang-error-handling` (`errors.Join`, sentinel errors), `golang-design-patterns` (global-state avoidance), `golang-security` (secrets must not reach an error or a log line), `golang-testing` (`t.Setenv` forbids `t.Parallel` here)._

- [x] 3.1 Create `internal/config` with `Load() (Config, error)`, no `sync.Once` and no package-level instance, keeping `godotenv.Load()` best-effort so a missing `.env` is not an error; verify `go build ./...` succeeds
- [x] 3.2 Accumulate validation failures with `errors.Join` so every missing required variable is reported in one error, and drop the duplicate `LAVALINK_HOST` and `LAVALINK_PORT` reads by keeping only the joined address field; verify a test with three variables missing names all three
- [x] 3.3 Validate that the Lavalink port parses as an integer in 1..65535 and that the guild ID parses as a snowflake, each error naming the variable; verify the table-driven test covers non-numeric, zero, out-of-range, and malformed-snowflake cases
- [x] 3.4 Ensure no error message includes the token or Lavalink password value; verify with a test that sets recognisable secret values and asserts they are absent from the joined error string
- [x] 3.5 Write the remaining `internal/config/config_test.go` cases using `t.Setenv` and `t.TempDir` for the happy path, the `.env`-absent path, the process-environment-wins-over-`.env` path, and the secure-flag default of `false`; verify `go test ./internal/config/` passes

## 4. Music package: errors, formatting, embeds

_Skills: `golang-error-handling` (sentinels, `%w`, `errors.Is`), `golang-safety` (nil `URI` and `ArtworkURL` derefs), `golang-testing` (table-driven, fuzz targets), `golang-pkg-go-dev` (verify disgo embed and lavalink signatures via `godig`)._

- [x] 4.1 Create `internal/music/errors.go` with the sentinels `ErrNoPlayer`, `ErrNotInVoice`, `ErrNothingPlaying`, `ErrQueueEmpty`, `ErrNoResults` and a table mapping each to its German user message; verify a test asserts every sentinel has a non-empty message
- [x] 4.2 Move duration formatting into `internal/music/embed.go` and fix it to render `h:mm:ss` at one hour and above; verify the table-driven test covers `0` → `0:00`, 45s → `0:45`, 3m07s → `3:07`, and 1h05m30s → `1:05:30`
- [x] 4.3 Move URL detection out of the play handler into a tested helper; verify the table-driven test covers `http://`, `https://`, a bare search phrase, a phrase containing a URL mid-string, and the empty string
- [x] 4.4 Bound the `/queue` embed: cap the listed tracks at 20, append an "and N more" line, and add a defensive check that the description cannot exceed 4096 characters even with pathologically long titles; verify a test builds a 200-track queue embed and asserts the description length and the residual count
- [x] 4.5 Extract the embed builders (now playing, queue list, track confirmation, queued confirmation, error, success, info) as pure functions returning `discord.Embed`, with the colour literals replaced by named constants; verify `go build ./...` succeeds
- [x] 4.6 Fix the track confirmation embed so duration is a single field with both a name and a value, replacing the current two fields where the duration is used as a field name with no value; verify a test asserts the field's `Name` and `Value` are both non-empty
- [x] 4.7 Sweep every `*track.Info.URI` and `*track.Info.ArtworkURL` dereference and guard it, since both are pointers that Lavalink may leave nil; verify `rg '\*[a-z]+\.Info\.(URI|ArtworkURL)' internal/` shows no unguarded deref
- [x] 4.8 Add fuzz targets for the duration formatter and the URL detector; verify `go test -fuzz -fuzztime=30s` on each finds no panic and that the seed corpus covers the table cases
- [x] 4.9 Write `internal/music/embed_test.go` asserting titles, colours, artwork fallback to the YouTube thumbnail when `ArtworkURL` is nil, and that a track with a nil `URI` does not panic; verify `go test ./internal/music/` passes

## 5. Music package: service and seams

_Skills: `golang-dependency-injection` (constructor injection for testability), `golang-design-patterns` (constructor APIs), `golang-context` (propagation, timeouts), `golang-safety` (typed-nil interfaces), `golang-testing` (fakes, table-driven), `golang-pkg-go-dev`._

- [x] 5.1 Declare the consumer-side `Lavalink`, `Player`, and `Node` interfaces in `internal/music` carrying only the methods the service uses; verify `go build ./...` succeeds
- [x] 5.2 Write hand-rolled fakes for those interfaces in `internal/music` test files, with settable return values and recorded calls; verify a smoke test constructs a `Service` against the fakes
- [x] 5.3 Implement `Service` holding the single `Queue`, the `Lavalink` seam, the configured guild ID, and the injected `*slog.Logger`, constructed by a single constructor with no exported mutable fields and no per-method `guildID` parameter; verify `go build ./...` succeeds
- [x] 5.4 Implement `Pause` returning the resulting paused state and `ErrNoPlayer` when there is no active player; verify the table-driven test covers pause, resume, and the no-player path
- [x] 5.5 Implement `Stop` clearing the queue, nulling the track, and leaving the voice channel, returning `ErrNoPlayer` when there is no active player; verify the test asserts the queue is empty afterwards and that a failure leaves no partial state
- [x] 5.6 Implement `Skip` taking the next queued track, returning `ErrQueueEmpty` when there is none and `ErrNoPlayer` when there is no player; verify the test asserts the current track keeps playing on the empty-queue path
- [x] 5.7 Implement `NowPlaying` returning the current track and position, with `ErrNoPlayer` and `ErrNothingPlaying` distinguished; verify the table-driven test covers all three outcomes
- [x] 5.8 Implement `Queue` returning a bounded view of the queue contents for the embed builder; verify the test asserts an empty queue reports empty and that the returned slice is a copy
- [x] 5.9 Implement `Enqueue` covering the voice-channel requirement, search-prefix decision, track and playlist and search results, the empty-playlist and empty-search guards, `ErrNoResults`, load errors, and the queue-versus-play-now branch; verify the table-driven test covers each of those branches
- [x] 5.10 Bound the track load with a context timeout derived from the caller's context and verify a test using an already-cancelled context returns without hanging
- [x] 5.11 Wrap every error crossing a package boundary with `%w` and verify `rg 'fmt\.Errorf\([^)]*%v' internal/` returns nothing, since there is no `errorlint` to catch this mechanically

## 6. Music package: router and event handlers

_Skills: `golang-context` (`Background` vs `TODO`, the context-in-a-struct exception), `golang-error-handling` (single handling rule, `slog`), `golang-testing`, `golang-pkg-go-dev` (verify `handler.Mux` API)._

- [x] 6.1 Move the slash command definitions to `internal/music/commands.go`, removing the commented-out `source` option; verify `go build ./...` succeeds
- [x] 6.2 Build the `handler.Mux` router registering all six commands as thin adapters that read options, call the service, and send an embed; delete `Bot.InitHandlers` and the `map[string]func(...)` dispatch and verify no references to the old map remain
- [x] 6.3 Set `mux.Error` to reply with an ephemeral embed using the sentinel message table, falling back to a generic message plus an error-level log for unrecognised errors; verify a test asserts each sentinel maps to its message and an unknown error maps to the generic one
- [x] 6.4 Set `mux.NotFound` to log the unknown command name without crashing; verify by asserting on a captured `slog` record
- [x] 6.5 Add a logging middleware recording command name and guild, and apply `middleware.Defer` to `/play` so the deferred acknowledgement is no longer hand-rolled in the handler; verify `/play` still edits its acknowledgement rather than sending a second message
- [x] 6.6 Move the gateway and player event handlers to `internal/music/events.go`, taking their context from a struct field set at construction with a comment stating why, and replace every `context.TODO()`; verify `rg 'context.TODO' internal/ main.go` returns nothing
- [x] 6.7 Add the guild guard at all three edges: router middleware refusing a command from any guild other than the configured one with an ephemeral error, and `OnVoiceStateUpdate` and `OnVoiceServerUpdate` returning early with a debug log; verify a table-driven test covers the configured guild and a foreign guild for each edge
- [x] 6.8 Confirm the bot takes no action in an unconfigured guild and does not leave it; verify a test asserts a foreign-guild voice event leaves the queue and the configured player untouched
- [x] 6.9 Confirm `OnTrackEnd` does not consume a queued track when the end reason forbids advancing; verify the table-driven test covers an advancing reason, a non-advancing reason, and the exhausted-queue path that leaves the voice channel
- [x] 6.10 Handle the error returns currently discarded as `_, _ =` in the deferred reply helpers by logging the failure, since no `errcheck` will catch them; verify `rg '_, _ =|_ = ' internal/` returns only assignments with a comment justifying the discard

## 7. Composition root and entry point

_Skills: `golang-design-patterns` (graceful shutdown, resource lifecycle, `init()` avoidance), `golang-safety` (typed-nil adapter, `defer` in loops), `golang-testing` (`goleak`), `golang-dependency-injection`, `golang-context`, `golang-error-handling` (structured logging with `slog`)._

- [x] 7.1 Create `internal/app` with `Run(ctx, cfg, logger) error` as the single composition root; verify `go build ./...` succeeds
- [x] 7.2 Write the adapter wrapping the real `disgolink.Client` and `disgolink.Player` to satisfy the consumer-side interfaces; verify with a compile-time `var _ music.Lavalink = (*adapter)(nil)` assertion
- [x] 7.3 Make the adapter's `ExistingPlayer` return an untyped nil interface when the underlying client returns no player, never a non-nil interface holding a nil pointer; verify a test asserts the returned value `== nil` so the service's `player == nil` checks keep working
- [x] 7.4 Implement ordered startup registering a cleanup per successfully built dependency, and unwind the cleanups registered so far on any failure; verify no `log.Fatal` remains outside `main` and that `rg 'log\.Fatal' .` returns nothing under `internal/`
- [x] 7.5 Bound gateway open and Lavalink node registration with timeouts derived from the lifetime context, and log a warning rather than failing when the bot's own user cannot be fetched; verify the warning path by asserting on a captured `slog` record
- [x] 7.6 Implement graceful shutdown leaving voice channels, then closing Lavalink, then the gateway, under a bounded shutdown context; verify shutdown returns within the timeout when a close call blocks
- [x] 7.7 Slim `main.go` to flag parsing, logger construction, `signal.NotifyContext`, `config.Load`, `app.Run`, and `os.Exit(1)` on error; verify `main.go` is under 40 lines and `go build ./...` succeeds
- [x] 7.8 Define the `-reset-commands` flag the README already documents and wire it to the reset path so guild and global commands are cleared before registration; verify `go run . -reset-commands` reaches the reset path and `go run . -bogus` exits non-zero with a usage message
- [x] 7.9 Replace the package-level `charm.land/log/v2` calls with the injected `*slog.Logger`, constructed in `main.go` as `slog.New(log.New(os.Stderr))`; verify `rg 'charm.land/log' .` matches only `main.go`
- [x] 7.10 Delete the now-empty `bot/` and `config/` directories and verify `go build ./... && go vet ./...` succeeds with no stale imports

## 8. Docker, docs, and final verification

_Skills: `golang-documentation` (README, doc comments), `golang-dependency-management` (`go mod tidy`, `govulncheck`), `golang-security` (secrets and PII in logs), `golang-continuous-integration`, `golang-modernize`._

- [x] 8.1 Harden the Dockerfile with `-trimpath`, `-ldflags="-s -w"`, `--mount=type=cache` for the module and build caches, a non-root user, and a current Alpine base; verify `docker build .` succeeds and `docker run --rm --entrypoint id <image>` reports a non-root user
- [x] 8.2 Correct the README environment variable names to match `.env.example` and the code (`LAVALINK_HOST`, `LAVALINK_PORT`, `LAVALINK_PASSWORD`, `NODE_NAME`, `NODE_SECURE`), and verify every variable named in the README appears in `internal/config`
- [x] 8.3 Update the README run and development sections to document the `mise run` tasks and the working `-reset-commands` flag; verify each documented command runs
- [x] 8.4 Move `github.com/stretchr/testify` from indirect to a direct requirement and verify `go mod tidy` leaves `go.mod` unchanged
- [x] 8.5 Run `mise run check` and verify build, `gofmt`, `go vet`, and `go test -race ./...` all pass clean
- [x] 8.6 Add `go.uber.org/goleak` as a test-only dependency and assert the shutdown path leaves no goroutines behind; verify the leak check fails when a cleanup step is deliberately removed
- [x] 8.7 Run `govulncheck ./...` and verify it reports no actionable findings in the module tree
- [ ] 8.8 Start the bot against a live Lavalink node and manually exercise `/play` with a search query and with a URL, `/pause` twice, `/queue`, `/now-playing`, `/skip`, and `/stop`, then confirm SIGINT leaves the voice channel and exits zero
