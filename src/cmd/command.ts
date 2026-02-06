import type { Command } from "@types"
import { play } from "./music/play"

export const commands = new Map<string, Command>()

// Commands
commands.set(play.Metadata.name, play)
