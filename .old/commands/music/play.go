package music

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/utils"
)

var (
	urlPattern = regexp.MustCompile(`^https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)[a-zA-Z0-9_-]{11}`)
)

var PlayCommand = discord.SlashCommandCreate{
	Name:        "play",
	Description: "Starte die Wiedergabe eines Songs.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "query",
			Description: "Nutze eine YouTube-URL oder führe eine Suche durch.",
			Required:    true,
		},
	},
}

func PlayCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	data := event.SlashCommandInteractionData()
	query := data.String("query")
	if !urlPattern.MatchString(query) {
		query = "ytsearch:" + strings.TrimSpace(query)
	}
	voiceState, ok := event.Client().Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Dieser Befehl erfordert, dass du dich in einem Voice-Channel befindest.",
		})
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	song, err := b.Enqueue(query)
	if err != nil {
		event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
			Content: "Es gab einen Fehler beim Laden der Songdaten.",
		})
		return err
	}

	// ADD TO QUEUE CASE
	if b.BotStatus == mybot.Playing {
		_, err = event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Author: &discord.EmbedAuthor{
						Name:    event.User().Username,
						IconURL: utils.GetAvatarUrl(event.User()),
					},
					Title:       "📃 - Warteschlange",
					Description: fmt.Sprintf("Added Song: [`%s`](%s)", song.Title, song.URL),
					Thumbnail: &discord.EmbedResource{
						URL: song.ThumbnailURL,
					},
					Footer: &discord.EmbedFooter{
						Text:    "source: youtube",
						IconURL: "https://upload.wikimedia.org/wikipedia/commons/e/ef/Youtube_logo.png",
					},
				},
			},
		})
		return err
	}

	// Play SONG CASE
	// CONNECT TO VOICE, IS A BLOCKING CALL SO RUN IN GO ROUTINE
	go func() {
		conn := b.Client.VoiceManager().CreateConn(*event.GuildID())
		if err = conn.Open(context.TODO(), *voiceState.ChannelID, false, false); err != nil {
			slog.Error("connecting to voice channel", "err:", err.Error())
		}
		if err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone); err != nil {
			slog.Error("setting bot to speaking", "err:", err.Error())
		}
	}()

	go b.PlayQueue(*event.GuildID())
	event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Author: &discord.EmbedAuthor{
					Name:    event.User().Username,
					IconURL: utils.GetAvatarUrl(event.User()),
				},
				Color:       0xff0000,
				Title:       "🔊 - Playing",
				Description: fmt.Sprintf("Loaded Song: [`%s`](%s)", song.Title, song.URL),
				Fields: []discord.EmbedField{
					{
						Name:  "Duration",
						Value: song.Duration + " min",
					},
				},
				Image: &discord.EmbedResource{
					URL: song.ThumbnailURL,
				},
				Footer: &discord.EmbedFooter{
					Text:    "source: youtube",
					IconURL: "https://upload.wikimedia.org/wikipedia/commons/e/ef/Youtube_logo.png",
				},
			},
		},
	})

	return nil
}
