package models

import (
	"github.com/TopHatCroat/CryptoChat-server/database"
)

type UserSession struct {
	SessionKey   string
	UserID       int64
	LoginTime    int64
	LastSeenTime int64
}

func (ses *UserSession) Save() error {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("INSERT OR REPLACE INTO user_sessions (session_key, user_id, login_time," +
		" last_seen_time) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(ses.SessionKey, ses.UserID, ses.LoginTime, ses.LastSeenTime)
	if err != nil {
		return err
	}

	return nil
}

func (ses *UserSession) Delete() error {
	db := database.GetDatabase()
	defer db.Close()

	preparedStatement, err := db.Prepare("DELETE FROM user_sessions WHERE session_key = ?")
	if err != nil {
		return err
	}

	_, err = preparedStatement.Exec(ses.SessionKey)
	if err != nil {
		return err
	}

	return nil
}
