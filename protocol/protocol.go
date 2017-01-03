package protocol

import (
	//"encoding/json"
	//"github.com/TopHatCroat/CryptoChat-server/constants"
	//"fmt"
)
import (
	"time"
)

type CompleteMessage struct {
	Type 	 string 		`json: "type"`
	Content  interface{}	`json: "content"`
	Meta 	 Meta 			`json: "metadata`
}

type Message struct {
	Sender   []byte	`json: "sender"`
	Reciever []byte	`json: "reciever"`
	Content  []byte `json: "content"`
}

type Meta struct {
	SentAt   int64	`json: "sent_at"`
	Hash 	 []byte `json: "hash"`
}

type ConnectRequest struct {
	UserName string `json: "user_name"`
	Password string `json: "pass_hash"`
}

type ConnectResponse struct {
	Type 	 string `json: "type"`
	Token    string `json: "token"`
	Error 	 string `json: "error"`
}

type ErrorResponse struct {
	Error string 	`json: "error"`
}

func ResolveRequest(data []byte) (returnData []byte, err error) {
	returnData = data
	return data, err
}

func ConstructMetaData(fullMsg *CompleteMessage) () {
	timeStamp := time.Now();
	fullMsg.Meta.SentAt = timeStamp.UnixNano()
}

