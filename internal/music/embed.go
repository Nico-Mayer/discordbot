package music

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

const (
	colorError   = 0xED4245
	colorSuccess = 0x57F287
	colorInfo    = 0x5865F2
)

const (
	// Discord rejects an embed whose description exceeds this many characters.
	embedDescriptionLimit = 4096
	// Room reserved for the "und N weitere" line appended after the listing.
	residualLineBudget = 64
	// How many queued tracks /queue lists before summarising the rest.
	queueListLimit = 20
)

// The status icons. They live here rather than with the copy, so a reply takes
// its icon as a parameter and cannot end up carrying two.
const (
	iconError     = "❌"
	iconSuccess   = "✅"
	iconInfo      = "ℹ️"
	iconPaused    = "⏸️"
	iconPlaying   = "▶️"
	iconStopped   = "⏹️"
	iconSkipped   = "⏭️"
	iconQueue     = "📋"
	iconMusicNote = "🎶"
)

// iconPad separates the leading status icon from the message text.
const iconPad = "     "

// statusEmbed is the shape every one-line reply takes: the icon leads, the text
// follows, and the text itself never carries one.
func statusEmbed(icon string, color int, text string) discord.Embed {
	return discord.Embed{
		Description: icon + iconPad + text,
		Color:       color,
	}
}

func errorEmbed(msg string) discord.Embed {
	return statusEmbed(iconError, colorError, msg)
}

func successEmbed(msg string) discord.Embed {
	return statusEmbed(iconSuccess, colorSuccess, msg)
}

func infoEmbed(msg string) discord.Embed {
	return statusEmbed(iconInfo, colorInfo, msg)
}

// The confirmations. Each names its own icon, so /skip no longer trails one.
func pausedEmbed() discord.Embed  { return statusEmbed(iconPaused, colorInfo, replyPaused) }
func resumedEmbed() discord.Embed { return statusEmbed(iconPlaying, colorInfo, replyResumed) }
func stoppedEmbed() discord.Embed { return statusEmbed(iconStopped, colorSuccess, replyStopped) }
func skippedEmbed() discord.Embed { return statusEmbed(iconSkipped, colorSuccess, replySkipped) }

func nowPlayingEmbed(track lavalink.Track, position lavalink.Duration) discord.Embed {
	return discord.Embed{
		Title:       iconMusicNote + " " + titleNowPlaying,
		Description: fmt.Sprintf("%s\nvon **%s**", trackLink(track), track.Info.Author),
		Color:       colorInfo,
		Thumbnail:   &discord.EmbedResource{URL: artworkURL(track)},
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%s / %s", formatDuration(position), formatDuration(track.Info.Length)),
		},
	}
}

// trackEmbed confirms that a track started playing immediately.
func trackEmbed(track lavalink.Track) discord.Embed {
	embed := discord.Embed{
		// The author line is what states the outcome; the title is the track.
		Author:      &discord.EmbedAuthor{Name: iconPlaying + " " + authorNowPlaying},
		Title:       track.Info.Title,
		Description: fmt.Sprintf("von **%s**", track.Info.Author),
		Color:       colorSuccess,
		Image:       &discord.EmbedResource{URL: artworkURL(track)},
		Fields: []discord.EmbedField{
			{Name: fieldDuration, Value: formatDuration(track.Info.Length), Inline: new(true)},
		},
	}
	if track.Info.URI != nil {
		embed.URL = *track.Info.URI
	}
	return embed
}

// queuedEmbed confirms that a track was appended to the queue at position.
func queuedEmbed(track lavalink.Track, position int) discord.Embed {
	return discord.Embed{
		Title:       iconQueue + " " + titleQueued,
		Description: fmt.Sprintf("%s\nvon **%s**", trackLink(track), track.Info.Author),
		Color:       colorInfo,
		Thumbnail:   &discord.EmbedResource{URL: artworkURL(track)},
		Fields: []discord.EmbedField{
			{Name: fieldPosition, Value: fmt.Sprintf("#%d", position), Inline: new(true)},
			{Name: fieldDuration, Value: formatDuration(track.Info.Length), Inline: new(true)},
		},
	}
}

// queueEmbed lists at most queueListLimit tracks and summarises the rest, so a
// long queue cannot breach the embed description limit and fail the interaction.
func queueEmbed(tracks []lavalink.Track) discord.Embed {
	if len(tracks) == 0 {
		return statusEmbed(iconQueue, colorInfo, replyQueueEmpty)
	}

	var b strings.Builder
	listed := 0
	for i, track := range tracks[:min(len(tracks), queueListLimit)] {
		line := fmt.Sprintf("`%d.` %s · %s\n", i+1, trackLink(track), formatDuration(track.Info.Length))
		if b.Len()+len(line) > embedDescriptionLimit-residualLineBudget {
			break
		}
		b.WriteString(line)
		listed++
	}

	if remaining := len(tracks) - listed; remaining > 0 {
		fmt.Fprintf(&b, "\n%s", lineQueueResidual(remaining))
	}

	return discord.Embed{
		Title:       iconQueue + " " + titleQueue,
		Description: truncate(b.String(), embedDescriptionLimit),
		Color:       colorInfo,
		Footer: &discord.EmbedFooter{
			Text: footerQueueCount(len(tracks)),
		},
	}
}

// trackLink renders a track as a masked link, or as plain bold text when
// Lavalink left the URI unset.
func trackLink(track lavalink.Track) string {
	if track.Info.URI == nil {
		return fmt.Sprintf("**%s**", track.Info.Title)
	}
	return fmt.Sprintf("[%s](<%s>)", track.Info.Title, *track.Info.URI)
}

// artworkURL falls back to the YouTube thumbnail when Lavalink reports no artwork.
func artworkURL(track lavalink.Track) string {
	if track.Info.ArtworkURL != nil {
		return *track.Info.ArtworkURL
	}
	return fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", track.Info.Identifier)
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

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	const ellipsis = "…"
	cut := limit - len(ellipsis)
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + ellipsis
}
