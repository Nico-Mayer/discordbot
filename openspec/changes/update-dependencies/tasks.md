## 1. Pre-flight

- [x] 1.1 Confirm the working tree is clean via `git status --short` producing no output, so the dependency diff is reviewable in isolation
- [x] 1.2 Record the pre-upgrade baseline with `go build ./...` exiting 0, so any later failure is attributable to this change
- [x] 1.3 Confirm `golang:1.27-alpine` exists in the registry. If it does not, skip task 3.2 and 3.3 and note that the `go` directive stays at `1.26.1`

## 2. Log v2 migration

- [x] 2.1 Run `go get charm.land/log/v2@v2.0.0` and verify `go.mod` gains a `charm.land/log/v2 v2.0.0` requirement
- [x] 2.2 Rewrite the import path from `github.com/charmbracelet/log` to `charm.land/log/v2` in `main.go`, `config/config.go`, `bot/events.go`, and `bot/commands.go`, then verify `rg 'github.com/charmbracelet/log'` returns no matches
- [x] 2.3 Run `go mod tidy` and verify `go.mod` no longer requires `github.com/charmbracelet/log`, and that `muesli/termenv` and `aymanbagabas/go-osc52/v2` are gone from the indirect block
- [x] 2.4 Review the `go.mod`/`go.sum` diff and confirm every added indirect entry traces to log v2's tree (`charm.land/lipgloss/v2`, `charmbracelet/ultraviolet`, and their transitive deps). Flag anything that does not
- [x] 2.5 Verify the code compiles unchanged apart from imports by running `go build ./...` and `go vet ./...`, both exiting 0, with no edits to any log call site

## 3. Go version pinning

- [x] 3.1 Add `go = "1.27.1"` to the `[tools]` section of `mise.toml`, then verify `mise exec -- go version` reports `go1.27.1`
- [x] 3.2 Raise the `go` directive in `go.mod` from `1.26.1` to `1.27.1` and verify `go build ./...` still exits 0
- [x] 3.3 Bump the Dockerfile build stage from `golang:1.26-alpine` to `golang:1.27-alpine`, and verify no `toolchain` directive was added to `go.mod` and the `alpine:3.21` runtime line is untouched
- [x] 3.4 Verify the container builds by running `docker build -t discordbot-deptest .` and confirming it succeeds. If no Docker daemon is available, state that the Dockerfile edit is unverified rather than assuming it works

## 4. Verification

- [x] 4.1 Confirm the other four direct dependencies are unchanged and still latest by running `go list -m -u github.com/disgoorg/disgo github.com/disgoorg/disgolink/v3 github.com/disgoorg/snowflake/v2 github.com/joho/godotenv` and verifying no `[vX.Y.Z]` update markers appear
- [x] 4.2 Start the bot against Discord with a real config and verify it connects to the gateway, registers slash commands, joins a voice channel, and plays audio through Lavalink
- [x] 4.3 Inspect the actual log output from that run and confirm level colors, timestamps, and key/value formatting still render correctly - log v2 changed color profile detection, which a compile-clean build does not cover
- [x] 4.4 Commit as two atomic conventional commits: the log v2 migration, and the Go version pinning
