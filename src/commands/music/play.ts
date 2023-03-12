import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	GuildMember,
	EmbedBuilder,
} from 'discord.js'
import { QueryType, Track, useMasterPlayer } from 'discord-player'

export default {
	data: new SlashCommandBuilder()
		.setName('play')
		.setDescription('Spielt einen Song von YouTube ab.')
		.addSubcommand((subcommand) =>
			subcommand
				.setName('song')
				.setDescription('Spielt song von youtube URL ab.')
				.addStringOption((option) =>
					option
						.setName('url')
						.setDescription('youtube url')
						.setRequired(true)
				)
		)
		.addSubcommand((subcommand) => {
			return subcommand
				.setName('search')
				.setDescription(
					'Sucht nach einem Song auf YouTube und spielt diesen ab.'
				)
				.addStringOption((option) =>
					option
						.setName('searchterms')
						.setDescription('Suchbegriffe')
						.setRequired(true)
				)
		}),
	async execute(interaction: ChatInputCommandInteraction) {
		const command = interaction.options.getSubcommand()
		const player = useMasterPlayer()
		const member = interaction.member as GuildMember
		let search = ''

		if (command === 'song') {
			search = interaction.options.getString('url', true)
		} else if (command === 'search') {
			search = interaction.options.getString('searchterms', true)
		}

		if (!member.voice.channel)
			return interaction.reply('Du musst in einem Sprachkanal sein!')

		let embed = new EmbedBuilder()

		const response = await player?.search(search, {
			requestedBy: member,
			searchEngine: QueryType.AUTO,
		})

		if (!response || response.tracks.length < 1)
			return interaction.reply('No results found')

		const song = response.tracks[0]

		embed = createEmbed(song)

		if (!song) return interaction.reply('No results found')
		await player?.play(member.voice.channel, song, {
			nodeOptions: {
				metadata: {
					channel: interaction.channel,
					client: interaction.guild?.members.me,
					requestedBy: interaction.user,
				},
				selfDeaf: true,
				volume: 80,
				leaveOnEmpty: true,
				leaveOnEmptyCooldown: 100000,
				leaveOnEnd: true,
				leaveOnEndCooldown: 100000,
			},
		})
		await interaction.reply({ embeds: [embed] })
	},
}
function createEmbed(item: Track) {
	const embed = new EmbedBuilder()
		.setTitle(item.title)
		.setDescription(
			`<@${item.requestedBy?.id}> hat diesen Song zur Warteschlange hinzugefügt.`
		)
		.setColor('#ff0000')
		.setURL(item.url)
		.setImage(item.thumbnail)
		.setFooter({ text: `Dauer: ${item.duration}` })

	return embed
}
