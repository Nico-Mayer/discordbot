import type { Command, CommandCollection } from "@types"
import rollCmd from "./roll"

export const generalCollection: CommandCollection = {
	name: "General",
	description: "",
	commands: new Map<string, Command>([[rollCmd.metadata.name, rollCmd]]),
}
