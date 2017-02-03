package protocol

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"time"
)

const (
	KEY_SIZE   = 32
	NONCE_SIZE = 24
)

type CompleteMessage struct {
	Type    string           `json:"type"`
	Content *json.RawMessage `json:"content"`
	Meta    Meta             `json:"metadata"`
}

type CompleteMessageInterface struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
	Meta    Meta        `json:"metadata"`
}

type Message struct {
	Receiver  int64  `json:"receiver"`
	Content   string `json:"content"`
	KeyHash   string `json:"key_hash"`
	Timestamp int64  `json:"timestamp"`
}

type MessageData struct {
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	KeyHash   string `json:"key_hash"`
	Timestamp int64  `json:"timestamp"`
}

type UserData struct {
	APIID     int64  `json:"api_id"`
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

type Meta struct {
	SentAt int64  `json:"sent_at"`
	Hash   string `json:"hash"`
	Token  string `json:"token"`
}

type ConnectRequest struct {
	UserName  string `json:"user_name"`
	Password  string `json:"pass_hash"`
	PublicKey string `json:"public_key"`
}

type ConnectResponse struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Error string `json:"error"`
}

type GetMessagesRequest struct {
	LastMessageTimestamp int64 `json:"last_time"`
}

type GetMessagesResponse struct {
	Messages []MessageData `json:"messages"`
	Error    string        `json:"error"`
}

type FriendRequest struct {
	Username string `json:"username"`
}

type FriendResponse struct {
	User  UserData `json:"user"`
	Error string   `json:"error"`
}

type KeyRequest struct {
	KeyData
}

type KeyResponse struct {
	Status string `json:"status"`
	KeyData
	Error string `json:"error"`
}

type KeyData struct {
	Key       string `json:"key"`
	Hash      string `json:"hash"`
	UserID    int64  `json:"user"`
	CreatedAt int64  `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"Message"`
	Error   string `json:"error"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusUpdate struct {
	Message string `json:"msg"`
}

func ResolveRequest(data []byte) (returnData []byte, err error) {
	returnData = data
	return data, err
}

func ConstructMetaData(fullMsg *CompleteMessageInterface) {
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

func GenerateNonce() (*[NONCE_SIZE]byte, error) {
	nonce := new([NONCE_SIZE]byte)
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

func Encrypt(privateKey, publicKey, message string) (result string, err error) {
	keyBytes, err := helpers.DecodeB64(privateKey)
	if err != nil {
		return result, err
	}

	publicKeyBytes, err := helpers.DecodeB64(publicKey)
	if err != nil {
		return result, err
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return result, errors.New(constants.GENERATE_NONCE_ERROR)
	}

	out := make([]byte, len(nonce))
	copy(out, nonce[:])
	var keyBytesProper [KEY_SIZE]byte
	copy(keyBytesProper[:], keyBytes[0:KEY_SIZE])
	var publicKeyBytesProper [KEY_SIZE]byte
	copy(publicKeyBytesProper[:], publicKeyBytes[0:KEY_SIZE])
	out = box.Seal(out, []byte(message), nonce, &publicKeyBytesProper, &keyBytesProper)
	return helpers.EncodeB64(out), nil
}

func Decrypt(key, publicKey, message string) (result string, err error) {
	keyBytes, err := helpers.DecodeB64(key)
	if err != nil {
		return result, err
	}
	publicKeyBytes, err := helpers.DecodeB64(publicKey)
	if err != nil {
		return result, err
	}

	messageByte, err := helpers.DecodeB64(message)
	if err != nil {
		return result, err
	}

	if len(messageByte) < (NONCE_SIZE + secretbox.Overhead) {
		return result, errors.New(constants.NONCE_ERROR)
	}

	var nonce [NONCE_SIZE]byte
	copy(nonce[:], messageByte[:NONCE_SIZE])
	var out []byte
	var keyBytesProper [KEY_SIZE]byte
	copy(keyBytesProper[:], keyBytes[0:KEY_SIZE])
	var publicKeyBytesProper [KEY_SIZE]byte
	copy(publicKeyBytesProper[:], publicKeyBytes[0:KEY_SIZE])
	out, ok := box.Open(nil, messageByte[NONCE_SIZE:], &nonce, &publicKeyBytesProper, &keyBytesProper)
	if !ok {
		return result, errors.New(constants.DECRYPT_ERROR)
	}

	return string(out), nil
}

func EncryptMessage(key *string, message *string) (*string, error) {
	keyBytes, err := helpers.DecodeB64(*key)
	if err != nil {
		return nil, err
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return nil, errors.New(constants.ENCRYPT_ERROR)
	}

	out := make([]byte, len(nonce))
	copy(out, nonce[:])
	var keyBytesProper [KEY_SIZE]byte
	copy(keyBytesProper[:], keyBytes[0:KEY_SIZE])
	out = secretbox.Seal(out, []byte(*message), nonce, &keyBytesProper)

	result := helpers.EncodeB64(out)
	return &result, nil
}

func DecryptMessage(key *string, message *string) (*string, error) {
	keyBytes, err := helpers.DecodeB64(*key)
	if err != nil {
		return nil, err
	}

	messageByte, err := helpers.DecodeB64(*message)
	if err != nil {
		return nil, err
	}

	if len(messageByte) < (NONCE_SIZE + secretbox.Overhead) {
		return nil, errors.New(constants.INVALID_DECYPT_ERROR)
	}

	var nonce [NONCE_SIZE]byte
	copy(nonce[:], messageByte[:NONCE_SIZE])
	var keyBytesProper [KEY_SIZE]byte
	copy(keyBytesProper[:], keyBytes[0:KEY_SIZE])
	result, ok := secretbox.Open(nil, messageByte[NONCE_SIZE:], &nonce, &keyBytesProper)
	if !ok {
		return nil, errors.New(constants.DECRYPT_ERROR)
	}

	stringResult := string(result)
	return &stringResult, nil
}
