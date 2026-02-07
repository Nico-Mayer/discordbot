import config from "@config/env"
import consola from "consola"
import type { ChatInputCommandInteraction, GuildMember } from "discord.js"
import type { Track } from "lavalink-client"

export function getYtThumbnailUrl(track: Track) {
	return `https://img.youtube.com/vi/${track.info.identifier}/hqdefault.jpg`
}

// export function formatDuration(d: number) {
// 	const secondsTotal = d / 1000
// 	const minutes = Math.floor(secondsTotal / 60)
// 	const restSeconds = secondsTotal - minutes * 60

// 	return `${minutes}min ${restSeconds}s`
// }

export function formatDuration(ms: number): string {
	const seconds = Math.floor(ms / 1000)
	const minutes = Math.floor(seconds / 60)
	const hours = Math.floor(minutes / 60)

	const secs = seconds % 60
	const mins = minutes % 60

	const time =
		hours > 0
			? `${hours}:${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`
			: `${mins}:${secs.toString().padStart(2, "0")}`

	return `\`${time}\``
}

export function userInVoiceAndGuild(
	interaction: ChatInputCommandInteraction,
): { ok: true; guildId: string; voiceId: string } | { ok: false; errorMsg: string } {
	if (!interaction.guildId || interaction.guildId !== config.SERVER_ID) {
		consola.warn(`[${interaction.commandName}] No guildId in interaction or not in the same guild.`)
		return { ok: false, errorMsg: "Du bist nicht auf dem Richtigen server!" }
	}

	const sender = interaction.member as GuildMember
	const voiceChannelId = sender.voice.channelId
	if (!voiceChannelId) {
		consola.log(`[${interaction.commandName}] User ${interaction.user.username} not in a voice channel.`)
		return { ok: false, errorMsg: "Du musst in einem Voice Channel sein!" }
	}

	return { ok: true, guildId: interaction.guildId, voiceId: voiceChannelId }
}
