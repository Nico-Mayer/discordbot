import { Keys } from '../types'
import * as dotenv from 'dotenv'
import path from 'path'

dotenv.config({ path: path.resolve(__dirname, '../../.env') })

export const keys: Keys = {
	clientToken: process.env.CLIENT_TOKEN ?? 'nil',
	clientId: process.env.CLIENT_ID ?? 'nil',
	guildId: process.env.GUILD_ID ?? 'nil',
	nasenUrl: process.env.NASEN_URL ?? 'nil',
}

if (Object.values(keys).includes('nil')) {
	throw new Error('Not all environment variables are set.')
}

export default keys
