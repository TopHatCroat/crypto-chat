package models

import (
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
)

type Friend struct {
	ID        int64
	APIID     int64
	Username  string
	PublicKey string
}

func (f *Friend) Save() (err error) {
	db := database.GetDatabase()
	if f.ID == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO friends(api_id, username, public_key) VALUES(?,?,?)")
		if err != nil {
			return err
		}
		result, err := preparedStatement.Exec(f.APIID, f.Username, f.PublicKey)
		if err != nil {
			return err
		}
		f.ID, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE friends set api_id = ?, username = ?, public_key = ? WHERE id = ?")
		if err != nil {
			return err
		}
		if _, err = preparedStatement.Exec(f.APIID, f.Username, f.PublicKey, f.ID); err != nil {
			return err
		}
	}

	return nil
}

func (f *Friend) Delete() (err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("DELETE FROM friends WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(f.ID)
	if err != nil {
		return err
	}

	return nil
}

func (f *Friend) GetDecyptionKeyByHash(hash string) (key *Key, err error) {
	key, err = FindKeyByHash(hash)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func FindFriendByCreds(username string) (f Friend, e error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM friends WHERE username = ? ")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(username)
	if err != nil {
		return f, err
	}

	row.Next()
	err = row.Scan(&f.ID, &f.APIID, &f.Username, &f.PublicKey)
	defer row.Close()
	if err != nil {
		return f, errors.New(constants.NO_SUCH_USER_ERROR)
	}
	return f, nil
}
