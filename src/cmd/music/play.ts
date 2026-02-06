import { formatDuration, getYtThumbnailUrl } from "@lib/utils"
import type { Command } from "@types"
import {
	type ChatInputCommandInteraction,
	EmbedBuilder,
	type GuildMember,
	MessageFlags,
	SlashCommandBuilder,
} from "discord.js"
import type { Player, Track } from "lavalink-client"

const meta = new SlashCommandBuilder()
	.setName("play")
	.setDescription("Spielt einen Song ab")
	.addStringOption((option) =>
		option.setName("query").setDescription("Song URL oder Suchbegriff").setRequired(true),
	)

export const play: Command = {
	metadata: meta as SlashCommandBuilder,
	handler: async (interaction) => {
		if (!interaction.guildId) return

		const sender = interaction.member as GuildMember
		const voiceChannelId = sender?.voice.channelId
		if (!voiceChannelId) {
			return interaction.reply({
				content: "Du musst in einem Voice Channel sein!",
				flags: [MessageFlags.Ephemeral],
			})
		}
		await interaction.deferReply()
		const lavalink = interaction.client.lavalink
		const player =
			lavalink.getPlayer(interaction.guildId) ||
			lavalink.createPlayer({
				guildId: interaction.guildId,
				voiceChannelId: voiceChannelId,
				textChannelId: interaction.channelId,
				selfDeaf: true,
				selfMute: false,
				volume: 100,
			})

		if (!player.connected) await player.connect()

		const query = interaction.options.getString("query", true)
		const result = await player.search({ query }, interaction.user)

		if (!result.tracks.length || !result.tracks || !result.tracks[0]) {
			return interaction.editReply({ content: "keine tracks gefunden!" })
		}

		await player.queue.add(result.loadType === "playlist" ? result.tracks : result.tracks[0])

		if (player.playing) {
			return interaction.editReply({
				embeds: [queueReply(result.tracks[0] as Track, interaction, player)],
			})
		}

		await player.play()
		const track = result.tracks[0]

		// TODO: evtl auch playlist stuff bei need implementieren
		await interaction.editReply({
			embeds: [singleSongReply(track as Track, interaction)],
		})
	},
}

function singleSongReply(track: Track, interaction: ChatInputCommandInteraction) {
	return new EmbedBuilder()
		.setColor(0x5865f2)
		.setAuthor({ name: "🎶 Jetzt spielt", iconURL: interaction.client.user.displayAvatarURL() })
		.setTitle(track.info.title)
		.setURL(track.info.uri)
		.setDescription(`von **${track.info.author}**`)
		.setImage(getYtThumbnailUrl(track))
		.addFields({ name: "⏱️ Dauer", value: formatDuration(track.info.duration), inline: true })
		.setFooter({ text: "Let it rip, euer Rosine" })
}

function queueReply(track: Track, interaction: ChatInputCommandInteraction, player: Player) {
	return new EmbedBuilder()
		.setColor(0x5865f2) // Discord blurple
		.setAuthor({ name: "📝 Zur Warteschlange hinzugefügt" })
		.setTitle(track.info.title)
		.setURL(track.info.uri)
		.setThumbnail(track.info.artworkUrl || getYtThumbnailUrl(track))
		.addFields(
			{ name: "👤 Künstler", value: track.info.author, inline: true },
			{ name: "⏱️ Dauer", value: formatDuration(track.info.duration), inline: true },
			{ name: "📋 Position", value: `\`#${player.queue.tracks.length}\``, inline: true },
		)
		.setFooter({
			text: `Angefordert von ${interaction.user.username}`,
			iconURL: interaction.user.displayAvatarURL(),
		})
		.setTimestamp()
}
