import config from "@config/env"
import consola from "consola"
import { REST, Routes } from "discord.js"
import { allCommands, commandCollections } from "./command"

const rest = new REST().setToken(config.TOKEN)
export async function registerCommands() {
	commandCollections.forEach((collection) => {
		consola.info(`Registering Collection "${collection.name}" commands:`)
		collection.commands.forEach((command) => {
			consola.log(`   - ${command.metadata.name}`)
		})
	})

	consola.start(`Registering ${allCommands.size} command(s)`)

	await rest.put(Routes.applicationCommands(config.APP_ID), {
		body: Array.from(allCommands.values()).map((cmd) => cmd.metadata.toJSON()),
	})

	consola.success("Commands registriert!")
}

export async function resetCommands() {
	await rest
		.put(Routes.applicationGuildCommands(config.APP_ID, config.SERVER_ID), {
			body: [],
		})
		.then(() => consola.info("Successfully deleted all guild commands."))
		.catch(consola.error)
	await rest
		.put(Routes.applicationCommands(config.APP_ID), { body: [] })
		.then(() => consola.info("Successfully deleted all application commands."))
		.catch(consola.error)
}
