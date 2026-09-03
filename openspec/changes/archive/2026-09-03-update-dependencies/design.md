## Context

See proposal.md - Why. The relevant constraints:

- The logging dependency changes both module path and major version at once: `github.com/charmbracelet/log` v1.0.0 -> `charm.land/log/v2` v2.0.0. Go treats these as unrelated modules, so this is an import rewrite, not a version bump.
- `charm.land/log/v2` and `github.com/charmbracelet/log/v2` both resolve to v2.0.0 from the same repo, but the module's own `go.mod` declares its path as `charm.land/log/v2`. Only that path is canonical.
- Three places state a Go version and currently disagree: the `go` directive (`1.26.1`), the Dockerfile build image (`golang:1.26-alpine`), and `mise.toml` (silent - Go is inherited from the global mise config, which has `1.27.1` as `latest`).
- Registry tag lookups were blocked in this session, so `golang:1.27-alpine` is unverified and must be confirmed during apply.
- The codebase uses only package-level log functions and no styling API, which is what makes the migration mechanical.

## Goals / Non-Goals

**Goals:**

- Complete the log v2 migration with `go build` and a real bot run as the gate.
- Make all three Go version sources state `1.27.1` explicitly.

**Non-Goals:**

- No adoption of new log v2 features (custom styles, `colorprofile`, structured output changes). The migration is import-path-only.
- No explicit bumping of indirect modules, the Alpine runtime base, or the OpenSpec CLI pin.
- No switch away from package-level logging to an injected `*log.Logger`, tempting as v2 makes that cleaner. Separate change.

## Decisions

**Use `charm.land/log/v2`, not `github.com/charmbracelet/log/v2`.**
Both paths serve v2.0.0, but the module declares `charm.land/log/v2` in its `go.mod` and the upgrade guide names it as the vanity domain to migrate to. Requiring the GitHub path would work today but leaves the project on a non-canonical alias that upstream may stop publishing. Alternative considered: stay on the `github.com` path to keep the diff visually smaller. Rejected - the diff is four lines either way.

**Rewrite imports with a single mechanical substitution, then let `go mod tidy` settle the graph.**
`go get charm.land/log/v2@v2.0.0` followed by rewriting the four import lines and `go mod tidy` is the sequence the upgrade guide prescribes. Ordering matters: rewriting imports before the `go get` leaves the build broken between steps, which makes a compile failure ambiguous. Alternative considered: a `replace` directive to alias the old path onto v2. Rejected - `replace` is for temporary local overrides, and it would hide the real dependency from anyone reading `go.mod`.

**Pin `go = "1.27.1"` in `mise.toml` as an exact version, not `"latest"` or `"1.27"`.**
An exact pin is the point of the request: `latest` reintroduces the drift this change exists to remove, and `1.27` silently absorbs patch releases. Trade-off: patch updates now need a deliberate edit. That is the intended behavior for a pin.

**Move the `go` directive, the Dockerfile image, and the mise pin together in one commit.**
These three must agree or the build misbehaves: a `go` directive of `1.27.1` against a `golang:1.26-alpine` image forces an implicit toolchain download inside the Docker build, and fails outright under `GOTOOLCHAIN=local`. So the Dockerfile edit is a consequence of the Go pin, not independent scope. Alternative considered: pin mise only and leave `go.mod` and the Dockerfile at 1.26. Rejected - it leaves the local dev toolchain and the release build on different language versions.

**Still no `toolchain` directive.**
With `mise.toml` pinning `1.27.1` for local work and the Docker tag pinning the build, a `toolchain` line is a third copy of the same fact with its own upkeep. Alternative considered: add `toolchain go1.27.1` for anyone building without mise. Rejected as redundant given two existing pins.

## Risks / Trade-offs

- **`go mod tidy` pulls log v2's new tree (`charm.land/lipgloss/v2`, `charmbracelet/ultraviolet`) and this conflicts with an existing indirect requirement** → Read the `go.mod` diff before committing. If Minimal Version Selection cannot settle it, the failure is explicit at tidy time, not latent. Rollback is `git checkout go.mod go.sum`.
- **Log output changes appearance even though the API is source-compatible** (v2 detects color profiles automatically where v1 used termenv) → Run the bot and look at actual log lines for level colors, timestamps, and key/value formatting. A clean build proves nothing about rendering.
- **A missed import leaves the old module in the graph** → After tidy, `rg 'github.com/charmbracelet/log'` must return nothing and `go.mod` must contain no `charmbracelet/log` requirement.
- **`golang:1.27-alpine` does not exist yet** → Resolve the tag as the first apply step. If unavailable, pin mise to `1.27.1`, hold the `go` directive at `1.26.1`, and leave the Dockerfile alone rather than guessing a tag.
- **Docker build cannot be verified without a running daemon** → Say the Dockerfile edit is unverified rather than implying it builds.
- **The four outdated indirect modules stay outdated** → Accepted. This change is scoped to direct dependencies by request; the indirect bumps remain available as a follow-up.

## Open Questions

None that block implementation.
