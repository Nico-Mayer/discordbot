import config from "@config/env"
import { generateLavalinkClient } from "@config/lavalink"
import consola from "consola"
import { Client, Events, GatewayIntentBits } from "discord.js"
import type { LavalinkManager } from "lavalink-client"
import p from "../package.json"
import { commands } from "./cmd/command"
import { registerCommands } from "./cmd/register"

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

await registerCommands()

// Events
client.on(Events.Raw, (d) => client.lavalink.sendRawData(d))
client.once(Events.ClientReady, (c) => {
	consola.success(`Logged in as ${c.user.tag}`)
	client.lavalink.init({ ...c.user })
})
client.on("interactionCreate", async (interaction) => {
	if (!interaction.isChatInputCommand()) return
	if (interaction.commandName === "play") {
		commands.get("play")?.Handler(interaction)
	}
})

client.lavalink.nodeManager.on("connect", (node) => {
	consola.success(`Lavalink Node "${node.id}" connected!`)
})

// Bot start
client.login(config.TOKEN)
