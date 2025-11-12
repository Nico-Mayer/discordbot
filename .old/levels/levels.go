package levels

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nico-mayer/discordbot/db"
)

const (
	EXP_PER_MESSAGE         int = 10
	EXP_PER_VOICE_JOIN      int = 25
	EXP_PER_SLASH_COMMAND   int = 5
	EXP_NEEDED_FOR_LEVEL_UP int = 250
)

var levelMapping = map[int]string{
	5:  "1196937224634257479",
	15: "1196936971776438302",
	25: "1196936450306998322",
	40: "696672218117177437",
	60: "696671960507351100",
}

func GrantExpToUser(botClient bot.Client, dbUser db.DBUser, exp int) error {
	err := dbUser.GrantExp(exp)
	if err != nil {
		return err
	}

	oldLevel := CalcUserLevel(dbUser.Exp)
	newLevel := CalcUserLevel(dbUser.Exp + exp)
	levelUp := oldLevel != newLevel

	if levelUp {
		handleLevelUp(botClient, dbUser.ID, newLevel)
	}

	return err
}

func CalcUserLevel(exp int) int {
	if exp < EXP_NEEDED_FOR_LEVEL_UP {
		return 1
	}

	return exp / EXP_NEEDED_FOR_LEVEL_UP
}

func handleLevelUp(botClient bot.Client, userId snowflake.ID, level int) {
	newRank := levelMapping[level]
	if newRank != "" {
		err := botClient.Rest().AddMemberRole(snowflake.GetEnv("GUILD_ID"), userId, snowflake.MustParse(newRank))
		if err != nil {
			slog.Error("giving role to user", "err:", err.Error())
		}
		return
	}
}
