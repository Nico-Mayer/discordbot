package mybot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"os/exec"

	"github.com/disgoorg/ffmpeg-audio"
	"github.com/disgoorg/snowflake/v2"
)

type Song struct {
	Title        string
	URL          string
	ID           string
	ThumbnailURL string
	Duration     string
	Query        string
}

func (b *Bot) Enqueue(query string) (Song, error) {
	song, err := getSongData(query)
	if err != nil {
		return Song{}, err
	}

	b.Queue = append(b.Queue, song)

	return song, nil
}

func (b *Bot) Dequeue() Song {
	song := b.Queue[0]
	b.Queue = b.Queue[1:]
	return song
}

func (b *Bot) ClearQueue() {
	b.Queue = []Song{}
}

func (b *Bot) PlayQueue(guildID snowflake.ID) error {
	cmd := exec.Command(
		"yt-dlp", b.Queue[0].Query,
		"--extract-audio",
		"--audio-format", "opus",
		"--no-playlist",
		"-o", "-",
		"--quiet",
		"--ignore-errors",
		"--no-warnings",
	)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("creating stdout pipe", "err:", err.Error())
		return err
	}

	b.BotStatus = Playing

	if err = cmd.Start(); err != nil {
		slog.Error("stating yt-dlp command", "err:", err.Error())
		return err
	}

	opusProvider, err := ffmpeg.New(context.TODO(), bufio.NewReader(stdout))
	if err != nil {
		slog.Error("creating opus provider", "err:", err.Error())
		return err
	}

	conn := b.Client.VoiceManager().GetConn(guildID)
	conn.SetOpusFrameProvider(opusProvider)

	defer func() {
		opusProvider.Close()
		b.Dequeue()
		if len(b.Queue) > 0 {
			go b.PlayQueue(guildID)
		} else {
			conn.Close(context.TODO())
			b.BotStatus = Resting
		}
	}()

	done := make(chan bool, 1)

	go func(doneChan chan bool) {
		if err = cmd.Wait(); err != nil {
			slog.Error("waiting for yt-dlp command", "err:", err.Error())
		}
		time.Sleep(7 * time.Second)
		doneChan <- true
	}(done)

	select {
	case <-done:
		return nil

	case <-b.SkipInterrupt:
		return nil
	}
}

func getSongData(query string) (Song, error) {
	cmd := exec.Command("yt-dlp",
		"--get-title",
		"--get-id",
		"--get-thumbnail",
		"--get-duration",
		"--ignore-errors",
		"--no-warnings",
		"--skip-download",
		query,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("loading songdata with yt-dlp command", "err:", err.Error())
		return Song{}, err
	}

	metadata := strings.Split(string(output), "\n")
	if len(metadata) < 4 {
		return Song{}, errors.New("loading song metadata")
	}

	var song Song
	song.Title = metadata[0]
	song.ID = metadata[1]
	song.ThumbnailURL = metadata[2]
	song.Duration = metadata[3]
	song.Query = query
	song.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.ID)
	return song, nil
}
