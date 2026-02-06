import type {
	ChatInputCommandInteraction,
	InteractionResponse,
	Message,
	SlashCommandBuilder,
} from "discord.js"

export type Command = {
	metadata: SlashCommandBuilder
	handler: (interaction: ChatInputCommandInteraction) => Promise<InteractionResponse | Message | undefined>
}
