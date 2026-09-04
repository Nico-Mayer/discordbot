## Context

See proposal.md - Why. The relevant current state:

- `isURL` is two literal `strings.HasPrefix` checks; `resolveIdentifier` returns the raw value for a URL and `lavalink.SearchTypeYouTubeMusic.Apply(identifier)` otherwise (`internal/music/identifier.go`).
- `Service.loadTrack` calls `node.LoadTracks(ctx, resolveIdentifier(identifier))` but builds `NoResultsError` and `LoadError` from the *raw* `identifier` (`internal/music/service.go`), so today the raw and the loaded value can already differ.
- `identifier_test.go` pins four cases against `isURL` directly. Only one inverts: `HTTPS://example.com`. `" https://example.com"` keeps expecting false, because `isURL` is not where trimming happens; the `ftp://` and mid-string cases must keep failing outright.
- `FuzzResolveIdentifier` asserts `require.Contains(t, resolved, identifier)` for non-URLs. That invariant survives untouched, because normalisation sits in `Enqueue` rather than in `resolveIdentifier`, so the fuzz target never sees a value it did not receive whole. What is missing is any property coverage of `normalizeIdentifier` itself.
- `improve-copy-and-logs` has landed: the option is `titel`, German strings live in `internal/music/copy.go`, and `NoResultsError.UserMessage` already clamps the quoted value to `quotedInputLimit`.

## Goals / Non-Goals

**Goals:**

- A link a member can see is a link is loaded as one, whatever whitespace or wrapping came with the paste.
- A blank value fails with an answer, not with a search for nothing.
- The value the bot quotes back on failure is the value it actually searched for.

**Non-Goals:**

- Parsing or validating the URL itself. Whether a URL resolves to audio is the audio node's judgement, not the bot's.
- Extracting a URL from inside a longer sentence. `listen to https://… now` stays a search phrase; a member who typed a sentence meant a sentence.
- Accepting schemes other than `http` and `https`.
- Stripping Discord markdown beyond wrapping angle brackets, or unescaping entities.
- Supporting a second search source or a way to choose one.

## Decisions

### Normalisation is one function, applied once, at the edge of the service

A single `normalizeIdentifier(string) string` trims whitespace, then strips wrapping `<…>` pairs until none is left, trimming again after each. The repeated trim matters because `< https://… >` is a real paste shape.

`Service.Enqueue` normalises the raw option value once and uses the result for everything after: the empty check, `resolveIdentifier`, `LoadTracks`, and the identifier carried in `NoResultsError` and `LoadError`. Nothing downstream sees the raw value.

*Alternative considered.* Normalising inside `resolveIdentifier`. Rejected: `resolveIdentifier` returns a *search-prefixed* string for a phrase, so the normalised value could not be recovered from it for the error messages, and `loadTrack` would still quote the raw value back at the member.

A consequence worth stating, because it shapes the tests: `isURL` and `resolveIdentifier` keep operating on whatever they are handed. Only their case-sensitivity changes. A direct unit call such as `isURL(" https://example.com")` therefore still reports false, and that case stays in the table as a statement of where trimming does and does not live.

### Angle brackets are stripped as matched pairs, repeatedly

`<…>` is removed only when the value both starts with `<` and ends with `>`. A lone `<` or a value with brackets in the middle is left alone - it is far more likely to be part of a search phrase than a malformed wrapper.

Stripping repeats until no matched pair is left, and each pass trims the whitespace it exposes, so `< <url> >` reduces the same way `<url>` does.

*Alternative considered.* Stripping one pair only. Rejected: it makes `normalizeIdentifier` non-idempotent, because `<<url>>` reduces to `<url>`, which normalises further. A normalisation function that does not reach a fixed point is a trap for any second caller, and the property is worth more than the case it costs: a member searching for a phrase they deliberately wrapped in two bracket pairs. That phrase searches better unwrapped anyway.

### Scheme matching is case-insensitive on a bounded prefix

`isURL` lowercases only the first 8 bytes before comparing, rather than the whole value. A member is allowed to paste 6000 characters, and the scheme cannot appear past byte 8. This also avoids allocating a lowercase copy of a long value on every call.

### An empty value gets its own error, not "nothing found"

A new `ErrEmptyIdentifier` sentinel, checked in `Enqueue` before anything else, mapping to a German message that tells the member what to supply. Reusing `ErrNoResults` would be a lie - nothing was searched for - and its recovery advice ("check the link") does not fit a member who supplied nothing.

Proposed copy, following the `interface-copy` standard: **Gib einen Link oder einen Suchbegriff ein.**

### `MaxLength` is a second line of defence, not the fix

Setting `MaxLength` on the option makes Discord reject an absurd value in the client, which is a better experience than a doomed round trip. It does not replace bounding what a reply quotes: the clamp in `NoResultsError.UserMessage` must hold regardless of what the client enforces, since the option limit is published asynchronously with the command set and an older client can still send a longer value. 1000 characters is comfortably above any real link or search phrase.

That clamp's comment currently justifies itself with "/play sets no maximum length", which this change makes false. The comment is rewritten to say the clamp does not depend on the client-side limit; the clamp itself stays.

### The fuzz target grows a normalisation half

`FuzzResolveIdentifier` keeps its current invariants unchanged. A second target covers the function that is actually new:

- `normalizeIdentifier` is idempotent.
- A normalised value never has leading or trailing whitespace.
- `resolveIdentifier(normalizeIdentifier(x))` never panics, and never returns the empty string when the normalised value is non-empty.

## Risks / Trade-offs

- Trimming changes what a member searched for, so a phrase with deliberate leading whitespace searches differently → whitespace-only differences do not change a search result in practice, and the error message now quotes the normalised value, so what the bot searched for is what the member is shown.
- Angle-bracket stripping could eat a genuine search phrase wrapped in brackets → only matched leading-and-trailing pairs are stripped, which is not a shape a member types by accident, and a phrase that loses them searches better than one that keeps them.
- One existing test case inverts, which reads as a regression in the diff → it is rewritten with a name stating the new expectation, and the seed corpora gain the new shapes.
- Normalisation living in `Enqueue` means the unit tests for `isURL` and the behaviour a member sees no longer line up case for case, which can read as an incomplete fix → the padded-value case is renamed to say `isURL` does not trim, and the member-visible behaviour is covered by service-level tests and by the spec's scenarios instead.

## Migration Plan

Deploy is a restart. `MaxLength` is published with the command set on the next sync. Values that used to fail as searches now succeed as links; no value that used to work stops working. Rollback is a redeploy of the previous image.
