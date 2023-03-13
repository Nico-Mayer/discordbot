import { SlashCommandBuilder, ChatInputCommandInteraction } from 'discord.js'

export default {
	data: new SlashCommandBuilder()
		.setName('ping')
		.setDescription('Antwortet mit Pong!'),
	async execute(interaction: ChatInputCommandInteraction) {
		await interaction.reply('Pong!')
	},
}
