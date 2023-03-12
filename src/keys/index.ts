import { Keys } from '../types'

export const keys: Keys = {
	clientToken: process.env.CLIENT_TOKEN ?? 'nil',
	clientId: process.env.CLIENT_ID ?? 'nil',
	guildId: process.env.GUILD_ID ?? 'nil',
}

if (Object.values(keys).includes('nil')) {
	throw new Error('Not all environment variables are set.')
}

export default keys
