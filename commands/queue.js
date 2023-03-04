const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')
const { useMasterPlayer } = require('discord-player')
const player = useMasterPlayer()

module.exports = {
	data: new SlashCommandBuilder()
		.setName('queue')
		.setDescription('Displays the current song queue')
		.addNumberOption((option) =>
			option
				.setName('page')
				.setDescription('Page number of the queue')
				.setMinValue(1)
		),

	run: async ({ interaction }) => {
		const queue = player.nodes.get(interaction.guildId)
		if (!queue || !queue.node.isPlaying()) {
			return await interaction.editReply(
				'There are no songs in the queue'
			)
		}
		const totalPages = Math.ceil(queue.tracks.size / 10) || 1
		const page = (interaction.options.getNumber('page') || 1) - 1

		if (page > totalPages) {
			return await interaction.editReply(
				`The queue only has ${totalPages} pages`
			)
		}

		const queueString = queue.tracks
			.toArray()
			.slice(page * 10, page * 10 + 10)
			.map((song, index) => {
				return `\n**${page * 10 + index + 1}.** \`[${
					song.duration
				}]\` ${song.title} -- <@${song.requestedBy.id}>`
			})

		const currentSong = queue.currentTrack

		await interaction.editReply({
			embeds: [
				new EmbedBuilder()
					.setDescription(
						`**Currently Playing**\n` +
							(currentSong
								? `\`[${currentSong.duration}]\` ${currentSong.title} -- <@${currentSong.requestedBy.id}>`
								: 'None') +
							`\n\n**Queue**\n${queueString}`
					)
					.setFooter({
						text: `Page ${page + 1} of ${totalPages}`,
					})
					.setThumbnail(currentSong.setThumbnail),
			],
		})
	},
}
