package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"net/http"
	// "regexp"
	// "github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/models"
)

var (
	USERNAME           = "username"
	PASSWORD           = "password"
	LOGIN_SUCCESS      = "Successful login"
	REGISTER_SUCCESS   = "Registration has been successful"
	NO_SUCH_USER_ERROR = "No such user found"
	WRONG_CREDS_ERROR  = "Wrong username and/or password"
)

type Message struct {
	Sender   []byte
	Reciever []byte
	Content  []byte
}

func sendHandler(rw http.ResponseWriter, req *http.Request) {

	response, _ := json.Marshal(req.URL.Path[1:])
	fmt.Fprintf(rw, string(response))
}

func loginHandler(rw http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get(USERNAME)
	password := req.URL.Query().Get(PASSWORD)

	_, err := models.FindUserByCreds(username, password)
	if err != nil {
		response, _ := json.Marshal(WRONG_CREDS_ERROR)
		fmt.Fprintf(rw, string(response))
		return
	}

	response, _ := json.Marshal(map[string]interface{}{"msg": LOGIN_SUCCESS})
	fmt.Fprintf(rw, string(response))
}

func registerHandler(rw http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get(USERNAME)
	password := req.URL.Query().Get(PASSWORD)

	user, err := models.CreateUser(username, password)
	if err != nil {
		response, _ := json.Marshal(WRONG_CREDS_ERROR)
		fmt.Fprintf(rw, string(response))
		return
	}

	response, _ := json.Marshal(map[string]interface{}{"msg": LOGIN_SUCCESS, "creds": user})
	fmt.Fprintf(rw, string(response))
}

func main() {
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/register/", registerHandler)
	http.HandleFunc("/", sendHandler)
	http.ListenAndServe(":8080", nil)
}
