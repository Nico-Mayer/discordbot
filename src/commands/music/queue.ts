import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import keys from '../../keys'
import { useMasterPlayer } from 'discord-player'

export default {
	data: new SlashCommandBuilder()
		.setName('queue')
		.setDescription('Zeigt alle Songs in der Warteschlange')
		.addNumberOption((option) =>
			option
				.setName('page')
				.setDescription('Seite der Warteschlange')
				.setMinValue(1)
		),
	async execute(interaction: ChatInputCommandInteraction) {
		const player = useMasterPlayer()
		const queue = player?.nodes.get(keys.guildId)
		if (!queue || !queue.node.isPlaying()) {
			return await interaction.reply('Keine Songs in der Warteschlange')
		}
		const totalPages = Math.ceil(queue.tracks.size / 10) || 1
		const page = (interaction.options.getNumber('page') || 1) - 1

		if (page > totalPages) {
			return await interaction.reply(
				`Warteschlange hat nur ${totalPages} Seiten`
			)
		}

		const queueString = queue.tracks
			.toArray()
			.slice(page * 10, page * 10 + 10)
			.map((song, index) => {
				return `\n**${page * 10 + index + 1}.** \`[${
					song.duration
				}]\` ${song.title} -- <@${song.requestedBy?.id}>`
			})

		const currentSong = queue.currentTrack

		await interaction.reply({
			embeds: [
				new EmbedBuilder()
					.setDescription(
						`**Spielt gerade**\n` +
							(currentSong
								? `\`[${currentSong.duration}]\` ${currentSong.title} -- <@${currentSong.requestedBy?.id}>`
								: 'None') +
							`\n\n**Warteschlange**\n${queueString}`
					)
					.setFooter({
						text: `Seite ${page + 1} of ${totalPages}`,
					})
					.setColor('#00FF00')

					.setThumbnail(currentSong?.thumbnail ?? ''),
			],
		})
	},
}
