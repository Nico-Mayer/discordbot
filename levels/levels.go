package levels

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nico-mayer/go_discordbot/db"
)

const (
	minLevel          = 1
	expNeededPerLevel = 50
	expPerMessage     = 10
	expPerVoiceJoin   = 20
	commandsChannelId = "692820682425237615"
)

var levelMapping = map[int]string{
	20:  "1196937224634257479",
	40:  "1196936971776438302",
	60:  "1196936450306998322",
	80:  "696672218117177437",
	100: "696671960507351100",
}

func Init(s *discordgo.Session) {
	// Message Sent
	s.AddHandler(func(
		s *discordgo.Session,
		m *discordgo.MessageCreate,
	) {
		var user db.User

		if m.Author.Bot {
			return
		}

		user, err := db.GetUser(m.Author.ID)
		if err != nil {
			fmt.Println(err)
		}

		if !user.InDatabase() {
			err := db.InsertUser(m.Author.ID, m.Author.Username)
			if err != nil {
				fmt.Println(err)
			}
			user, err = db.GetUser(m.Author.ID)
			if err != nil {
				fmt.Println(err)
			}
		}

		user.GiveExp(expPerMessage)
		newLevel, _, levelUp, err := user.CalcLevel(expNeededPerLevel)
		if err != nil {
			fmt.Println(err)
		}

		if levelUp {
			handleLevelUp(newLevel, s, m.GuildID, user)
		}

	})

	// Voice Channel Joined
	s.AddHandler(func(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
		if v.BeforeUpdate == nil {
			var user db.User

			if v.Member.User.Bot {
				return
			}

			user, err := db.GetUser(v.UserID)
			if err != nil {
				fmt.Println(err)
			}

			if !user.InDatabase() {
				err := db.InsertUser(v.UserID, v.Member.User.Username)
				if err != nil {
					fmt.Println(err)
				}
				user, err = db.GetUser(v.UserID)
				if err != nil {
					fmt.Println(err)
				}
			}

			user.GiveExp(expPerVoiceJoin)
			newLevel, _, levelUp, err := user.CalcLevel(expNeededPerLevel)
			if err != nil {
				fmt.Println(err)
			}

			if levelUp {
				handleLevelUp(newLevel, s, v.GuildID, user)
			}
		}
	})
}

func handleLevelUp(newLevel int, s *discordgo.Session, guildID string, user db.User) {
	switch newLevel {
	case 20:
		s.GuildMemberRoleAdd(guildID, user.ID, levelMapping[newLevel])
		congratulate(s, user, newLevel)
	case 40:
		s.GuildMemberRoleAdd(guildID, user.ID, levelMapping[newLevel])
		congratulate(s, user, newLevel)
	case 60:
		s.GuildMemberRoleAdd(guildID, user.ID, levelMapping[newLevel])
		congratulate(s, user, newLevel)
	case 80:
		s.GuildMemberRoleAdd(guildID, user.ID, levelMapping[newLevel])
		congratulate(s, user, newLevel)
	case 100:
		s.GuildMemberRoleAdd(guildID, user.ID, levelMapping[newLevel])
		congratulate(s, user, newLevel)
	default:
		return
	}
}

func congratulate(s *discordgo.Session, user db.User, newLevel int) {
	message := fmt.Sprintf("<@%s> ist jetzt Level %d! 🎉", user.ID, newLevel)

	s.ChannelMessageSend(commandsChannelId, message)
}
