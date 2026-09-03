package music

import (
	"context"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

// Lavalink is the slice of the Lavalink client the service uses. It is declared
// here, by the consumer, so a fake stays small and an adapter in internal/app
// wraps the real client.
type Lavalink interface {
	Player(guildID snowflake.ID) Player
	// ExistingPlayer returns nil when the guild has no player. Implementations
	// must return an untyped nil, never an interface holding a nil pointer.
	ExistingPlayer(guildID snowflake.ID) Player
	BestNode() Node
}

// Player is the slice of a Lavalink player the service uses.
type Player interface {
	Update(ctx context.Context, opts ...lavalink.PlayerUpdateOpt) error
	Track() *lavalink.Track
	Paused() bool
	Position() lavalink.Duration
}

// Node is the slice of a Lavalink node the service uses.
type Node interface {
	LoadTracks(ctx context.Context, identifier string) (*lavalink.LoadResult, error)
}

// Voice is the slice of the Discord gateway the service uses to join and leave
// voice channels. A nil channelID leaves the current channel.
type Voice interface {
	UpdateVoiceState(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, selfMute bool, selfDeaf bool) error
}
