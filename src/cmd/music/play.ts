import { formatDuration, getYtThumbnailUrl, userInVoiceAndGuild } from "@lib/utils"
import type { Command } from "@types"
import consola from "consola"
import { type ChatInputCommandInteraction, EmbedBuilder, MessageFlags, SlashCommandBuilder } from "discord.js"
import type { Player, Track } from "lavalink-client"

const meta = new SlashCommandBuilder()
	.setName("play")
	.setDescription("Spielt einen Song ab")
	.addStringOption((option) =>
		option.setName("query").setDescription("Song URL oder Suchbegriff").setRequired(true),
	)

export const playCmd: Command = {
	metadata: meta as SlashCommandBuilder,
	handler: async (interaction) => {
		const { ok, errorMsg, guildId, voiceId } = userInVoiceAndGuild(interaction)
		if (!ok) {
			return interaction.reply({
				content: errorMsg,
				flags: [MessageFlags.Ephemeral],
			})
		}

		consola.log(`[play] User ${interaction.user.username} requested play in guild ${interaction.guildId}.`)
		await interaction.deferReply()
		const lavalink = interaction.client.lavalink
		const player =
			lavalink.getPlayer(guildId) ||
			lavalink.createPlayer({
				guildId: guildId,
				voiceChannelId: voiceId,
				textChannelId: interaction.channelId,
				selfDeaf: true,
				selfMute: false,
				volume: 100,
			})

		if (!player.connected) {
			consola.log(`[play] Connecting player to guild ${interaction.guildId}.`)
			await player.connect()
		}

		const query = interaction.options.getString("query", true)
		consola.log(`[play] Searching for query: ${query}`)
		const result = await player.search({ query }, interaction.user)

		if (!result.tracks.length || !result.tracks || !result.tracks[0]) {
			consola.log(`[play] No tracks found for query: ${query}`)
			return interaction.editReply({ content: "keine tracks gefunden!" })
		}

		await player.queue.add(result.loadType === "playlist" ? result.tracks : result.tracks[0])
		consola.log(`[play] Added ${result.tracks.length} track(s) to queue for guild ${interaction.guildId}.`)

		if (player.playing) {
			consola.log(`[play] Player already playing in guild ${interaction.guildId}.`)
			return interaction.editReply({
				embeds: [queueReply(result.tracks[0] as Track, interaction, player)],
			})
		}

		await player.play()
		const track = result.tracks[0]
		consola.log(`[play] Started playing: ${track.info.title} in guild ${interaction.guildId}.`)

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
