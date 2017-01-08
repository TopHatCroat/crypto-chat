package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/constants"
)

var (
	DATABASE_NAME  = "baza"
	DATABASE_PATH = "./" + DATABASE_NAME + ".db"
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

func openDatabase(dbName string) {
	db, err = sql.Open("sqlite3", dbName)
	helpers.HandleError(err)
}

func createDatabase(dbName string) {
	openDatabase(dbName)

	edition := os.Getenv(constants.EDITION_VAR)
	if len(edition) == 0 {
		panic(constants.EDITION_VAR + " must be set ('" + constants.CLIENT_EDITION + "' or '" + constants.SERVER_EDITION + "')")
	}

	if edition == constants.SERVER_EDITION {
		db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username VARCHAR(64), " +
			"password VARCHAR(255), gcm VARCHAR(255) NULL, public_key VARCHAR(257))")
		db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, sender_id INTEGER, " +
			"reciever_id INTEGER, content TEXT NULL, created_at INTEGER, " +
			"FOREIGN KEY (sender_id) REFERENCES users(id), " +
			"FOREIGN KEY (reciever_id) REFERENCES users(id) " +
			")")
		db.Exec("CREATE TABLE user_sessions (session_key TEXT PRIMARY KEY, user_id INTEGER NOT NULL, " +
			"login_time INTEGER NOT NULL, last_seen_time INTEGER NOT NULL, " +
			"FOREIGN KEY (user_id) REFERENCES users(id) " +
			")")
	} else if edition == constants.CLIENT_EDITION {
		db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username VARCHAR(64), " +
			"password VARCHAR(255), gcm VARCHAR(255) NULL, public_key VARCHAR(257))")
		db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, sender_id INTEGER, " +
			"reciever_id INTEGER, content TEXT NULL)")

	}

	db.Close() //close it to write changes
	openDatabase(dbName)
}

func CloseDatabase() {
	db.Close()
}
