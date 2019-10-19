package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var validTokens map[string]string

// Responds true on valid secutity token, false otherwise
func isValidToken(token string) bool {
	_, tokenValid := validTokens[token]
	return tokenValid
}

func validateToken(r *http.Request) bool {
	params := mux.Vars(r)
	token, hasToken := params["token"]

	if !hasToken {
		log.Println("Rejecting connection without token.")
		return false
	}

	if !isValidToken(token) {
		log.Println("Invalid token.")
		return false
	}
	return true
}

func readTokens() error {
	var tokensFile string
	flag.StringVar(&tokensFile, "tokens-file", "", "File to recover valid token assignments.")
	flag.Parse()

	if tokensFile == "" {
		return fmt.Errorf("Empty tokens file")
	}

	// Open settings file
	file, err := os.Open(tokensFile)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&validTokens)
	if err != nil {
		return err
	}

	return nil
}
