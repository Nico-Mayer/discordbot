package app

import (
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/nico-mayer/discordbot/internal/music"
)

var _ music.Lavalink = (*lavalinkAdapter)(nil)

// lavalinkAdapter narrows the real disgolink client down to the seams that
// internal/music declares.
type lavalinkAdapter struct {
	client disgolink.Client
}

func newLavalinkAdapter(client disgolink.Client) *lavalinkAdapter {
	return &lavalinkAdapter{client: client}
}

func (a *lavalinkAdapter) Player(guildID snowflake.ID) music.Player {
	return a.client.Player(guildID)
}

// ExistingPlayer returns an untyped nil when the guild has no player. The
// disgolink value is returned as-is rather than wrapped, because wrapping a nil
// player in a struct pointer would yield an interface that compares != nil and
// would silently disable every `player == nil` check in the service.
func (a *lavalinkAdapter) ExistingPlayer(guildID snowflake.ID) music.Player {
	player := a.client.ExistingPlayer(guildID)
	if player == nil {
		return nil
	}
	return player
}

// BestNode returns an untyped nil when no node is registered, for the same reason.
func (a *lavalinkAdapter) BestNode() music.Node {
	node := a.client.BestNode()
	if node == nil {
		return nil
	}
	return node
}
