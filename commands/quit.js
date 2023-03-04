const { SlashCommandBuilder } = require('discord.js')
const { useMasterPlayer } = require('discord-player')
const player = useMasterPlayer()

module.exports = {
	data: new SlashCommandBuilder()
		.setName('quit')
		.setDescription('Stops the Bot and clears the queue'),
	run: async ({ interaction }) => {
		const queue = player.nodes.get(interaction.guildId)
		if (!queue)
			return await interaction.editReply('There is no songs in the queue')
		queue.delete()
		await interaction.editReply('Tschüss! Euer DJ Rosine')
	},
}
