export interface Command {
	data: {
		name: string
		description: string
		options?: {}
	}
	execute: (interaction: any) => void
}
