package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"os/exec"
)

type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
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
	}

	if fullMsg.Type == "L" {
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
	}

	return nil
}

func (fn appHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := fn(rw, req); err != nil {
		encoder := json.NewEncoder(rw)
		encoder.Encode(map[string]string{"error": err.Message})
		//http.Error(rw, err.Message, err.Code)
	}
}

func init() {
	_, err := helpers.ReadFromFile("token.pem")
	if err != nil {
		err := exec.Command("./tools").Run()
		if err != nil {
			panic(err)
		}


		err = os.Rename("key.pem",  "token.pem")
		if err != nil {
			panic(err)
		}
		err = os.Rename("cert.pem",  "token_cert.pem")
		if err != nil {
			panic(err)
		}
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
	mux.Handle("/register", appHandler(registerHandler))

	server := &http.Server{
		Addr:         ":44333",
		Handler:      mux,
		TLSConfig:    configuration,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	database.GetDatabase()
	log.Fatal(server.ListenAndServeTLS("server.cert", "server.key"))

}
