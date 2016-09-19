package models

import (
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/database"
)

var (
)

type Entity interface {
	Save() int64
	Delete() int64
}

type User struct {
	id   int64
	username string
	password string
	gcm  string
}

func (u *User) Save() int64 {
	db := database.GetDatabase();
	if u.id == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO users(username, password, gcm) VALUES(?,?,?)")
		helpers.HandleError(err)
		result, err := preparedStatement.Exec(u.username, u.password, u.gcm)
		helpers.HandleError(err)
		u.id, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE users set username = ?, password = ?, gcm = ? WHERE id = ?")
		helpers.HandleError(err)
		_, err = preparedStatement.Exec(u.username, u.password, u.gcm, u.id)
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
	err = row.Scan(&u.id, &u.username, &u.password, &u.gcm)
	helpers.HandleError(err)

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
	err = row.Scan(&u.id, &u.username, &u.password, &u.gcm)
	helpers.HandleError(err)

	return u, nil
}

func CreateUser(nick string, pass string) (u User, e error) {
	user := User{username: nick, password: pass, gcm: "0"}
	id := user.Save()
	user = FindUserById(id)

	return user, nil
}
