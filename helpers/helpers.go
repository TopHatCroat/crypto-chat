package helpers

import "net/http"

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}
func HandleServerError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
}
