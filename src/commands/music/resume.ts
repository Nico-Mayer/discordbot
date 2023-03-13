import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import { useMasterPlayer } from 'discord-player'
import keys from '../../keys'

export default {
	data: new SlashCommandBuilder()
		.setName('resume')
		.setDescription('Setzt den Pausierten Song fort'),
	async execute(interaction: ChatInputCommandInteraction) {
		const player = useMasterPlayer()
		const queue = player?.nodes.get(keys.guildId)

		if (!queue)
			return await interaction.reply(
				'Aktuell kein Song in der Warteschlange'
			)

		if (queue.node.isPlaying()) {
			return await interaction.reply('Song wird bereits abgespielt')
		}
		queue.node.resume()
		const embed = new EmbedBuilder().setTitle(':play_pause: - Resume')
		embed.setColor('#ff0000')
		embed.setDescription(`Song wird fortgesetzt \`/pause\` zum pausieren`)
		await interaction.reply({ embeds: [embed] })
	},
}
