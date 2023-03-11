import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import test from 'node:test'
import { getChampions } from '../utils'

export default {
	data: new SlashCommandBuilder()
		.setName('random_champ')
		.setDescription('Get a random League of Legends champion'),
	async execute(interaction: ChatInputCommandInteraction) {
		const champions = await getChampions()
		const randomIndex = Math.floor(Math.random() * champions.length)
		const randomChampion = champions[randomIndex]

		const name = randomChampion.name
		const title = randomChampion.title
		const img = randomChampion.image.full
		const embed = new EmbedBuilder()
		embed.setColor(0x0099ff)
		embed.setTitle(name)
		embed.setDescription(title)
		embed.setThumbnail(
			`http://ddragon.leagueoflegends.com/cdn/13.4.1/img/champion/${img}`
		)

		await interaction.reply({ embeds: [embed] })
	},
}
