import { Collection } from '../../types'
import play from './play'
import queue from './queue'
import pause from './pause'
import resume from './resume'
import quit from './quit'
import skip from './skip'

export default <Collection>{
	name: 'Music',
	description: 'Commands für Musik Bot',
	emoji: '🎵',
	commands: [pause, queue, play, quit, resume, skip],
}
