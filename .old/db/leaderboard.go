package db

import "github.com/disgoorg/snowflake/v2"

type LeaderboardEntry struct {
	UserID     snowflake.ID
	NasenCount int
}

func GetLeaderboard() ([]LeaderboardEntry, error) {
	query := `
		SELECT u.id, COUNT(n.id) AS nasenCount
		FROM users u
		LEFT JOIN nasen n ON u.id = n.userid
		GROUP BY u.id
		ORDER BY nasenCount DESC
		LIMIT 10
	`

	rows, err := DB.Query(query)
	if err != nil {
		return []LeaderboardEntry{}, err
	}

	var leaderboard []LeaderboardEntry

	for rows.Next() {
		var userID snowflake.ID
		var nasenCount int

		err := rows.Scan(&userID, &nasenCount)
		if err != nil {
			return []LeaderboardEntry{}, err
		}

		leaderboard = append(leaderboard, LeaderboardEntry{UserID: userID, NasenCount: nasenCount})
	}

	return leaderboard, nil
}
