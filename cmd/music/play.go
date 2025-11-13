package music

import (
	"context"
	"log/slog"
	"os"
	"os/exec"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
)

var PlayCmdMeta = discord.SlashCommandCreate{
	Name:        "play",
	Description: "Play music from url",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "url",
			Description: "Youtube url",
			Required:    true,
		},
	},
}

func PlayCmdHandler(event *events.ApplicationCommandInteractionCreate) error {
	query := event.SlashCommandInteractionData().String("url")

	voiceState, ok := event.Client().Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "You need to be in a voice chanel to use this command",
		})
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	cmd := exec.Command(
		"yt-dlp", query,
		"--extract-audio",
		"--audio-format", "opus",
		"--no-playlist",
		"-o", "-",
		"--quiet",
		"--ignore-errors",
		"--no-warnings",
	)
	cmd.Stderr = os.Stderr

	_, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("error on yt-dlp command", "err", err)
	}

	go func() {
		conn := event.Client().VoiceManager().CreateConn(*event.GuildID())
		if err = conn.Open(context.TODO(), *voiceState.ChannelID, false, false); err != nil {
			slog.Error("connecting to voice channel", "err:", err.Error())
		}
		if err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone); err != nil {
			slog.Error("setting bot to speaking", "err:", err.Error())
		}
	}()

	return nil
}
