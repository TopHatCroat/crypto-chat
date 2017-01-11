package models

import (
	"github.com/TopHatCroat/CryptoChat-server/database"
	"time"
)

type Log struct {
	ID         int64
	SourceAddr string
	Params     string
	Method     string
	Cipher     uint16
	Timestamp  int64
}

func (log *Log) Log() (err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("INSERT INTO log(source_addr, params, method, cipher, timestamp) VALUES(?,?,?,?,?)")
	if err != nil {
		return err
	}
	result, err := preparedStatement.Exec(log.SourceAddr, log.Params, log.Method, log.Cipher, log.Timestamp)
	if err != nil {
		return err
	}
	log.ID, _ = result.LastInsertId()

	return nil
}

func (log *Log) TimeRequest() (err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("UPDATE log SET request_time = ? WHERE id = ?")
	if err != nil {
		return err
	}

	currentTime := time.Now().UnixNano()

	_, err = preparedStatement.Exec(currentTime - log.Timestamp, log.ID)
	if err != nil {
		return err
	}

	return nil
}
