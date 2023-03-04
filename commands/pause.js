const { SlashCommandBuilder } = require('discord.js')
const { useMasterPlayer } = require('discord-player')
const player = useMasterPlayer()

module.exports = {
	data: new SlashCommandBuilder()
		.setName('pause')
		.setDescription('Pauses the current song'),
	run: async ({ interaction }) => {
		const queue = player.nodes.get(interaction.guildId)
		if (!queue || !queue.node.isPlaying()) {
			return await interaction.editReply('There is no song playing')
		}
		if (queue.node.isPaused()) {
			return await interaction.editReply('The song is already paused')
		}
		queue.node.pause()
		await interaction.editReply(
			':play_pause:  Song has been paused use `/resume` to resume the song'
		)
	},
}
