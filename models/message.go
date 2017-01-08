package models

import (
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
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

func GetNewMessagesForUser(user User, timestamp int64) (messages []protocol.MessageData, err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT users.username, messages.content, messages.created_at " +
		"FROM messages JOIN users on messages.sender_id = users.id " +
		"WHERE messages.reciever_id = ? AND messages.created_at > ?")
	if err != nil {
		return messages, err
	}

	rows, err := preparedStatement.Query(user.ID, timestamp)
	if err != nil {
		return messages, err
	}

	for rows.Next() {
		var message protocol.MessageData
		rows.Scan(&message.Sender, &message.Content, &message.Timestamp)
		messages = append(messages, message)
	}
	rows.Close()
	if rows.Err() != nil {
		return messages, rows.Err()
	}

	return messages, nil
}
