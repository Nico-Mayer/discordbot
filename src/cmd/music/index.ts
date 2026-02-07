import type { Command, CommandCollection } from "@types"
import nextCmd from "./next"
import pauseCmd from "./pause"
import playCmd from "./play"
import resumeCmd from "./resume"
import stopCmd from "./stop"

export const musicCollection: CommandCollection = {
	name: "Music",
	description: "",
	commands: new Map<string, Command>([
		[playCmd.metadata.name, playCmd],
		[nextCmd.metadata.name, nextCmd],
		[pauseCmd.metadata.name, pauseCmd],
		[resumeCmd.metadata.name, resumeCmd],
		[stopCmd.metadata.name, stopCmd],
	]),
}
