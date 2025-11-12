package bot

import (
	"context"
	"log/slog"
	"sync"

	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/nico-mayer/discordbot/config"
)

type bot struct {
	Client        disgobot.Client
	SlashCommands map[string]func(event *events.ApplicationCommandInteractionCreate, b *bot) error
}

var instance *bot
var once sync.Once

func Get() *bot {
	once.Do(func() {
		instance = newBot()
	})
	return instance
}

func newBot() *bot {
	config := config.Get()

	client, err := disgo.New(config.TOKEN,
		disgobot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
			),
		),
	)

	if err != nil {
		slog.Error("errors while connecting to gateway", slog.Any("err", err))
		panic(err)
	}

	return &bot{
		Client: client,
	}
}

func (b *bot) Start() error {
	err := b.Client.OpenGateway(context.TODO())
	if err != nil {
		slog.Error("errors while connecting to gateway", slog.Any("err", err))
	}
	return err
}
