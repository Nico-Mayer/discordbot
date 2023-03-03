const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')
const { QueryType, SearchResult } = require('discord-player')

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

		let embed = new EmbedBuilder()
		let result = null

		if (interaction.options.getSubcommand() === 'song') {
			let url = interaction.options.getString('url')
			const res = await client.player.search(url, {
				requestedBy: interaction.user,
				searchEngine: QueryType.YOUTUBE_VIDEO,
			})
			if (res.tracks.length < 1)
				return interaction.editReply('No results found')
			const song = res.tracks[0]
			result = song
			embed.setDescription(`Added ${song.title} (${song.url}) to queue`)
			embed.setThumbnail(song.thumbnail)
			embed.setFooter({ text: `Duration: ${song.duration}` })
		} else if (interaction.options.getSubcommand() === 'playlist') {
			let url = interaction.options.getString('url')
			const res = await client.player.search(url, {
				requestedBy: interaction.user,
				searchEngine: QueryType.YOUTUBE_PLAYLIST,
			})

			if (res.tracks.length < 1)
				return interaction.editReply('No results found')

			const playlist = res.playlist
			const searchResult = new SearchResult(client.player, {
				playlist,
				tracks: playlist.tracks,
			})

			result = searchResult
			embed.setDescription(
				`Added ${res.tracks.length} songs from ${playlist.title} (${playlist.url}) to queue`
			)
			embed.setThumbnail(playlist.tracks[0].thumbnail)
		} else if (interaction.options.getSubcommand() === 'search') {
			let url = interaction.options.getString('searchterms')
			const res = await client.player.search(url, {
				requestedBy: interaction.user,
				searchEngine: QueryType.AUTO,
			})
			if (res.tracks.length === 0)
				return interaction.editReply('No results found')
			const song = res.tracks[0]
			result = song
			embed.setDescription(`Added ${song.title} (${song.url}) to queue`)
			embed.setThumbnail(song.thumbnail)
			embed.setFooter({ text: `Duration: ${song.duration}` })
		}
		if (!result) return interaction.editReply('No results found')
		await client.player.play(interaction.member.voice.channel, result, {
			nodeOptions: {
				metadata: {
					channel: interaction.channel,
					client: interaction.guild.members.me,
					requestedBy: interaction.user,
				},
				selfDeaf: true,
				volume: 80,
				leaveOnEmpty: true,
				leaveOnEmptyCooldown: 300000,
				leaveOnEnd: true,
				leaveOnEndCooldown: 300000,
			},
		})
		return interaction.editReply({ embeds: [embed] })
	},
}
