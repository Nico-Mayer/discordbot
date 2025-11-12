package general

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/db"
	"github.com/nico-mayer/discordbot/levels"
	"github.com/nico-mayer/discordbot/utils"
)

var UserCommand = discord.SlashCommandCreate{
	Name:        "user",
	Description: "Zeige Informationen über einen User.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionUser{
			Name:        "user",
			Description: "Wähle einen User aus.",
			Required:    true,
		},
	},
}

func UserCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	data := event.SlashCommandInteractionData()
	target := data.User("user")

	if target.Bot {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Bot-Informationen sind nicht abrufbar.",
		})
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	dbUser, err := db.ValidateAndFetchUser(target.ID, target.Username)
	if err != nil {
		return err
	}

	userNasenCount, err := db.GetNasenCountForUser(dbUser.ID)
	if err != nil {
		return err
	}

	_, err = event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title:       dbUser.Name,
				Description: "User stats:",
				Color:       0x00ff00,
				Thumbnail: &discord.EmbedResource{
					URL: utils.GetAvatarUrl(target),
				},
				Fields: []discord.EmbedField{
					{
						Name: "Level",
						// Todo add level calculation
						Value: fmt.Sprintf("```%d```", levels.CalcUserLevel(dbUser.Exp)),
					}, {
						Name:  "Exp",
						Value: fmt.Sprintf("```%d```", dbUser.Exp),
					}, {
						Name:  "Nasen",
						Value: fmt.Sprintf("```%d```", userNasenCount),
					},
				},
				Footer: &discord.EmbedFooter{
					Text: "Für eine Liste aller Clownsnasen des Users, verwende `/nasen`.",
				},
			},
		},
	})
	return err
}
