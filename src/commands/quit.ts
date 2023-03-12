import { SlashCommandBuilder, ChatInputCommandInteraction } from 'discord.js'
import { useMasterPlayer } from 'discord-player'
import keys from '../keys'

export default {
	data: new SlashCommandBuilder()
		.setName('quit')
		.setDescription('Stoppt den Musik Bot'),
	async execute(interaction: ChatInputCommandInteraction) {
		const player = useMasterPlayer()
		const queue = player?.nodes.get(keys.guildId)
		if (!queue)
			return await interaction.editReply('There is no songs in the queue')
		queue.delete()
		await interaction.editReply('Tschüss! Euer DJ Rosine')
	},
}
