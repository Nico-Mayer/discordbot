import type { Command } from "@types"
import { pauseCmd } from "./music/pause"
import { playCmd } from "./music/play"

export const commands = new Map<string, Command>()

// Commands
commands.set(playCmd.metadata.name, playCmd)
commands.set(pauseCmd.metadata.name, pauseCmd)
