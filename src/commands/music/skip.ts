import { useMasterPlayer } from 'discord-player'
import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import keys from '../../keys'

export default {
	data: new SlashCommandBuilder()
		.setName('skip')
		.setDescription('Überspringt den aktuellen Song'),
	async execute(interaction: ChatInputCommandInteraction) {
		const player = useMasterPlayer()
		const queue = player?.nodes.get(keys.guildId)

		if (!queue) return interaction.reply('Keine Songs in der Warteschlange')
		if (!queue.node.isPlaying())
			return interaction.reply('Kein Song wird abgespielt')
		if (queue.tracks.toArray().length < 1)
			return interaction.reply('Kein Song in der Warteschlange')
		else {
			const embed = new EmbedBuilder().setTitle(':fast_forward: - Skip')
			embed.setColor('#ff0000')
			embed.setDescription(
				`\`${queue.currentTrack?.title}\` wurde übersprungen`
			)
			queue.node.skip()
			await interaction.reply({ embeds: [embed] })
		}
	},
}
