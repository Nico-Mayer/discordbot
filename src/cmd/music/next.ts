import { formatDuration, getYtThumbnailUrl, userInVoiceAndGuild } from "@lib/utils"
import type { Command } from "@types"
import consola from "consola"
import { EmbedBuilder, MessageFlags, SlashCommandBuilder } from "discord.js"
import type { Track } from "lavalink-client"

const meta = new SlashCommandBuilder()
	.setName("next")
	.setDescription("Springt zum nächsten Song in der Warteschlange")

export const nextCmd: Command = {
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
		consola.log(`[next] User ${interaction.user.username} requested next in guild ${interaction.guildId}.`)
		await interaction.deferReply()

		const lavalink = interaction.client.lavalink
		const player = lavalink.getPlayer(guildId)

		if (!player || !player.playing) {
			consola.log(`[next] No active player or nothing playing in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Aktuell wird nichts abgespielt.",
			})
		}

		if (player.voiceChannelId !== voiceId) {
			consola.log(
				`[next] User ${interaction.user.username} requested next outside of the bot's voice channel in guild ${interaction.guildId}.`,
			)
			return interaction.editReply({
				content: "Du bist nicht im selben Voicechannel wie ich du clown.",
			})
		}

		if (!player.queue.tracks.length) {
			consola.log(`[next] No more tracks in queue in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Es gibt keine weiteren Songs in der Warteschlange.",
			})
		}

		const nextTrack = player.queue.tracks[0]
		if (!nextTrack) {
			consola.log(`[next] Next track is undefined in guild ${interaction.guildId}.`)
			return interaction.editReply({
				content: "Es gibt keine weiteren Songs in der Warteschlange.",
			})
		}

		await player.skip()
		consola.log(`[next] Skipped to next track: ${nextTrack.info.title} in guild ${interaction.guildId}.`)

		const embed = new EmbedBuilder()
			.setColor(0x5865f2)
			.setAuthor({ name: "⏭️ Nächster Song", iconURL: interaction.client.user.displayAvatarURL() })
			.setTitle(nextTrack.info.title)
			.setURL(nextTrack.info.uri || "")
			.setDescription(`von **${nextTrack.info.author}**`)
			.setImage(getYtThumbnailUrl(nextTrack as Track))
			.addFields({
				name: "⏱️ Dauer",
				value: formatDuration(nextTrack.info.duration || 0),
				inline: true,
			})
			.setFooter({
				text: `Angefordert von ${interaction.user.username}`,
				iconURL: interaction.user.displayAvatarURL(),
			})

		return interaction.editReply({ embeds: [embed] })
	},
}
