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

// Tries to access database file only once, if there is no file it will create it
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
		db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, sender_id INTEGER NOT NULL, " +
			"reciever_id INTEGER NOT NULL, content TEXT NULL, key_hash VARCHAR(100) NOT NULL, created_at INTEGER NOT NULL, " +
			"FOREIGN KEY (sender_id) REFERENCES users(id), " +
			"FOREIGN KEY (reciever_id) REFERENCES users(id) " +
			")")
		db.Exec("CREATE TABLE keys (id INTEGER PRIMARY KEY AUTOINCREMENT, key VARCHAR(255) NOT NULL, " +
			"hash VARCHAR(100) NOT NULL, friend_id INTEGER NOT NULL, created_at INTEGER NOT NULL, " +
			"FOREIGN KEY (friend_id) REFERENCES friends(id) " +
			")")
		db.Exec("CREATE TABLE user_sessions (session_key TEXT PRIMARY KEY, user_id INTEGER NOT NULL, " +
			"login_time INTEGER NOT NULL, last_seen_time INTEGER NOT NULL, " +
			"FOREIGN KEY (user_id) REFERENCES users(id) " +
			")")
		db.Exec("CREATE TABLE log (id INTEGER PRIMARY KEY AUTOINCREMENT, source_addr VARCHAR(45) NOT NULL, " +
			"params VARCHAR(128) NOT NULL, method VARCHAR(10) NOT NULL, cipher INTEGER NOT NULL, timestamp INTEGER NOT NULL, " +
			"request_time INTEGER" +
			")")
	} else if edition == constants.CLIENT_EDITION {
		db.Exec("CREATE TABLE friends (id INTEGER PRIMARY KEY AUTOINCREMENT, api_id INTEGER NOT NULL, " +
			"username VARCHAR(64) NOT NULL, public_key VARCHAR(257) NOT NULL)")
		db.Exec("CREATE TABLE settings (key VARCHAR(100) PRIMARY KEY, value TEXT )")
		db.Exec("CREATE TABLE keys (id INTEGER PRIMARY KEY AUTOINCREMENT, key VARCHAR(255) NOT NULL, " +
			"hash VARCHAR(100) NOT NULL, friend_id INTEGER NOT NULL, created_at INTEGER NOT NULL, " +
			"FOREIGN KEY (friend_id) REFERENCES friends(id) " +
			")")

	}

	db.Close() //close it to write changes
	openDatabase(dbName)
}

func CloseDatabase() {
	db.Close()
}
