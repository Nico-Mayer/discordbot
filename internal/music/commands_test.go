package music

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/stretchr/testify/require"
)

func TestCommandDescriptionsComeFromTheCopy(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"play":        descPlay,
		"pause":       descPause,
		"stop":        descStop,
		"skip":        descSkip,
		"now-playing": descNowPlaying,
		"queue":       descQueue,
	}
	require.Len(t, Commands, len(want))

	for _, command := range Commands {
		slash, ok := command.(discord.SlashCommandCreate)
		require.True(t, ok)

		t.Run(slash.Name, func(t *testing.T) {
			t.Parallel()

			description, known := want[slash.Name]
			require.True(t, known, "%s has no entry in the copy", slash.Name)
			require.NotEmpty(t, description)
			require.Equal(t, description, slash.Description)
		})
	}
}

func TestPlayOptionIsLabelledForTheReader(t *testing.T) {
	t.Parallel()

	play, ok := Commands[0].(discord.SlashCommandCreate)
	require.True(t, ok)
	require.Equal(t, "play", play.Name, "the command name stays English")
	require.Len(t, play.Options, 1)

	option, ok := play.Options[0].(discord.ApplicationCommandOptionString)
	require.True(t, ok)
	require.Equal(t, "titel", option.Name)
	require.Equal(t, "Link oder Suchbegriff", option.Description, "the label alone does not say a link is accepted")
	require.True(t, option.Required)
}
