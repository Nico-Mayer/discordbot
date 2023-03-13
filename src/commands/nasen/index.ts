import { Collection } from '../../types'
import nasen from './nasen'
import clownsnase from './clownsnase'
export default <Collection>{
	name: 'Nasen',
	description: 'Nasen Commands',
	emoji: '👃',
	commands: [clownsnase, nasen],
}
