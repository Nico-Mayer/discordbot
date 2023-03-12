import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import { CnUser } from '../types'

export default {
	data: new SlashCommandBuilder()
		.setName('nasen')
		.setDescription('Zeigt Clownsnasen Leaderboard!'),
	async execute(interaction: ChatInputCommandInteraction) {
		const embed = new EmbedBuilder()
		const url = 'https://expressjs-postgres-production-733d.up.railway.app'

		try {
			const res = await fetch(url + '/stats').then((res) => {
				if (res.ok) {
					return res.json()
				}
			})
			embed.setColor('#fff000')
			embed.setTitle('Clownsnasen Leaderboard  🤡')
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
			return interaction.reply({ embeds: [embed] })
		} catch (err) {
			console.log(err)
			return interaction.reply('Fehler beim abrufen der Clownsnasen')
		}
	},
}

function buildLeaderboard(data: CnUser[]) {
	const sortedData = data.sort((a, b) => b.nasen - a.nasen)

	let leaderboard = ''
	sortedData.forEach((user, index) => {
		leaderboard += `**${index + 1}.** <@${user.id}> = ${user.nasen} Nasen\n`
	})
	return leaderboard
}
