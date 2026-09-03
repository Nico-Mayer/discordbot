## Context

See `proposal.md` - Why.

Constraints established by reading the current code and the pinned `disgo v0.19.6`:

- The bot requests `IntentGuilds` and `IntentGuildVoiceStates`, and caches only `cache.FlagVoiceStates`.
- `Caches.AudioChannelMembers(channel)` looks like the right API but its implementation calls `c.Member(...)`, and the member cache is not enabled. It would return an empty slice. **Unusable here.**
- `Caches.VoiceStates(guildID) iter.Seq[discord.VoiceState]` *is* populated and is enough to count occupants.
- `discord.VoiceState` carries `GuildID`, `ChannelID *snowflake.ID`, `UserID`, `SessionID`, and mute/deaf flags. It has **no member and no bot flag**, so occupants cannot be classified.
- Today `OnTrackEnd` calls `UpdateVoiceState(..., nil, ...)` directly when the queue is exhausted. That is the line the queue-dry trigger replaces.

## Goals / Non-Goals

**Goals:**

- Two independent idle conditions that cannot disagree, because both resolve to the same idempotent action.
- No new gateway intents. Nothing here should require a privileged intent or a Discord approval step.
- Both timers must be provably dead after shutdown.

**Non-Goals:**

- Distinguishing bots from humans, which the available data does not support.
- Reacting to self-deaf or server-deaf as a form of not listening. A deafened user is still present and may undeafen; treating them as absent would make the bot leave on someone muting themselves.

## Decisions

### Occupancy is counted from voice states, excluding only our own ID

```go
func (s *Service) listeners(channelID snowflake.ID) int {
    var n int
    for st := range s.caches.VoiceStates(s.guildID) {
        if st.ChannelID != nil && *st.ChannelID == channelID && st.UserID != s.applicationID {
            n++
        }
    }
    return n
}
```

This works with the intents and cache flags already configured. Nothing needs to change in the gateway setup, which is the main reason to prefer it.

**Accepted limitation:** another bot sitting in the channel counts as a listener, so the bot would not leave. Correcting this needs the member cache, which needs the privileged `IntentGuildMembers`, which needs Discord approval for a bot that has no other use for it. On a private server the limitation is theoretical. It is written down here so it is not rediscovered as a bug.

### Two timers, two timeouts, one idempotent action

An earlier draft of this design used a single timer shared by both triggers, on the grounds that two timers raise the question of which one wins. That reasoning only held while both triggers shared one timeout. With independent timeouts it is wrong, and the objection dissolves once the action is seen for what it is: **leaving is idempotent**. There is no contest to resolve. Each condition independently says "I have been true for my configured patience, so leave", and whichever says it first is simply right.

```
   listeners == 0 ---------> [ alone timer: IDLE_ALONE_SECONDS ]-------+
                                                                       |
                                                                       +--> leave:
                                                                       |      cancel both
   nothing playing --------> [ empty timer: IDLE_EMPTY_QUEUE_SECONDS ]-+      stop player
   AND queue empty                                                            clear queue
                                                                              leave voice
   listener rejoins -------> cancel alone timer
   track starts -----------> cancel empty timer
```

Both timers can be armed at once, which is the common case: the last track ends and everyone leaves. Whichever fires first performs the leave and cancels the other. The second firing, if it races in, finds the service already idle and returns.

```go
type Service struct {
    // ...
    idleMu   sync.Mutex
    aloneTmr *time.Timer
    emptyTmr *time.Timer
}
```

Each `arm` is idempotent: an already-running timer is left alone rather than restarted, so join/leave churn cannot extend a countdown. Each condition owns exactly one timer, so the two never need to compare durations or recompute a minimum.

*Alternative considered:* one timer whose duration is recomputed as the minimum of the currently-active conditions. Rejected: it produces the same observable behaviour as two timers while requiring the code to track which conditions are live and re-derive the deadline on every event. Two timers make the state directly readable.

### `0` means immediately, `off` means never

Splitting into two variables is not by itself enough to reproduce the current behaviour, and it is worth being precise about why. The bot today leaves the *instant* the queue is exhausted. If `0` meant "disabled", no setting would express instant-leave, and the split would not have delivered what it was meant to deliver.

So each variable takes one of these forms:

| Value | Meaning |
|---|---|
| unset | 60 seconds |
| `0` | leave immediately when the condition becomes true |
| positive integer | wait that many seconds |
| `off` | never leave for this reason |

`off` is a literal rather than a magic negative number. `-1` would work as a sentinel but requires the reader to know the convention, whereas `off` is self-describing in `.env.example` and cannot be misread as a duration.

`IDLE_EMPTY_QUEUE_SECONDS=0` therefore reproduces today's behaviour exactly, and every combination is now reachable:

```
  IDLE_ALONE_SECONDS=60   IDLE_EMPTY_QUEUE_SECONDS=0     <- old queue behaviour, new alone behaviour
  IDLE_ALONE_SECONDS=off  IDLE_EMPTY_QUEUE_SECONDS=300   <- never leave when alone, linger 5 min when dry
  IDLE_ALONE_SECONDS=off  IDLE_EMPTY_QUEUE_SECONDS=off   <- feature off
```

Validation follows the pattern the refactor establishes: parsed in `config.Load`, contributing to the joined error on failure. A value that is neither `off` nor a non-negative integer is an error rather than a silent fallback, because silently ignoring a misconfigured value is how the current config bugs happened.

A `0` timeout still goes through the timer rather than calling leave inline, so there is exactly one code path to the leave action. `time.AfterFunc` with a zero duration fires at the next scheduler opportunity, which is the desired semantics and keeps the cancel-if-the-condition-clears logic uniform.

### Where the triggers are evaluated

| Event | Check | Action |
|---|---|---|
| `OnVoiceStateUpdate`, any user in the guild | is the bot in a channel, and is its listener count now 0? | `armAlone()` |
| `OnVoiceStateUpdate`, any user in the guild | listener count now >= 1? | `cancelAlone()` |
| `OnVoiceStateUpdate`, bot's own state, `ChannelID == nil` | bot was disconnected | cancel both, clear queue |
| `OnTrackEnd`, queue exhausted | nothing left to play | `armEmpty()` instead of leaving immediately |
| `OnTrackStart` | playback resumed | `cancelEmpty()` |

The existing handler already filters `OnVoiceStateUpdate` to the bot's own user ID. That filter has to be **relaxed**, because the whole point is reacting to *other* users' movements. The guild guard from `refactor-bot-architecture` stays, and the bot's-own-state branch keeps its current behaviour.

This is the subtlety most likely to be got wrong: widening that filter without keeping the self branch intact would break voice server handoff.

### Shutdown stops both timers

`time.AfterFunc` holds a goroutine until it fires. Either pending timer at shutdown means a goroutine outliving the process, or a callback running against a closed client.

The service exposes `Close()` which cancels both timers, and the composition root registers it as a cleanup alongside the gateway and Lavalink teardown. The `goleak` check from `refactor-bot-architecture` is what verifies this, and it will fail loudly if the cleanup is forgotten. That is the intended relationship: the earlier change built the detector, this change is the first thing it usefully detects.

## Risks / Trade-offs

- **Queue-dry lingering is a behaviour change on upgrade** → An existing deployment gets a 60 second linger it did not have. Mitigation: it is the documented default, and `IDLE_EMPTY_QUEUE_SECONDS=0` restores the old instant-leave for that case alone, independently of the nobody-listening timeout.
- **Two variables and a three-form value grammar is more configuration surface than one variable** → Accepted, and the reason to accept it is that one variable could not express the combination an existing deployment actually wants. The cost is a longer `.env.example` and one more validation path.
- **Both timers armed at once could double-leave** → The leave action cancels both timers and re-checks state under the mutex, so a second callback returns early. Worth an explicit test, since it is the common case rather than an edge case.
- **Another bot in the channel prevents leaving** → Accepted, documented above. Needs a privileged intent to fix.
- **A timer callback racing shutdown could act on a closed client** → The callback re-checks the idle condition under the mutex and returns early if the service is closed, rather than assuming the state it was armed with still holds.
- **Rapid join/leave churn** → `arm()` being idempotent means churn cannot extend the countdown, but it also means a user who leaves and rejoins within the window does not reset it: the bot may leave shortly after they return. Acceptable at a 60 second timeout; would need a rethink at 10 seconds.
- **Discord voice state updates can arrive out of order or be missed on reconnect** → The count is recomputed from the cache on every event rather than incrementally tracked, so a missed event self-corrects on the next one. The failure mode is a delayed leave, not a wrong one.

## Open Questions

- Should the idle timeout apply while the player is explicitly *paused* with listeners present? Current design says no, since a paused player with people in the channel is a deliberate state. Worth confirming after using it.
