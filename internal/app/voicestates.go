package app

import (
	"iter"

	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/snowflake/v2"

	"github.com/nico-mayer/discordbot/internal/music"
)

var _ music.VoiceStates = (*voiceStateAdapter)(nil)

// voiceStateAdapter narrows the disgo cache down to the seam internal/music
// declares. Caches.AudioChannelMembers would read more naturally but resolves
// through the member cache, which needs a privileged intent the bot does not
// request, so it would always come back empty.
type voiceStateAdapter struct {
	caches cache.Caches
}

func newVoiceStateAdapter(caches cache.Caches) *voiceStateAdapter {
	return &voiceStateAdapter{caches: caches}
}

func (a *voiceStateAdapter) VoiceStates(guildID snowflake.ID) iter.Seq[music.VoiceState] {
	return func(yield func(music.VoiceState) bool) {
		for state := range a.caches.VoiceStates(guildID) {
			if !yield(music.VoiceState{UserID: state.UserID, ChannelID: state.ChannelID}) {
				return
			}
		}
	}
}
