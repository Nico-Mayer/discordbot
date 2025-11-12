package lol

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/KnutZuidema/golio/datadragon"
	"github.com/KnutZuidema/golio/riot/lol"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	mybot "github.com/nico-mayer/discordbot/bot"
	"github.com/nico-mayer/discordbot/db"
)

const (
	RED_SIDE        = 100
	BLUE_SIDE       = 200
	SPECTATOR_DELAY = 240
)

var LiveGameCommand = discord.SlashCommandCreate{
	Name:        "live_game",
	Description: "Zeige Live-Spieldaten für einen User an.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionUser{
			Name:        "user",
			Description: "Der Benutzer, dessen Live-Spieldaten du sehen möchtest.",
			Required:    true,
		},
	},
}

func LiveGameCommandHandler(event *events.ApplicationCommandInteractionCreate, b *mybot.Bot) error {
	data := event.SlashCommandInteractionData()
	target := data.User("user")
	champs, err := GolioClient.DataDragon.GetChampions()
	if err != nil {
		slog.Error("error fetching champs", "err:", err.Error())
	}

	if target.Bot {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "Ich bin in keinem game, ich bin ein bot.",
		})
	}

	dbuser, err := db.ValidateAndFetchUser(target.ID, target.Username)
	if err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: "ERROR getting user from database",
		})
	}

	if dbuser.RiotPUUID.String == "" {
		return event.CreateMessage(discord.MessageCreate{
			Flags:   discord.MessageFlagEphemeral,
			Content: fmt.Sprintf("<@%s> has no riot account connected", target.ID),
		})
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	targetRiotAccount, err := GolioClient.Riot.Account.GetByPUUID(dbuser.RiotPUUID.String)
	if err != nil {
		event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
			Content: fmt.Sprintf("ERROR account not found %s", target.ID),
		})
		return err
	}

	liveGame, err := GolioClient.Riot.LoL.Spectator.GetCurrent(targetRiotAccount.Puuid)
	if err != nil {
		event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
			Content: fmt.Sprintf("<@%s> ist aktuell in keinem spiel", target.ID),
		})
		return err
	}

	if liveGame.GameMode != "CLASSIC" {
		event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
			Content: fmt.Sprintf("<@%s> ist in einem spielmodus der aktuell nicht unterstuetzt wird.", target.ID),
		})
		return err
	}

	var inline bool = true
	_, err = event.Client().Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title:       fmt.Sprintf("🔎 - Live Game [%s#%s]", targetRiotAccount.GameName, targetRiotAccount.TagLine),
				Description: fmt.Sprintf("Ingame seit [`%s min`]", secToMinutesAndSec(liveGame.GameLength+SPECTATOR_DELAY)),
				Fields: []discord.EmbedField{
					{
						Name:   "🔵 Blue Team",
						Value:  getTeam(liveGame, BLUE_SIDE, champs),
						Inline: &inline,
					}, {
						Name:   "🔴 Red Team",
						Value:  getTeam(liveGame, RED_SIDE, champs),
						Inline: &inline,
					},
				},
			},
		},
	})
	return err
}

func getTeam(liveGame *lol.GameInfo, teamID int, champs []datadragon.ChampionData) string {
	var res strings.Builder

	for _, participant := range liveGame.Participants {
		if participant.TeamID == teamID {
			var champName string
			var urlExtension string

			participantRiotAccount, err := GolioClient.Riot.Account.GetByPUUID(participant.PUUID)
			if err != nil {
				urlExtension = "unknown"
				slog.Error("get participant account", "err:", err.Error())
			} else {
				urlExtension = participantRiotAccount.GameName + "-" + participantRiotAccount.TagLine
				urlExtension = strings.ReplaceAll(urlExtension, " ", "%20")
			}

			var playedChamp datadragon.ChampionData

			for _, champion := range champs {
				if champion.Key == strconv.Itoa(participant.ChampionID) {
					playedChamp = champion
				}
			}

			if err != nil {
				champName = "unknown"
				slog.Error("get participant played champion", "err:", err.Error())
			} else {
				champName = playedChamp.Name
			}

			res.WriteString(fmt.Sprintf(
				"`%s` - [%s#%s](https://www.op.gg/summoners/euw/%s) \n",
				champName,
				participantRiotAccount.GameName,
				participantRiotAccount.TagLine,
				urlExtension,
			))
		}
	}

	return res.String()
}

func secToMinutesAndSec(seconds int) string {
	minutes := seconds / 60
	sec := seconds % 60

	return fmt.Sprintf("%d:%02d", minutes, sec)
}
