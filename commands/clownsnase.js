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

	run: async ({ interaction }) => {
		let embed = new EmbedBuilder()
		const target = interaction.options.getUser('target')
		const author = interaction.user
		let url = false
			? 'http://localhost:3333'
			: 'https://expressjs-postgres-production-733d.up.railway.app'
		const res = await fetch(url + `/nasen?id=${target.id}`).then((res) => {
			if (res.ok) {
				return res.json()
			}
		})

		try {
			await fetch(url + '/add', {
				method: 'put',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					id: target.id,
					name: target.username,
					from: author.id,
				}),
			})
		} catch (err) {
			console.log(err)
		}

		embed.setColor('#fff000')
		embed.setTitle('Clownsnase Kassiert  🤡')
		embed.setThumbnail(target.displayAvatarURL())
		embed.addFields(
			{ name: 'Von', value: `<@${author.id}>`, inline: true },
			{ name: 'An', value: `<@${target.id}>`, inline: true },
			{ name: 'Total', value: `${res.count + 1}`, inline: true },
			{
				name: ' ',
				value: ` \`/nasen\` in den chat um leaderboard anzuzeigen `,
			}
		)

		return interaction.editReply({ embeds: [embed] })
	},
}
