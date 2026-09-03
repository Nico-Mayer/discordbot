## Why

`charmbracelet/log` shipped v2 under Charm's new vanity domain, so the project's logging dependency is now a major version and a module path behind. At the same time the `go` directive (`1.26.1`) trails the installed toolchain (`1.27.1`) and no Go version is pinned in `mise.toml`, so the language version depends on whatever the machine happens to have.

## What Changes

- Migrate the logging dependency from `github.com/charmbracelet/log` v1.0.0 to `charm.land/log/v2` v2.0.0. **BREAKING** at the module level: the import path changes in all four files that use it (`main.go`, `config/config.go`, `bot/events.go`, `bot/commands.go`).
- Re-verify the other four direct dependencies. As of this proposal they are already at their latest published versions, so no upgrade is expected:
  `disgoorg/disgo` v0.19.6, `disgoorg/disgolink/v3` v3.1.0, `disgoorg/snowflake/v2` v2.0.3, `joho/godotenv` v1.5.1
- Pin Go to `1.27.1` in `mise.toml`, which currently pins only the OpenSpec CLI and inherits Go from the global mise config.
- Raise the `go` directive in `go.mod` from `1.26.1` to `1.27.1` and bump the Dockerfile build stage from `golang:1.26-alpine` to `golang:1.27-alpine` so all three Go version sources agree.

## Capabilities

### New Capabilities

None. This change touches dependency pins and build configuration only.

### Modified Capabilities

None. The log v2 migration is an import path swap with no change to log output or runtime behavior, so `skip_specs: true` is set in this change's `.openspec.yaml`.

## Impact

- `main.go`, `config/config.go`, `bot/events.go`, `bot/commands.go`: one import line each. No call sites change - every package-level function in use (`SetReportTimestamp`, `Info`, `Warn`, `Error`, `Debug`, `Fatal`) exists in v2 with the same signature.
- `go.mod`, `go.sum`: the direct requirement, plus indirect churn from v2's new dependency tree (see below).
- `mise.toml`: new `go` pin.
- `Dockerfile`: build-stage image tag only.
- Indirect tree shifts as a side effect of log v2, not as a goal: it drops `muesli/termenv` and `aymanbagabas/go-osc52/v2`, and adds `charm.land/lipgloss/v2` and `charmbracelet/ultraviolet`. This is unavoidable via `go mod tidy` and is not scope creep.
- Out of scope, deliberately: the four outdated **indirect** modules (`klauspost/compress`, `mattn/go-runewidth`, `golang.org/x/crypto`, `golang.org/x/exp`), the `alpine:3.21` runtime base, and the `@fission-ai/openspec` CLI pin. Whatever of these moves will move transitively through `go mod tidy`, not by explicit bump.
- Risk is low. The v2 upgrade guide estimates 5-10 minutes, and its two documented breaking changes - the `SetColorProfile` type switch to `colorprofile.Profile` and the `Styles` struct moving to Lip Gloss v2 - do not apply here: the codebase has no reference to `lipgloss`, `termenv`, `colorprofile`, `SetStyles`, or `DefaultStyles`.
