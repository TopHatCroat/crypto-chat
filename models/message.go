package models

import "github.com/TopHatCroat/CryptoChat-server/database"

type Message struct {
	ID         int64
	SenderID   int64
	RecieverID int64
	Content    string
}

func (msg *Message) Save() error {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("INSERT OR REPLACE INTO messages (sender_id, reciever_id," +
		" content) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(msg.SenderID, msg.RecieverID, msg.Content)
	if err != nil {
		return err
	}

	return nil
}

func (msg *Message) Delete() error {
	db := database.GetDatabase()
	defer db.Close()

	preparedStatement, err := db.Prepare("DELETE FROM messages WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = preparedStatement.Exec(msg.ID)
	if err != nil {
		return err
	}

	return nil
}
