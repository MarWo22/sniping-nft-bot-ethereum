package main

import (
	"NFT_Bot/src"
	"net/http"
)

// Starts the websocket server
func main() {
	http.HandleFunc("/traits", src.TraitsEndpoint)
	http.ListenAndServe(":8080", nil)
}
