const { SlashCommandBuilder, EmbedBuilder } = require('discord.js')

module.exports = {
	data: new SlashCommandBuilder()
		.setName('champ_stats')
		.setDescription('Display champion stats')
		.addStringOption((option) => {
			return option
				.setName('champion')
				.setDescription('Champion name')
				.setRequired(true)
		}),

	run: async ({ interaction, champs }) => {
		let embed = new EmbedBuilder()
		let input = interaction.options.getString('champion')

		let result = champs.filter((champ) => {
			return champ.id.toLowerCase() === input.toLowerCase()
		})

		if (result.length === 0)
			return interaction.editReply('Champion not found!')

		embed.setColor('#fff000')
		embed.setTitle(result[0].name)

		const stats = `
HP = ${result[0].stats.hp}
HPLevel = ${result[0].stats.hpperlevel}

MP = ${result[0].stats.mp}
MPLevel = ${result[0].stats.mpperlevel}

MoveSpeed = ${result[0].stats.movespeed}

Armor = ${result[0].stats.armor}
ArmorLevel = ${result[0].stats.armorperlevel}

MagicResist = ${result[0].stats.spellblock}
MagicResistLevel = ${result[0].stats.spellblockperlevel}

AttackRange = ${result[0].stats.attackrange}

HPregen = ${result[0].stats.hpregen}
HPregenLevel = ${result[0].stats.hpregenperlevel}

MPregen = ${result[0].stats.mpregen}
MPregenLevel = ${result[0].stats.mpregenperlevel}

Crit = ${result[0].stats.crit}
CritLevel = ${result[0].stats.critperlevel}

AttackDamage = ${result[0].stats.attackdamage}
AttackDamageLevel = ${result[0].stats.attackdamageperlevel}

AttackSpeed = ${result[0].stats.attackspeed}
AttackSpeedLevel = ${result[0].stats.attackspeedperlevel}
			`

		//embed.setDescription(stats)
		embed.setDescription('```javascript' + stats + '```')
		embed.setThumbnail(
			`http://ddragon.leagueoflegends.com/cdn/13.4.1/img/champion/${result[0].image.full}`
		)

		return interaction.editReply({ embeds: [embed] })
	},
}
