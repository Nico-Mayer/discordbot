package app

import (
	"testing"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"
)

const testGuildID = snowflake.ID(111111111111111111)

// fakeDisgolinkClient overrides only the methods the adapter uses. The embedded
// nil interface makes any other call panic loudly rather than pass silently.
type fakeDisgolinkClient struct {
	disgolink.Client

	player         disgolink.Player
	existingPlayer disgolink.Player
	node           disgolink.Node
}

func (f *fakeDisgolinkClient) Player(snowflake.ID) disgolink.Player {
	return f.player
}

func (f *fakeDisgolinkClient) ExistingPlayer(snowflake.ID) disgolink.Player {
	return f.existingPlayer
}

func (f *fakeDisgolinkClient) BestNode() disgolink.Node {
	return f.node
}

// fakeDisgolinkPlayer is a non-nil disgolink.Player for the populated cases.
type fakeDisgolinkPlayer struct {
	disgolink.Player
}

type fakeDisgolinkNode struct {
	disgolink.Node
}

func TestExistingPlayerReturnsAnUntypedNil(t *testing.T) {
	t.Parallel()

	adapter := newLavalinkAdapter(&fakeDisgolinkClient{existingPlayer: nil})

	got := adapter.ExistingPlayer(testGuildID)

	// This is the check the service relies on. A non-nil interface holding a nil
	// pointer would pass `got != nil` and make every no-player path panic.
	require.Nil(t, got)
	require.True(t, got == nil, "ExistingPlayer must return an untyped nil interface")
}

func TestExistingPlayerPassesThroughARealPlayer(t *testing.T) {
	t.Parallel()

	player := &fakeDisgolinkPlayer{}
	adapter := newLavalinkAdapter(&fakeDisgolinkClient{existingPlayer: player})

	got := adapter.ExistingPlayer(testGuildID)
	require.NotNil(t, got)
	require.False(t, got == nil)
}

func TestBestNodeReturnsAnUntypedNil(t *testing.T) {
	t.Parallel()

	adapter := newLavalinkAdapter(&fakeDisgolinkClient{node: nil})

	got := adapter.BestNode()
	require.True(t, got == nil, "BestNode must return an untyped nil interface")
}

func TestBestNodePassesThroughARealNode(t *testing.T) {
	t.Parallel()

	adapter := newLavalinkAdapter(&fakeDisgolinkClient{node: &fakeDisgolinkNode{}})

	require.False(t, adapter.BestNode() == nil)
}

func TestPlayerPassesThroughTheClientsPlayer(t *testing.T) {
	t.Parallel()

	player := &fakeDisgolinkPlayer{}
	adapter := newLavalinkAdapter(&fakeDisgolinkClient{player: player})

	require.NotNil(t, adapter.Player(testGuildID))
}
