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
	"github.com/gorilla/websocket"
	"log"
)

var (
	registerOption    = flag.Bool("r", false, "registers with server using username and password specified")
	loginOption       = flag.Bool("l", false, "logs in on the server using username and password")
	sendOption        = flag.Bool("s", false, "send a message to user")
	getMessagesOption = flag.Bool("g", false, "[username], get all new messages")
	realOption        = flag.Bool("rt", false, "register on realtime chat (testing)")

	sigs chan os.Signal
)

func init() {
	if err := os.Setenv(constants.EDITION_VAR, constants.CLIENT_EDITION); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	database.GetDatabase()

	sigs = make(chan os.Signal, 1)
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
	} else if *realOption {
		registerReal()
	} else {
		//flag.Usage()
		//login()
		registerReal()
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

func SecureSendGet(path string, token *string) {
	rootCertificates := x509.NewCertPool()
	certificate, err := helpers.ReadFromFile("server.cert")
	helpers.HandleError(err)
	ok := rootCertificates.AppendCertsFromPEM(certificate)
	if !ok {
		panic("Failed parsing certificate")
	}

	TLSConfig := &tls.Config{RootCAs: rootCertificates, }
	TLSConfig.BuildNameToCertificate()

	dialer := websocket.Dialer{
		TLSClientConfig:TLSConfig,
	}

	conn, res, err := dialer.Dial("wss://localhost:44333/" + path + "?token=" + *token, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Printf("%s \n", res)

	go func() {

		for {
			var message protocol.Message
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Println("websocket read: ", err)
				return
			}

			log.Printf("rescv: %s", message.Content)
		}
	}()


	//conn.SetReadDeadline(time.Now().Add(protocol.PongWait))
	//conn.SetPingHandler(func(string) error {
	//	conn.SetReadDeadline(time.Now().Add(protocol.PongWait))
	//	return nil
	//})
	//
	//ticker := time.NewTicker(protocol.PingPeriod)
	//defer ticker.Stop()

	for {
		select {
		case <-sigs:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
		}
	}

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

func registerReal() {
	token, err := models.GetSetting(constants.TOKEN_KEY)
	if err != nil {
		panic(err)
	}

	var connectRequest protocol.ConnectRequest
	connectRequest.UserName = "Ja"

	SecureSendGet("real", &token.Value)

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

				friend, err := models.FindFriendByCreds(msg.Sender)
				if err != nil {
					friend, err = getFriend(msg.Sender)
					if err != nil {
						panic(err)
					}
				}

				decryptionKey, err := models.GetValidEncryptionKey(friend)
				if err != nil {
					if err.Error() == constants.NO_SUCH_KEY_ERROR {
						decryptionKey, err = requestKey(&friend, &msg.KeyHash)
						if err != nil {
							panic(err)
						}
					} else {
						panic(err)
					}
				}

				decryptedMessage, err := protocol.DecryptMessage(&decryptionKey.Key, &msg.Content)
				if err != nil {
					panic(err)
				}
				fmt.Print(*decryptedMessage)

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

	friend, err := models.FindFriendByCreds(receiverUserName)
	if err != nil {
		friend, err = getFriend(receiverUserName)
		if err != nil {
			panic(err)
		}
	}

	encryptionKey, err := models.GetValidEncryptionKey(friend)
	if err != nil {
		if err.Error() == constants.NO_SUCH_KEY_ERROR {
			encryptionKey, err = models.GenerateEncryptionKey(friend)

			if err != nil {
				panic(err)
			}

			err = submitKey(encryptionKey, friend)
			if err != nil {
				panic(err)
			}
		}
	}

	textEncrypted, err := protocol.EncryptMessage(&encryptionKey.Key, &text)
	if err != nil {
		panic(err)
	}

	var messageRequest protocol.Message
	messageRequest.Content = *textEncrypted
	messageRequest.Receiver = friend.APIID
	messageRequest.KeyHash = encryptionKey.Hash
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

func submitKey(key *models.Key, reciever models.Friend) error {

	privateKey, err := models.GetSetting(constants.PRIVATE_KEY)
	if err != nil {
		return err
	}

	token, err := models.GetSetting(constants.TOKEN_KEY)
	if err != nil {
		return err
	}

	encryptedKey, err := protocol.Encrypt(privateKey.Value, reciever.PublicKey, key.Key)
	if err != nil {
		return err
	}

	var keyRequest protocol.KeyRequest
	keyRequest.Key = encryptedKey
	keyRequest.Hash = key.Hash
	keyRequest.UserID = key.FriendID
	keyRequest.CreatedAt = key.CreatedAt

	buffer, err := buildRequestWithToken("K", keyRequest, &token.Value)
	if err != nil {
		return err
	}

	keyResponse := &protocol.KeyResponse{}
	SecureSend("keys", buffer, keyResponse)

	if keyResponse.Error != "" {
		fmt.Println(keyResponse.Error)
	} else {
		fmt.Println(keyResponse.Status)
	}
	return nil
}

func requestKey(friend *models.Friend, keyHash *string) (*models.Key, error) {
	privateKey, err := models.GetSetting(constants.PRIVATE_KEY)
	if err != nil {
		return nil, err
	}

	token, err := models.GetSetting(constants.TOKEN_KEY)
	if err != nil {
		return nil, err
	}

	var keyRequest protocol.KeyRequest
	keyRequest.Hash = *keyHash

	buffer, err := buildRequestWithToken("KR", keyRequest, &token.Value)
	if err != nil {
		return nil, err
	}

	keyResponse := &protocol.KeyResponse{}
	SecureSend("keys", buffer, keyResponse)

	decryptedKey, err := protocol.Decrypt(privateKey.Value, friend.PublicKey, keyResponse.Key)
	if err != nil {
		return nil, err
	}

	key := &models.Key{
		Key:       decryptedKey,
		Hash:      keyResponse.Hash,
		FriendID:  friend.ID,
		CreatedAt: keyResponse.CreatedAt,
	}

	if err = key.Save(); err != nil {
		return nil, err
	}

	if keyResponse.Error != "" {
		fmt.Println(keyResponse.Error)
	} else {
		fmt.Println(keyResponse.Status)
	}
	return key, nil
}
