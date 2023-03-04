const { SlashCommandBuilder } = require('discord.js')
const { useMasterPlayer } = require('discord-player')
const player = useMasterPlayer()

module.exports = {
	data: new SlashCommandBuilder()
		.setName('resume')
		.setDescription('Resumes the current song'),
	run: async ({ interaction }) => {
		const queue = player.nodes.get(interaction.guildId)

		if (!queue)
			return await interaction.editReply('There is no song in queue')

		if (queue.node.isPlaying()) {
			return await interaction.editReply('The song is already playing')
		}
		queue.node.resume()
		await interaction.editReply(
			':play_pause:  Song has been resumed use `/pause` to pause the song'
		)
	},
}
