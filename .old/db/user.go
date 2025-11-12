package db

import (
	"database/sql"

	"github.com/disgoorg/snowflake/v2"
)

type DBUser struct {
	ID             snowflake.ID   `json:"id"`
	Name           string         `json:"name"`
	Exp            int            `json:"exp"`
	RiotPUUID      sql.NullString `json:"riot_puuid"`
	MessageCount   int            `json:"sent_messages_count"`
	VoiceJoinCount int            `json:"voice_join_count"`
}

func InsertDBUser(discordUserID snowflake.ID, username string) error {
	query := "INSERT INTO users (id, name) VALUES ($1, $2)"
	_, err := DB.Exec(query, discordUserID, username)
	if err != nil {
		return err
	}

	return nil
}

func GetUser(id snowflake.ID) (DBUser, error) {
	var user DBUser
	query := "SELECT * FROM users WHERE id = $1"
	err := DB.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Exp, &user.RiotPUUID, &user.MessageCount, &user.VoiceJoinCount)
	if err != nil {
		return user, err
	}
	return user, nil
}

func UserInDatabase(id snowflake.ID) bool {
	query := "SELECT id FROM users WHERE id = $1"
	err := DB.QueryRow(query, id).Scan(&id)

	return err == nil
}

// Validates if user is inside database, if not it inserts the user. Returns the corresponding database user.
func ValidateAndFetchUser(userId snowflake.ID, username string) (DBUser, error) {
	if !UserInDatabase(userId) {
		err := InsertDBUser(userId, username)
		if err != nil {
			return DBUser{}, err
		}
	}

	dbUser, err := GetUser(userId)
	if err != nil {
		return DBUser{}, err
	}

	return dbUser, nil
}

func (user *DBUser) SetRiotPUUID(puuid string) error {
	query := "UPDATE users SET riot_puuid = $1 WHERE id = $2"

	_, err := DB.Exec(query, puuid, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (user *DBUser) GrantExp(exp int) error {
	query := "UPDATE users SET exp = exp + $1 WHERE id = $2"

	_, err := DB.Exec(query, exp, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (user *DBUser) IncrementMessageSentCount() error {
	query := "UPDATE users SET sent_messages_count = sent_messages_count + 1 WHERE id = $1"

	_, err := DB.Exec(query, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (user *DBUser) IncrementVoiceJoinCount() error {
	query := "UPDATE users SET voice_join_count = voice_join_count + 1 WHERE id = $1"

	_, err := DB.Exec(query, user.ID)
	if err != nil {
		return err
	}

	return nil
}
