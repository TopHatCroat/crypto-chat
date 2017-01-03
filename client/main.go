package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"flag"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"net/http"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"log"
)

var (
	registerOption = flag.Bool("register", false, "registers with server using username and password specified")
	loginOption = flag.Bool("login", false, "logs in on the server using username and password")
)

func main() {
	flag.Parse()
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

	//conn, err := net.Dial("tcp", "localhost:2000")
	//helpers.HandleError(err)


	if *registerOption {
		var userName = flag.Arg(0)
		var password = flag.Arg(1)
		var connectRequest protocol.ConnectRequest

		public, private, err := protocol.GenerateAsyncKeyPair()
		helpers.HandleError(err)

		log.Println(helpers.EncodeB64(public[:]))
		log.Println(helpers.EncodeB64(private[:]))

		connectRequest.UserName = userName
		connectRequest.Password = password
		connectRequest.PublicKey = helpers.EncodeB64(public[:])
		var fullMsg protocol.CompleteMessage
		fullMsg.Type = "R"
		protocol.ConstructMetaData(&fullMsg)
		fullMsg.Content = connectRequest
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(fullMsg)

		resp, err := http.Post("http://localhost:8080/register", "application/json", buffer)
		//defer resp.Close()
		helpers.HandleError(err)

		var connectResponse protocol.ConnectResponse
		body, err := ioutil.ReadAll(resp.Body)
		helpers.HandleError(err)
		err = json.Unmarshal(body, &connectResponse)

		if connectResponse.Error != "" {
			fmt.Println(connectResponse.Error)
		} else {
			fmt.Println(connectResponse.Type)
		}

		//TODO: find out why this doesn't work
		//json.NewDecoder(resp.Body).Decode(&connectRequest)

	}

	if *loginOption {
		var userName = flag.Arg(0)
		var password = flag.Arg(1)
		var connectRequest protocol.ConnectRequest

		connectRequest.UserName = userName
		connectRequest.Password = password

		var fullMsg protocol.CompleteMessage
		fullMsg.Type = "L"
		protocol.ConstructMetaData(&fullMsg)
		fullMsg.Content = connectRequest
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(fullMsg)

		resp, err := http.Post("http://localhost:8080/register", "application/json", buffer)
		//defer resp.Close()
		helpers.HandleError(err)

		var connectResponse protocol.ConnectResponse
		body, err := ioutil.ReadAll(resp.Body)
		helpers.HandleError(err)
		err = json.Unmarshal(body, &connectResponse)

		if connectResponse.Error != "" {
			fmt.Println(connectResponse.Error)
		} else {
			fmt.Println(connectResponse.Type)
		}

	}

}