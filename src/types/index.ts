import type {
	ChatInputCommandInteraction,
	InteractionResponse,
	Message,
	SlashCommandBuilder,
} from "discord.js"

export type Command = {
	Metadata: SlashCommandBuilder

	Handler: (interaction: ChatInputCommandInteraction) => Promise<InteractionResponse | Message | undefined>
}
