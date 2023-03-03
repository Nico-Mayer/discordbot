const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('quit')
		.setDescription('Stops the Bot and clears the queue'),
	run: async ({ client, interaction }) => {
		const queue = client.player.getQueue(interaction.guildId)
		if (!queue)
			return await interaction.editReply('There is no songs in the queue')
		queue.destroy()
		await interaction.editReply('Tschüss! Euer DJ Rosine')
	},
}
