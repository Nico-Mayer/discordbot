import { Client, Events, GatewayIntentBits } from 'discord.js'
import commands from '../commands'
import keys from '../keys'
import { registerCommands } from '../utils'
import { MyClient } from '../types'

const client: MyClient = new Client({
	intents: [
		GatewayIntentBits.Guilds,
		GatewayIntentBits.GuildVoiceStates,
		GatewayIntentBits.GuildMembers,
	],
})

registerCommands(client, commands)

client.once(Events.ClientReady, (c) => {
	console.log(`Ready! Logged in as ${c.user.tag}`)
})

client.login(keys.clientToken).catch((error) => {
	console.error('[Login Error]', error)
	process.exit(1)
})
