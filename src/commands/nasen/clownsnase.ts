import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import keys from '../../keys'
export default {
	data: new SlashCommandBuilder()
		.setName('clownsnase')
		.setDescription('Gib einem Bre eine Clownsnase!')
		.addUserOption((option) => {
			return option
				.setName('bre')
				.setDescription('Kassiert eine Clownsnase!')
				.setRequired(true)
		}),

	async execute(interaction: ChatInputCommandInteraction) {
		const embed = new EmbedBuilder()
		const target = interaction.options.getUser('bre', true)
		const author = interaction.user
		const url = keys.nasenUrl

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

			return interaction.reply({ embeds: [embed] })
		} catch (err) {
			console.log(err)
			return interaction.reply('Fehler beim Clownsnasen verteilen')
		}
	},
}
