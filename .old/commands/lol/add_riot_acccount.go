package lol

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/db"
	"github.com/nico-mayer/discordbot/utils"
)

var AddRiotAccountCommand = discord.SlashCommandCreate{
	Name:        "add_riot_account",
	Description: "Füge deinen Riot-Account hinzu.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "name",
			Description: "Riot ID name",
			Required:    true,
		},
		discord.ApplicationCommandOptionString{
			Name:        "tagline",
			Description: "Riot ID Tagline (ohne #)",
			Required:    true,
		},
	},
}

func AddRiotAccountCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	data := event.SlashCommandInteractionData()
	name := data.String("name")
	tagLine := data.String("tagline")
	author := event.User()

	account, err := GolioClient.Riot.Account.GetByRiotID(name, tagLine)
	if err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: fmt.Sprintf("Account [`%s#%s`] not found.", name, tagLine),
		})
	}

	dbuser, err := db.ValidateAndFetchUser(author.ID, author.Username)
	if err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "ERROR fetching user data from database",
		})
	}

	err = dbuser.SetRiotPUUID(account.Puuid)
	if err != nil {
		slog.Error("setting riot puuid in database", "err:", err.Error())
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "ERROR setting riot puuid in database",
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Flags: discord.MessageFlagEphemeral,
		Embeds: []discord.Embed{
			{
				Author: &discord.EmbedAuthor{
					Name:    author.Username,
					IconURL: utils.GetAvatarUrl(author),
				},
				Title:       "✅ - Account added",
				Description: fmt.Sprintf("Account: [`%s#%s`]", account.GameName, account.TagLine),
			},
		},
	})

}
