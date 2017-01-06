package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"bufio"
)

var (
	registerOption = flag.Bool("register", false, "registers with server using username and password specified")
	loginOption    = flag.Bool("login", false, "logs in on the server using username and password")
)

func main() {
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		fmt.Println("SecureChat client closed...")
		os.Exit(0)
	}()

	rootCertificates := x509.NewCertPool()
	certificate, err := helpers.ReadFromFile("server.cert")
	helpers.HandleError(err)
	ok := rootCertificates.AppendCertsFromPEM(certificate)
	if !ok {
		panic("Failed parsing certificate")
	}

	TLSConfig := &tls.Config{RootCAs: rootCertificates}
	TLSConfig.BuildNameToCertificate()
	transportLayer := &http.Transport{TLSClientConfig: TLSConfig}
	client := &http.Client{Transport: transportLayer}

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

		resp, err := client.Post("https://localhost:44333/register", "application/json", buffer)
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

		resp, err := client.Post("https://localhost:44333/register", "application/json", buffer)
		//defer resp.Close()
		helpers.HandleError(err)

		var connectResponse protocol.ConnectResponse
		body, err := ioutil.ReadAll(resp.Body)
		helpers.HandleError(err)
		err = json.Unmarshal(body, &connectResponse)
		helpers.HandleError(err)

		if connectResponse.Error != "" {
			fmt.Println(connectResponse.Error)
			panic(connectResponse.Error)
		} else {
			fmt.Println(connectResponse.Type)
		}

		reader := bufio.NewReader(os.Stdin)
		for true {
			fmt.Print("Enter text: ")
			text, err := reader.ReadString('\n')
			helpers.HandleError(err)

			var fullNewMsg protocol.CompleteMessage
			var messageRequest protocol.Message
			messageRequest.Content = text
			messageRequest.Reciever = 0

			fullNewMsg.Type = "S"
			protocol.ConstructMetaData(&fullNewMsg)
			fullNewMsg.Content = messageRequest
			fullNewMsg.Meta.Token = connectResponse.Token
			buffer := new(bytes.Buffer)
			json.NewEncoder(buffer).Encode(fullNewMsg)

			resp, err := client.Post("https://localhost:44333/send", "application/json", buffer)
			helpers.HandleError(err)

			var messageResponse protocol.MessageResponse
			body, err := ioutil.ReadAll(resp.Body)
			helpers.HandleError(err)
			err = json.Unmarshal(body, &messageResponse)
			helpers.HandleError(err)

			if messageResponse.Error != "" {
				fmt.Println(messageResponse.Error)
				break
			} else {
				fmt.Println(messageResponse.Message)
			}
		}


	}

	//flag.Usage()

}
