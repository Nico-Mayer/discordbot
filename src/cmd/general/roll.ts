import type { Command } from "@types"
import { type ChatInputCommandInteraction, SlashCommandBuilder } from "discord.js"

const meta = new SlashCommandBuilder()
	.setName("roll")
	.setDescription("Roll a dice (0-100)")
	.addNumberOption((option) => option.setName("max").setDescription("Maximale augenzahl").setRequired(false))

const handler = async (interaction: ChatInputCommandInteraction) => {
	const max = Math.floor(interaction.options.getNumber("max") ?? 100)
	return interaction.reply({
		content: `${Math.floor(Math.random() * max)}`,
	})
}

const rollCmd: Command = {
	metadata: meta as SlashCommandBuilder,
	handler,
}

export default rollCmd
