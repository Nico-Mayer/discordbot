package cmd

import (
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nico-mayer/discordbot/cmd/music"
)

type Cmd struct {
	Meta       discord.SlashCommandCreate
	Collection string
	Handler    func(event *events.ApplicationCommandInteractionCreate) error
}

var commands []*Cmd

func init() {
	registerCommand(music.PlayCmdMeta, music.PlayCmdHandler)
}

func registerCommand(meta discord.SlashCommandCreate, handler func(event *events.ApplicationCommandInteractionCreate) error) {
	var cmd = &Cmd{
		Meta:    meta,
		Handler: handler,
	}

	commands = append(commands, cmd)
}

func GetAll() []*Cmd {
	return commands
}

func GetCmdByName(name string) *Cmd {
	for _, cmd := range commands {
		if strings.EqualFold(cmd.Meta.Name, name) {
			return cmd
		}
	}
	return nil
}

func RegisterSlashCommands(client bot.Client, guildID snowflake.ID) {
	metadata := make([]discord.ApplicationCommandCreate, len(commands))
	for i, cmd := range commands {
		metadata[i] = cmd.Meta
	}

	if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), guildID, metadata); err != nil {
		slog.Error("error while registering commands", slog.Any("err", err))
	}
}
