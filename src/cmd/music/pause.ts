import { userInVoiceAndGuild } from "@lib/utils"
import type { Command } from "@types"
import consola from "consola"
import { EmbedBuilder, MessageFlags, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder().setName("pause").setDescription("Pausiert die Wiedergabe")

const pauseCmd: Command = {
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

		consola.log(`[pause] User ${interaction.user.username} requested pause in guild ${interaction.guildId}.`)
		await interaction.deferReply()

		const lavalink = interaction.client.lavalink
		const player = lavalink.getPlayer(guildId)
		if (!player || !player.playing) {
			consola.log(`[pause] No active player or nothing playing in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Aktuell wird nichts abgespielt.",
			})
		}
		if (player.voiceChannelId !== voiceId) {
			consola.log(
				`[pause] User ${interaction.user.username} requested pause outside of the bot's voice channel in guild ${interaction.guildId}.`,
			)
			return interaction.editReply({
				content: "Du bist nicht im selben Voicechannel wie ich du clown.",
			})
		}
		await player.pause()
		consola.log(`[pause] Paused playback in guild ${interaction.guildId}.`)

		const embed = new EmbedBuilder()
			.setColor(0xfee75c)
			.setAuthor({ name: "⏸️ Wiedergabe pausiert", iconURL: interaction.client.user.displayAvatarURL() })
			.setDescription(`Die Wiedergabe wurde pausiert.`)
			.setFooter({
				text: `Angefordert von ${interaction.user.username}`,
				iconURL: interaction.user.displayAvatarURL(),
			})

		return interaction.editReply({ embeds: [embed] })
	},
}

export default pauseCmd
