package bot

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

type Bot struct {
	Client   *bot.Client
	Lavalink disgolink.Client
	Queues   *QueueManager
	handlers map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error
}

func New() *Bot {
	return &Bot{
		Queues: &QueueManager{
			queues: make(map[snowflake.ID]*Queue),
		},
	}
}

func (b *Bot) InitHandlers() {
	b.handlers = map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error{
		"play":        b.play,
		"pause":       b.pause,
		"stop":        b.stop,
		"skip":        b.skip,
		"now-playing": b.nowPlaying,
		"queue":       b.queue,
	}
}
