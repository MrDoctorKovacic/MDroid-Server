package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

var share []byte

var (
	urlRegex   *regexp.Regexp
	mapsRegex  *regexp.Regexp
	phoneRegex *regexp.Regexp
)

func init() {
	mapsRegex = regexp.MustCompile(`https:\/\/maps.*`)
	urlRegex = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	phoneRegex = regexp.MustCompile(`^([0-9]( |-)?)?(\(?[0-9]{3}\)?|[0-9]{3})( |-)?([0-9]{3}( |-)?[0-9]{4}|[a-zA-Z0-9]{7})$`)
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
		if len(share) > 0 {
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

		if len(share) > 0 {
			log.Println("Popped " + string(share) + " from the queue")
			w.Write(share)
			w.WriteHeader(http.StatusOK)
			share = []byte("")
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
		messageStr := string(message)
		messageStr = strings.Replace(messageStr, "\n", " ", -1)
		messageStr = urlRegex.ReplaceAllString(messageStr, "")
		messageStr = phoneRegex.ReplaceAllString(messageStr, "")
		share = []byte(messageStr)

		log.Println("Added " + string(share) + " to the queue")
		w.Write(share)

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
