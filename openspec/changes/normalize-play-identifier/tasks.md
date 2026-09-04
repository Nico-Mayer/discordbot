## 1. Normalisation

- [ ] 1.1 Add `normalizeIdentifier` to `internal/music/identifier.go`, trimming whitespace, stripping one matched `<…>` pair, and trimming again; verify a table test covers a plain value, a padded value, `<url>`, `< url >`, a lone `<`, and brackets in the middle
- [ ] 1.2 Make `isURL` match the scheme case-insensitively over the first 8 bytes only; verify `HTTPS://example.com` and `Http://example.com` return true while `ftp://example.com` and a mid-string URL still return false
- [ ] 1.3 Add a test that `normalizeIdentifier` is idempotent and never leaves surrounding whitespace; verify it passes

## 2. Wiring the normalised value through

- [ ] 2.1 Normalise the raw option value once in `Service.Enqueue` and pass only the normalised value to `loadTrack`; verify a service test asserts the fake node received the trimmed value for a padded URL
- [ ] 2.2 Build `NoResultsError` and `LoadError` from the normalised value; verify a test asserts the "nothing found" message quotes the trimmed value, not the raw one

## 3. Empty values

- [ ] 3.1 Add an `ErrEmptyIdentifier` sentinel with the German message "Gib einen Link oder einen Suchbegriff ein." in the copy file; verify `UserMessage` returns it and reports the error as known
- [ ] 3.2 Reject an empty normalised value in `Enqueue` before the node is contacted; verify a service test asserts the fake node received no call and the queue was untouched for the inputs `""`, `"   "`, and `"<>"`

## 4. Client-side bound

- [ ] 4.1 Set `MaxLength` to 1000 on the `/play` option in `commands.go`; verify a test asserts the registered option carries the limit

## 5. Tests that assert the old behaviour

- [ ] 5.1 Rewrite the `scheme uppercased` and `leading space` cases in `identifier_test.go` to expect the new result, renaming them to state it; verify `go test ./internal/music -run TestIsURL` passes
- [ ] 5.2 Restate `FuzzResolveIdentifier`'s invariant against the normalised identifier and add `" https://example.com "`, `"<https://example.com>"`, `"HTTPS://example.com"`, and `"   "` to the seed corpus; verify `go test -run FuzzResolveIdentifier -fuzz FuzzResolveIdentifier -fuzztime 30s ./internal/music` finds no failure

## 6. Whole-change verification

- [ ] 6.1 Run `go build ./... && go test ./... && go vet ./...`; verify all pass
- [ ] 6.2 Run `/play` in the guild with a padded link, a link copied out of one of the bot's own replies, an uppercase scheme, and a blank value; verify the first three play and the fourth returns the empty-value message without a node round trip
