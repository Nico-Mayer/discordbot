package nasen

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/jedib0t/go-pretty/table"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/db"
)

var NasenCommand = discord.SlashCommandCreate{
	Name:        "nasen",
	Description: "Zeige eine Liste aller Nasen des Users.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionUser{
			Name:        "user",
			Description: "Wähle einen User aus",
			Required:    true,
		},
	},
}

func NasenCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	data := event.SlashCommandInteractionData()
	target := data.User("user")

	nasen, err := db.GetNasenForUser(target.ID)
	if err != nil {
		return err
	}

	if len(nasen) == 0 {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: fmt.Sprintf("<@%s> hat noch keine Clownsnase. Du kannst ihm eine mit `/clownsnase` geben.", target.ID),
		})
	}

	description, err := formatDescription(target, nasen)
	if err != nil {
		return err
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	_, err = event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
		Content: description,
	})
	return err
}

func formatDescription(user discord.User, nasen []db.Nase) (string, error) {
	var sb strings.Builder

	heading := fmt.Sprintf("Alle Clownsnasen von <@%s> \n", user.ID)
	sb.WriteString(heading)
	sb.WriteString("```\n")

	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.AppendHeader(table.Row{"Datum", "Von", "Grund"})
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.SetAutoIndex(true)

	for i := len(nasen) - 1; i >= 0; i-- {
		nase := nasen[i]
		author, err := db.GetUser(nase.AuthorID)
		if err != nil {
			return "", err
		}

		date := fmt.Sprintf("%v", nase.Created.Format("02-Jan-06"))
		t.AppendRow(table.Row{date, author.Name, nase.Reason})
	}

	t.Render()
	sb.WriteString("```")

	return sb.String(), nil
}
