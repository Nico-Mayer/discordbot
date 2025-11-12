package mybot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/nico-mayer/discordbot/db"
	"github.com/nico-mayer/discordbot/levels"
)

type BotStatus int32

const (
	Resting BotStatus = 0
	Playing BotStatus = 1
)

type Bot struct {
	Client        bot.Client
	BotStatus     BotStatus
	Handlers      map[string]func(event *events.ApplicationCommandInteractionCreate, b *Bot) error
	Queue         []Song
	SkipInterrupt chan bool
}

func NewBot() *Bot {
	return &Bot{
		BotStatus:     Resting,
		SkipInterrupt: make(chan bool, 1),
	}
}

func (b *Bot) SetupBot() {
	var err error

	// Initialize bot client
	b.Client, err = disgo.New(
		os.Getenv("TOKEN"),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates, cache.FlagMembers, cache.FlagChannels),
		),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentGuildVoiceStates,
			),
		),
		// Slash command listener
		bot.WithEventListenerFunc(b.onApplicationCommand),

		// Message create listener
		bot.WithEventListenerFunc(b.onMessageCreate),

		// Voice join listener
		bot.WithEventListenerFunc(b.onVoiceJoin),
	)
	if err != nil {
		log.Fatal("FATAL: setting up bot client", err)
	}

	defer b.Client.Close(context.TODO())
}

func (b *Bot) onApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	author := event.User()

	dbUser, err := db.ValidateAndFetchUser(author.ID, author.Username)
	if err != nil {
		slog.Error("fetching db user on slashcommand execute")
	}

	err = levels.GrantExpToUser(b.Client, dbUser, levels.EXP_PER_SLASH_COMMAND)
	if err != nil {
		slog.Error("granting exp to user on slash command exec")
	}

	handler, ok := b.Handlers[data.CommandName()]
	if !ok {
		slog.Info("unknown command", slog.String("command", data.CommandName()))
		return
	}
	err = handler(event, b)
	if err != nil {
		slog.Error("executing slash command", slog.String("command", data.CommandName()), "err:", err.Error())
	}
}

func (b *Bot) onMessageCreate(event *events.MessageCreate) {
	author := event.Message.Member.User
	if author.Bot {
		return
	}

	dbUser, err := db.ValidateAndFetchUser(author.ID, author.Username)
	if err != nil {
		slog.Error("validating and fetching user on message create")
	}

	err = levels.GrantExpToUser(b.Client, dbUser, levels.EXP_PER_MESSAGE)
	if err != nil {
		slog.Error("granting exp to user on message create")
	}

	err = dbUser.IncrementMessageSentCount()
	if err != nil {
		slog.Error("incrementing message count for user", "err:", err.Error())
	}
}

func (b *Bot) onVoiceJoin(event *events.GuildVoiceJoin) {
	author := event.Member.User
	if event.Member.User.Bot {
		return
	}
	dbUser, err := db.ValidateAndFetchUser(author.ID, author.Username)
	if err != nil {
		slog.Error("validating and fetching user on voice join")
	}

	err = levels.GrantExpToUser(b.Client, dbUser, levels.EXP_PER_VOICE_JOIN)
	if err != nil {
		slog.Error("granting exp to user on voice join")
	}

	err = dbUser.IncrementVoiceJoinCount()
	if err != nil {
		slog.Error("incrementing voice join count", "err:", err.Error())
	}
}

func (b *Bot) SetStatus(status string) error {
	slog.Info(fmt.Sprintf("bot status set to: %s", status))
	return b.Client.SetPresence(context.TODO(), gateway.WithCustomActivity(status))
}
