package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"os"
	"syscall"
	"os/signal"
	"net"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"log"
)

func sendHandler(rw http.ResponseWriter, req *http.Request) {

	response, _ := json.Marshal(req.URL.Path[1:])
	fmt.Fprintf(rw, string(response))
}

func loginHandler(rw http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get(constants.USERNAME)
	password := req.URL.Query().Get(constants.PASSWORD)

	_, err := models.FindUserByCreds(username, password)
	if err != nil {
		response, _ := json.Marshal(constants.WRONG_CREDS_ERROR)
		fmt.Fprintf(rw, string(response))
		return
	}

	response, _ := json.Marshal(map[string]interface{}{"msg": constants.LOGIN_SUCCESS})
	fmt.Fprintf(rw, string(response))
}

func registerHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var msg json.RawMessage
	fullMsg := protocol.CompleteMessage{
		Content: &msg,
	}
	if err := decoder.Decode(&fullMsg); err != nil {
		log.Fatal(err)
	}

	//decoder := json.NewDecoder(req.Body)
	//var fullMsg protocol.CompleteMessage
	//err := decoder.Decode(&fullMsg)
	//helpers.HandleServerError(err, rw)

	if fullMsg.Type == "R" {
		var connectRequest protocol.ConnectRequest
		json.Unmarshal(msg, &connectRequest)
		println("Recieved register request")
		println(connectRequest.UserName)
		println(connectRequest.Password)
		var user models.User
		user, err := models.CreateUser(connectRequest.UserName, connectRequest.Password)
		helpers.HandleServerError(err, rw)

		//encoder := json.NewEncoder(rw)
		//
		//if err != nil {
		//	//encoder.Encode(map[string]string {"error": constants.WRONG_CREDS_ERROR})
		//	helpers.HandleServerError(err, rw)
		//	return
		//}

		var connectResponse protocol.ConnectResponse
		connectResponse.Type = constants.REGISTER_SUCCESS
		connectResponse.Token = user.Username

		encoder := json.NewEncoder(rw)
		encoder.Encode(connectResponse)
	}

	if fullMsg.Type == "L" {
		var connectRequest protocol.ConnectRequest
		println("Recieved login request")
		var user models.User
		user, err := models.FindUserByCreds(connectRequest.UserName, connectRequest.Password)
		helpers.HandleServerError(err, rw)
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
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", sendHandler)
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
