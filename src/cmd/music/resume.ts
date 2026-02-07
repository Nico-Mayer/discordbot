import { userInVoiceAndGuild } from "@lib/utils"
import type { Command } from "@types"
import consola from "consola"
import { EmbedBuilder, MessageFlags, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder().setName("resume").setDescription("Wiedergabe fortsetzen")

export const resumeCmd: Command = {
	metadata: meta,
	handler: async (interaction) => {
		const validation = userInVoiceAndGuild(interaction)
		if (!validation.ok) {
			return interaction.reply({
				content: validation.errorMsg,
				flags: [MessageFlags.Ephemeral],
			})
		}

		const { guildId, voiceId } = validation
		consola.log(
			`[resume] User ${interaction.user.username} requested resume in guild ${interaction.guildId}.`,
		)
		await interaction.deferReply()

		const lavalink = interaction.client.lavalink
		const player = lavalink.getPlayer(guildId)

		if (!player) {
			consola.log(`[resume] No player found in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Aktuell gibt es keinen Musikplayer.",
			})
		}
		if (!player.paused) {
			consola.log(`[resume] Player is not paused in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Die Wiedergabe ist nicht pausiert.",
			})
		}
		if (player.voiceChannelId !== voiceId) {
			consola.log(
				`[resume] User ${interaction.user.username} requested resume outside of the bot's voice channel in guild ${interaction.guildId}.`,
			)
			return interaction.editReply({
				content: "Du bist nicht im selben Voicechannel wie ich du clown.",
			})
		}
		await player.resume()
		consola.log(`[resume] Resumed playback in guild ${interaction.guildId}.`)

		const embed = new EmbedBuilder()
			.setColor(0x57f287)
			.setAuthor({ name: "▶️ Wiedergabe fortgesetzt", iconURL: interaction.client.user.displayAvatarURL() })
			.setDescription(`Die Wiedergabe wurde fortgesetzt.`)
			.setFooter({
				text: `Angefordert von ${interaction.user.username}`,
				iconURL: interaction.user.displayAvatarURL(),
			})

		return interaction.editReply({ embeds: [embed] })
	},
}
