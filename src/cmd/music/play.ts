import type { Command } from "@types"
import { type GuildMember, MessageFlags, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder()
	.setName("play")
	.setDescription("Spielt einen Song ab")
	.addStringOption((option) =>
		option.setName("query").setDescription("Song URL oder Suchbegriff").setRequired(true),
	)

export const play: Command = {
	Metadata: meta as SlashCommandBuilder,
	Handler: async (interaction) => {
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
		// Player erstellen oder holen
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

		// Mit Voice Channel verbinden
		if (!player.connected) await player.connect()

		const query = interaction.options.getString("query", true)
		const result = await player.search({ query }, interaction.user)

		if (!result.tracks.length || !result.tracks || !result.tracks[0]) {
			return interaction.editReply({ content: "keine tracks gefunden!" })
		}
		// Track zur Queue hinzufügen
		await player.queue.add(result.loadType === "playlist" ? result.tracks : result.tracks[0])

		// Abspielen wenn nicht bereits am Spielen
		if (!player.playing) await player.play()

		const track = result.tracks[0]
		await interaction.editReply({
			content:
				result.loadType === "playlist"
					? `✅ Playlist mit ${result.tracks.length} Tracks hinzugefügt!`
					: `✅ Jetzt spielt: **${track.info.title}** von **${track.info.author}**`,
		})
	},
}
