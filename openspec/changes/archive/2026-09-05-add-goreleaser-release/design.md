## Context

See proposal.md - Why.

Current state: `ci.yml` runs `mise run check` on push and pull request and exposes `workflow_call`. `docker.yml` calls it, then builds and pushes a multi-platform image on `main`, on `v*` tags, and (build only) on pull requests. The `Dockerfile` is multi-stage: it compiles the program with the Go toolchain inside the image, then copies the binary into `alpine:3.23` running as uid 10001, entrypoint `/bin/bot`.

Constraints:

- The toolchain is pinned in `mise.toml`; CI installs it with `jdx/mise-action`. Anything new must be pinned the same way.
- The quality gate is `mise run check`. No golangci-lint (see `.claude/CLAUDE.md`).
- The bot is deployed as a container. Binaries and archives are a by-product, not the point.
- One image consumer exists in practice (the deployment), and the repository is private-use, so a tag policy change is cheap but should still be stated.

## Goals / Non-Goals

**Goals:**

- One file describes a release. Reading `.goreleaser.yaml` answers "what does cutting a tag produce".
- The program is compiled once per release, natively, not once per architecture under QEMU.
- A release is reproducible locally before it is published.
- The published image keeps its current name, platforms, and entrypoint.

**Non-Goals:**

- Release-please or semantic-release. The version is computed from the commits, but nothing else about the release is automated: a human still decides when to cut one.
- Publishing from a maintainer's machine as the normal path. Local runs stay snapshots.
- Darwin and Windows binaries. The bot runs in Linux containers.

## Decisions

### GoReleaser owns the image build, not `docker/build-push-action`

`dockers_v2` builds the multi-platform image from binaries GoReleaser has already cross-compiled, using buildx to assemble the manifest. Chosen over keeping `docker/build-push-action` and merely restricting its trigger to tags, because that leaves two build paths for the same artifact: Go cross-compiles for the archives, then the Dockerfile compiles again inside QEMU for each platform. One config, one compile, and the archive binary is byte-for-byte the image binary.

`dockers_v2` requires GoReleaser v2.12+ and buildx. Both are available; the pinned version is well past that.

Alternative rejected: `dockers` plus `docker_manifests`. It needs per-architecture image blocks and intermediate `-amd64`/`-arm64` tags. `dockers_v2` exists to replace that.

### The Dockerfile becomes runtime-only

The build stage disappears; the image copies the binary from the build context GoReleaser prepares:

```dockerfile
FROM alpine:3.23
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 bot
COPY $TARGETOS/$TARGETARCH/bot /bin/bot
USER bot
ENTRYPOINT ["/bin/bot"]
```

Base image, user, uid, and entrypoint stay exactly as they are, so the deployment does not change.

Trade-off: `docker build .` at the repository root stops working on its own, because the binary is no longer produced inside the image. `mise run release-snapshot` replaces it and is the same command CI runs, minus publishing. This is recorded in the README.

GoReleaser stages the binaries into the build context as `<os>/<arch>/bot`, so the `COPY` uses the platform arguments buildx sets. A flat `COPY bot /bin/bot` fails, which is what the snapshot run in the task list catches before any tag is pushed.

### Trigger on pushing a `v*` tag, gated by the existing CI workflow

`release.yml` runs on `push: tags: ["v*"]`, and its first job is `uses: ./.github/workflows/ci.yml`, exactly as `docker.yml` does today. The release job needs the gate to pass.

Chosen over triggering on GitHub's `release: published` event, which would require creating a release before the artifacts that belong to it exist and would leave GoReleaser editing a release it did not create. Tag push keeps GoReleaser the sole author of the release.

Permissions are set at job level: `contents: write` to create the release, `packages: write` to push to GHCR. The CI gate job keeps the default read-only token.

Nothing publishes on pull requests, and no pull-request path can reach a job holding a write token, because the workflow only triggers on tags.

### CI runs GoReleaser through mise, not `goreleaser-action`

`goreleaser` is added to `[tools]` in `mise.toml`, and the release workflow installs it with the `jdx/mise-action` step the project already uses. Chosen over `goreleaser/goreleaser-action`, which would pin the version a second time and let the local and CI versions drift apart. One pin, one source of truth, and `mise run release-snapshot` locally is the same binary CI uses.

### `mise run release` tags, CI publishes

```
mise run release
```

verifies the branch, the working tree, and that `main` matches `origin/main`, works out the version, runs `mise run check`, creates the annotated tag, and pushes it. Everything after that happens in CI. The task never builds or pushes artifacts itself, so there is no path where a maintainer's machine publishes an image that CI never validated.

`mise run release-snapshot` runs `goreleaser release --snapshot --clean`: binaries, archives, and a local image, no tag required, nothing pushed. It is the pre-flight check and the replacement for a local `docker build`.

### `svu` computes the version, GoReleaser does not

GoReleaser reads the version from the git tag and has no opinion on what the next one should be. Its own cookbook points at `svu` for that, which is by the same author, is a single binary in the mise registry, and reads the conventional commits since the last tag.

`mise run release` calls `svu next --v0`. `--v0` keeps the major at zero while the project is pre-1.0, so a breaking change bumps the minor instead of jumping to 1.0.0. Without it, the `refactor(config)!` and `feat!` commits already in the history would make the next release 1.0.0. Reaching a stable major stays a deliberate act: `mise run release 1.0.0`.

`mise run next-version` prints the same computation without doing anything, and `release` echoes the version and the current one before it runs the checks, so the number is never a surprise.

The task refuses to release when `svu` reports no change since the current tag, which is the case where every commit since the last release was a `chore` or `docs`.

Alternatives rejected: release-please and semantic-release, both of which want to own the changelog, the release notes, and a release PR. GoReleaser already produces the changelog and the release from the same conventional commits, so those tools would duplicate it and add a Node toolchain. `git-cliff` was not needed for the same reason.

The changelog groups commits by conventional-commit prefix and the version is derived from the same prefixes, so one commit convention drives both.

### Image tags: `X.Y.Z`, `X.Y`, `latest`

Matches what `docker/metadata-action` produces today for semver tags. Dropped: `type=ref,event=branch` (the `main` tag) and `type=sha`, which are exactly the per-commit builds this change removes. `latest` is unconditional, which is correct while releases are strictly increasing; a patch to an old version would need `{{ if not .IsNightly }}`-style guarding, and that case does not exist here.

### Archives and changelog

`linux/amd64` and `linux/arm64` binaries, `tar.gz` archives, `checksums.txt`. Binaries are needed for the image regardless; publishing them costs one config block and makes the GitHub release inspectable without pulling the image.

The changelog groups commits by conventional-commit prefix and excludes `docs`, `test`, `ci`, `chore`, and `style`. The project already writes conventional commits, so this produces a usable changelog with no extra discipline.

`project_name` is set to `discordbot` (the module name and the repository name) and the binary stays `bot`, so the entrypoint and the image name are unchanged.

## Risks / Trade-offs

- **A computed version is wrong because a commit was mislabelled** → `mise run next-version` and the line `release` prints before it does anything both show the number first. An explicit version overrides it, and a wrong tag is recoverable by releasing the next version rather than by moving a published one.
- **A bad tag publishes a bad release** → Tags are cheap to delete but a pushed GHCR image and a published release are visible. Mitigated by the CI gate running before GoReleaser and by the snapshot dry run. Recovery is to delete the tag and release and cut the next patch version, not to overwrite a published one.
- **`main` no longer produces an image, so a regression on `main` is only found at release time** → `mise run check` still runs on every push and pull request; only the image build moves. The image is now a thin copy over a binary CI already built, so the surface that stopped being exercised per-commit is small.
- **Consumers pinned to the `main` or `sha-` image tags break** → Only the deployment consumes the image; it moves to `latest` or a version tag. Called out in the README.
- **`goreleaser` in `mise.toml` slows down `mise install` for anyone who only wants to build the bot** → It is a single prebuilt binary from the mise registry, so the cost is one download, and it keeps local and CI versions identical.
- **GoReleaser is a new external dependency in the release path** → Confined to release time. A failure there never affects the running bot, and the fallback is a manual `docker buildx build --push` with the old multi-stage Dockerfile recoverable from git history.

## Migration Plan

1. Land the configuration, workflow, Dockerfile, and mise changes together. `docker.yml` is deleted in the same commit that adds `release.yml`, so there is never a window where both build images.
2. Run `mise run release-snapshot` locally and confirm the image runs and the binary layout is right. Nothing is published.
3. Cut the first release with `mise run release` and confirm the GitHub release, the archives, and the version, minor, and `latest` image tags. With the current history `svu` computes `v0.2.0`.
4. Point the deployment at a version tag or `latest`.
5. Rollback: restore `docker.yml` and the multi-stage `Dockerfile` from git history. The published `v0.1.0` image is untouched by any of this and remains pullable.
