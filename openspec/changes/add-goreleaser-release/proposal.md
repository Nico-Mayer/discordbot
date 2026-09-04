## Why

Every push to `main` and every pull request currently builds a multi-platform Docker image, which compiles the whole program twice under QEMU for work that is thrown away on merge. There is also no defined way to cut a release: the single existing tag `v0.1.0` produced an image, but no GitHub release, no changelog, and no repeatable local command.

## What Changes

- Add GoReleaser (`.goreleaser.yaml`) as the single owner of what a release produces: linux binaries, archives, checksums, the multi-platform Docker image, and the GitHub release with a changelog built from conventional commits.
- Add `.github/workflows/release.yml`, triggered only by pushing a `v*` tag. It runs the existing `mise run check` gate first, then GoReleaser.
- Remove `.github/workflows/docker.yml`. Images are no longer built on `main` pushes or pull requests, only on tags.
- Replace the multi-stage `Dockerfile` with a runtime-only image that copies the binary GoReleaser already built. The compile stops happening inside the image build.
- Add mise tasks: `release` (verify, tag, push), `next-version` (print the version the next release would get), and `release-snapshot` (local dry run producing binaries and a local image without publishing anything).
- Derive the version number from the conventional commits since the last tag with `svu`, so `mise run release` needs no argument. An explicit version stays available as an override.
- Pin `goreleaser` and `svu` in `mise.toml` so local runs and CI use the same versions.
- Document the release flow in `README.md`.

Not in scope: version and commit stamped into the binary and reported at startup, signing and SBOM attestations, Docker Hub as a second registry, and any change to the bot's runtime behaviour.

## Capabilities

### New Capabilities

None. This change is release tooling only and adds no requirement to the bot's behaviour, so `.openspec.yaml` sets `skip_specs: true`. This follows the archived `update-dependencies` change, which did the same for a dependency and toolchain change.

### Modified Capabilities

None.

## Impact

- **Removed**: `.github/workflows/docker.yml`.
- **Added**: `.goreleaser.yaml`, `.github/workflows/release.yml`.
- **Modified**: `Dockerfile` (runtime-only), `mise.toml` (`goreleaser` and `svu` tools, `release`, `next-version` and `release-snapshot` tasks), `README.md`, `.gitignore` (`dist/`).
- **Removed**: `.dockerignore`. GoReleaser stages its own build context, so a repository-root ignore file has no consumer.
- **Unchanged**: `.github/workflows/ci.yml` keeps running on every push and pull request, and stays the gate the release workflow calls.
- **Image**: still `ghcr.io/nico-mayer/discordbot`, still `linux/amd64` and `linux/arm64`, still entrypoint `/bin/bot`. Branch and per-commit `sha-` tags stop being published; released versions get `X.Y.Z`, `X.Y`, and `latest`.
- **Consumers**: anything pulling the `main` tag or an `sha-` tag must move to a version tag or `latest`.
- **Permissions**: the release workflow needs `contents: write` (GitHub release) and `packages: write` (GHCR push). Both use the built-in `GITHUB_TOKEN`; no new secret.
