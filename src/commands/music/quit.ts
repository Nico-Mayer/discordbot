import { SlashCommandBuilder, ChatInputCommandInteraction } from 'discord.js'
import { useMasterPlayer } from 'discord-player'
import keys from '../../keys'

export default {
	data: new SlashCommandBuilder()
		.setName('quit')
		.setDescription('Stoppt den Musik Bot'),
	async execute(interaction: ChatInputCommandInteraction) {
		const player = useMasterPlayer()
		const queue = player?.nodes.get(keys.guildId)
		if (!queue)
			return await interaction.reply(
				'Aktuell kein song in der Warteschlange'
			)
		queue.delete()
		await interaction.reply('Tschüss! Euer DJ Rosine')
	},
}
