## Why

`/play` decides between "load this link" and "search for this phrase" with two literal prefix checks: `strings.HasPrefix(identifier, "http://")` and `"https://"` (`internal/music/identifier.go:12`). Anything that is not byte-for-byte one of those is searched for as a phrase, so four ordinary ways of supplying a link silently turn into a doomed search:

- A pasted link with a leading or trailing space or newline. `identifier_test.go:26` pins this as current behaviour.
- A link wrapped in angle brackets. The bot's own replies render `[Titel](<url>)`, so copying a link out of a bot reply produces exactly this.
- A scheme the phone's keyboard capitalised, `HTTPS://…` (`identifier_test.go:27`).
- A value that is only whitespace, which passes Discord's required-option check and is then searched for, producing a "nothing found" reply quoting a blank.

Each failure looks identical to the member: a search that found nothing, for a link they can see is correct.

## What Changes

- Normalise the `/play` value before deciding what it is: trim surrounding whitespace, strip a single pair of wrapping angle brackets, and match the scheme case-insensitively.
- Load the normalised value rather than the raw one, so a trimmed link is what reaches the audio node.
- Reject a value that is empty once normalised, with its own message telling the member what to supply, instead of searching for nothing.
- Set a maximum length on the option so Discord rejects an absurd value at the client rather than the bot searching for it.
- Update the four test cases that currently assert the old behaviour, and extend the fuzz target's invariant to hold for the normalised value.

## Capabilities

### Modified Capabilities

- `music-playback`: the `/play` value is normalised before it is classified as a link or a search phrase, a value that is empty after normalisation is rejected, and the option has a maximum length.

## Impact

- `internal/music/identifier.go`: `isURL` and `resolveIdentifier`.
- `internal/music/service.go`: `loadTrack` loads the normalised value; `Enqueue` rejects an empty one.
- `internal/music/errors.go` and the copy file added by `improve-copy-and-logs`: one new sentinel and its German message.
- `internal/music/commands.go`: `MaxLength` on the option.
- `internal/music/identifier_test.go`: four table cases invert; the fuzz invariant changes.
- Depends on `improve-copy-and-logs` for where the new German message lives. Sequence it second, or put the message inline and move it when the two meet.
