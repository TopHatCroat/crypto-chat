package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"

	"github.com/TopHatCroat/CryptoChat-server/helpers"
)

var (
	DATABASE_NAME = "baza"
	DATABASE_PATH = "./"+DATABASE_NAME+".db"
	db *sql.DB
	err error
)

func GetDatabase() (*sql.DB) {
	if file, err := os.Stat(DATABASE_PATH); (file != nil && file.Size() < 10) || os.IsNotExist(err) {
		createDatabase()
	} else if db == nil {
		openDatabase()
	}
	return db

}

func openDatabase() {
	db, err = sql.Open("sqlite3", DATABASE_PATH);
	helpers.HandleError(err)
}

func nesr() {

}

func createDatabase() {
	openDatabase()

	db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username VARCHAR(64), " +
		"password VARCHAR(255), gcm VARCHAR(255) NULL)");
	db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, sender_id INTEGER, " +
		"reciever_id INTEGER, content BLOB NULL)");
	db.Close() //close it to write changes
	openDatabase()
}

func CloseDatabase(){
	db.Close()
}

