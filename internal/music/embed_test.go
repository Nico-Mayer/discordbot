package music

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/stretchr/testify/require"
)

func testTrack(title string) lavalink.Track {
	uri := "https://example.com/" + title
	artwork := "https://example.com/art/" + title
	return lavalink.Track{
		Info: lavalink.TrackInfo{
			Identifier: title + "-id",
			Author:     title + "-author",
			Title:      title,
			Length:     3*lavalink.Minute + 7*lavalink.Second,
			URI:        &uri,
			ArtworkURL: &artwork,
		},
	}
}

func testTrackWithoutURI(title string) lavalink.Track {
	track := testTrack(title)
	track.Info.URI = nil
	return track
}

func testTrackWithoutArtwork(title string) lavalink.Track {
	track := testTrack(title)
	track.Info.ArtworkURL = nil
	return track
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   lavalink.Duration
		want string
	}{
		{name: "zero", in: 0, want: "0:00"},
		{name: "negative", in: -1 * lavalink.Second, want: "0:00"},
		{name: "sub second", in: 999 * lavalink.Millisecond, want: "0:00"},
		{name: "45 seconds", in: 45 * lavalink.Second, want: "0:45"},
		{name: "one minute", in: lavalink.Minute, want: "1:00"},
		{name: "3m07s", in: 3*lavalink.Minute + 7*lavalink.Second, want: "3:07"},
		{name: "59m59s", in: 59*lavalink.Minute + 59*lavalink.Second, want: "59:59"},
		{name: "exactly one hour", in: lavalink.Hour, want: "1:00:00"},
		{name: "1h05m30s", in: lavalink.Hour + 5*lavalink.Minute + 30*lavalink.Second, want: "1:05:30"},
		{name: "over a day", in: 25*lavalink.Hour + 2*lavalink.Minute + 3*lavalink.Second, want: "25:02:03"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, formatDuration(tt.in))
		})
	}
}

func TestFormatDurationNeverRendersMinutesBeyondSixty(t *testing.T) {
	t.Parallel()

	got := formatDuration(lavalink.Hour + 5*lavalink.Minute + 30*lavalink.Second)
	require.Equal(t, "1:05:30", got)
	require.NotEqual(t, "65:30", got)
}

var durationPattern = regexp.MustCompile(`^(\d+:)?\d+:\d{2}$`)

func FuzzFormatDuration(f *testing.F) {
	for _, seed := range []int64{
		0,
		-1,
		int64(45 * lavalink.Second),
		int64(3*lavalink.Minute + 7*lavalink.Second),
		int64(lavalink.Hour + 5*lavalink.Minute + 30*lavalink.Second),
		1,
		-9223372036854775808,
		9223372036854775807,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, ms int64) {
		got := formatDuration(lavalink.Duration(ms))
		require.NotEmpty(t, got)
		require.Regexp(t, durationPattern, got)
	})
}

func TestErrorEmbed(t *testing.T) {
	t.Parallel()

	embed := errorEmbed("kaputt")
	require.Equal(t, colorError, embed.Color)
	require.Contains(t, embed.Description, "kaputt")
	require.Contains(t, embed.Description, "❌")
}

func TestSuccessEmbed(t *testing.T) {
	t.Parallel()

	embed := successEmbed("erledigt")
	require.Equal(t, colorSuccess, embed.Color)
	require.Contains(t, embed.Description, "erledigt")
}

func TestInfoEmbed(t *testing.T) {
	t.Parallel()

	embed := infoEmbed("hinweis")
	require.Equal(t, colorInfo, embed.Color)
	require.Contains(t, embed.Description, "hinweis")
}

func TestNowPlayingEmbed(t *testing.T) {
	t.Parallel()

	embed := nowPlayingEmbed(testTrack("song"), 30*lavalink.Second)

	require.Equal(t, "🎶 Läuft gerade", embed.Title)
	require.Equal(t, colorInfo, embed.Color)
	require.Contains(t, embed.Description, "song")
	require.Contains(t, embed.Description, "song-author")
	require.Contains(t, embed.Description, "https://example.com/song")
	require.NotNil(t, embed.Footer)
	require.Equal(t, "0:30 / 3:07", embed.Footer.Text)
	require.NotNil(t, embed.Thumbnail)
	require.Equal(t, "https://example.com/art/song", embed.Thumbnail.URL)
}

func TestTrackEmbed(t *testing.T) {
	t.Parallel()

	embed := trackEmbed(testTrack("song"))

	require.Equal(t, "song", embed.Title)
	require.Equal(t, colorSuccess, embed.Color)
	require.Equal(t, "https://example.com/song", embed.URL)
	require.Contains(t, embed.Description, "song-author")
	require.NotNil(t, embed.Image)
	require.Equal(t, "https://example.com/art/song", embed.Image.URL)

	require.Len(t, embed.Fields, 1)
	require.NotEmpty(t, embed.Fields[0].Name, "duration field must have a name")
	require.NotEmpty(t, embed.Fields[0].Value, "duration field must have a value")
	require.Equal(t, "3:07", embed.Fields[0].Value)
}

func TestQueuedEmbed(t *testing.T) {
	t.Parallel()

	embed := queuedEmbed(testTrack("song"), 4)

	require.Equal(t, "📋 Zur Warteschlange hinzugefügt", embed.Title)
	require.Equal(t, colorInfo, embed.Color)
	require.Len(t, embed.Fields, 2)
	require.Equal(t, "#4", embed.Fields[0].Value)
	require.Equal(t, "3:07", embed.Fields[1].Value)
	for _, field := range embed.Fields {
		require.NotEmpty(t, field.Name)
		require.NotEmpty(t, field.Value)
	}
}

func TestQueueEmbedEmpty(t *testing.T) {
	t.Parallel()

	embed := queueEmbed(nil)
	require.Equal(t, colorInfo, embed.Color)
	require.Contains(t, embed.Description, "leer")
	require.Nil(t, embed.Footer)
}

func TestQueueEmbedListsTracks(t *testing.T) {
	t.Parallel()

	tracks := []lavalink.Track{testTrack("first"), testTrack("second")}
	embed := queueEmbed(tracks)

	require.Equal(t, "📋 Warteschlange", embed.Title)
	require.Equal(t, colorInfo, embed.Color)
	require.Contains(t, embed.Description, "`1.`")
	require.Contains(t, embed.Description, "first")
	require.Contains(t, embed.Description, "`2.`")
	require.Contains(t, embed.Description, "second")
	require.Contains(t, embed.Description, "3:07")
	require.NotContains(t, embed.Description, "weitere")
	require.NotNil(t, embed.Footer)
	require.Equal(t, "2 Titel in der Warteschlange", embed.Footer.Text)
}

func TestQueueEmbedIsBounded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		title        func(int) string
		wantListed   int
		wantResidual int
	}{
		{
			name:         "200 ordinary tracks are capped at the listing limit",
			title:        func(i int) string { return fmt.Sprintf("track-%03d", i) },
			wantListed:   queueListLimit,
			wantResidual: 200 - queueListLimit,
		},
		{
			name:  "pathologically long titles are cut before the description limit",
			title: func(i int) string { return strings.Repeat("x", 5000) + fmt.Sprintf("%03d", i) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tracks := make([]lavalink.Track, 0, 200)
			for i := range 200 {
				tracks = append(tracks, testTrack(tt.title(i)))
			}

			embed := queueEmbed(tracks)

			require.LessOrEqual(t, len(embed.Description), embedDescriptionLimit)
			require.Equal(t, "200 Titel in der Warteschlange", embed.Footer.Text)

			listed := strings.Count(embed.Description, "`") / 2
			if tt.wantListed > 0 {
				require.Equal(t, tt.wantListed, listed)
				require.Contains(t, embed.Description, fmt.Sprintf("… und %d weitere", tt.wantResidual))
			} else {
				require.Contains(t, embed.Description, "weitere")
			}
		})
	}
}

func TestTrackLinkHandlesNilURI(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		require.Equal(t, "**song**", trackLink(testTrackWithoutURI("song")))
	})
	require.Equal(t, "[song](<https://example.com/song>)", trackLink(testTrack("song")))
}

func TestEmbedsDoNotPanicOnNilURI(t *testing.T) {
	t.Parallel()

	track := testTrackWithoutURI("song")

	require.NotPanics(t, func() {
		require.Empty(t, trackEmbed(track).URL)
		require.Contains(t, nowPlayingEmbed(track, 0).Description, "**song**")
		require.Contains(t, queuedEmbed(track, 1).Description, "**song**")
		require.Contains(t, queueEmbed([]lavalink.Track{track}).Description, "**song**")
	})
}

func TestArtworkURLFallsBackToYouTubeThumbnail(t *testing.T) {
	t.Parallel()

	require.Equal(t, "https://example.com/art/song", artworkURL(testTrack("song")))
	require.Equal(t, "https://img.youtube.com/vi/song-id/hqdefault.jpg", artworkURL(testTrackWithoutArtwork("song")))
}

func TestEmbedsFallBackToYouTubeThumbnail(t *testing.T) {
	t.Parallel()

	track := testTrackWithoutArtwork("song")
	const want = "https://img.youtube.com/vi/song-id/hqdefault.jpg"

	require.Equal(t, want, trackEmbed(track).Image.URL)
	require.Equal(t, want, nowPlayingEmbed(track, 0).Thumbnail.URL)
	require.Equal(t, want, queuedEmbed(track, 1).Thumbnail.URL)
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	require.Equal(t, "abc", truncate("abc", 3))
	require.Equal(t, "a…", truncate("abcde", 4))
	// Cutting inside a multi-byte rune must not produce invalid UTF-8.
	require.Equal(t, "…", truncate(strings.Repeat("ä", 10), 4))
}
