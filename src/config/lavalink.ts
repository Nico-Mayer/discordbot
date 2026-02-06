import type { Client } from "discord.js"
import { LavalinkManager } from "lavalink-client"
import config from "./env"

export function generateLavalinkClient(botClient: Client) {
	return new LavalinkManager({
		nodes: [
			{
				authorization: config.LAVALINK_PASSWORD,
				host: config.LAVALINK_HOST,
				port: config.LAVALINK_PORT,
				id: "Main Node",
			},
		],
		// A function to send voice server updates to the Lavalink client
		sendToShard: (guildId, payload) => {
			const guild = botClient.guilds.cache.get(guildId)
			if (guild) guild.shard.send(payload)
		},
		autoSkip: true,
		client: {
			id: config.APP_ID,
			username: "MyBot",
		},
	})
}
