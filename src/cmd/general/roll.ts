import type { Command } from "@types"
import consola from "consola"
import { type ChatInputCommandInteraction, EmbedBuilder, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder()
	.setName("roll")
	.setDescription("Würfle einen Würfel (0-100)")
	.addIntegerOption((option) =>
		option
			.setName("max")
			.setDescription("Maximaler Wert (Standard: 100)")
			.setRequired(false)
			.setMinValue(1)
			.setMaxValue(1000000),
	)

const handler = async (interaction: ChatInputCommandInteraction) => {
	const maxInput = interaction.options.getInteger("max") ?? 100
	const minInput = 0

	const result = Math.floor(Math.random() * (maxInput + 1))
	consola.log(`[roll] User ${interaction.user.username} rolled ${result} (range: ${minInput}-${maxInput})`)

	const embed = new EmbedBuilder()
		.setColor(0x5865f2)
		.setAuthor({ name: "🎲 Roll", iconURL: interaction.client.user.displayAvatarURL() })
		.setDescription(`# ${result}`)
		.addFields({ name: "Range", value: `\`${minInput}\` - \`${maxInput}\``, inline: false })
		.setFooter({
			text: `Gewürfelt von ${interaction.user.username}`,
			iconURL: interaction.user.displayAvatarURL(),
		})

	return interaction.reply({ embeds: [embed] })
}

const rollCmd: Command = {
	metadata: meta as SlashCommandBuilder,
	handler,
}

export default rollCmd
