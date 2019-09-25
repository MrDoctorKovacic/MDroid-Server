package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var validTokens map[string]string

// Responds true on valid secutity token, false otherwise
func isValidToken(token string) bool {
	_, tokenValid := validTokens[token]
	return tokenValid
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
