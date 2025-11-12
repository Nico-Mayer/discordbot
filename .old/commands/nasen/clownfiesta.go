package nasen

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/db"
)

var ClownfiestaCommand = discord.SlashCommandCreate{
	Name:        "clownfiesta",
	Description: "Verteile Clownsnasen an alle User im Voice-Channel.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "grund",
			Description: "Grund für die Clownfiesta.",
			Required:    false,
		},
	},
}

func ClownfiestaCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	data := event.SlashCommandInteractionData()

	reason, ok := data.OptString("grund")
	if !ok {
		reason = "Clownfiesta 🤡"
	}

	voiceState, ok := event.Client().Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Um diesen Befehl zu nutzen, musst du dich in einem Voice-Channel befinden.",
		})
	}

	voiceChannel, ok := event.Client().Caches().GuildAudioChannel(*voiceState.ChannelID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "ERROR: voice channel is not existing",
		})
	}

	usersInChannel := event.Client().Caches().AudioChannelMembers(voiceChannel)

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	author := event.User()

	var wg sync.WaitGroup
	errChan := make(chan error, len(usersInChannel))

	for _, user := range usersInChannel {
		wg.Add(1)

		if user.User.Bot {
			defer wg.Done()
			continue
		}

		go func(target discord.Member) {
			defer wg.Done()
			var err error
			var nase db.Nase = db.Nase{
				ID:       snowflake.New(time.Now()),
				UserID:   target.User.ID,
				AuthorID: author.ID,
				Reason:   reason,
				Created:  time.Now(),
			}

			if !db.UserInDatabase(target.User.ID) {
				err = db.InsertDBUser(target.User.ID, target.User.Username)
				if err != nil {
					errChan <- err
					return
				}
			}
			err = db.InsertNase(nase)

			errChan <- err
		}(user)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	_, err := event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title: "Clownfiesta! 🤡",
				Thumbnail: &discord.EmbedResource{
					URL: "https://i.kym-cdn.com/photos/images/newsfeed/001/480/336/e0a.gif",
				},
				Description: fmt.Sprintf("**Komplette Clownfiesta in <#%s>!**", voiceChannel.ID()) + buildList(usersInChannel, reason),
			},
		},
	})
	return err
}

func buildList(users []discord.Member, reason string) string {
	var sb strings.Builder
	reasonRow := fmt.Sprintf("\n`Grund:` %s\n", reason)

	sb.WriteString(reasonRow)
	for _, user := range users {
		row := fmt.Sprintf("\n- <@%s> +1 Clownsnase", user.User.ID)
		sb.WriteString(row)
	}
	return sb.String()
}
