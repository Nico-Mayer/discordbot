module.exports = async function fetchChamps() {
	const res = await fetch(
		'http://ddragon.leagueoflegends.com/cdn/13.4.1/data/de_DE/champion.json'
	)
	if (res.ok) {
		const data = await res.json()
		return Object.values(data.data)
	}
}
