package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"net/http"
	// "regexp"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"os"
	"syscall"
	"os/signal"
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

	decoder := json.NewDecoder(req.Body)
	var user models.User
	err := decoder.Decode(&user)
	helpers.HandleError(err)

	user, err = models.CreateUser(user.Username, user.Password)
	if err != nil {
		response, _ := json.Marshal(WRONG_CREDS_ERROR)
		fmt.Fprintf(rw, string(response))
		return
	}

	response, _ := json.Marshal(map[string]interface{}{"msg": LOGIN_SUCCESS, "creds": user})
	fmt.Fprintf(rw, string(response))
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

	database.GetDatabase()
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/register/", registerHandler)
	http.HandleFunc("/", sendHandler)
	http.ListenAndServe(":8080", nil)




	database.CloseDatabase()
}
