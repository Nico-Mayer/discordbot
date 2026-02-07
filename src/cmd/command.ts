import type { Command } from "@types"
import { pauseCmd } from "./music/pause"
import { playCmd } from "./music/play"
import { resumeCmd } from "./music/resume"
import { stopCmd } from "./music/stop"

export const commands = new Map<string, Command>()

// Commands
commands.set(playCmd.metadata.name, playCmd)
commands.set(pauseCmd.metadata.name, pauseCmd)
commands.set(resumeCmd.metadata.name, resumeCmd)
commands.set(stopCmd.metadata.name, stopCmd)
