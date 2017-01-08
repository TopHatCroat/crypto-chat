package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	registerOption    = flag.Bool("r", false, "registers with server using username and password specified")
	loginOption       = flag.Bool("l", false, "logs in on the server using username and password")
	sendOption        = flag.Bool("s", false, "send a message to user")
	getMessagesOption = flag.Bool("g", false, "send a message to user")
	newMessages       = make(chan protocol.Message, 1)
)

func init() {
	if err := os.Setenv(constants.EDITION_VAR, constants.CLIENT_EDITION); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	database.GetDatabase()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		database.CloseDatabase()
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
		userName, password := helpers.GetCredentials()
		var connectRequest protocol.ConnectRequest

		public, private, err := protocol.GenerateAsyncKeyPair()
		helpers.HandleError(err)

		var privateKey models.Setting
		var publicKey models.Setting

		privateKey.Key = constants.PRIVATE_KEY
		privateKey.Value = helpers.EncodeB64(private[:])
		err = privateKey.Save()
		if err != nil {
			panic(err)
		}

		publicKey.Key = constants.PUBLIC_KEY
		publicKey.Value = helpers.EncodeB64(public[:])
		err = publicKey.Save()
		if err != nil {
			panic(err)
		}

		connectRequest.UserName = userName
		connectRequest.Password = password
		connectRequest.PublicKey = publicKey.Value
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

	} else if *loginOption {
		userName, password := helpers.GetCredentials()

		println(userName, password)
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

		var token models.Setting
		token.Key = constants.TOKEN_KEY
		token.Value = connectResponse.Token
		err = token.Save()
		if err != nil {
			panic(err)
		}
	} else if *sendOption {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, err := reader.ReadString('\n')
		helpers.HandleError(err)

		var fullNewMsg protocol.CompleteMessage
		var messageRequest protocol.Message
		messageRequest.Content = text
		messageRequest.Reciever = 2
		messageRequest.Timestamp = time.Now().UnixNano()

		token, err := models.GetSetting(constants.TOKEN_KEY)
		if err != nil {
			panic(err)
		}

		fullNewMsg.Type = "S"
		protocol.ConstructMetaData(&fullNewMsg)
		fullNewMsg.Content = messageRequest
		fullNewMsg.Meta.Token = token.Value
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
		} else {
			fmt.Println(messageResponse.Message)
		}
	} else if *getMessagesOption {
		receiveNewMessages(*client)
	}

	//flag.Usage()

}

func receiveNewMessages(client http.Client) {
	timestamp := int64(0)
	for ; ; time.Sleep(1 * time.Second) {
		var fullNewMsg protocol.CompleteMessage
		var getMessagesRequest protocol.GetMessagesRequest
		getMessagesRequest.LastMessageTimestamp = timestamp

		token, err := models.GetSetting(constants.TOKEN_KEY)
		if err != nil {
			panic(err)
		}

		fullNewMsg.Type = "M"
		protocol.ConstructMetaData(&fullNewMsg)
		fullNewMsg.Content = getMessagesRequest
		fullNewMsg.Meta.Token = token.Value
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(fullNewMsg)

		resp, err := client.Post("https://localhost:44333/messages", "application/json", buffer)
		helpers.HandleError(err)

		var getMessagesResponse protocol.GetMessagesResponse
		body, err := ioutil.ReadAll(resp.Body)
		helpers.HandleError(err)
		err = json.Unmarshal(body, &getMessagesResponse)
		helpers.HandleError(err)

		if getMessagesResponse.Error != "" {
			fmt.Println(getMessagesResponse.Error)
			panic(err)
		} else {
			for _, msg := range getMessagesResponse.Messages {
				sentAt := time.Unix(0, msg.Timestamp)

				fmt.Print(msg.Sender)
				fmt.Print(" (")
				fmt.Print(sentAt.Format("2006-01-02 15:04:05"))
				fmt.Print("): ")
				fmt.Print(msg.Content)

				timestamp = sentAt.UnixNano()
			}
		}
	}
}
