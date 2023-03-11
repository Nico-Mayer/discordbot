import { Client, Collection } from 'discord.js'
export interface MyClient extends Client {
	commands?: Collection<string, any>
	champs?: Promise<unknown[]>
}
