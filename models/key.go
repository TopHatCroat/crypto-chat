package models

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"io"
	"time"
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

func FindKeyByHash(hash string) (*Key, error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM keys WHERE hash = ? ")
	if err != nil {
		return nil, err
	}
	row, err := preparedStatement.Query(hash)
	if err != nil {
		return nil, err
	}

	var key Key
	row.Next()
	err = row.Scan(&key.ID, &key.Key, &key.Hash, &key.FriendID, &key.CreatedAt)
	defer row.Close()
	if err != nil {
		return nil, errors.New(constants.NO_SUCH_KEY_ERROR)
	}
	return &key, nil
}

func GetValidEncryptionKey(friend Friend) (*Key, error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM keys WHERE keys.friend_id = ? ")
	if err != nil {
		return nil, err
	}
	row, err := preparedStatement.Query(friend.ID)
	if err != nil {
		return nil, err
	}

	var key Key
	row.Next()
	err = row.Scan(&key.ID, &key.Key, &key.Hash, &key.FriendID, &key.CreatedAt)
	defer row.Close()
	if err != nil {
		return nil, errors.New(constants.NO_SUCH_KEY_ERROR)
	}
	return &key, nil

}

func GenerateEncryptionKey(user Friend) (*Key, error) {
	keyBytes := new([constants.KEY_SIZE]byte)

	_, err := io.ReadFull(rand.Reader, keyBytes[:])
	if err != nil {
		return nil, err
	}

	encodedKey := helpers.EncodeB64(keyBytes[:])

	keyHash := sha512.Sum512([]byte(encodedKey))
	encodedKeyHash := helpers.EncodeB64(keyHash[:])

	key := &Key{
		Key:       encodedKey,
		Hash:      encodedKeyHash,
		FriendID:  user.ID,
		CreatedAt: time.Now().UnixNano(),
	}

	err = key.Save()
	if err != nil {
		return nil, err
	}

	return key, nil
}
