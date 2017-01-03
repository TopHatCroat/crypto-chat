package models

import (
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
)

var (
)

type Entity interface {
	Save() int64
	Delete() int64
}

type User struct {
	id       int64
	Username string
	Password string
	Gcm      string
}

func (u *User) Save() int64 {
	db := database.GetDatabase();
	if u.id == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO users(username, password, gcm) VALUES(?,?,?)")
		helpers.HandleError(err)
		result, err := preparedStatement.Exec(u.Username, u.Password, u.Gcm)
		helpers.HandleError(err)
		u.id, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE users set username = ?, password = ?, gcm = ? WHERE id = ?")
		helpers.HandleError(err)
		_, err = preparedStatement.Exec(u.Username, u.Password, u.Gcm, u.id)
		helpers.HandleError(err)
	}
	return u.id
}

func (u *User) Delete() int64 {
	db := database.GetDatabase();

	preparedStatement, err := db.Prepare("DELETE FROM users WHERE id = ?")
	helpers.HandleError(err)
	result, err := preparedStatement.Exec(u.id)
	helpers.HandleError(err)

	defer db.Close()
	count, _ := result.RowsAffected()

	return count
}

func FindUserById(id int64) (u User) {
	db := database.GetDatabase();

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE id = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(id)

	row.Next()
	err = row.Scan(&u.id, &u.Username, &u.Password, &u.Gcm)
	helpers.HandleError(err)
	row.Close()
	return u
}

func FindUserByCreds(username string, password string) (u User, e error) {
	db := database.GetDatabase();

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE username = ? AND password = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(username, password)
	if err != nil {
		return u, err
	}

	row.Next()
	err = row.Scan(&u.id, &u.Username, &u.Password, &u.Gcm)
	defer row.Close()
	if err != nil {
		return u, errors.New(constants.WRONG_CREDS_ERROR)
	}
	return u, nil
}

func usernameExists(username string) (exists bool) {
	db := database.GetDatabase()
	preparedStatement, err := db.Prepare("SELECT COUNT(*) FROM users WHERE username = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(username)
	defer row.Close()
	row.Next()
	var count int
	err = row.Scan(&count)
	if count == 0 {
		return false
	} else {
		return true
	}
}

func CreateUser(nick string, pass string) (u User, e error) {
	if usernameExists(nick) {
		return u, errors.New(constants.ALREADY_EXISTS)
	}
	user := User{Username: nick, Password: pass, Gcm: "0"}
	id := user.Save()
	user = FindUserById(id)
	return user, nil
}
