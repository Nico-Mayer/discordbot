import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import { collections } from '../index'

export default {
	data: new SlashCommandBuilder()
		.setName('help')
		.setDescription('Zeigt liste aller Commands'),
	async execute(interaction: ChatInputCommandInteraction) {
		await interaction.reply({ embeds: [createEmbed()] })
	},
}

function createEmbed() {
	const embed = new EmbedBuilder()
		.setTitle('Commands:')
		.setDescription(createDescription())
		.setColor('#0099ff')
	return embed
}

function createDescription() {
	let description = ''
	collections.forEach((collection) => {
		description += `**${collection.emoji} - ${collection.name}:**\n`
		collection.commands.forEach((command) => {
			description += `\`/${command.data.name}\` - ${command.data.description}\n`
		})
		description += '\n'
	})
	return description
}
