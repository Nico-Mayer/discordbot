const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')

function getRandomChamp(champs) {
	return champs[Math.floor(Math.random() * champs.length)]
}

module.exports = {
	data: new SlashCommandBuilder()
		.setName('random_champ')
		.setDescription('Get a random lol champion'),

	run: async ({ interaction, champs }) => {
		let embed = new EmbedBuilder()
		let champ = getRandomChamp(champs)
		embed.setColor(0x0099ff)
		embed.setTitle(champ.name)
		embed.setDescription(champ.title)
		embed.setThumbnail(
			`http://ddragon.leagueoflegends.com/cdn/13.4.1/img/champion/${champ.image.full}`
		)

		return interaction.editReply({ embeds: [embed] })
	},
}
