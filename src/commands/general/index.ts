import { Collection } from '../../types'
import ping from './ping'
import help from './help'

export default <Collection>{
	name: 'General',
	description: 'General commands',
	emoji: '📜',
	commands: [ping, help],
}
