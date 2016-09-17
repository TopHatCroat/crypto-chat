package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"net/http"
	// "regexp"
)

type Message struct {
	sender   []byte
	reciever []byte
	content  []byte
}

func sendHandler(rw http.ResponseWriter, req *http.Request) {

	response, _ := json.Marshal(req.URL.Path[1:])
	fmt.Fprintf(rw, string(response))
}

func main() {
	http.HandleFunc("/", sendHandler)
	http.ListenAndServe(":8080", nil)
}
