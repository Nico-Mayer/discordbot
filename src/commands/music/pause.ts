import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import { useMasterPlayer } from 'discord-player'
import keys from '../../keys'

export default {
	data: new SlashCommandBuilder()
		.setName('pause')
		.setDescription('Pausiert den Aktuellen Song'),
	async execute(interaction: ChatInputCommandInteraction) {
		const player = useMasterPlayer()
		const queue = player?.nodes.get(keys.guildId)
		if (!queue || !queue.node.isPlaying()) {
			return await interaction.reply(
				'Aktuell kein Song in der Warteschlange'
			)
		}
		if (queue.node.isPaused()) {
			return await interaction.reply('Song ist bereits pausiert')
		}
		queue.node.pause()
		const embed = new EmbedBuilder().setTitle(':pause_button: - Pause')
		embed.setColor('#ff0000')
		embed.setDescription(
			`Song pausiert \`/resume\` um wiedergabe fortzusetzen`
		)
		await interaction.reply({ embeds: [embed] })
	},
}
