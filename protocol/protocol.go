package protocol

import (
	//"encoding/json"
	//"github.com/TopHatCroat/CryptoChat-server/constants"
	//"fmt"
)

type Message struct {
	Sender   []byte
	Reciever []byte
	Content  []byte
}

type Meta struct {
	SentAt   int64
	Hash 	 []byte
}

type ConnectRequest struct {
	UserName string
	Password string
}

type ConnectResponse struct {
	Type 	 string
	Token    string
}

func ResolveRequest(data []byte) (returnData []byte, err error) {
	returnData = data
	return data, err
}

