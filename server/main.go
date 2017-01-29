package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"github.com/TopHatCroat/CryptoChat-server/tools"
	"github.com/gorilla/context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AppHandler func(http.ResponseWriter, *http.Request) *appData

type Adapter func(http.Handler) http.Handler

type appData struct {
	Data   interface{}
	Error  error
	Status string
	Code   int
}

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log := models.Log{
			SourceAddr: req.RemoteAddr,
			Params:     req.RequestURI,
			Method:     req.Method,
			Cipher:     req.TLS.CipherSuite,
			Timestamp:  time.Now().UnixNano(),
		}

		if err := log.Log(); err != nil {
			encoder := json.NewEncoder(rw)
			encoder.Encode(map[string]string{"error": constants.REQUEST_REJECTED})
			return
		}

		handler.ServeHTTP(rw, req)

		log.TimeRequest()

		if err := GetError(req); err != nil {
			encoder := json.NewEncoder(rw)
			encoder.Encode(map[string]string{"error": constants.REQUEST_REJECTED})
			return
		}
	})
}

func JSONDecoder(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		//var msg json.RawMessage
		fullMsg := protocol.CompleteMessage{
		//Content: &msg,
		}
		if err := decoder.Decode(&fullMsg); err != nil {
			context.Set(req, "err", err)
			return
		}

		context.Set(req, "json", fullMsg)

		handler.ServeHTTP(rw, req)
	})
}

func (fn AppHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	data := fn(rw, req)

	if data.Error != nil {
		encoder := json.NewEncoder(rw)
		encoder.Encode(map[string]string{"error": data.Error.Error(), "status": data.Status})
	} else {
		encoder := json.NewEncoder(rw)
		encoder.Encode(data.Data)
	}
}

func GetJSON(r *http.Request) protocol.CompleteMessage {
	if rv := context.Get(r, "json"); rv != nil {
		return rv.(protocol.CompleteMessage)
	}
	return protocol.CompleteMessage{}
}

func GetError(r *http.Request) error {
	if rv := context.Get(r, "err"); rv != nil {
		return rv.(error)
	}
	return nil
}

func validResponse(data interface{}) *appData {
	return &appData{
		Data: data,
	}
}

func errorResponse(err error, status string, code int) *appData {
	return &appData{
		Error:  err,
		Status: status,
		Code:   code,
	}
}

func sendHandler(rw http.ResponseWriter, req *http.Request) *appData {
	fullMsg := GetJSON(req)

	user, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return errorResponse(err, err.Error(), 500)
	}

	if fullMsg.Type == "S" {
		var messageRequest protocol.Message
		log.Println("Recieved message request")
		json.Unmarshal(*fullMsg.Content, &messageRequest)

		message := &models.Message{
			RecieverID: int64(messageRequest.Receiver),
			SenderID:   user.ID,
			Content:    messageRequest.Content,
			KeyHash:    messageRequest.KeyHash,
			CreatedAt:  messageRequest.Timestamp,
		}

		err := message.Save()
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		var messageResponse protocol.MessageResponse
		messageResponse.Message = constants.MESSAGE_SENT
		return validResponse(messageResponse)

	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return errorResponse(err, err.Error(), 500)
	}
}

func messagesHandler(rw http.ResponseWriter, req *http.Request) *appData {
	fullMsg := GetJSON(req)

	user, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return errorResponse(err, err.Error(), 500)
	}

	if fullMsg.Type == "M" {
		var getMessagesRequest protocol.GetMessagesRequest
		json.Unmarshal(*fullMsg.Content, &getMessagesRequest)

		messages, err := models.GetNewMessagesForUser(user, getMessagesRequest.LastMessageTimestamp)
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		var getMessagesResponse protocol.GetMessagesResponse
		getMessagesResponse.Messages = messages

		return validResponse(getMessagesResponse)
	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return errorResponse(err, err.Error(), 500)
	}
}

func keyHandler(rw http.ResponseWriter, req *http.Request) *appData {
	fullMsg := GetJSON(req)

	_, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return errorResponse(err, err.Error(), 500)
	}

	//key submit
	if fullMsg.Type == "K" {
		var keyRequest protocol.KeyRequest
		json.Unmarshal(*fullMsg.Content, &keyRequest)

		key := &models.Key{
			Key:       keyRequest.Key,
			Hash:      keyRequest.Hash,
			FriendID:  keyRequest.UserID,
			CreatedAt: keyRequest.CreatedAt,
		}

		err := key.Save()
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		var keyResponse protocol.KeyResponse
		keyResponse.Status = constants.KEY_SUBMIT_SUCCESS
		keyResponse.Hash = key.Hash

		return validResponse(keyResponse)
	} else if fullMsg.Type == "KR" {
		var keyRequest protocol.KeyRequest
		json.Unmarshal(*fullMsg.Content, &keyRequest)

		key, err := models.FindKeyByHash(keyRequest.Hash)
		if err != nil {
			return errorResponse(err, err.Error(), 400)
		}

		var keyResponse protocol.KeyResponse
		keyResponse.Status = constants.KEY_FOUND_SUCCESS
		keyResponse.Key = key.Key
		keyResponse.Hash = key.Hash
		keyResponse.UserID = key.FriendID
		keyResponse.CreatedAt = key.CreatedAt

		return validResponse(keyResponse)
	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return errorResponse(err, err.Error(), 500)
	}
}

func userHandler(rw http.ResponseWriter, req *http.Request) *appData {
	fullMsg := GetJSON(req)

	_, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return errorResponse(err, err.Error(), 500)
	}

	if fullMsg.Type == "U" {
		var friendRequest protocol.FriendRequest
		json.Unmarshal(*fullMsg.Content, &friendRequest)

		user, err := models.FindUserByCreds(friendRequest.Username)
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		var friendResponse protocol.FriendResponse
		friendResponse.User.APIID = user.ID
		friendResponse.User.Username = user.Username
		friendResponse.User.PublicKey = user.PublicKey

		return validResponse(friendResponse)
	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return errorResponse(err, err.Error(), 500)
	}

	return nil
}

func registerHandler(rw http.ResponseWriter, req *http.Request) *appData {
	fullMsg := GetJSON(req)
	if fullMsg.Type == "R" {
		var connectRequest protocol.ConnectRequest
		log.Println("Recieved register request")
		json.Unmarshal(*fullMsg.Content, &connectRequest)
		var user models.User
		user, err := models.CreateUser(connectRequest.UserName, connectRequest.Password, connectRequest.PublicKey)
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		var connectResponse protocol.ConnectResponse
		connectResponse.Type = constants.REGISTER_SUCCESS
		connectResponse.Token = user.Username
		return validResponse(connectResponse)

	} else if fullMsg.Type == "L" {
		var connectRequest protocol.ConnectRequest
		log.Println("Recieved login request")
		json.Unmarshal(*fullMsg.Content, &connectRequest)
		var user models.User
		user, err := models.FindUserByCreds(connectRequest.UserName)
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		userToken, err := user.LogIn(connectRequest.Password)
		if err != nil {
			return errorResponse(err, err.Error(), 500)
		}

		var connectResponse protocol.ConnectResponse
		connectResponse.Type = constants.LOGIN_SUCCESS
		connectResponse.Token = userToken
		return validResponse(connectResponse)

	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return errorResponse(err, err.Error(), 500)
	}
}

func init() {
	if _, err := os.Stat(constants.TOKEN_KEY_FILE); os.IsNotExist(err) {
		tools.GenerateTokenKey()
	}

	if err := os.Setenv(constants.EDITION_VAR, constants.SERVER_EDITION); err != nil {
		panic(err)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		database.CloseDatabase()
		fmt.Println("SecureChat server closed...")
		os.Exit(0)
	}()

	configuration := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/register", Logger(JSONDecoder(AppHandler(registerHandler))))
	mux.Handle("/send", Logger(JSONDecoder(AppHandler(sendHandler))))
	mux.Handle("/messages", Logger(JSONDecoder(AppHandler(messagesHandler))))
	mux.Handle("/user", Logger(JSONDecoder(AppHandler(userHandler))))
	mux.Handle("/keys", Logger(JSONDecoder(AppHandler(keyHandler))))

	server := &http.Server{
		Addr:         ":44333",
		Handler:      mux,
		TLSConfig:    configuration,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	database.GetDatabase()
	log.Fatal(server.ListenAndServeTLS("server.cert", "server.key"))

}
