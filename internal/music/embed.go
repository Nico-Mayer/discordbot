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

// iconPad separates the leading status icon from the message text.
const iconPad = "     "

func errorEmbed(msg string) discord.Embed {
	return discord.Embed{
		Description: "❌" + iconPad + msg,
		Color:       colorError,
	}
}

func successEmbed(msg string) discord.Embed {
	return discord.Embed{
		Description: "✅" + iconPad + msg,
		Color:       colorSuccess,
	}
}

func infoEmbed(msg string) discord.Embed {
	return discord.Embed{
		Description: "ℹ️" + iconPad + msg,
		Color:       colorInfo,
	}
}

func nowPlayingEmbed(track lavalink.Track, position lavalink.Duration) discord.Embed {
	return discord.Embed{
		Title:       "🎶 Läuft gerade",
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
		Title:       track.Info.Title,
		Description: fmt.Sprintf("von **%s**", track.Info.Author),
		Color:       colorSuccess,
		Image:       &discord.EmbedResource{URL: artworkURL(track)},
		Fields: []discord.EmbedField{
			{Name: "⏱️ Dauer", Value: formatDuration(track.Info.Length), Inline: new(true)},
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
		Title:       "📋 Zur Warteschlange hinzugefügt",
		Description: fmt.Sprintf("%s\nvon **%s**", trackLink(track), track.Info.Author),
		Color:       colorInfo,
		Thumbnail:   &discord.EmbedResource{URL: artworkURL(track)},
		Fields: []discord.EmbedField{
			{Name: "Position", Value: fmt.Sprintf("#%d", position), Inline: new(true)},
			{Name: "Dauer", Value: formatDuration(track.Info.Length), Inline: new(true)},
		},
	}
}

// queueEmbed lists at most queueListLimit tracks and summarises the rest, so a
// long queue cannot breach the embed description limit and fail the interaction.
func queueEmbed(tracks []lavalink.Track) discord.Embed {
	if len(tracks) == 0 {
		return infoEmbed("📋 Die Warteschlange ist leer")
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
		fmt.Fprintf(&b, "\n… und %d weitere", remaining)
	}

	return discord.Embed{
		Title:       "📋 Warteschlange",
		Description: truncate(b.String(), embedDescriptionLimit),
		Color:       colorInfo,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%d Titel in der Warteschlange", len(tracks)),
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
