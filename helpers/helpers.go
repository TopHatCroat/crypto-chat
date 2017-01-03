package helpers

import (
	"net/http"
	"encoding/json"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"log"
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
