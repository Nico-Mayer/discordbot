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

export function userInVoiceAndGuild(interaction: ChatInputCommandInteraction) {
	type resp = {
		ok: boolean
		errorMsg: string
		guildId: string
		voiceId: string
	}
	let result: resp
	if (!interaction.guildId || interaction.guildId !== config.SERVER_ID) {
		consola.warn(`[${interaction.commandName}] No guildId in interaction or not in the same guild.`)
		result = { ok: false, errorMsg: "Du bist nicht auf dem Richtigen server!", guildId: "", voiceId: "" }
		return result
	}

	const sender = interaction.member as GuildMember
	const voiceChannelId = sender.voice.channelId
	if (!voiceChannelId) {
		consola.log(`[${interaction.commandName}] User ${interaction.user.username} not in a voice channel.`)
		result = { ok: false, errorMsg: "Du musst in einem Voice Channel sein!", guildId: "", voiceId: "" }
		return result
	}
	result = { ok: true, errorMsg: "", guildId: interaction.guildId, voiceId: voiceChannelId }
	return result
}
