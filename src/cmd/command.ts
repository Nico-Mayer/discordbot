import type { Command } from "@types"
import { playCmd } from "./music/play"

export const commands = new Map<string, Command>()

// Commands
commands.set(playCmd.metadata.name, playCmd)
