## Why

`/play` decides between "load this link" and "search for this phrase" with two literal prefix checks: `strings.HasPrefix(identifier, "http://")` and `"https://"` (`internal/music/identifier.go:12`). Anything that is not byte-for-byte one of those is searched for as a phrase, so four ordinary ways of supplying a link silently turn into a doomed search:

- A pasted link with a leading or trailing space or newline. `identifier_test.go:27` pins this as current behaviour.
- A link wrapped in angle brackets. The bot's own replies render `[Titel](<url>)`, so copying a link out of a bot reply produces exactly this.
- A scheme the phone's keyboard capitalised, `HTTPS://…` (`identifier_test.go:26`).
- A value that is only whitespace, which passes Discord's required-option check and is then searched for, producing a "nothing found" reply quoting a blank.

Each failure looks identical to the member: a search that found nothing, for a link they can see is correct.

## What Changes

- Normalise the `/play` value before deciding what it is: trim surrounding whitespace, strip wrapping angle brackets, and match the scheme case-insensitively.
- Load the normalised value rather than the raw one, so a trimmed link is what reaches the audio node.
- Reject a value that is empty once normalised, with its own message telling the member what to supply, instead of searching for nothing.
- Set a maximum length on the option so Discord rejects an absurd value at the client rather than the bot searching for it.
- Invert the single test case that pins an uppercase scheme as a search, keep the padded-value case as a statement that `isURL` itself does not trim, and add fuzz coverage for the new normalisation function.

## Capabilities

### Modified Capabilities

- `music-playback`: the `/play` value is normalised before it is classified as a link or a search phrase, a value that is empty after normalisation is rejected, and the option has a maximum length.

## Impact

- `internal/music/identifier.go`: `isURL` and `resolveIdentifier`.
- `internal/music/service.go`: `Enqueue` normalises once and rejects an empty value; `loadTrack` only ever sees a normalised one.
- `internal/music/errors.go` and `internal/music/copy.go`: one new sentinel and its German message.
- `internal/music/commands.go`: `MaxLength` on the option.
- `internal/music/identifier_test.go`: one table case inverts; the fuzz target gains normalisation properties.
- `improve-copy-and-logs` has landed, so the option is already `titel` and the German strings already live in `copy.go`.
