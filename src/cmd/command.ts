import type { Command, CommandCollection } from "@types"
import { generalCollection } from "./general"
import { musicCollection } from "./music"

export const commandCollections = new Map<string, CommandCollection>()
export const allCommands = new Map<string, Command>()

function loadCommands() {
	commandCollections.set(generalCollection.name, generalCollection)
	commandCollections.set(musicCollection.name, musicCollection)

	for (const collection of commandCollections.values()) {
		for (const [name, cmd] of collection.commands) {
			allCommands.set(name, cmd)
		}
	}
}

loadCommands()
