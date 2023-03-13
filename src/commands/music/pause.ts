import { SlashCommandBuilder, ChatInputCommandInteraction } from 'discord.js'
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
		await interaction.reply(
			':play_pause:  Song pausiert `/resume` um wiedergabe fortzusetzen'
		)
	},
}
