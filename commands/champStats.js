const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('champ_stats')
		.setDescription('Display champion stats')
		.addStringOption((option) => {
			return option
				.setName('champion')
				.setDescription('Champion name')
				.setRequired(true)
		}),

	run: async ({ interaction, champs }) => {
		let embed = new EmbedBuilder()
		let input = interaction.options.getString('champion')

		let result = champs.filter((champ) => {
			return champ.id.toLowerCase() === input.toLowerCase()
		})

		if (result.length === 0)
			return interaction.editReply('Champion not found!')

		embed.setColor('#fff000')
		embed.setTitle(result[0].name)
		embed.setDescription(
			JSON.stringify(result[0].stats, null, 2).slice(1, -1)
		)
		embed.setThumbnail(
			`http://ddragon.leagueoflegends.com/cdn/13.4.1/img/champion/${result[0].image.full}`
		)

		return interaction.editReply({ embeds: [embed] })
	},
}
