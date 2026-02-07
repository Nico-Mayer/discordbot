import { parseArgs } from "node:util"
import config from "@config/env"
import { generateLavalinkClient } from "@config/lavalink"
import consola from "consola"
import { Client, Events, GatewayIntentBits } from "discord.js"
import type { LavalinkManager } from "lavalink-client"
import p from "../package.json"
import { commands } from "./cmd/command"
import { registerCommands, resetCommands } from "./cmd/register"

const { values: flags } = parseArgs({
	args: Bun.argv,
	options: {
		debug: {
			type: "boolean",
			short: "d",
		},
		register: {
			type: "boolean",
			short: "r",
		},
		"reset-commands": {
			type: "boolean",
		},
	},
	strict: false,
	allowPositionals: true,
})

const discordjsVersion = p.dependencies["discord.js"]
consola.info("Running DiscordJs Version:", discordjsVersion)

// Extend the Client type to include the lavalink manager
declare module "discord.js" {
	interface Client {
		lavalink: LavalinkManager
	}
}

const client = new Client({
	intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildVoiceStates],
})
client.lavalink = generateLavalinkClient(client)

if (flags["reset-commands"]) {
	await resetCommands()
}
if (flags.register || flags["reset-commands"]) {
	await registerCommands()
}

// Events
client.on(Events.Raw, (d) => client.lavalink.sendRawData(d))
client.once(Events.ClientReady, (c) => {
	consola.success(`Logged in as ${c.user.tag}`)
	client.lavalink.init({ ...c.user })
})
client.on("interactionCreate", async (interaction) => {
	if (!interaction.isChatInputCommand()) return
	commands.get(interaction.commandName)?.handler(interaction)
})

client.lavalink.nodeManager.on("connect", (node) => {
	consola.success(`Lavalink Node "${node.id}" connected!`)
})
client.lavalink.on("queueEnd", async (player) => {
	consola.log("Playback ended, disconnecting player.")
	await player.destroy()
})

// Bot start
client.login(config.TOKEN)
