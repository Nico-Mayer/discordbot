const { SlashCommandBuilder } = require('discord.js')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('play')
		.setDescription('Plays song from youtube'),
	async execute(interaction) {
		await interaction.reply('Playing song!')
	},
}
