import { userInVoiceAndGuild } from "@lib/utils"
import type { Command } from "@types"
import consola from "consola"
import { MessageFlags, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder().setName("pause").setDescription("Pausiert die Wiedergabe")

export const pauseCmd: Command = {
	metadata: meta,
	handler: async (interaction) => {
		const { ok, errorMsg, guildId, voiceId } = userInVoiceAndGuild(interaction)
		if (!ok) {
			return interaction.reply({
				content: errorMsg,
				flags: [MessageFlags.Ephemeral],
			})
		}

		consola.log(`[pause] User ${interaction.user.username} requested pause in guild ${interaction.guildId}.`)
		await interaction.deferReply()

		const lavalink = interaction.client.lavalink
		const player = lavalink.getPlayer(guildId)
		if (!player || !player.playing) {
			consola.warn(`[pause] Went wrong nothing is playing`)
			return interaction.editReply({
				content: "Aktuell wird nichts abgespielt.",
			})
		}
		if (player.voiceChannelId !== voiceId) {
			consola.warn(
				`[pause] User ${interaction.user.username} requested pause outside of the bots voicechannel`,
			)
			return interaction.editReply({
				content: "Du bist nicht im selben Voicechannel wie ich du clown.",
			})
		}
		consola.log(`[pause] pausing playback`)
		await player.pause()
	},
}
