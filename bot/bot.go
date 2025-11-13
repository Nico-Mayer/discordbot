package bot

import (
	"context"
	"log/slog"
	"sync"

	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/nico-mayer/discordbot/cmd"
	"github.com/nico-mayer/discordbot/config"
)

type bot struct {
	Client        disgobot.Client
	SlashCommands map[string]*cmd.Cmd
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
		disgobot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates, cache.FlagMembers, cache.FlagChannels),
		),
		disgobot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentGuildVoiceStates,
			),
		),

		disgobot.WithEventListenerFunc(onApplicationCommand),
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
	b.SlashCommands = cmd.GetAll()
	cmd.RegisterSlashCommands(b.Client, config.Get().GUILD_ID)
	return err
}

func onApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	b := Get()

	cmd, ok := b.SlashCommands[data.CommandName()]
	if !ok {
		slog.Info("unknown command", slog.String("command", data.CommandName()))
		return
	}
	err := cmd.Handler(event)
	if err != nil {
		slog.Error("executing slash command", slog.String("command", data.CommandName()), "err:", err.Error())
	}
}
