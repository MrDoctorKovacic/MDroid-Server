package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Read tokens file
	err := readTokens()
	if err != nil {
		log.Panicln(err.Error())
	}

	// Init hub
	hub := newHub()
	go hub.run()

	// Init router
	router := mux.NewRouter()

	//
	// Main routes
	//

	// Handles a low data ping, responding 200 if a connection is waiting and 204 otherwise
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if len(hub.clients) > 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}).Methods("GET")

	// Handles upgrading a client to the websocket.
	router.HandleFunc("/ws/{token}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		token, hasToken := params["token"]

		if !hasToken {
			log.Println("Rejecting connection without token.")
			json.NewEncoder(w).Encode("Invalid token")
		}

		log.Println(fmt.Sprintf("Client attempting to connect with token %s", token))
		if isValidToken(token) {
			log.Println("Accepting client connection...")
			serveWs(hub, w, r)
		} else {
			log.Println("Token is invalid")
			json.NewEncoder(w).Encode("Invalid token")
		}
	})

	log.Fatal(http.ListenAndServe(":5353", router))
}
