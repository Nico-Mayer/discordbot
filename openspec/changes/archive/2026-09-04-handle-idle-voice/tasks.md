_Applies to every group: `golang-how-to` (orchestrator), `golang-gopls` (find references before changing a signature, diagnostics after an edit), `golang-naming`, `golang-code-style`, `golang-project-layout`, `golang-testing`. `golang-lint` and `golang-performance` are deliberately not used, following `refactor-bot-architecture`._

## 1. Configuration

_Skills: `golang-error-handling` (sentinel errors joined by `errors.Join`), `golang-testing` (`t.Setenv` forbids `t.Parallel` here)._

- [x] 1.1 Add an exported `IdleTimeout` type to `internal/config` carrying a `time.Duration` and an enabled flag, so `off` is representable without a sentinel duration; verify `go build ./...` succeeds
- [x] 1.2 Add `envIdleTimeout(key string) (IdleTimeout, error)` parsing unset or empty as 60 seconds, `off` as disabled, and a non-negative whole number as that many seconds, returning `ErrInvalid` naming the key and the offending value otherwise; verify a table-driven test covers unset, `0`, `60`, `off`, `-1`, `30s`, and `never`
- [x] 1.3 Add `IdleAlone` and `IdleEmptyQueue` fields to `Config` and parse `IDLE_ALONE_SECONDS` and `IDLE_EMPTY_QUEUE_SECONDS` in `Load`, joining their errors alongside the existing ones; verify a test asserts a single load reports both a bad idle value and a missing required variable together
- [x] 1.4 Confirm the two variables are independent by testing that setting only one leaves the other at the 60 second default; verify `go test ./internal/config/` passes

## 2. Occupancy seam

_Skills: `golang-design-patterns` (consumer-declared interfaces), `golang-safety` (nil `ChannelID` handling)._

- [x] 2.1 Declare a `VoiceStates` seam and a plain `VoiceState` value type (`UserID`, nullable `ChannelID`) in `internal/music/seams.go`, so `internal/music` does not import the `discord` package and a fake stays trivial; verify `go build ./...` succeeds
- [x] 2.2 Add a `voiceStateAdapter` in `internal/app` wrapping `client.Caches.VoiceStates(guildID)` into the seam, asserted against the interface with a `var _` declaration; verify `go build ./...` succeeds
- [x] 2.3 Add a `fakeVoiceStates` to `internal/music/fakes_test.go` that returns a configurable set of voice states; verify it satisfies the seam at compile time

## 3. Idle state on the service

_Skills: `golang-concurrency` (one mutex guarding both timers, `time.AfterFunc` lifetime), `golang-safety` (callback must not act on a closed service)._

- [x] 3.1 Replace the positional `NewService` parameters with a `ServiceConfig` struct carrying guild ID, application ID, the Lavalink, voice, and voice-state seams, both idle timeouts, and the logger; verify `go build ./...` succeeds and every existing call site and test compiles
- [x] 3.2 Add `listeners(channelID)` counting cached voice states in that channel excluding the bot's own application ID, and `botChannel()` returning the channel the bot is currently in or nil; verify tests cover an empty channel, a channel holding only the bot, a channel holding another user, and the bot not being connected
- [x] 3.3 Add the `idleMu` mutex, the two `*time.Timer` fields, and a `closed` flag to `Service`, with `armAlone`/`armEmptyQueue` that leave an already-running timer untouched and `cancelAlone`/`cancelEmptyQueue` that stop and clear theirs; verify a test asserts re-arming a running timer does not move its deadline
- [x] 3.4 Make arming a disabled timeout a no-op so `off` never starts a countdown; verify a test asserts no leave happens for a disabled timeout however long its condition holds
- [x] 3.5 Implement the single `leaveIdle` action: under the mutex return early when the service is closed, stop both timers, then stop the player, discard the queue, and leave the voice channel; verify a test asserts it is safe to call twice and issues exactly one voice-leave
- [x] 3.6 Make `leaveIdle` return early when the bot is no longer connected to a voice channel, so a countdown that elapses after the bot has already left does nothing; verify a test covers it
- [x] 3.7 Add `EvaluateOccupancy(ctx)` that recomputes the count from the cache on every call and arms or cancels the alone timer accordingly, doing nothing when the bot is not connected; verify tests cover last user leaving, a user rejoining, and the bot being unconnected
- [x] 3.8 Add `ArmEmptyQueue(ctx)` guarded on the player holding no current track and the queue being empty, so a paused player with a current track does not arm it; verify tests cover queue exhausted, queue still holding tracks, and paused with a current track
- [x] 3.9 Add `CancelIdle()` stopping both timers, and `Close()` marking the service closed and stopping both; verify a test asserts `Close` is safe to call twice and that a timer armed before `Close` never fires its action
- [x] 3.10 Add a test arming both timers at once with a zero and a longer timeout and assert the leave happens exactly once and the other timer is cancelled; verify `go test -race ./internal/music/` passes

## 4. Event wiring

_Skills: `golang-concurrency` (handlers run concurrently), `golang-testing` (handler tests through the existing `handleVoiceStateUpdate` seam)._

- [x] 4.1 Relax the self-only filter in `handleVoiceStateUpdate` so other users' movements reach the service, while keeping the guild guard and forwarding only the bot's own state to Lavalink; verify existing voice-server-handoff tests still pass
- [x] 4.2 Make the bot's own disconnect (`ChannelID == nil`) cancel both timers and discard the queue, and make every other in-guild voice state change call `EvaluateOccupancy`, including the bot moving to another channel; verify tests cover a foreign guild, the bot disconnecting, another user leaving, and the bot changing channel
- [x] 4.3 Replace the direct `service.Leave` call in `handleTrackEnd` with `ArmEmptyQueue`, so the queue-exhausted path no longer leaves inline; verify the existing queue-exhausted test is updated to assert a countdown starts and no voice-leave is issued
- [x] 4.4 Make `OnTrackStart` cancel the empty-queue countdown, behind the same guild guard as the other player events; verify a test asserts a track starting during the countdown cancels it
- [x] 4.5 Add a test asserting a track queued during the empty-queue countdown starts playing without a second `UpdateVoiceState` join; verify `go test ./internal/music/` passes

## 5. Composition root and shutdown

_Skills: `golang-design-patterns` (composition root), `golang-safety` (no goroutine outliving the process)._

- [x] 5.1 Build the voice-state adapter in `app.Run` and pass it, the application ID, and both configured idle timeouts into `NewService` via `ServiceConfig`; verify `go build ./...` succeeds
- [x] 5.2 Register `service.Close` as a cleanup that runs before the voice-leave step, so no countdown can fire against a closing client; verify the shutdown order is asserted by a test on the cleanup list
- [x] 5.3 Confirm a countdown cannot act after shutdown: `Close` must set the `closed` flag and stop both timers under the mutex the callback takes, so a callback already past the timer fires and returns without touching Lavalink or the gateway; verify by mutation that gutting `Close` fails both `TestServiceCloseStopsPendingCountdowns` and `TestServiceDoesNotArmAfterClose`, and that dropping the `closed` check alone fails `TestServiceDoesNotArmAfterClose`. `goleak` is deliberately not the detector here - a pending `time.AfterFunc` is a runtime timer, not a goroutine, so it reports nothing (see design.md)

## 6. Documentation

- [x] 6.1 Add `IDLE_ALONE_SECONDS=60` and `IDLE_EMPTY_QUEUE_SECONDS=60` to `.env.example`; verify the file lists every variable the README documents
- [x] 6.2 Document both variables in the README configuration section, including the `0` and `off` forms and that `IDLE_EMPTY_QUEUE_SECONDS=0` restores the previous leave-immediately behaviour; verify the README no longer claims `NODE_SECURE` is the only optional variable

## 7. Verification

- [x] 7.1 Run `mise run check` and verify build, `gofmt`, `go vet`, and `go test -race ./...` all pass clean
- [x] 7.2 Review the diff for comments that only restate the code and delete them, keeping only the ones recording the accepted bot-counts-as-listener limitation and the timer lifetime constraint; verify the diff carries no comment that could be removed without losing context
