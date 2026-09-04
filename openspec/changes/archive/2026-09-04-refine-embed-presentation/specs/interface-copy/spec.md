## ADDED Requirements

### Requirement: Text states the outcome without help from an icon

Every user-facing reply SHALL state its outcome in words. An icon MUST NOT be the only carrier of meaning, so a reply still reads correctly to a screen reader, in a client that fails to render the icon, and where the icon is unfamiliar to the reader.

How many icons a reply may carry and where they are placed is defined by `embed-presentation`. This requirement governs only the words: the copy MUST be written so that it survives every icon being removed.

#### Scenario: The queue is empty

- **WHEN** a member runs `/queue` while the queue holds no tracks
- **THEN** the text MUST state that the queue is empty without relying on the icon

#### Scenario: A skip is confirmed

- **WHEN** the bot confirms a skip
- **THEN** the text alone MUST say that the track was skipped

#### Scenario: Every icon is removed

- **WHEN** every icon is removed from any reply the bot can send
- **THEN** the remaining text MUST still state the outcome unambiguously

## REMOVED Requirements

### Requirement: A reply carries exactly one status icon

**Reason**: The requirement mixed a copy rule with a layout rule. Icon count, placement and inventory are layout, and are now owned by `embed-presentation`, which also drops the icon from track cards and list replies - so "exactly one" is no longer true of every reply. The copy half of it, that text must state the outcome without an icon, is kept as the new requirement above.

**Migration**: Read `embed-presentation` for where an icon may appear and which icons exist. No user-facing string changes as a result of the removal itself.
