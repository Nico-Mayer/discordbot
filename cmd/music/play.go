package music

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/ffmpeg-audio"
	"github.com/nico-mayer/discordbot/utils"
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

	voiceState, ok := event.Client().Caches.VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "You need to be in a voice chanel to use this command",
		})
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
		// "--buffer-size", "16K",
	)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "Error creating stdout pipe (yt-dlp): " + err.Error(),
		})
	}

	if err = event.DeferCreateMessage(false); err != nil {
		log.Error("creating defer message", "err", err)
	}

	go func() {
		conn := event.Client().VoiceManager.CreateConn(*event.GuildID())
		if err = conn.Open(context.TODO(), *voiceState.ChannelID, false, false); err != nil {
			slog.Error("connecting to voice channel", "err:", err.Error())
		}
		defer conn.Close(context.TODO())
		if err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone); err != nil {
			slog.Error("setting bot to speaking", "err:", err.Error())
		}

		if err = cmd.Start(); err != nil {
			fmt.Print(err)
		}

		opusProvider := ffmpeg.New(context.TODO(), bufio.NewReader(stdout))
		defer opusProvider.Close()
		time.Sleep(1000 * time.Millisecond)

		conn.SetOpusFrameProvider(opusProvider)

		utils.CreateFollowupMessage(event, discord.MessageCreate{
			Content: fmt.Sprintf("playing %s", query),
		})

		if err = cmd.Wait(); err != nil {
			log.Error("waiting for yt-dlp: ", err)
		}

		if err = opusProvider.Wait(); err != nil {
			log.Info("track finished")
			conn.SetOpusFrameProvider(nil)
		}
		time.Sleep(time.Second * 10)
	}()

	return nil
}
