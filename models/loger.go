package models

import "github.com/TopHatCroat/CryptoChat-server/database"

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
