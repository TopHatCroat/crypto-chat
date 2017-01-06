package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
)

const (
	CLIENT_DB_NAME = "kli_baza"
	SERVER_DB_NAME = "ser_baza"
)

var (
	DatabaseName  = "default"
	DATABASE_PATH = "./" + SERVER_DB_NAME + ".db"
	DATABASE_CLIENT_PATH = "./" + CLIENT_DB_NAME + ".db"
	db            *sql.DB
	err           error
	once sync.Once
)

func GetDatabase() *sql.DB {
	once.Do(func() {
		if file, err := os.Stat(DATABASE_PATH); (file != nil && file.Size() < 10) || os.IsNotExist(err) {
			createDatabase(DATABASE_PATH)
		} else if db == nil {
			openDatabase(DATABASE_PATH)
		}
	})

	return db
}

func GetClientDatabase() *sql.DB {
	once.Do(func() {
		if file, err := os.Stat(DATABASE_CLIENT_PATH); (file != nil && file.Size() < 10) || os.IsNotExist(err) {
			createDatabase(DATABASE_CLIENT_PATH)
		} else if db == nil {
			openDatabase(DATABASE_CLIENT_PATH)
		}
	})

	return db
}

func openDatabase(dbName string) {
	db, err = sql.Open("sqlite3", dbName)
	helpers.HandleError(err)
}

func createDatabase(dbName string) {
	openDatabase(dbName)

	if dbName == DATABASE_PATH {
		db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username VARCHAR(64), " +
			"password VARCHAR(255), gcm VARCHAR(255) NULL, public_key VARCHAR(257))")
		db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, sender_id INTEGER, " +
			"reciever_id INTEGER, content TEXT NULL)")
		db.Exec("CREATE TABLE user_sessions (session_key TEXT PRIMARY KEY, user_id INTEGER NOT NULL, " +
			"login_time INTEGER NOT NULL, last_seen_time INTEGER NOT NULL)")
	} else {
		db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username VARCHAR(64), " +
			"password VARCHAR(255), gcm VARCHAR(255) NULL, public_key VARCHAR(257), public_key VARCHAR(257))")
		db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, sender_id INTEGER, " +
			"reciever_id INTEGER, content TEXT NULL)")
		db.Exec("CREATE TABLE user_sessions (session_key TEXT PRIMARY KEY, user_id INTEGER NOT NULL, " +
			"login_time INTEGER NOT NULL, last_seen_time INTEGER NOT NULL)")
	}

	db.Close() //close it to write changes
	openDatabase(dbName)
}

func CloseDatabase() {
	db.Close()
}
