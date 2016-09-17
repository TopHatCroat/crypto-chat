package models

import (
	"database/sql"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DATABASE_NAME = "baza"
)

type Entity interface {
	Save() int64
	Delete() int64
}

type User struct {
	id   int64
	nick string
	pass string
	gcm  string
}

func (u *User) Save() int64 {
	db, err := sql.Open("sqlite3", "./"+DATABASE_NAME+".db")
	helpers.HandleError(err)

	if u.id == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO users(nick, pass, gcm) VALUES(?,?,?)")
		helpers.HandleError(err)
		result, err := preparedStatement.Exec(u.nick, u.pass, u.gcm)
		helpers.HandleError(err)
		u.id, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE users set nick = ?, pass = ?, gcm = ? WHERE id = ?")
		helpers.HandleError(err)
		_, err = preparedStatement.Exec(u.nick, u.pass, u.gcm, u.id)
		helpers.HandleError(err)
	}
	db.Close()
	return u.id
}

func (u *User) Delete() int64 {
	db, err := sql.Open("sqlite3", "./"+DATABASE_NAME+".db")
	helpers.HandleError(err)

	preparedStatement, err := db.Prepare("DELETE FROM users WHERE id = ?")
	helpers.HandleError(err)
	result, err := preparedStatement.Exec(u.id)
	helpers.HandleError(err)

	defer db.Close()
	count, _ := result.RowsAffected()

	return count
}

func FindUserById(id int) (u User) {
	db, err := sql.Open("sqlite3", "./"+DATABASE_NAME+".db")
	helpers.HandleError(err)

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE id = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(id)

	row.Next()
	err = row.Scan(&u.id, &u.nick, &u.pass, &u.gcm)
	helpers.HandleError(err)

	db.Close()
	return u
}

func FindUserByCreds(nick string, pass string) (u User, e error) {
	db, err := sql.Open("sqlite3", "./"+DATABASE_NAME+".db")
	helpers.HandleError(err)

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE nick = ? AND pass = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(nick, pass)
	if err != nil {
		return u, err
	}

	row.Next()
	err = row.Scan(&u.id, &u.nick, &u.pass, &u.gcm)
	helpers.HandleError(err)

	db.Close()
	return u, nil
}
