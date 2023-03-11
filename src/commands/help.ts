import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import commands from './index'

export default {
	data: new SlashCommandBuilder()
		.setName('help')
		.setDescription('Show a list of all commands'),
	async execute(interaction: ChatInputCommandInteraction) {
		await interaction.reply({ embeds: [createEmbed()] })
	},
}

function createEmbed() {
	const embed = new EmbedBuilder()
		.setTitle('All Commands:')
		.setDescription(createDescription())
		.setColor('#0099ff')
	return embed
}

function createDescription() {
	let description = ''
	commands.forEach((command) => {
		description += `\`/${command.data.name}\` - ${command.data.description}\n`
	})
	return description
}
