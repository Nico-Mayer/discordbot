package utils

import (
	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func CreateFollowupMessage(event *events.ApplicationCommandInteractionCreate, msg discord.MessageCreate) {
	_, err := event.Client().Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(), msg)
	if err != nil {
		log.Error("creating followup message", "err", err)
	}
}
