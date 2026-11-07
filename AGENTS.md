# Suggestions

- Use the Context7 `/find-docs` skill to check the latest Discord and Disgo API

Before any Go coding, review, debugging, troubleshooting, or setup task, load the `samber/cc-skills-golang@golang-how-to` skill first - it routes to whichever other Go skills the task needs.

## Tooling

No `golangci-lint` in this project - it is overkill here. Do not add it or a `.golangci.yml`. The quality gate is `mise run check` (build, gofmt, `go vet`, `go test -race`, coverage report). For modern-Go suggestions use `go fix -diff ./...` instead of a linter.
