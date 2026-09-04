package music

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

// Colour carries meaning on two axes: a status reply is coloured by the state it
// reports, and a track card by the service the track came from. Nothing picks a
// colour per command.
const (
	colorError   = 0xED4245
	colorSuccess = 0x57F287
	colorPaused  = 0xFEE75C
	colorAccent  = 0x5865F2
	colorNeutral = 0x4E5058
)

// sourceColors stripes a track card in the colour of the service it came from,
// so a reader recognises where a track was resolved from before reading the
// footer, and two tracks from the same service read as a group.
var sourceColors = map[string]int{
	"applemusic":   0xFA243C,
	"bandcamp":     0x629AA9,
	"deezer":       0xA238FF,
	"soundcloud":   0xFF5500,
	"spotify":      0x1DB954,
	"twitch":       0x9146FF,
	"vimeo":        0x1AB7EA,
	"youtube":      0xFF0000,
	"youtubemusic": 0xFF0000,
}

func sourceColor(source string) int {
	if color, ok := sourceColors[strings.ToLower(source)]; ok {
		return color
	}
	return colorAccent
}

// Discord answers 400 when an embed field is longer than its limit, or when the
// embed as a whole exceeds limitTotal, which would fail the very reply that
// carries the outcome.
const (
	limitTitle       = 256
	limitDescription = 4096
	limitFieldName   = 256
	limitFieldValue  = 1024
	limitFooter      = 2048
	limitAuthorName  = 256
	limitTotal       = 6000
)

const (
	// How many queued tracks /queue lists before summarising the rest.
	queueListLimit = 20
	// Room reserved for the residual line appended after the listing.
	residualLineBudget = 64
	// Cells in the /now-playing progress bar.
	progressBarWidth = 12
)

// The status icons. Each one names a state the reader would otherwise have to
// read a sentence to tell apart - paused from resumed, queued from playing - and
// a reply carries at most one. There is no icon for decoration: none marks a
// title, a field or a footer.
const (
	iconError   = "❌"
	iconInfo    = "ℹ️"
	iconPlaying = "▶️"
	iconPaused  = "⏸️"
	iconStopped = "⏹️"
	iconSkipped = "⏭️"
	iconQueued  = "➕"
)

const (
	barFilled = "━"
	barKnob   = "●"
	barEmpty  = "─"
)

// statusEmbed is the shape every one-line reply takes: the icon leads, the text
// follows, and the text itself never carries one.
func statusEmbed(icon string, color int, text string) discord.Embed {
	return bound(discord.Embed{
		Description: icon + " " + text,
		Color:       color,
	})
}

func errorEmbed(msg string) discord.Embed {
	return statusEmbed(iconError, colorError, msg)
}

func infoEmbed(msg string) discord.Embed {
	return statusEmbed(iconInfo, colorNeutral, msg)
}

// The confirmations. Each names the state it moved playback to, in its icon and
// in its colour: paused is a state to notice, stopped leaves nothing active.
func pausedEmbed() discord.Embed  { return statusEmbed(iconPaused, colorPaused, replyPaused) }
func resumedEmbed() discord.Embed { return statusEmbed(iconPlaying, colorSuccess, replyResumed) }
func stoppedEmbed() discord.Embed { return statusEmbed(iconStopped, colorNeutral, replyStopped) }
func skippedEmbed() discord.Embed { return statusEmbed(iconSkipped, colorSuccess, replySkipped) }

// card is what distinguishes one track reply from another. The layout itself
// lives in trackCard, so the three replies cannot drift apart again.
type card struct {
	icon   string
	state  string
	fields []discord.EmbedField
	extra  string
}

func trackCard(track lavalink.Track, c card) discord.Embed {
	description := fmt.Sprintf("von **%s**", track.Info.Author)
	if c.extra != "" {
		description += "\n\n" + c.extra
	}

	embed := discord.Embed{
		Author:      &discord.EmbedAuthor{Name: c.icon + " " + c.state},
		Title:       track.Info.Title,
		Description: description,
		Color:       sourceColor(track.Info.SourceName),
		Fields:      c.fields,
	}
	if track.Info.URI != nil {
		embed.URL = *track.Info.URI
	}
	if url, ok := artworkURL(track); ok {
		embed.Thumbnail = &discord.EmbedResource{URL: url}
	}
	if source := sourceLabel(track.Info.SourceName); source != "" {
		embed.Footer = &discord.EmbedFooter{Text: source}
	}
	return bound(embed)
}

func startedEmbed(track lavalink.Track) discord.Embed {
	return trackCard(track, card{
		icon:   iconPlaying,
		state:  authorNowPlaying,
		fields: []discord.EmbedField{durationField(track)},
	})
}

// queuedEmbed confirms that a track was appended to the queue at position.
func queuedEmbed(track lavalink.Track, position int) discord.Embed {
	return trackCard(track, card{
		icon:  iconQueued,
		state: titleQueued,
		fields: []discord.EmbedField{
			{Name: fieldPosition, Value: fmt.Sprintf("#%d", position), Inline: new(true)},
			durationField(track),
		},
	})
}

func nowPlayingEmbed(track lavalink.Track, position lavalink.Duration) discord.Embed {
	c := card{icon: iconPlaying, state: titleNowPlaying}
	if track.Info.IsStream {
		c.fields = []discord.EmbedField{durationField(track)}
	} else {
		c.extra = progressLine(position, track.Info.Length)
	}
	return trackCard(track, c)
}

func durationField(track lavalink.Track) discord.EmbedField {
	return discord.EmbedField{Name: fieldDuration, Value: trackLength(track), Inline: new(true)}
}

// queueEmbed names the track playing now, lists at most queueListLimit waiting
// tracks and summarises the rest, so a long queue cannot breach the embed
// description limit and fail the interaction.
func queueEmbed(current *lavalink.Track, tracks []lavalink.Track) discord.Embed {
	if len(tracks) == 0 {
		return infoEmbed(replyQueueEmpty)
	}

	var b strings.Builder
	used := 0
	write := func(s string) {
		b.WriteString(s)
		used += utf8.RuneCountInString(s)
	}

	if current != nil {
		write(fmt.Sprintf("%s **%s**\n%s · %s\n\n**%s**\n",
			iconPlaying, titleNowPlaying, trackLink(*current), trackLength(*current), headingUpNext))
	}

	listed := 0
	for i, track := range tracks[:min(len(tracks), queueListLimit)] {
		line := fmt.Sprintf("`%d.` %s · %s\n", i+1, trackLink(track), trackLength(track))
		if used+utf8.RuneCountInString(line) > limitDescription-residualLineBudget {
			break
		}
		write(line)
		listed++
	}

	if remaining := len(tracks) - listed; remaining > 0 {
		write("\n" + lineQueueResidual(remaining))
	}

	total, partial := queueDuration(tracks)

	return bound(discord.Embed{
		Title:       titleQueue,
		Description: b.String(),
		Color:       colorNeutral,
		Footer:      &discord.EmbedFooter{Text: footerQueueSummary(len(tracks), formatDuration(total), partial)},
	})
}

// queueDuration sums how long the queue will play. A livestream has no length to
// contribute, so a queue holding one reports a lower bound.
func queueDuration(tracks []lavalink.Track) (total lavalink.Duration, partial bool) {
	for _, track := range tracks {
		if track.Info.IsStream {
			partial = true
			continue
		}
		total += track.Info.Length
	}
	return total, partial
}

// trackLink renders a track as a masked link, or as plain bold text when
// Lavalink left the URI unset.
func trackLink(track lavalink.Track) string {
	if track.Info.URI == nil {
		return fmt.Sprintf("**%s**", track.Info.Title)
	}
	return fmt.Sprintf("[%s](<%s>)", track.Info.Title, *track.Info.URI)
}

// trackLength renders how long a track runs. A livestream has no end, so it
// takes the live marker where a duration would otherwise go.
func trackLength(track lavalink.Track) string {
	if track.Info.IsStream {
		return markerLive
	}
	return formatDuration(track.Info.Length)
}

// progressLine renders the elapsed position as a fixed-width bar between the two
// times. A track of unknown length gets the times alone rather than a bar that
// would have to divide by zero to place its knob.
func progressLine(position, length lavalink.Duration) string {
	if length <= 0 {
		return fmt.Sprintf("%s / %s", formatDuration(position), formatDuration(length))
	}

	// Computed in float, because a position multiplied by the bar width
	// overflows int64 for the sentinel lengths Lavalink can report.
	knob := int(float64(position) / float64(length) * float64(progressBarWidth-1))
	knob = min(max(knob, 0), progressBarWidth-1)

	bar := strings.Repeat(barFilled, knob) + barKnob + strings.Repeat(barEmpty, progressBarWidth-1-knob)
	return fmt.Sprintf("%s %s %s", formatDuration(position), bar, formatDuration(length))
}

// artworkURL reports the thumbnail to show, and false when there is none. The
// derived URL is built from a YouTube video id, so it is only valid for a
// YouTube track - for any other source it would render as a broken image.
func artworkURL(track lavalink.Track) (string, bool) {
	if track.Info.ArtworkURL != nil {
		return *track.Info.ArtworkURL, true
	}
	if strings.HasPrefix(strings.ToLower(track.Info.SourceName), "youtube") {
		return fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", track.Info.Identifier), true
	}
	return "", false
}

// sourceNames spells the sources Lavalink reports the way they spell themselves.
var sourceNames = map[string]string{
	"applemusic":   "Apple Music",
	"bandcamp":     "Bandcamp",
	"deezer":       "Deezer",
	"soundcloud":   "SoundCloud",
	"spotify":      "Spotify",
	"twitch":       "Twitch",
	"vimeo":        "Vimeo",
	"youtube":      "YouTube",
	"youtubemusic": "YouTube Music",
}

func sourceLabel(source string) string {
	if name, ok := sourceNames[strings.ToLower(source)]; ok {
		return name
	}
	return source
}

// formatDuration renders a duration as m:ss, or h:mm:ss from one hour upwards.
func formatDuration(d lavalink.Duration) string {
	if d <= 0 {
		return "0:00"
	}
	if hours := d.Hours(); hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, d.MinutesPart(), d.SecondsPart())
	}
	return fmt.Sprintf("%d:%02d", d.Minutes(), d.SecondsPart())
}

// bound brings an embed within what Discord accepts. Every builder returns
// through it, so a field added later cannot be the one that is left unbounded.
func bound(embed discord.Embed) discord.Embed {
	embed.Title = clamp(embed.Title, limitTitle)
	embed.Description = clamp(embed.Description, limitDescription)

	if embed.Author != nil {
		author := *embed.Author
		author.Name = clamp(author.Name, limitAuthorName)
		embed.Author = &author
	}
	if embed.Footer != nil {
		footer := *embed.Footer
		footer.Text = clamp(footer.Text, limitFooter)
		embed.Footer = &footer
	}
	if len(embed.Fields) > 0 {
		fields := slices.Clone(embed.Fields)
		for i := range fields {
			fields[i].Name = clamp(fields[i].Name, limitFieldName)
			fields[i].Value = clamp(fields[i].Value, limitFieldValue)
		}
		embed.Fields = fields
	}

	if over := embedLength(embed) - limitTotal; over > 0 {
		embed.Description = clamp(embed.Description, max(utf8.RuneCountInString(embed.Description)-over, 0))
	}
	return embed
}

// embedLength counts the characters Discord counts towards limitTotal.
func embedLength(embed discord.Embed) int {
	n := utf8.RuneCountInString(embed.Title) + utf8.RuneCountInString(embed.Description)
	if embed.Author != nil {
		n += utf8.RuneCountInString(embed.Author.Name)
	}
	if embed.Footer != nil {
		n += utf8.RuneCountInString(embed.Footer.Text)
	}
	for _, field := range embed.Fields {
		n += utf8.RuneCountInString(field.Name) + utf8.RuneCountInString(field.Value)
	}
	return n
}

// clamp shortens s to limit characters, marking that it was shortened. Discord
// counts characters rather than bytes, so counting bytes would cut a German or
// CJK value at a fraction of the allowance it actually has.
func clamp(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= limit {
		return s
	}

	const ellipsis = "…"
	var b strings.Builder
	for i, r := range []rune(s) {
		if i >= limit-1 {
			break
		}
		b.WriteRune(r)
	}
	return b.String() + ellipsis
}
