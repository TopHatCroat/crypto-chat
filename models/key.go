package models

import (
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
)

type Key struct {
	ID        int64
	Key       string
	Hash      string
	FriendID  int64
	CreatedAt int64
}

func (f *Key) Save() (err error) {
	db := database.GetDatabase()
	if f.ID == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO keys (key, hash, friend_id, created_at) VALUES(?,?,?,?)")
		if err != nil {
			return err
		}
		result, err := preparedStatement.Exec(f.Key, f.Hash, f.FriendID, f.CreatedAt)
		if err != nil {
			return err
		}
		f.ID, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE keys set key = ?, hash = ?, friend_id = ? WHERE id = ?")
		if err != nil {
			return err
		}
		if _, err = preparedStatement.Exec(f.Key, f.Hash, f.FriendID, f.ID); err != nil {
			return err
		}
	}

	return nil
}

func (f *Key) Delete() (err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("DELETE FROM keys WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(f.ID)
	if err != nil {
		return err
	}

	return nil
}

func FindKeyByHash(hash string) (k *Key, e error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM keys WHERE hash = ? ")
	if err != nil {
		return nil, err
	}
	row, err := preparedStatement.Query(hash)
	if err != nil {
		return nil, err
	}

	row.Next()
	err = row.Scan(&k.ID, &k.Key, &k.Hash, &k.FriendID, &k.CreatedAt)
	defer row.Close()
	if err != nil {
		return nil, errors.New(constants.NO_SUCH_KEY_ERROR)
	}
	return k, nil
}
