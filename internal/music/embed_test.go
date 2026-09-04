package music

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/disgoorg/disgo/discord"
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

func testTrackFromSource(title, source string) lavalink.Track {
	track := testTrack(title)
	track.Info.SourceName = source
	return track
}

func testStream(title string) lavalink.Track {
	track := testTrack(title)
	track.Info.IsStream = true
	return track
}

// allIcons is every status icon a reply can carry. No icon is a substring of
// another, so counting occurrences counts icons.
var allIcons = []string{
	iconError, iconInfo,
	iconPlaying, iconPaused, iconStopped, iconSkipped, iconQueued,
}

func countIcons(s string) int {
	var n int
	for _, icon := range allIcons {
		n += strings.Count(s, icon)
	}
	return n
}

func stripIcons(s string) string {
	for _, icon := range allIcons {
		s = strings.ReplaceAll(s, icon, "")
	}
	return strings.TrimSpace(s)
}

// embedIcons counts the icons across every part of an embed a reader sees.
func embedIcons(embed discord.Embed) int {
	n := countIcons(embed.Title) + countIcons(embed.Description)
	if embed.Author != nil {
		n += countIcons(embed.Author.Name)
	}
	if embed.Footer != nil {
		n += countIcons(embed.Footer.Text)
	}
	for _, field := range embed.Fields {
		n += countIcons(field.Name) + countIcons(field.Value)
	}
	return n
}

// iconsOutsideTheAuthorLine counts the icons in the parts of an embed that must
// never carry one, whatever the reply is.
func iconsOutsideTheAuthorLine(embed discord.Embed) int {
	n := countIcons(embed.Title)
	if embed.Footer != nil {
		n += countIcons(embed.Footer.Text)
	}
	for _, field := range embed.Fields {
		n += countIcons(field.Name) + countIcons(field.Value)
	}
	return n
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

func TestClamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		limit int
		want  string
	}{
		{name: "under the limit", in: "abc", limit: 5, want: "abc"},
		{name: "exactly at the limit", in: "abc", limit: 3, want: "abc"},
		{name: "over the limit", in: "abcde", limit: 4, want: "abc…"},
		{name: "german stays counted in characters", in: "übersprungen", limit: 6, want: "übers…"},
		{name: "cjk stays counted in characters", in: "曲名がとても長い", limit: 4, want: "曲名が…"},
		{name: "limit of one is the marker alone", in: "abc", limit: 1, want: "…"},
		{name: "limit of zero yields nothing", in: "abc", limit: 0, want: ""},
		{name: "negative limit yields nothing", in: "abc", limit: -1, want: ""},
		{name: "empty input", in: "", limit: 5, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := clamp(tt.in, tt.limit)
			require.Equal(t, tt.want, got)
			require.True(t, utf8.ValidString(got), "clamping must not cut inside a rune")
			require.LessOrEqual(t, utf8.RuneCountInString(got), max(tt.limit, 0))
		})
	}
}

// TestClampCountsCharactersNotBytes is the regression: the byte-based predecessor
// cut a multi-byte value at a fraction of the allowance it actually had.
func TestClampCountsCharactersNotBytes(t *testing.T) {
	t.Parallel()

	in := strings.Repeat("ä", 100)
	got := clamp(in, 100)
	require.Equal(t, in, got, "100 characters fit a 100 character limit")
	require.Len(t, got, 200, "the same value is 200 bytes long")
}

// requireBounded asserts an embed is within every limit Discord enforces.
func requireBounded(t *testing.T, embed discord.Embed) {
	t.Helper()

	require.LessOrEqual(t, utf8.RuneCountInString(embed.Title), limitTitle)
	require.LessOrEqual(t, utf8.RuneCountInString(embed.Description), limitDescription)
	if embed.Author != nil {
		require.LessOrEqual(t, utf8.RuneCountInString(embed.Author.Name), limitAuthorName)
	}
	if embed.Footer != nil {
		require.LessOrEqual(t, utf8.RuneCountInString(embed.Footer.Text), limitFooter)
	}
	for _, field := range embed.Fields {
		require.LessOrEqual(t, utf8.RuneCountInString(field.Name), limitFieldName)
		require.LessOrEqual(t, utf8.RuneCountInString(field.Value), limitFieldValue)
	}
	require.LessOrEqual(t, embedLength(embed), limitTotal)
}

func TestBoundBringsEveryFieldWithinItsLimit(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", 9000)
	embed := bound(discord.Embed{
		Title:       long,
		Description: long,
		Author:      &discord.EmbedAuthor{Name: long},
		Footer:      &discord.EmbedFooter{Text: long},
		Fields: []discord.EmbedField{
			{Name: long, Value: long},
			{Name: long, Value: long},
		},
	})

	requireBounded(t, embed)
	require.Contains(t, embed.Title, "…", "shortening must be visible")
}

func TestBoundLeavesAFittingEmbedUntouched(t *testing.T) {
	t.Parallel()

	in := discord.Embed{
		Title:       "Warteschlange",
		Description: "kurz",
		Author:      &discord.EmbedAuthor{Name: "Läuft jetzt"},
		Footer:      &discord.EmbedFooter{Text: "YouTube"},
		Fields:      []discord.EmbedField{{Name: fieldDuration, Value: "3:07"}},
	}

	got := bound(in)

	require.Equal(t, in.Title, got.Title)
	require.Equal(t, in.Description, got.Description)
	require.Equal(t, in.Author.Name, got.Author.Name)
	require.Equal(t, in.Footer.Text, got.Footer.Text)
	require.Equal(t, in.Fields, got.Fields)
	require.NotContains(t, got.Description, "…")
}

func TestBoundDoesNotMutateItsInput(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", 9000)
	in := discord.Embed{
		Author: &discord.EmbedAuthor{Name: long},
		Footer: &discord.EmbedFooter{Text: long},
		Fields: []discord.EmbedField{{Name: long, Value: long}},
	}

	bound(in)

	require.Len(t, in.Author.Name, 9000)
	require.Len(t, in.Footer.Text, 9000)
	require.Len(t, in.Fields[0].Value, 9000)
}

// adversarialTrack is metadata long enough to breach every embed limit at once.
func adversarialTrack() lavalink.Track {
	long := strings.Repeat("ä", 5000)
	uri := "https://example.com/" + long
	artwork := "https://example.com/art/" + long
	return lavalink.Track{
		Info: lavalink.TrackInfo{
			Identifier: long,
			Author:     long,
			Title:      long,
			Length:     3*lavalink.Minute + 7*lavalink.Second,
			SourceName: long,
			URI:        &uri,
			ArtworkURL: &artwork,
		},
	}
}

// TestEveryBuilderIsBounded is what makes the limits enforceable: a builder that
// forgets to return through bound fails here rather than in production, where the
// reply Discord rejects is the one carrying the outcome.
func TestEveryBuilderIsBounded(t *testing.T) {
	t.Parallel()

	track := adversarialTrack()
	long := strings.Repeat("ä", 9000)
	queue := make([]lavalink.Track, 0, 200)
	for range 200 {
		queue = append(queue, track)
	}

	tests := []struct {
		name  string
		embed discord.Embed
	}{
		{name: "error", embed: errorEmbed(long)},
		{name: "info", embed: infoEmbed(long)},
		{name: "paused", embed: pausedEmbed()},
		{name: "resumed", embed: resumedEmbed()},
		{name: "stopped", embed: stoppedEmbed()},
		{name: "skipped", embed: skippedEmbed()},
		{name: "started", embed: startedEmbed(track)},
		{name: "queued", embed: queuedEmbed(track, 4)},
		{name: "now playing", embed: nowPlayingEmbed(track, 30*lavalink.Second)},
		{name: "now playing stream", embed: nowPlayingEmbed(testStream(long), 0)},
		{name: "queue listing", embed: queueEmbed(&track, queue)},
		{name: "queue listing without a current track", embed: queueEmbed(nil, queue)},
		{name: "queue empty", embed: queueEmbed(nil, nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireBounded(t, tt.embed)
		})
	}
}

func TestStatusEmbedLeadsWithTheIcon(t *testing.T) {
	t.Parallel()

	embed := statusEmbed(iconSkipped, colorSuccess, "erledigt")
	require.Equal(t, iconSkipped+" erledigt", embed.Description, "one space separates the icon from the text")
	require.Equal(t, colorSuccess, embed.Color)
	require.Equal(t, 1, countIcons(embed.Description))
}

func TestErrorEmbed(t *testing.T) {
	t.Parallel()

	embed := errorEmbed("kaputt")
	require.Equal(t, colorError, embed.Color)
	require.Contains(t, embed.Description, "kaputt")
	require.True(t, strings.HasPrefix(embed.Description, iconError))
}

func TestInfoEmbed(t *testing.T) {
	t.Parallel()

	embed := infoEmbed("hinweis")
	require.Equal(t, colorNeutral, embed.Color)
	require.Contains(t, embed.Description, "hinweis")
	require.True(t, strings.HasPrefix(embed.Description, iconInfo))
}

// TestRepliesStateTheirOutcomeWithoutTheirIcon strips the icons from every
// confirmation and checks the remaining text still says what happened, so a
// reply reads correctly where the icon does not render.
func TestRepliesStateTheirOutcomeWithoutTheirIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		embed    discord.Embed
		wantText string
		wantWord string
	}{
		{name: "pause", embed: pausedEmbed(), wantText: replyPaused, wantWord: "pausiert"},
		{name: "resume", embed: resumedEmbed(), wantText: replyResumed, wantWord: "fortgesetzt"},
		{name: "stop", embed: stoppedEmbed(), wantText: replyStopped, wantWord: "gestoppt"},
		{name: "skip", embed: skippedEmbed(), wantText: replySkipped, wantWord: "übersprungen"},
		{name: "empty queue", embed: queueEmbed(nil, nil), wantText: replyQueueEmpty, wantWord: "leer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, 1, embedIcons(tt.embed), "a status reply carries exactly one icon")

			text := stripIcons(tt.embed.Description)
			require.Equal(t, tt.wantText, text, "the copy itself must carry no icon")
			require.Contains(t, text, tt.wantWord, "the text alone must name the outcome")
		})
	}
}

// TestConfirmationsAreToldApartByIconAndColour is the point of keeping the
// transport icons: "Wiedergabe pausiert" and "Wiedergabe fortgesetzt" are two
// words apart in text, and a reader scanning a channel should not have to read
// them to tell which one happened.
func TestConfirmationsAreToldApartByIconAndColour(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		embed     discord.Embed
		wantIcon  string
		wantColor int
	}{
		{name: "pause", embed: pausedEmbed(), wantIcon: iconPaused, wantColor: colorPaused},
		{name: "resume", embed: resumedEmbed(), wantIcon: iconPlaying, wantColor: colorSuccess},
		{name: "stop", embed: stoppedEmbed(), wantIcon: iconStopped, wantColor: colorNeutral},
		{name: "skip", embed: skippedEmbed(), wantIcon: iconSkipped, wantColor: colorSuccess},
	}

	seen := map[string]string{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.True(t, strings.HasPrefix(tt.embed.Description, tt.wantIcon), "the icon names the state")
			require.Equal(t, tt.wantColor, tt.embed.Color)
			require.Equal(t, 1, embedIcons(tt.embed))
		})
		require.NotContains(t, seen, tt.wantIcon, "two states must not share an icon")
		seen[tt.wantIcon] = tt.name
	}

	require.NotEqual(t, pausedEmbed().Color, resumedEmbed().Color, "pause and resume must not look alike")
}

func TestSourceColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "youtube", in: "youtube", want: 0xFF0000},
		{name: "youtube music", in: "youtubemusic", want: 0xFF0000},
		{name: "spotify", in: "spotify", want: 0x1DB954},
		{name: "soundcloud", in: "soundcloud", want: 0xFF5500},
		{name: "mixed case", in: "SoundCloud", want: 0xFF5500},
		{name: "unknown source", in: "bandcamp-mirror", want: colorAccent},
		{name: "no source", in: "", want: colorAccent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, sourceColor(tt.in))
		})
	}
}

// TestCardsAreColouredByTheirSource is the grouping the colour buys: same
// service, same stripe; different service, different stripe.
func TestCardsAreColouredByTheirSource(t *testing.T) {
	t.Parallel()

	spotify := testTrackFromSource("song", "spotify")
	youtube := testTrackFromSource("song", "youtube")

	require.Equal(t, startedEmbed(spotify).Color, queuedEmbed(spotify, 2).Color, "one source, one colour")
	require.Equal(t, startedEmbed(spotify).Color, nowPlayingEmbed(spotify, 0).Color)
	require.NotEqual(t, startedEmbed(spotify).Color, startedEmbed(youtube).Color, "two sources, two colours")
	require.Equal(t, colorAccent, startedEmbed(testTrack("song")).Color, "an unknown source falls back")
}

// TestTheFailureColourIsReservedForFailures keeps the two colour axes from
// colliding: no reply that is not a failure may look like one.
func TestTheFailureColourIsReservedForFailures(t *testing.T) {
	t.Parallel()

	current := testTrack("current")
	for name, embed := range map[string]discord.Embed{
		"paused":      pausedEmbed(),
		"resumed":     resumedEmbed(),
		"stopped":     stoppedEmbed(),
		"skipped":     skippedEmbed(),
		"info":        infoEmbed("hinweis"),
		"queue empty": queueEmbed(nil, nil),
		"queue":       queueEmbed(&current, []lavalink.Track{current}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.NotEqual(t, colorError, embed.Color)
		})
	}

	for _, source := range slices.Sorted(maps.Keys(sourceColors)) {
		t.Run("card from "+source, func(t *testing.T) {
			t.Parallel()
			require.NotEqual(t, colorError, startedEmbed(testTrackFromSource("song", source)).Color)
		})
	}

	require.Equal(t, colorError, errorEmbed(msgGeneric).Color)
}

func TestTrackLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		track lavalink.Track
		want  string
	}{
		{name: "ordinary track", track: testTrack("song"), want: "3:07"},
		{name: "livestream", track: testStream("radio"), want: markerLive},
		{name: "zero length", track: func() lavalink.Track {
			track := testTrack("song")
			track.Info.Length = 0
			return track
		}(), want: "0:00"},
		{name: "livestream with a sentinel length", track: func() lavalink.Track {
			track := testStream("radio")
			track.Info.Length = lavalink.Duration(9223372036854775807)
			return track
		}(), want: markerLive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, trackLength(tt.track))
		})
	}
}

func TestArtworkURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		track   lavalink.Track
		want    string
		wantHas bool
	}{
		{
			name:    "the source reports artwork",
			track:   testTrack("song"),
			want:    "https://example.com/art/song",
			wantHas: true,
		},
		{
			name: "youtube without artwork falls back to the video thumbnail",
			track: func() lavalink.Track {
				track := testTrackWithoutArtwork("song")
				track.Info.SourceName = "youtube"
				return track
			}(),
			want:    "https://img.youtube.com/vi/song-id/hqdefault.jpg",
			wantHas: true,
		},
		{
			name:    "reported artwork wins over the fallback",
			track:   testTrackFromSource("song", "youtube"),
			want:    "https://example.com/art/song",
			wantHas: true,
		},
		{
			name: "youtube music without artwork falls back too",
			track: func() lavalink.Track {
				track := testTrackWithoutArtwork("song")
				track.Info.SourceName = "youtubemusic"
				return track
			}(),
			want:    "https://img.youtube.com/vi/song-id/hqdefault.jpg",
			wantHas: true,
		},
		{
			name: "spotify without artwork gets no thumbnail",
			track: func() lavalink.Track {
				track := testTrackWithoutArtwork("song")
				track.Info.SourceName = "spotify"
				return track
			}(),
			wantHas: false,
		},
		{
			name:    "unknown source without artwork gets no thumbnail",
			track:   testTrackWithoutArtwork("song"),
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := artworkURL(tt.track)
			require.Equal(t, tt.wantHas, ok)
			if tt.wantHas {
				require.Equal(t, tt.want, got)
				return
			}
			require.Empty(t, got, "no thumbnail must mean no URL, not a broken one")
		})
	}
}

// TestArtworkFallbackIsNotBuiltForForeignSources is the regression: the previous
// fallback built a YouTube URL out of any source's identifier.
func TestArtworkFallbackIsNotBuiltForForeignSources(t *testing.T) {
	t.Parallel()

	track := testTrackWithoutArtwork("song")
	track.Info.SourceName = "spotify"

	require.Nil(t, startedEmbed(track).Thumbnail)
	require.Nil(t, queuedEmbed(track, 1).Thumbnail)
	require.Nil(t, nowPlayingEmbed(track, 0).Thumbnail)
}

func TestSourceLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "youtube", in: "youtube", want: "YouTube"},
		{name: "youtube music", in: "youtubemusic", want: "YouTube Music"},
		{name: "spotify", in: "spotify", want: "Spotify"},
		{name: "soundcloud", in: "soundcloud", want: "SoundCloud"},
		{name: "mixed case", in: "SoundCloud", want: "SoundCloud"},
		{name: "unknown source keeps its name", in: "bandcamp-mirror", want: "bandcamp-mirror"},
		{name: "no source", in: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, sourceLabel(tt.in))
		})
	}
}

// progressBar returns the bar out of a progress line.
func progressBar(t *testing.T, line string) string {
	t.Helper()

	parts := strings.Fields(line)
	require.Len(t, parts, 3, "a progress line is elapsed, bar, total")
	return parts[1]
}

func TestProgressLine(t *testing.T) {
	t.Parallel()

	const length = 3*lavalink.Minute + 7*lavalink.Second

	tests := []struct {
		name         string
		position     lavalink.Duration
		length       lavalink.Duration
		wantElapsed  string
		wantTotal    string
		wantFilled   int
		wantRemained int
	}{
		{
			name: "just started", position: 0, length: length,
			wantElapsed: "0:00", wantTotal: "3:07", wantFilled: 0, wantRemained: progressBarWidth - 1,
		},
		{
			name: "halfway", position: length / 2, length: length,
			wantElapsed: "1:33", wantTotal: "3:07", wantFilled: 5, wantRemained: progressBarWidth - 6,
		},
		{
			name: "exactly at the length", position: length, length: length,
			wantElapsed: "3:07", wantTotal: "3:07", wantFilled: progressBarWidth - 1, wantRemained: 0,
		},
		{
			name: "past the length", position: 2 * length, length: length,
			wantElapsed: "6:14", wantTotal: "3:07", wantFilled: progressBarWidth - 1, wantRemained: 0,
		},
		{
			name: "negative position", position: -1 * lavalink.Second, length: length,
			wantElapsed: "0:00", wantTotal: "3:07", wantFilled: 0, wantRemained: progressBarWidth - 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			line := progressLine(tt.position, tt.length)
			require.True(t, strings.HasPrefix(line, tt.wantElapsed), "the elapsed time leads: %q", line)
			require.True(t, strings.HasSuffix(line, tt.wantTotal), "the total closes: %q", line)

			bar := progressBar(t, line)
			require.Equal(t, progressBarWidth, utf8.RuneCountInString(bar), "the bar has a fixed width: %q", bar)
			require.Equal(t, tt.wantFilled, strings.Count(bar, barFilled))
			require.Equal(t, tt.wantRemained, strings.Count(bar, barEmpty))
			require.Equal(t, 1, strings.Count(bar, barKnob))
		})
	}
}

func TestProgressLineOmitsTheBarWhenTheLengthIsUnknown(t *testing.T) {
	t.Parallel()

	line := progressLine(30*lavalink.Second, 0)

	require.Equal(t, "0:30 / 0:00", line)
	require.NotContains(t, line, barKnob, "a length of zero cannot place a knob")
	require.NotContains(t, line, barFilled)
}

func TestProgressLineSurvivesASentinelLength(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		line := progressLine(lavalink.Duration(9223372036854775807), lavalink.Duration(9223372036854775807))
		require.Equal(t, progressBarWidth, utf8.RuneCountInString(progressBar(t, line)))
	})
}

// TestTrackCardsShareOneShape is the structural contract: the three track replies
// differ in their author line and fields, never in their layout.
func TestTrackCardsShareOneShape(t *testing.T) {
	t.Parallel()

	track := testTrackFromSource("song", "youtube")
	tests := []struct {
		name  string
		embed discord.Embed
		state string
		icon  string
	}{
		{name: "started", embed: startedEmbed(track), state: authorNowPlaying, icon: iconPlaying},
		{name: "queued", embed: queuedEmbed(track, 4), state: titleQueued, icon: iconQueued},
		{name: "now playing", embed: nowPlayingEmbed(track, 30*lavalink.Second), state: titleNowPlaying, icon: iconPlaying},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, "song", tt.embed.Title, "the track title is the embed title")
			require.Equal(t, "https://example.com/song", tt.embed.URL, "the title links to the source")
			require.Equal(t, sourceColor("youtube"), tt.embed.Color, "the colour names the source")

			require.NotNil(t, tt.embed.Author, "the author line states the outcome")
			require.Equal(t, tt.icon+" "+tt.state, tt.embed.Author.Name, "the icon leads the state it marks")

			require.NotNil(t, tt.embed.Thumbnail, "artwork is a thumbnail")
			require.Equal(t, "https://example.com/art/song", tt.embed.Thumbnail.URL)
			require.Nil(t, tt.embed.Image, "a track card never uses a full-width image")

			require.Contains(t, tt.embed.Description, "song-author")
			require.Equal(t, 1, embedIcons(tt.embed), "a card carries one icon, in its author line")
			require.Zero(t, iconsOutsideTheAuthorLine(tt.embed), "no icon decorates the title, fields or footer")
		})
	}
}

func TestStartedEmbed(t *testing.T) {
	t.Parallel()

	embed := startedEmbed(testTrack("song"))

	require.Equal(t, iconPlaying+" "+authorNowPlaying, embed.Author.Name)
	require.Len(t, embed.Fields, 1)
	require.Equal(t, fieldDuration, embed.Fields[0].Name)
	require.Equal(t, "3:07", embed.Fields[0].Value)
}

func TestQueuedEmbed(t *testing.T) {
	t.Parallel()

	embed := queuedEmbed(testTrack("song"), 4)

	require.Equal(t, iconQueued+" "+titleQueued, embed.Author.Name)
	require.Len(t, embed.Fields, 2)
	require.Equal(t, fieldPosition, embed.Fields[0].Name)
	require.Equal(t, "#4", embed.Fields[0].Value)
	require.Equal(t, fieldDuration, embed.Fields[1].Name)
	require.Equal(t, "3:07", embed.Fields[1].Value)
	for _, field := range embed.Fields {
		require.NotEmpty(t, field.Name)
		require.NotEmpty(t, field.Value)
	}
}

func TestNowPlayingEmbedShowsProgress(t *testing.T) {
	t.Parallel()

	embed := nowPlayingEmbed(testTrack("song"), 30*lavalink.Second)

	require.Equal(t, iconPlaying+" "+titleNowPlaying, embed.Author.Name)
	require.Contains(t, embed.Description, progressLine(30*lavalink.Second, 3*lavalink.Minute+7*lavalink.Second))
	require.Contains(t, embed.Description, "0:30")
	require.Contains(t, embed.Description, "3:07")
	require.Empty(t, embed.Fields, "the progress line carries the times, so no duration field is needed")
}

func TestNowPlayingEmbedMarksALivestream(t *testing.T) {
	t.Parallel()

	embed := nowPlayingEmbed(testStream("radio"), 30*lavalink.Second)

	require.NotContains(t, embed.Description, barKnob, "a livestream has no progress to render")
	require.Len(t, embed.Fields, 1)
	require.Equal(t, fieldDuration, embed.Fields[0].Name)
	require.Equal(t, markerLive, embed.Fields[0].Value)
}

func TestTrackCardNamesItsSource(t *testing.T) {
	t.Parallel()

	embed := startedEmbed(testTrackFromSource("song", "spotify"))
	require.NotNil(t, embed.Footer)
	require.Equal(t, "Spotify", embed.Footer.Text)
}

func TestTrackCardOmitsAnUnknownSource(t *testing.T) {
	t.Parallel()

	require.Nil(t, startedEmbed(testTrack("song")).Footer, "an empty source leaves no footer behind")
}

func TestTrackCardsHandleAMissingURI(t *testing.T) {
	t.Parallel()

	track := testTrackWithoutURI("song")

	require.NotPanics(t, func() {
		for name, embed := range map[string]discord.Embed{
			"started":     startedEmbed(track),
			"queued":      queuedEmbed(track, 1),
			"now playing": nowPlayingEmbed(track, 0),
		} {
			require.Empty(t, embed.URL, name)
			require.Equal(t, "song", embed.Title, name)
		}
	})
}

func TestTrackCardsHandleMissingArtwork(t *testing.T) {
	t.Parallel()

	track := testTrackWithoutArtwork("song")

	require.Nil(t, startedEmbed(track).Thumbnail)
	require.Nil(t, queuedEmbed(track, 1).Thumbnail)
	require.Nil(t, nowPlayingEmbed(track, 0).Thumbnail)
}

func TestQueueEmbedEmpty(t *testing.T) {
	t.Parallel()

	embed := queueEmbed(nil, nil)

	require.Equal(t, colorNeutral, embed.Color)
	require.Contains(t, embed.Description, "leer")
	require.Nil(t, embed.Footer)
	require.Empty(t, embed.Title, "the empty reply has no title to carry a second icon")
	require.Equal(t, 1, embedIcons(embed))
}

func TestQueueEmbedNamesTheTrackPlayingNow(t *testing.T) {
	t.Parallel()

	current := testTrack("current")
	embed := queueEmbed(&current, []lavalink.Track{testTrack("first"), testTrack("second")})

	require.Equal(t, titleQueue, embed.Title)
	require.Equal(t, colorNeutral, embed.Color)
	require.Contains(t, embed.Description, iconPlaying+" **"+titleNowPlaying+"**", "the icon marks which track is playing")
	require.Contains(t, embed.Description, "current")
	require.Contains(t, embed.Description, "**"+headingUpNext+"**")
	require.Contains(t, embed.Description, "`1.`")
	require.Contains(t, embed.Description, "first")
	require.Contains(t, embed.Description, "`2.`")
	require.Contains(t, embed.Description, "second")
	require.NotContains(t, embed.Description, "weitere")
	require.Equal(t, 1, embedIcons(embed), "the current-track line is the only icon in a listing")
	require.Zero(t, iconsOutsideTheAuthorLine(embed), "the queue title carries no icon")

	require.Less(t,
		strings.Index(embed.Description, "current"),
		strings.Index(embed.Description, "first"),
		"the current track is named above the waiting ones",
	)
}

func TestQueueEmbedOmitsTheCurrentLineWhenNothingPlays(t *testing.T) {
	t.Parallel()

	embed := queueEmbed(nil, []lavalink.Track{testTrack("first")})

	require.NotContains(t, embed.Description, titleNowPlaying)
	require.NotContains(t, embed.Description, headingUpNext)
	require.True(t, strings.HasPrefix(embed.Description, "`1.`"), "the listing starts the description")
}

func TestQueueEmbedFooterStatesCountAndDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tracks []lavalink.Track
		want   string
	}{
		{
			name:   "one track",
			tracks: []lavalink.Track{testTrack("a")},
			want:   "1 Titel · Gesamtdauer 3:07",
		},
		{
			name:   "three tracks add up",
			tracks: []lavalink.Track{testTrack("a"), testTrack("b"), testTrack("c")},
			want:   "3 Titel · Gesamtdauer 9:21",
		},
		{
			name:   "a livestream makes the total a lower bound",
			tracks: []lavalink.Track{testTrack("a"), testStream("radio")},
			want:   "2 Titel · Gesamtdauer mindestens 3:07",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			embed := queueEmbed(nil, tt.tracks)
			require.NotNil(t, embed.Footer)
			require.Equal(t, tt.want, embed.Footer.Text)
		})
	}
}

func TestQueueEmbedShowsALivestreamAsLive(t *testing.T) {
	t.Parallel()

	embed := queueEmbed(nil, []lavalink.Track{testStream("radio")})
	require.Contains(t, embed.Description, markerLive)
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

			embed := queueEmbed(nil, tracks)

			requireBounded(t, embed)
			require.Contains(t, embed.Footer.Text, "200 Titel")

			listed := strings.Count(embed.Description, "`") / 2
			if tt.wantListed > 0 {
				require.Equal(t, tt.wantListed, listed)
				require.Contains(t, embed.Description, lineQueueResidual(tt.wantResidual))
			} else {
				require.Contains(t, embed.Description, "weitere")
			}
		})
	}
}

// TestQueueEmbedListsTheFullLimitOfMultibyteTitles is the regression for the
// byte-based budget: 20 CJK-titled tracks are far under the character limit but
// were cut short when the budget counted bytes.
func TestQueueEmbedListsTheFullLimitOfMultibyteTitles(t *testing.T) {
	t.Parallel()

	tracks := make([]lavalink.Track, 0, queueListLimit)
	for i := range queueListLimit {
		tracks = append(tracks, testTrack(strings.Repeat("曲", 40)+fmt.Sprintf("%02d", i)))
	}

	embed := queueEmbed(nil, tracks)

	require.Greater(t, len(embed.Description), limitDescription-residualLineBudget, "the listing is over the old byte budget")
	require.Equal(t, queueListLimit, strings.Count(embed.Description, "`")/2, "every track within the limit is listed")
	require.NotContains(t, embed.Description, "weitere")
	requireBounded(t, embed)
}

func TestQueueEmbedResidualLineAgreesInNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tracks int
		want   string
	}{
		{name: "one remaining", tracks: queueListLimit + 1, want: "… und 1 weiterer Titel"},
		{name: "two remaining", tracks: queueListLimit + 2, want: "… und 2 weitere Titel"},
		{name: "many remaining", tracks: 200, want: fmt.Sprintf("… und %d weitere Titel", 200-queueListLimit)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tracks := make([]lavalink.Track, 0, tt.tracks)
			for i := range tt.tracks {
				tracks = append(tracks, testTrack(fmt.Sprintf("track-%03d", i)))
			}

			require.Contains(t, queueEmbed(nil, tracks).Description, tt.want)
		})
	}
}

// TestEveryReplyCarriesAtMostOneIcon holds the icon diet in place: an icon marks
// a state, and no reply accumulates two.
func TestEveryReplyCarriesAtMostOneIcon(t *testing.T) {
	t.Parallel()

	track := testTrackFromSource("song", "youtube")
	current := testTrack("current")

	tests := []struct {
		name  string
		embed discord.Embed
		want  int
	}{
		{name: "play started", embed: startedEmbed(track), want: 1},
		{name: "play queued", embed: queuedEmbed(track, 2), want: 1},
		{name: "now playing", embed: nowPlayingEmbed(track, 0), want: 1},
		{name: "now playing a livestream", embed: nowPlayingEmbed(testStream("radio"), 0), want: 1},
		{name: "queue listing", embed: queueEmbed(&current, []lavalink.Track{track}), want: 1},
		{name: "queue listing without a current track", embed: queueEmbed(nil, []lavalink.Track{track}), want: 0},
		{name: "queue empty", embed: queueEmbed(nil, nil), want: 1},
		{name: "error", embed: errorEmbed(msgGeneric), want: 1},
		{name: "paused", embed: pausedEmbed(), want: 1},
		{name: "resumed", embed: resumedEmbed(), want: 1},
		{name: "stopped", embed: stoppedEmbed(), want: 1},
		{name: "skipped", embed: skippedEmbed(), want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, embedIcons(tt.embed))
			require.Zero(t, iconsOutsideTheAuthorLine(tt.embed), "an icon never decorates a title, field or footer")
		})
	}
}

// TestEveryReplyReadsWithoutItsIcons checks the accessibility rule across the
// whole reply surface, not only the confirmations.
func TestEveryReplyReadsWithoutItsIcons(t *testing.T) {
	t.Parallel()

	track := testTrack("song")
	embeds := map[string]discord.Embed{
		"started":     startedEmbed(track),
		"queued":      queuedEmbed(track, 2),
		"now playing": nowPlayingEmbed(track, 0),
		"queue":       queueEmbed(&track, []lavalink.Track{track}),
		"queue empty": queueEmbed(nil, nil),
		"error":       errorEmbed(msgGeneric),
		"paused":      pausedEmbed(),
	}

	for name, embed := range embeds {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			text := stripIcons(embed.Title) + " " + stripIcons(embed.Description)
			if embed.Author != nil {
				text += " " + stripIcons(embed.Author.Name)
			}
			require.NotEmpty(t, strings.TrimSpace(text), "the outcome must survive every icon being removed")
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

func TestQueueDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tracks      []lavalink.Track
		want        lavalink.Duration
		wantPartial bool
	}{
		{name: "empty queue", tracks: nil, want: 0},
		{
			name:   "two tracks",
			tracks: []lavalink.Track{testTrack("a"), testTrack("b")},
			want:   2 * (3*lavalink.Minute + 7*lavalink.Second),
		},
		{
			name:        "a livestream contributes no length",
			tracks:      []lavalink.Track{testTrack("a"), testStream("radio")},
			want:        3*lavalink.Minute + 7*lavalink.Second,
			wantPartial: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			total, partial := queueDuration(tt.tracks)
			require.Equal(t, tt.want, total)
			require.Equal(t, tt.wantPartial, partial)
		})
	}
}
