package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"os"
	"syscall"
	"os/signal"
	"net"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"log"
)

type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error error
	Message string
	Code int
}

func sendHandler(rw http.ResponseWriter, req *http.Request) *appError {

	response, _ := json.Marshal(req.URL.Path[1:])
	fmt.Fprintf(rw, string(response))

	return nil
}

func loginHandler(rw http.ResponseWriter, req *http.Request) *appError {
	username := req.URL.Query().Get(constants.USERNAME)
	password := req.URL.Query().Get(constants.PASSWORD)

	_, err := models.FindUserByCreds(username, password)
	if err != nil {
		return &appError{err, "whops", 500}
	}

	response, _ := json.Marshal(map[string]interface{}{"msg": constants.LOGIN_SUCCESS})
	fmt.Fprintf(rw, string(response))

	return nil
}

func registerHandler(rw http.ResponseWriter, req *http.Request) *appError {
	decoder := json.NewDecoder(req.Body)
	var msg json.RawMessage
	fullMsg := protocol.CompleteMessage{
		Content: &msg,
	}
	if err := decoder.Decode(&fullMsg); err != nil {
		return &appError{err, "lol", 500}
	}


	if fullMsg.Type == "R" {
		var connectRequest protocol.ConnectRequest
		log.Println("Recieved register request")
		json.Unmarshal(msg, &connectRequest)
		var user models.User
		user, err := models.CreateUser(connectRequest.UserName, connectRequest.Password)
		if err != nil {
			return &appError{err, "Something went tits up", 500}
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
		user, err := models.FindUserByCreds(connectRequest.UserName, connectRequest.Password)
		if err != nil {
			return &appError{err, err.Error(), 500}
		}
		//encoder := json.NewEncoder(rw)
		//
		//if err != nil {
		//	//encoder.Encode(map[string]string {"error": constants.WRONG_CREDS_ERROR})
		//	helpers.HandleServerError(err, rw)
		//	return
		//}

		var connectResponse protocol.ConnectResponse
		connectResponse.Type = constants.LOGIN_SUCCESS
		connectResponse.Token = user.Username

		encoder := json.NewEncoder(rw)
		encoder.Encode(connectResponse)
	}

	return nil
}

func handleClient(conn net.Conn) {
	// close connection on exit
	defer conn.Close()

	var buf [512]byte
	for {
		// read upto 512 bytes
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}

		fmt.Println(n);
		fmt.Println(buf);

		// write the n bytes read
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

func (fn appHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := fn(rw, req); err != nil {
		encoder := json.NewEncoder(rw)
		encoder.Encode(map[string]string {"error": err.Message})
		//http.Error(rw, err.Message, err.Code)
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

	database.GetDatabase()
	//http.Handle("/login", appHandler(loginHandler))
	http.Handle("/register", appHandler(registerHandler))
	//http.Handle("/", appHandler(sendHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))


	//listener, err := net.Listen("tcp", ":2000")
	//helpers.HandleError(err)

	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		continue
	//	}
	//	// run as a goroutine
	//	go handleClient(conn)
	//}


	//database.CloseDatabase()
}
