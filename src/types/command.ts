import { SlashCommandBuilder } from 'discord.js'

export interface Command {
	data: Omit<SlashCommandBuilder, 'addSubcommand' | 'addSubcommandGroup'>
	execute: (interaction: any) => void
}
