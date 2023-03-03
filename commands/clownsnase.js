const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('clownsnase')
		.setDescription('Gib einem Bre eine Clownsnase!')
		.addUserOption((option) => {
			return option
				.setName('target')
				.setDescription('Kassiert eine Clownsnase!')
				.setRequired(true)
		}),

	run: async ({ interaction, champs }) => {
		let embed = new EmbedBuilder()
		const target = interaction.options.getUser('target')

		embed.setColor('#fff000')
		embed.setTitle('Clownsnase 🤡')
		embed.setDescription(`${target} hat eine Clownsnase kassiert!`)

		return interaction.editReply({ embeds: [embed] })
	},
}
