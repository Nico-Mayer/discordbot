import commands from '../commands'

import { REST, Routes } from 'discord.js'
import keys from '../keys'

const rest = new REST({ version: '10' }).setToken(keys.clientToken)
const body = commands.map((command) => command.data.toJSON())

async function main() {
	const endpoint = Routes.applicationGuildCommands(
		keys.clientId,
		keys.guildId
	)

	await rest.put(endpoint, { body })
}

main()
	.then(() => {
		const response = `Successfully registered ${commands.length} commands`
		console.log(response)
	})
	.catch(console.error)
