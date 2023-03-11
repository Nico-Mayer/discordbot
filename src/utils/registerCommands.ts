import { MyClient } from '../types/myClient'
import { Collection, Events } from 'discord.js'
import { Command } from '../types/command'

export async function registerCommands(client: MyClient, commands: Command[]) {
	client.commands = new Collection()

	commands.forEach((command) => {
		if ('data' in command && 'execute' in command)
			client.commands?.set(command.data.name, command)
		else console.error(`Command is missing 'data' or 'execute': ${command}`)
	})

	client.on(Events.InteractionCreate, async (interaction) => {
		if (!interaction.isChatInputCommand()) return

		const clientInteraction = interaction.client as MyClient
		const command = clientInteraction.commands?.get(interaction.commandName)

		if (!command) {
			console.error(
				`No command matching ${interaction.commandName} was found.`
			)
			return
		}

		try {
			await command.execute(interaction)
		} catch (error) {
			console.error(error)
			if (interaction.replied || interaction.deferred) {
				await interaction.followUp({
					content: 'There was an error while executing this command!',
					ephemeral: true,
				})
			} else {
				await interaction.reply({
					content: 'There was an error while executing this command!',
					ephemeral: true,
				})
			}
		}
	})
}
