import { Player } from 'discord-player'
import { Client, Events, GatewayIntentBits } from 'discord.js'
import commands from '../commands'
import keys from '../keys'
import { MyClient } from '../types'
import { registerCommands } from '../utils'

const client: MyClient = new Client({
	intents: [
		GatewayIntentBits.Guilds,
		GatewayIntentBits.GuildVoiceStates,
		GatewayIntentBits.GuildMembers,
	],
})

new Player(client, {
	ytdlOptions: {
		filter: 'audioonly',
		highWaterMark: 1 << 30,
		dlChunkSize: 0,
	},
})

registerCommands(client, commands)

client.once(Events.ClientReady, (c) => {
	console.log(`Ready! Logged in as ${c.user.tag}`)
})

client.login(keys.clientToken).catch((error) => {
	console.error('[Login Error]', error)
	process.exit(1)
})
