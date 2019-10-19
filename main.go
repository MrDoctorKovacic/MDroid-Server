package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var shareQueue [][]byte

func init() {
	shareQueue = make([][]byte, 0)
}

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
	router.HandleFunc("/ws/ping", func(w http.ResponseWriter, r *http.Request) {
		if len(hub.clients) > 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}).Methods("GET")

	// Handles a low data ping, responding 200 if a connection is waiting and 204 otherwise
	router.HandleFunc("/share/ping", func(w http.ResponseWriter, r *http.Request) {
		if len(shareQueue) > 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}).Methods("GET")

	// Handles a low data ping, responding 200 if a connection is waiting and 204 otherwise
	router.HandleFunc("/share/{token}", func(w http.ResponseWriter, r *http.Request) {
		if !validateToken(r) {
			json.NewEncoder(w).Encode("Invalid token")
			return
		}

		if len(shareQueue) > 0 {
			var message []byte
			message, shareQueue = shareQueue[len(shareQueue)-1], shareQueue[:len(shareQueue)-1]
			log.Println("Popped " + string(message) + " from the queue")
			w.Write(message)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}).Methods("GET")

	// Handles a low data ping, responding 200 if a connection is waiting and 204 otherwise
	router.HandleFunc("/share/{token}", func(w http.ResponseWriter, r *http.Request) {
		if !validateToken(r) {
			json.NewEncoder(w).Encode("Invalid token")
			return
		}

		message, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading body: \n" + err.Error())
		}

		shareQueue = append(shareQueue, message)
		log.Println("Added " + string(message) + " to the queue")
		w.Write(message)

	}).Methods("POST")

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
