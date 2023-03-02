const Discord = require('discord.js')
const dotenv = require('dotenv')
const { Rest } = require('@discordjs/rest')
const { Routes } = require('discord-api-types/v9')
const fs = require('fs')
const { Player } = require('discord-player')

dotenv.config()
const TOKEN = process.env.DISCORD_TOKEN
const CLIENT_ID = process.env.CLIENT_ID
const GUILD_ID = process.env.GUILD_ID
const LOAD_SLASH = process.argv[2] === 'load'

const client = new Discord.Client({ intents: ['GUILDS', 'GUILD_VOICE_STATES'] })
client.player = new Player(client, {
	ytdlOptions: {
		quality: 'highestaudio',
		highWaterMark: 1 << 25,
	},
})
client.slashcommands = new Discord.Collection()
