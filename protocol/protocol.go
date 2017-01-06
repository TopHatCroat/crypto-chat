package protocol

import (
	"crypto/rand"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/nacl/box"
	"time"
)

const (
	KEY_SIZE = 32
)

type CompleteMessage struct {
	Type    string      `json: "type"`
	Content interface{} `json: "content"`
	Meta    Meta        `json: "metadata`
}

type Message struct {
	Reciever int64  `json: "reciever"`
	Content  string `json: "content"`
}

type Meta struct {
	SentAt int64  `json: "sent_at"`
	Hash   string `json: "hash"`
	Token  string `json; "token"`
}

type ConnectRequest struct {
	UserName  string `json: "user_name"`
	Password  string `json: "pass_hash"`
	PublicKey string `json: "public_key"`
}

type ConnectResponse struct {
	Type  string `json: "type"`
	Token string `json: "token"`
	Error string `json: "error"`
}

type MessageResponse struct {
	Message string `json: "message"`
	Error   string `json: "error"`
}

type ErrorResponse struct {
	Error string `json: "error"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func ResolveRequest(data []byte) (returnData []byte, err error) {
	returnData = data
	return data, err
}

func ConstructMetaData(fullMsg *CompleteMessage) {
	timeStamp := time.Now()
	fullMsg.Meta.SentAt = timeStamp.UnixNano()
}

func GenerateAsyncKeyPair() (publicKey, privateKey *[KEY_SIZE]byte, err error) {
	publicKey, privateKey, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, privateKey, nil
}
