import { readdirSync, statSync } from "node:fs"
import path from "node:path"
import type { Command, CommandCollection } from "@types"
import consola from "consola"

export const commandCollections = new Map<string, CommandCollection>()
export const allCommands = new Map<string, Command>()

async function loadCommands() {
	const folders = readdirSync(__dirname)

	for (const entry of folders) {
		const entryPath = path.join(__dirname, entry)
		const entryStat = statSync(entryPath)

		if (entryStat.isFile()) continue

		const collectionFiles = readdirSync(entryPath)
		const col: CommandCollection = {
			name: entry,
			description: "", // TODO: currently no use maybe fill later
			commands: new Map<string, Command>(),
		}

		for (const fileName of collectionFiles) {
			if (!fileName.endsWith(".ts")) continue

			const filePath = path.join(entryPath, fileName)
			const fileStat = statSync(filePath)

			if (!fileStat.isFile()) continue

			try {
				const module = await import(filePath)
				const cmd = module.default as Command

				if (cmd && "metadata" in cmd && "handler" in cmd) {
					if (allCommands.has(cmd.metadata.name)) {
						consola.error(
							`Command ${cmd.metadata.name} (${entry}/${fileName}), seems to be a duplicate, skipping addition`,
						)
						continue
					}
					col.commands.set(cmd.metadata.name, cmd)
					allCommands.set(cmd.metadata.name, cmd)
					consola.debug(`Loaded command: ${cmd.metadata.name} from ${entry}/${fileName}`)
				} else {
					consola.warn(`File ${entry}/${fileName} does not export a valid command, skip registering`)
				}
			} catch (error) {
				consola.error(`Failed to load command from ${entry}/${fileName}:`, error)
			}
		}

		if (col.commands.size > 0) {
			commandCollections.set(col.name, col)
			consola.debug(`Collection "${entry}" loaded with ${col.commands.size} command(s)`)
		}
	}
}

await loadCommands()
