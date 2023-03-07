const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('nasen')
		.setDescription('Zeit das Clownsnasen Leaderboard!'),
	run: async ({ interaction }) => {
		let embed = new EmbedBuilder()
		let url = false
			? 'http://localhost:3333'
			: 'https://expressjs-postgres-production-733d.up.railway.app'

		let res = []
		try {
			res = await fetch(url + '/stats').then((res) => {
				if (res.ok) {
					return res.json()
				}
			})
			embed.setTitle('Clownsnasen Leaderboard  🤡')
			embed.setColor('#fff000')
			if (res.length === 0) {
				embed.setDescription(
					'Bis jetzt hat noch niemand eine Clownsnase, verteile clownsnasen mit `/clownsnase`'
				)
			} else {
				embed.setThumbnail(
					'https://media.tenor.com/81u64lUzA_QAAAAi/clown-peepo.gif'
				)
				embed.setDescription(buildLeaderboard(res))
			}

			return interaction.editReply({ embeds: [embed] })
		} catch (err) {
			return interaction.editReply('Something went wrong')
		}
	},
}

function buildLeaderboard(data) {
	let sortedData = data.sort((a, b) => b.nasen - a.nasen)

	let leaderboard = ''
	sortedData.forEach((user, index) => {
		leaderboard += `**${index + 1}.** <@${user.id}> = ${user.nasen} Nasen`
	})
	return leaderboard
}
