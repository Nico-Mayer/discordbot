import { Collection } from '../../types'
import champStats from './champStats'
import randomChamp from './randomChamp'

export default <Collection>{
	name: 'League of Legends',
	description: 'Commands für LoL',
	emoji: '🪄',
	commands: [champStats, randomChamp],
}
