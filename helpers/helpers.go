package helpers

import (
	"net/http"
	"encoding/json"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"log"
	"encoding/base64"
)

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func HandleServerError(err error, w http.ResponseWriter) {
	if err != nil {
		encoder := json.NewEncoder(w)
		encoder.Encode(map[string]string {"error": constants.WRONG_CREDS_ERROR})
		log.Fatal(err)
	}
}

func EncodeB64(message []byte) (string) {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(base64Text, []byte(message))
	return string(base64Text)
}

func DecodeB64(message string) ([]byte) {
	resultSlice := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	length, _ := base64.StdEncoding.Decode(resultSlice, []byte(message))
	return resultSlice[:length]
}
