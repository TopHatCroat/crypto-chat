package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"flag"
	//"net"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"net/http"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"encoding/json"
	"bytes"
	"io/ioutil"
)

var (
	registerOption = flag.Bool("register", false, "registers with server using username and password specified")
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
		//conn.Write([]byte("lalala"))
		var userName = flag.Arg(0)
		var password = flag.Arg(1)
		var connectRequest protocol.ConnectRequest

		connectRequest.UserName = userName
		connectRequest.Password = password

		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(connectRequest)

		resp, err := http.Post("http://localhost:8080/register", "application/json", buffer)
		//defer resp.Close()
		helpers.HandleError(err)

		var connectResponse protocol.ConnectResponse
		body, err := ioutil.ReadAll(resp.Body)
		helpers.HandleError(err)
		err = json.Unmarshal(body, &connectResponse)

		//TODO: find out why this doesn't work
		//json.NewDecoder(resp).Decode(&connectRequest)


		println(connectResponse.Token)
		println(connectResponse.Type)

	}

}