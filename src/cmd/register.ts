import config from "@config/env"
import consola from "consola"
import { REST, Routes } from "discord.js"
import { commands } from "./command"

const rest = new REST().setToken(config.TOKEN)

export async function registerCommands() {
	consola.info(`Registering ${commands.size} command(s)`)

	commands.forEach((cmd) => {
		consola.log(`   - ${cmd.metadata.name}`)
	})

	await rest.put(Routes.applicationCommands(config.APP_ID), {
		body: Array.from(commands.values()).map((cmd) => cmd.metadata.toJSON()),
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
