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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AppHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn AppHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

	if err := fn(rw, req); err != nil {
		encoder := json.NewEncoder(rw)
		encoder.Encode(map[string]string{"error": err.Message})
		//http.Error(rw, err.Message, err.Code)
	}
}

func sendHandler(rw http.ResponseWriter, req *http.Request) *appError {
	decoder := json.NewDecoder(req.Body)
	var msg json.RawMessage
	fullMsg := protocol.CompleteMessage{
		Content: &msg,
	}
	if err := decoder.Decode(&fullMsg); err != nil {
		return &appError{err, err.Error(), 500}
	}

	user, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return &appError{err, err.Error(), 500}
	}

	if fullMsg.Type == "S" {
		var messageRequest protocol.Message
		log.Println("Recieved message request")
		json.Unmarshal(msg, &messageRequest)

		message := &models.Message{
			RecieverID: int64(messageRequest.Reciever),
			SenderID:   user.ID,
			Content:    messageRequest.Content,
			CreatedAt:  messageRequest.Timestamp,
		}

		err := message.Save()
		if err != nil {
			return &appError{err, err.Error(), 500}
		}

		var messageResponse protocol.MessageResponse
		messageResponse.Message = constants.MESSAGE_SENT

		encoder := json.NewEncoder(rw)
		encoder.Encode(messageResponse)

	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return &appError{err, err.Error(), 500}
	}

	return nil
}

func messagesHandler(rw http.ResponseWriter, req *http.Request) *appError {
	decoder := json.NewDecoder(req.Body)
	var msg json.RawMessage
	fullMsg := protocol.CompleteMessage{
		Content: &msg,
	}
	if err := decoder.Decode(&fullMsg); err != nil {
		return &appError{err, err.Error(), 500}
	}

	user, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return &appError{err, err.Error(), 500}
	}

	if fullMsg.Type == "M" {
		var getMessagesRequest protocol.GetMessagesRequest
		json.Unmarshal(msg, &getMessagesRequest)

		messages, err := models.GetNewMessagesForUser(user, getMessagesRequest.LastMessageTimestamp)
		if err != nil {
			return &appError{err, err.Error(), 500}
		}

		var getMessagesResponse protocol.GetMessagesResponse
		getMessagesResponse.Messages = messages

		encoder := json.NewEncoder(rw)
		encoder.Encode(getMessagesResponse)
	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return &appError{err, err.Error(), 500}
	}

	return nil
}

func userHandler(rw http.ResponseWriter, req *http.Request) *appError {
	decoder := json.NewDecoder(req.Body)
	var msg json.RawMessage
	fullMsg := protocol.CompleteMessage{
		Content: &msg,
	}
	if err := decoder.Decode(&fullMsg); err != nil {
		return &appError{err, err.Error(), 500}
	}

	_, err := models.FindUserByToken(fullMsg.Meta.Token)
	if err != nil {
		return &appError{err, err.Error(), 500}
	}

	if fullMsg.Type == "U" {
		var friendRequest protocol.FriendRequest
		json.Unmarshal(msg, &friendRequest)

		user, err := models.FindUserByCreds(friendRequest.Username)
		if err != nil {
			return &appError{err, err.Error(), 500}
		}

		var friendResponse protocol.FriendResponse
		friendResponse.User.APIID = user.ID
		friendResponse.User.Username = user.Username
		friendResponse.User.PublicKey = user.PublicKey

		encoder := json.NewEncoder(rw)
		encoder.Encode(friendResponse)
	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return &appError{err, err.Error(), 500}
	}

	return nil
}

func registerHandler(rw http.ResponseWriter, req *http.Request) *appError {
	decoder := json.NewDecoder(req.Body)
	var msg json.RawMessage
	fullMsg := protocol.CompleteMessage{
		Content: &msg,
	}
	if err := decoder.Decode(&fullMsg); err != nil {
		return &appError{err, err.Error(), 500}
	}

	if fullMsg.Type == "R" {
		var connectRequest protocol.ConnectRequest
		log.Println("Recieved register request")
		json.Unmarshal(msg, &connectRequest)
		var user models.User
		user, err := models.CreateUser(connectRequest.UserName, connectRequest.Password, connectRequest.PublicKey)
		if err != nil {
			return &appError{err, err.Error(), 500}
		}

		var connectResponse protocol.ConnectResponse
		connectResponse.Type = constants.REGISTER_SUCCESS
		connectResponse.Token = user.Username

		encoder := json.NewEncoder(rw)
		encoder.Encode(connectResponse)
	} else if fullMsg.Type == "L" {
		var connectRequest protocol.ConnectRequest
		log.Println("Recieved login request")
		json.Unmarshal(msg, &connectRequest)
		var user models.User
		user, err := models.FindUserByCreds(connectRequest.UserName)
		if err != nil {
			return &appError{err, err.Error(), 500}
		}

		userToken, err := user.LogIn(connectRequest.Password)
		if err != nil {
			return &appError{err, err.Error(), 500}
		}

		var connectResponse protocol.ConnectResponse
		connectResponse.Type = constants.LOGIN_SUCCESS
		connectResponse.Token = userToken

		encoder := json.NewEncoder(rw)
		encoder.Encode(connectResponse)
	} else {
		err := errors.New(constants.WRONG_REQUEST)
		return &appError{err, err.Error(), 500}
	}

	return nil
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
	mux.Handle("/register", AppHandler(registerHandler))
	mux.Handle("/send", AppHandler(sendHandler))
	mux.Handle("/messages", AppHandler(messagesHandler))
	mux.Handle("/user", AppHandler(userHandler))

	server := &http.Server{
		Addr:         ":44333",
		Handler:      mux,
		TLSConfig:    configuration,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	database.GetDatabase()
	log.Fatal(server.ListenAndServeTLS("server.cert", "server.key"))

}
