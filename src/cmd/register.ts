import config from "@config/env"
import consola from "consola"
import { REST, Routes } from "discord.js"
import { commands } from "./command"

const rest = new REST().setToken(config.TOKEN)

export async function registerCommands() {
	// await resetCommands()

	for (const cmd of Array.from(commands.values())) {
		await rest.put(Routes.applicationCommands(config.APP_ID), {
			body: [cmd.Metadata.toJSON()],
		})
	}

	consola.success("Commands registriert!")
}

async function _resetCommands() {
	rest
		.put(Routes.applicationGuildCommands(config.APP_ID, config.SERVER_ID), {
			body: [],
		})
		.then(() => consola.info("Successfully deleted all guild commands."))
		.catch(consola.error)
	rest
		.put(Routes.applicationCommands(config.APP_ID), { body: [] })
		.then(() => consola.info("Successfully deleted all application commands."))
		.catch(consola.error)
}
