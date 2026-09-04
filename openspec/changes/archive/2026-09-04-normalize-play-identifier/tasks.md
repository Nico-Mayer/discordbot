## 1. Normalisation

- [x] 1.1 Add `normalizeIdentifier` to `internal/music/identifier.go`, trimming whitespace and stripping matched `<…>` pairs until none is left, trimming again after each; verify a table test covers a plain value, a padded value, `<url>`, `< url >`, `<<url>>`, a lone `<`, and brackets in the middle
- [x] 1.2 Make `isURL` match the scheme case-insensitively over the first 8 bytes only; verify `HTTPS://example.com` and `Http://example.com` return true while `ftp://example.com` and a mid-string URL still return false
- [x] 1.3 Add table cases asserting `normalizeIdentifier` is idempotent and never leaves surrounding whitespace, for the same inputs as 1.1; verify they pass (the property version over arbitrary input is 5.2)

## 2. Wiring the normalised value through

- [x] 2.1 Normalise the raw option value once in `Service.Enqueue` and pass only the normalised value to `loadTrack`; verify a service test asserts the fake node received the trimmed value for a padded URL
- [x] 2.2 Build `NoResultsError` and `LoadError` from the normalised value; verify a test asserts the "nothing found" message quotes the trimmed value, not the raw one

## 3. Empty values

- [x] 3.1 Add an `ErrEmptyIdentifier` sentinel to `errors.go` and its German message "Gib einen Link oder einen Suchbegriff ein." to `copy.go`, wiring the pair into the `userMessages` table; verify `UserMessage` returns the message and reports the error as known
- [x] 3.2 Reject an empty normalised value in `Enqueue` before the node is contacted; verify a service test asserts the fake node received no call and the queue was untouched for the inputs `""`, `"   "`, and `"<>"`

## 4. Client-side bound

- [x] 4.1 Set `MaxLength` to 1000 on the `/play` option in `commands.go`; verify a test asserts the registered option carries the limit
- [x] 4.2 Rewrite the `NoResultsError.UserMessage` comment in `errors.go`, which still justifies the clamp with "/play sets no maximum length"; verify the existing clamp test still passes unchanged, so the clamp is shown not to depend on the option limit

## 5. Tests that assert the old behaviour

- [x] 5.1 Invert the `scheme uppercased` case in `identifier_test.go` to expect true and rename it to state that; rename `leading space` to say `isURL` does not trim, keeping `want: false`; verify `go test ./internal/music -run TestIsURL` passes
- [x] 5.2 Add `FuzzNormalizeIdentifier` asserting idempotence, no surrounding whitespace, and that `resolveIdentifier(normalizeIdentifier(x))` never panics nor returns empty for a non-empty normalised value, seeded with `" https://example.com "`, `"<https://example.com>"`, `"< https://example.com >"`, `"HTTPS://example.com"`, `"   "`, and `"<>"`; leave `FuzzResolveIdentifier`'s invariants as they are; verify `go test -run FuzzNormalizeIdentifier -fuzz FuzzNormalizeIdentifier -fuzztime 30s ./internal/music` finds no failure

## 6. Whole-change verification

- [x] 6.1 Run `go build ./... && go test ./... && go vet ./...`; verify all pass
- [x] 6.2 Run `/play` in the guild with a padded link, a link copied out of one of the bot's own replies, an uppercase scheme, and a blank value; verify the first three play and the fourth returns the empty-value message without a node round trip
