const Discord = require('discord.js')
const { GatewayIntentBits, Events } = require('discord.js')
const dotenv = require('dotenv')
const { REST } = require('@discordjs/rest')
const { Routes } = require('discord-api-types/v10')
const fs = require('fs')
const { Player } = require('discord-player')

dotenv.config()
const TOKEN = process.env.DISCORD_TOKEN
const CLIENT_ID = process.env.CLIENT_ID
const GUILD_ID = process.env.GUILD_ID
const LOAD_SLASH = process.argv[2] === 'load'
const fetchChamps = require('./utils/champs')
let champs = []
;(async () => {
	champs = Object.values(await fetchChamps())
})()

const client = new Discord.Client({
	intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildVoiceStates],
})
new Player(client, {
	ytdlOptions: {
		filter: 'audioonly',
		highWaterMark: 1 << 30,
		dlChunkSize: 0,
	},
})

client.slashcommands = new Discord.Collection()

let commands = []
const commandFiles = fs
	.readdirSync('./commands')
	.filter((file) => file.endsWith('.js'))
for (const file of commandFiles) {
	let slashcmd = require(`./commands/${file}`)
	client.slashcommands.set(slashcmd.data.name, slashcmd)
	if (LOAD_SLASH) commands.push(slashcmd.data.toJSON())
}

if (LOAD_SLASH) {
	loadCommands()
} else {
	setupHandlers()
	client.login(TOKEN)
}

function loadCommands() {
	const rest = new REST({ version: '10' }).setToken(TOKEN)
	console.log('Deploying slash commands...')
	rest.put(Routes.applicationGuildCommands(CLIENT_ID, GUILD_ID), {
		body: commands,
	})
		.then(() => {
			console.log('Successfully registered application commands.')
			process.exit(0)
		})
		.catch((err) => {
			console.error(err)
			process.exit(1)
		})
}

function setupHandlers() {
	client.once(Events.ClientReady, (c) => {
		console.log(`Ready! Logged in as ${c.user.tag}`)
	})
	client.on(Events.InteractionCreate, (interaction) => {
		async function handleCommand() {
			if (!interaction.isCommand()) return
			const slashcmd = client.slashcommands.get(interaction.commandName)
			if (!slashcmd) interaction.reply('Command not found')

			await interaction.deferReply()
			await slashcmd.run({ client, interaction, champs })
		}
		handleCommand()
	})
}
