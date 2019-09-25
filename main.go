package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
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
		if isValidToken(params["token"]) {
			log.Println("Responding to client connection...")
			serveWs(hub, w, r)
		} else {
			json.NewEncoder(w).Encode("Invalid token")
		}
	})

	log.Fatal(http.ListenAndServe(":5353", router))
}
