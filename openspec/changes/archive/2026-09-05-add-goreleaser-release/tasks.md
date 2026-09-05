## 1. Tooling

- [x] 1.1 Pin `goreleaser` (latest v2, currently 2.18.0) and `svu` (3.4.1) in `[tools]` in `mise.toml` and verify `goreleaser --version` and `svu --version` report the pinned versions
- [x] 1.2 Add `dist/` to `.gitignore` and verify `git status` stays clean after a snapshot build

## 2. GoReleaser configuration

- [x] 2.1 Write `.goreleaser.yaml` with `version: 2`, `project_name: discordbot`, and a build producing the `bot` binary for `linux/amd64` and `linux/arm64` with `CGO_ENABLED=0` and `-s -w`; verify `goreleaser check` passes
- [x] 2.2 Add `archives` (`formats: [tar.gz]`), `checksum`, and a `changelog` block grouping conventional commits and excluding `docs`, `test`, `ci`, `chore`, and `style`; verify `goreleaser check` still passes
- [x] 2.3 Add the `dockers_v2` block for `ghcr.io/nico-mayer/discordbot` with tags `{{ .Version }}`, `{{ .Major }}.{{ .Minor }}`, and `latest`, platforms `linux/amd64` and `linux/arm64`, and OCI annotations for source, version, and revision; verify `goreleaser check` passes

## 3. Runtime Dockerfile

- [x] 3.1 Replace `Dockerfile` with the runtime-only image from design.md (`alpine:3.23`, `ca-certificates`, user `bot` uid 10001, `COPY` the binary to `/bin/bot`, `ENTRYPOINT ["/bin/bot"]`) and verify no Go toolchain stage remains
- [x] 3.2 Remove `.dockerignore`, since GoReleaser stages its own build context from the repository root, and verify the snapshot build in 5.1 still succeeds

## 4. Release automation

- [x] 4.1 Add a `release-snapshot` task to `mise.toml` running `goreleaser release --snapshot --clean`, and verify `mise run release-snapshot` produces binaries, archives, and a local image under `dist/`
- [x] 4.2 Add a `release` task that fails on a dirty tree, a branch other than `main`, or a local `main` that differs from `origin/main`, runs `mise run check`, then creates and pushes the annotated tag; verify each guard refuses to run
- [x] 4.6 Compute the version in `release` with `svu next --v0` when no version argument is given, keep an explicit version as an override, and fail when no commit since the current tag bumps the version; verify all four paths in a scratch repository
- [x] 4.7 Add a `next-version` task printing `svu next --v0` and verify it reports `v0.2.0` for the current history
- [x] 4.3 Add `.github/workflows/release.yml` triggered on `push: tags: ["v*"]`, with a gate job calling `./.github/workflows/ci.yml`, and a release job needing it that installs mise, logs in to `ghcr.io` with `GITHUB_TOKEN`, and runs `goreleaser release --clean` with job-level `contents: write` and `packages: write`; verify the file parses with `gh workflow view` or `actionlint` if available
- [x] 4.4 Delete `.github/workflows/docker.yml` and verify no remaining workflow builds or pushes an image outside `release.yml`
- [x] 4.5 Confirm `.github/workflows/ci.yml` is unchanged and still exposes `workflow_call`

## 5. Verification

- [x] 5.1 Run `mise run release-snapshot` and verify the produced image starts and reports a configuration error rather than a missing-binary or exec-format error (`docker run --rm <snapshot-image>`); if the binary is not where the Dockerfile expects it, correct the `COPY` path and rerun
- [x] 5.2 Inspect the snapshot image with `docker inspect` and verify the entrypoint is `/bin/bot`, the user is `bot`, and both platforms are present in the manifest
- [x] 5.3 Run `mise run check` and verify build, gofmt, vet, tests, and coverage all still pass

## 6. Documentation

- [x] 6.1 Add a release section to `README.md` covering `mise run release`, `mise run next-version`, how the version is derived, `mise run release-snapshot` as the replacement for a local `docker build`, and what a tag publishes; verify the commands in it match `mise.toml`
- [x] 6.2 Note in `README.md` that images are published only for releases and that the `main` and `sha-` tags are gone, so deployments must use a version tag or `latest`

## 7. First release

- [x] 7.1 Cut the first release with `mise run release` and verify it tags the version `svu` computed and that the workflow runs the CI gate before publishing
- [x] 7.2 Verify the GitHub release exists with a grouped changelog, archives, and `checksums.txt`, and that `ghcr.io/nico-mayer/discordbot` has the version, minor, and `latest` tags
- [x] 7.3 Pull the published image on the deployment target and verify the bot starts
