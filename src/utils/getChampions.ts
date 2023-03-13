import { Champion } from '../types'

let cachedChampions: Champion[] | null = null

export async function getChampions(): Promise<Champion[]> {
	if (cachedChampions) {
		return cachedChampions
	}
	const response = await fetch(
		'https://ddragon.leagueoflegends.com/cdn/11.6.1/data/en_US/champion.json'
	)
	const data = await response.json()
	const champions = Object.values(data.data)
	cachedChampions = champions as Champion[]

	return cachedChampions
}
