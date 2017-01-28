package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/models"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"io"
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
	getMessagesOption = flag.Bool("g", false, "[username], get all new messages")
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

	if *registerOption {
		register()
	} else if *loginOption {
		login()
	} else if *sendOption {
		send()
	} else if *getMessagesOption {
		receiveNewMessages()
	} else {
		flag.Usage()
	}
}

func SecureSend(path string, buffer io.Reader, destination interface{}) {
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

	resp, err := client.Post("https://localhost:44333/"+path, "application/json", buffer)
	helpers.HandleError(err)

	body, err := ioutil.ReadAll(resp.Body)
	helpers.HandleError(err)
	err = json.Unmarshal(body, &destination)
	helpers.HandleError(err)
}

func buildRequestWithToken(typ string, message interface{}, token *string) (*bytes.Buffer, error) {
	var fullNewMsg protocol.CompleteMessageInterface
	fullNewMsg.Type = typ
	protocol.ConstructMetaData(&fullNewMsg)
	fullNewMsg.Content = message
	if token != nil {
		fullNewMsg.Meta.Token = *token
	}
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(fullNewMsg); err != nil {
		return buffer, err
	}
	return buffer, nil
}

func getFriend(friendUserName string) (friend models.Friend, err error) {
	var friendRequest protocol.FriendRequest

	token, err := models.GetSetting(constants.TOKEN_KEY)
	if err != nil {
		return friend, err
	}

	friendRequest.Username = friendUserName

	buffer, err := buildRequestWithToken("U", friendRequest, &token.Value)
	if err != nil {
		panic(err)
	}

	friendResponse := &protocol.FriendResponse{}
	SecureSend("user", buffer, friendResponse)

	if friendResponse.Error != "" {
		return friend, errors.New(friendResponse.Error)
	} else {
		friend := models.Friend{
			APIID:     friendResponse.User.APIID,
			Username:  friendResponse.User.Username,
			PublicKey: friendResponse.User.PublicKey}

		if err := friend.Save(); err != nil {
			return friend, err
		}

		return friend, nil
	}
}

func login() {
	userName, password := helpers.GetCredentials()

	var connectRequest protocol.ConnectRequest

	connectRequest.UserName = userName
	connectRequest.Password = password

	buffer, err := buildRequestWithToken("L", connectRequest, nil)
	if err != nil {
		panic(err)
	}

	connectResponse := &protocol.ConnectResponse{}
	SecureSend("register", buffer, connectResponse)

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
}

func receiveNewMessages() {
	timestamp := int64(0)
	for ; ; time.Sleep(1 * time.Second) {
		var getMessagesRequest protocol.GetMessagesRequest
		getMessagesRequest.LastMessageTimestamp = timestamp

		token, err := models.GetSetting(constants.TOKEN_KEY)
		if err != nil {
			panic(err)
		}

		privateKey, err := models.GetSetting(constants.PRIVATE_KEY)
		if err != nil {
			panic(err)
		}

		buffer, err := buildRequestWithToken("M", getMessagesRequest, &token.Value)
		if err != nil {
			panic(err)
		}

		getMessagesResponse := &protocol.GetMessagesResponse{}
		SecureSend("messages", buffer, getMessagesResponse)

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

				sender, err := models.FindFriendByCreds(msg.Sender)
				if err != nil {
					sender, err = getFriend(msg.Sender)
					if err != nil {
						panic(err)
					}
				}

				//decryptKey, err := sender.GetDecyptionKeyByHash(msg.KeyHash)
				//if err != nil {
				//	panic(err)
				//}

				decryptedMessage, err := protocol.Decrypt(privateKey.Value, sender.PublicKey, msg.Content)
				if err != nil {
					panic(err)
				}
				fmt.Print(decryptedMessage)

				timestamp = sentAt.UnixNano()
			}
		}
	}
}

func register() {
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

	buffer, err := buildRequestWithToken("R", connectRequest, nil)
	if err != nil {
		panic(err)
	}

	connectResponse := &protocol.ConnectResponse{}
	SecureSend("register", buffer, connectResponse)

	if connectResponse.Error != "" {
		fmt.Println(connectResponse.Error)
	} else {
		fmt.Println(connectResponse.Type)
	}
}

func send() {
	var receiverUserName = flag.Arg(0)
	if receiverUserName == "" {
		flag.Usage()
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, err := reader.ReadString('\n')
	helpers.HandleError(err)

	receiver, err := models.FindFriendByCreds(receiverUserName)
	if err != nil {
		receiver, err = getFriend(receiverUserName)
		if err != nil {
			panic(err)
		}
	}

	privateKey, err := models.GetSetting(constants.PRIVATE_KEY)
	if err != nil {
		panic(err)
	}

	textEncrypted, err := protocol.Encrypt(privateKey.Value, receiver.PublicKey, text)
	if err != nil {
		panic(err)
	}

	var messageRequest protocol.Message
	messageRequest.Content = textEncrypted
	messageRequest.Receiver = receiver.APIID
	messageRequest.Timestamp = time.Now().UnixNano()

	token, err := models.GetSetting(constants.TOKEN_KEY)
	if err != nil {
		panic(err)
	}

	buffer, err := buildRequestWithToken("S", &messageRequest, &token.Value)
	if err != nil {
		panic(err)
	}

	messageResponse := &protocol.MessageResponse{}
	SecureSend("send", buffer, messageResponse)

	if messageResponse.Error != "" {
		fmt.Println(messageResponse.Error)
	} else {
		fmt.Println(messageResponse.Message)
	}
}
