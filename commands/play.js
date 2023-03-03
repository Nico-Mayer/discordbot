const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')
const { QueryType } = require('discord-player')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('play')
		.setDescription('Play song from youtube')
		.addSubcommand((subcommand) => {
			return subcommand
				.setName('song')
				.setDescription('Loads a single song from youtube url')
				.addStringOption((option) =>
					option
						.setName('url')
						.setDescription('youtube url')
						.setRequired(true)
				)
		})
		.addSubcommand((subcommand) => {
			return subcommand
				.setName('playlist')
				.setDescription('Loads a playlist from youtube url')
				.addStringOption((option) =>
					option
						.setName('url')
						.setDescription('youtube url')
						.setRequired(true)
				)
		})
		.addSubcommand((subcommand) => {
			return subcommand
				.setName('search')
				.setDescription(
					'Searches a song from youtube based on provided keyword'
				)
				.addStringOption((option) =>
					option
						.setName('searchterms')
						.setDescription('search keywords')
						.setRequired(true)
				)
		}),
	run: async ({ client, interaction }) => {
		if (!interaction.member.voice.channel)
			return interaction.editReply(
				'You must be in a voice channel to use this command'
			)

		const queue = client.player.createQueue(interaction.guild)
		if (!queue.connection)
			await queue.connect(interaction.member.voice.channel)

		let embed = new EmbedBuilder()

		if (interaction.options.getSubcommand() === 'song') {
			let url = interaction.options.getString('url')
			const result = await client.player.search(url, {
				requestedBy: interaction.user,
				searchEngine: QueryType.YOUTUBE_VIDEO,
			})
			if (result.tracks.length < 1)
				return interaction.editReply('No results found')
			const song = result.tracks[0]
			await queue.addTrack(song)
			embed.setDescription(`Added ${song.title} (${song.url}) to queue`)
			embed.setThumbnail(song.thumbnail)
			embed.setFooter({ text: `Duration: ${song.duration}` })
		} else if (interaction.options.getSubcommand() === 'playlist') {
			let url = interaction.options.getString('url')
			const result = await client.player.search(url, {
				requestedBy: interaction.user,
				searchEngine: QueryType.YOUTUBE_PLAYLIST,
			})
			if (result.tracks.length < 1)
				return interaction.editReply('No results found')
			const playlist = result.playlist
			await queue.addTracks(result.tracks)
			embed.setDescription(
				`Added ${result.tracks.length} songs from ${playlist.title} (${playlist.url}) to queue`
			)
			embed.setThumbnail(playlist.thumbnail)
		} else if (interaction.options.getSubcommand() === 'search') {
			let url = interaction.options.getString('searchterms')
			const result = await client.player.search(url, {
				requestedBy: interaction.user,
				searchEngine: QueryType.AUTO,
			})
			if (result.tracks.length === 0)
				return interaction.editReply('No results found')
			const song = result.tracks[0]
			await queue.addTrack(song)
			embed.setDescription(`Added ${song.title} (${song.url}) to queue`)
			embed.setThumbnail(song.thumbnail)
			embed.setFooter({ text: `Duration: ${song.duration}` })
		}
		if (!queue.playing) await queue.play()
		return interaction.editReply({ embeds: [embed] })
	},
}
