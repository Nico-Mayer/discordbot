package db

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Nase struct {
	ID       snowflake.ID `json:"id"`
	UserID   snowflake.ID `json:"userid"`
	AuthorID snowflake.ID `json:"authorid"`
	Reason   string       `json:"reason"`
	Created  time.Time    `json:"created"`
}

func InsertNase(nase Nase) error {
	var query string
	var err error

	if nase.Reason == "" {
		query = "INSERT INTO nasen (id, userid, authorid, created) VALUES ($1, $2, $3, $4)"
		_, err = DB.Exec(query, nase.ID, nase.UserID, nase.AuthorID, nase.Created)
	} else {
		query = "INSERT INTO nasen (id, userid, authorid, reason, created) VALUES ($1, $2, $3, $4, $5)"
		_, err = DB.Exec(query, nase.ID, nase.UserID, nase.AuthorID, nase.Reason, nase.Created)
	}

	return err
}

func GetNasenForUser(dbUserID snowflake.ID) ([]Nase, error) {
	var nasen []Nase

	query := "SELECT * FROM nasen WHERE userid = $1 ORDER BY created DESC"

	rows, err := DB.Query(query, dbUserID)
	if err != nil {
		return []Nase{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var nase Nase

		err := rows.Scan(&nase.ID, &nase.UserID, &nase.AuthorID, &nase.Reason, &nase.Created)
		if err != nil {
			return []Nase{}, err
		}

		nasen = append(nasen, nase)
	}

	return nasen, nil
}

func GetNasenCountForUser(dbUserID snowflake.ID) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM nasen WHERE userid = $1"
	err := DB.QueryRow(query, dbUserID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
