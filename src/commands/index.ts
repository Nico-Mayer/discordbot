import { Collection, Command } from '../types'
import general from './general'
import lol from './lol'
import music from './music'
import nasen from './nasen'

export const collections = [general, lol, music, nasen]
export const commands = getCommands(collections)

function getCommands(collections: Collection[]) {
	let commands: Command[] = []
	for (const collection of collections) {
		commands = [...commands, ...collection.commands]
	}
	return commands
}
