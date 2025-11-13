package cmd

import (
	"github.com/charmbracelet/log"
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

var commands map[string]*Cmd = make(map[string]*Cmd)

func init() {
	registerCommand(music.PlayCmdMeta, music.PlayCmdHandler)
}

func registerCommand(meta discord.SlashCommandCreate, handler func(event *events.ApplicationCommandInteractionCreate) error) {
	var cmd = &Cmd{
		Meta:    meta,
		Handler: handler,
	}

	commands[meta.Name] = cmd
}

func GetAll() map[string]*Cmd {
	return commands
}

func RegisterSlashCommands(client bot.Client, guildID snowflake.ID) {
	log.Info("start command registration")
	metadata := make([]discord.ApplicationCommandCreate, 0, len(commands))
	for _, cmd := range commands {
		log.Info("register cmd", "name", cmd.Meta.Name)
		metadata = append(metadata, cmd.Meta)
	}

	if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), guildID, metadata); err != nil {
		log.Error("while registering commands", "error", err.Error())
	}
}
