## Context

See proposal.md - Why. The relevant current state:

- `isURL` is two literal `strings.HasPrefix` checks; `resolveIdentifier` returns the raw value for a URL and `lavalink.SearchTypeYouTubeMusic.Apply(identifier)` otherwise (`internal/music/identifier.go`).
- `Service.loadTrack` calls `node.LoadTracks(ctx, resolveIdentifier(identifier))` but builds `NoResultsError` and `LoadError` from the *raw* `identifier` (`internal/music/service.go`), so today the raw and the loaded value can already differ.
- `identifier_test.go` asserts the current behaviour for four of the cases this change inverts: `HTTPS://example.com`, `" https://example.com"`, and the `ftp://` and mid-string cases that must keep failing.
- `FuzzResolveIdentifier` asserts `require.Contains(t, resolved, identifier)` for non-URLs. Normalisation makes that false for any value with surrounding whitespace, so the invariant has to be restated against the normalised value.
- This change assumes `improve-copy-and-logs` has landed: the option is `titel`, and German strings live in `internal/music/copy.go`.

## Goals / Non-Goals

**Goals:**

- A link a member can see is a link is loaded as one, whatever whitespace or wrapping came with the paste.
- A blank value fails with an answer, not with a search for nothing.
- The value the bot quotes back on failure is the value it actually searched for.

**Non-Goals:**

- Parsing or validating the URL itself. Whether a URL resolves to audio is the audio node's judgement, not the bot's.
- Extracting a URL from inside a longer sentence. `listen to https://… now` stays a search phrase; a member who typed a sentence meant a sentence.
- Accepting schemes other than `http` and `https`.
- Stripping Discord markdown beyond the one angle-bracket pair, or unescaping entities.
- Supporting a second search source or a way to choose one.

## Decisions

### Normalisation is one function, applied once, at the edge of the service

A single `normalizeIdentifier(string) string` performs all three steps in a fixed order: trim whitespace, strip one wrapping `<…>` pair, trim whitespace again. The second trim matters because `< https://… >` is a real paste shape.

`Service.Enqueue` normalises the raw option value once and uses the result for everything after: the empty check, `resolveIdentifier`, `LoadTracks`, and the identifier carried in `NoResultsError` and `LoadError`. Nothing downstream sees the raw value.

*Alternative considered.* Normalising inside `resolveIdentifier`. Rejected: `resolveIdentifier` returns a *search-prefixed* string for a phrase, so the normalised value could not be recovered from it for the error messages, and `loadTrack` would still quote the raw value back at the member.

### Angle brackets are stripped as a matched pair only

`<…>` is removed only when the value both starts with `<` and ends with `>`. A lone `<` or a value with brackets in the middle is left alone - it is far more likely to be part of a search phrase than a malformed wrapper.

One pair, not repeated: `<<url>>` is not a shape Discord produces, and stripping greedily risks eating characters from a genuine phrase.

### Scheme matching is case-insensitive on a bounded prefix

`isURL` lowercases only the first 8 bytes before comparing, rather than the whole value. A member is allowed to paste 6000 characters, and the scheme cannot appear past byte 8. This also avoids allocating a lowercase copy of a long value on every call.

### An empty value gets its own error, not "nothing found"

A new `ErrEmptyIdentifier` sentinel, checked in `Enqueue` before anything else, mapping to a German message that tells the member what to supply. Reusing `ErrNoResults` would be a lie - nothing was searched for - and its recovery advice ("check the link") does not fit a member who supplied nothing.

Proposed copy, following the `interface-copy` standard: **Gib einen Link oder einen Suchbegriff ein.**

### `MaxLength` is a second line of defence, not the fix

Setting `MaxLength` on the option makes Discord reject an absurd value in the client, which is a better experience than a doomed round trip. It does not replace bounding what a reply quotes - that is `improve-copy-and-logs`, and it must hold regardless of what the client enforces. 1000 characters is comfortably above any real link or search phrase.

### The fuzz invariant is restated, not dropped

`FuzzResolveIdentifier`'s `Contains` assertion becomes: for a non-URL, the resolved value contains the *normalised* identifier. The properties worth adding, given normalisation is now involved:

- `normalizeIdentifier` is idempotent.
- A normalised value never has leading or trailing whitespace.
- `resolveIdentifier` never panics and never returns the empty string for a non-empty normalised input.

## Risks / Trade-offs

- Trimming changes what a member searched for, so a phrase with deliberate leading whitespace searches differently → whitespace-only differences do not change a search result in practice, and the error message now quotes the normalised value, so what the bot searched for is what the member is shown.
- Angle-bracket stripping could eat a genuine search phrase wrapped in brackets → only a matched leading-and-trailing pair is stripped, which is not a shape a member types by accident.
- Four existing test cases invert, which reads as a regression in the diff → the cases are rewritten with names stating the new expectation, and the fuzz seed corpus gains the new shapes.
- Depends on `improve-copy-and-logs` for the option name and the copy file → sequence it second. If it lands first instead, the option is still `identifier` and the new message goes inline, to be moved when the copy file arrives.

## Migration Plan

Deploy is a restart. `MaxLength` is published with the command set on the next sync. Values that used to fail as searches now succeed as links; no value that used to work stops working. Rollback is a redeploy of the previous image.
