import {
	SlashCommandBuilder,
	ChatInputCommandInteraction,
	EmbedBuilder,
} from 'discord.js'
import Fuse from 'fuse.js'
import { getChampions } from '../utils'
import { Champion } from '../types'

const options = {
	includeScore: true,
	keys: ['name'],
}

export default {
	data: new SlashCommandBuilder()
		.setName('champ_stats')
		.setDescription('Zeigt alle Stats eines LoL Champions')
		.addStringOption((option) =>
			option
				.setName('champion')
				.setDescription('Champion Name')
				.setRequired(true)
		),

	async execute(interaction: ChatInputCommandInteraction) {
		const champions = await getChampions()
		const fuse = new Fuse(champions, options)
		const championName = interaction.options.getString('champion', true)
		const result = fuse.search(championName)

		if (result.length <= 0)
			return interaction.reply({
				content: `Champion \`${championName}\` nicht gefunden.`,
				ephemeral: true,
			})
		else {
			const champ = result[0].item
			const embed = new EmbedBuilder()

			embed.setColor(0x0099ff)
			embed.setTitle(champ.name)
			embed.setDescription(
				'```javascript' + getStats(champ as unknown as Champion) + '```'
			)
			embed.setThumbnail(
				`http://ddragon.leagueoflegends.com/cdn/11.5.1/img/champion/${champ.image.full}`
			)
			await interaction.reply({ embeds: [embed] })
		}
	},
}

function getStats(champion: Champion) {
	const stats = `
HP = ${champion.stats.hp}
HPLevel = ${champion.stats.hpperlevel}\n
MP = ${champion.stats.mp}
MPLevel = ${champion.stats.mpperlevel}\n
MoveSpeed = ${champion.stats.movespeed}\n
Armor = ${champion.stats.armor}
ArmorLevel = ${champion.stats.armorperlevel}\n
MagicResist = ${champion.stats.spellblock}
MagicResistLevel = ${champion.stats.spellblockperlevel}\n
AttackRange = ${champion.stats.attackrange}\n
HPregen = ${champion.stats.hpregen}
HPregenLevel = ${champion.stats.hpregenperlevel}\n
MPregen = ${champion.stats.mpregen}
MPregenLevel = ${champion.stats.mpregenperlevel}\n
Crit = ${champion.stats.crit}
CritLevel = ${champion.stats.critperlevel}\n
AttackDamage = ${champion.stats.attackdamage}
AttackDamageLevel = ${champion.stats.attackdamageperlevel}\n
AttackSpeed = ${champion.stats.attackspeed}
AttackSpeedLevel = ${champion.stats.attackspeedperlevel}
			`

	return stats
}
