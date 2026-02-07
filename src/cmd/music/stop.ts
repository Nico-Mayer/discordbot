import { userInVoiceAndGuild } from "@lib/utils"
import type { Command } from "@types"
import consola from "consola"
import { EmbedBuilder, MessageFlags, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder()
	.setName("stop")
	.setDescription("Stopt die Wiedergabe und Leert die Warteschlange")

export const stopCmd: Command = {
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
		consola.log(`[stop] User ${interaction.user.username} requested stop in guild ${interaction.guildId}.`)
		await interaction.deferReply()

		const lavalink = interaction.client.lavalink
		const player = lavalink.getPlayer(guildId)

		if (!player) {
			consola.log(`[stop] No player found in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Aktuell gibt es keinen Musikplayer.",
			})
		}
		if (player.voiceChannelId !== voiceId) {
			consola.log(
				`[stop] User ${interaction.user.username} requested stop outside of the bot's voice channel in guild ${interaction.guildId}.`,
			)
			return interaction.editReply({
				content: "Du bist nicht im selben Voicechannel wie ich du clown.",
			})
		}
		await player.stopPlaying()
		consola.log(`[stop] Stopped playback and cleared queue in guild ${interaction.guildId}.`)

		const embed = new EmbedBuilder()
			.setColor(0xed4245)
			.setAuthor({ name: "⏹️ Wiedergabe gestoppt", iconURL: interaction.client.user.displayAvatarURL() })
			.setDescription(`Die Wiedergabe wurde gestoppt und die Warteschlange wurde geleert.`)
			.setFooter({
				text: `Angefordert von ${interaction.user.username}`,
				iconURL: interaction.user.displayAvatarURL(),
			})

		return interaction.editReply({ embeds: [embed] })
	},
}
