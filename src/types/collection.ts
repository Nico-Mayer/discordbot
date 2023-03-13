import { Command } from './command'
export interface Collection {
	name: string
	description: string
	emoji: string
	commands: Command[]
}
