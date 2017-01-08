package models

import (
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
)

type Setting struct {
	Key   string
	Value string
}

func (setting *Setting) Save() error {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("INSERT OR REPLACE INTO settings (key, value) VALUES(?,?)")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(setting.Key, setting.Value)
	if err != nil {
		return err
	}

	return nil
}

func (setting *Setting) Delete() error {
	db := database.GetDatabase()
	defer db.Close()

	preparedStatement, err := db.Prepare("DELETE FROM settings WHERE key = ?")
	if err != nil {
		return err
	}

	_, err = preparedStatement.Exec(setting.Key)
	if err != nil {
		return err
	}

	return nil
}

func GetSetting(key string) (setting Setting, err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM settings WHERE key = ?")
	if err != nil {
		return setting, err
	}
	row, err := preparedStatement.Query(key)

	row.Next()
	err = row.Scan(&setting.Key, &setting.Value)
	if err != nil {
		return setting, errors.New(constants.NO_SUCH_SETTING_ERROR)
	}
	row.Close()
	return setting, nil
}
