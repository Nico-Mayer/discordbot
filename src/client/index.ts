import { Client, Collection, Events, GatewayIntentBits } from 'discord.js'
import keys from '../keys'
import commands from '../commands'

interface MyClient extends Client {
	commands?: Collection<string, any>
}

const client: MyClient = new Client({
	intents: [
		GatewayIntentBits.Guilds,
		GatewayIntentBits.GuildVoiceStates,
		GatewayIntentBits.GuildMembers,
	],
})

registerCommands()

client.once(Events.ClientReady, (c) => {
	console.log(`Ready! Logged in as ${c.user.tag}`)
})

client.login(keys.clientToken).catch((error) => {
	console.error('[Login Error]', error)
	process.exit(1)
})

function registerCommands() {
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
