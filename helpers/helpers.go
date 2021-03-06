package helpers

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"
)

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func HandleServerError(err error, w http.ResponseWriter) {
	if err != nil {
		encoder := json.NewEncoder(w)
		encoder.Encode(map[string]string{"error": constants.WRONG_CREDS_ERROR})
		log.Fatal(err)
	}
}

func EncodeB64(message []byte) string {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(base64Text, []byte(message))
	return string(base64Text)
}

func DecodeB64(message string) ([]byte, error) {
	resultSlice := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	length, err := base64.StdEncoding.Decode(resultSlice, []byte(message))
	if err != nil {
		return resultSlice, err
	}
	return resultSlice[:length], nil
}

func ReadFromFile(fileName string) ([]byte, error) {
	var inputFile *os.File
	inputFile, err := os.Open(fileName)
	if err != nil {
		return nil, errors.New("File not found")
	}

	stats, statsErr := inputFile.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	result := make([]byte, size)

	reader := bufio.NewReader(inputFile)
	_, err = reader.Read(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func WriteToFile(fileName string, data []byte) error {
	//try to open the file with path fileName
	var outputFile *os.File
	outputFile, err := os.OpenFile(fileName, os.O_RDWR, os.ModeAppend)
	if err != nil {
		//if it fails with Not Exist error, we create the file
		if os.IsNotExist(err) {
			outputFile, err = os.Create(fileName)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := outputFile.Close(); err != nil {
			return
		}
	}()

	writen, err := outputFile.Write(data)
	if err != nil {
		return err
	}
	if writen != len(data) {
		return errors.New("Error writing to file")
	}

	outputFile.Sync()
	return nil
}

func GetCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}
