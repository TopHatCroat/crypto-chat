package models

import (
	"github.com/TopHatCroat/CryptoChat-server/database"
)

type Message struct {
	ID         int64
	SenderID   int64
	RecieverID int64
	Content    string
	CreatedAt  int64
}

func (msg *Message) Save() error {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("INSERT OR REPLACE INTO messages (sender_id, reciever_id," +
		" content, created_at) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(msg.SenderID, msg.RecieverID, msg.Content, msg.CreatedAt)
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

func GetNewMessagesForUser(user User, timestamp int64) (messages []Message, err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM messages WHERE reciever_id = ? AND created_at > ?")
	if err != nil {
		return messages, err
	}

	rows, err := preparedStatement.Query(user.ID, timestamp)
	if err != nil {
		return messages, err
	}

	for rows.Next() {
		var message Message
		rows.Scan(&message.ID, &message.SenderID, &message.RecieverID, &message.Content, &message.CreatedAt)
		messages = append(messages, message)
	}
	rows.Close()
	if rows.Err() != nil {
		return messages, rows.Err()
	}

	return messages, nil
}
