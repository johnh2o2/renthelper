package main

import (
	"fmt"
	"os"

	"github.com/chuckha/renthelper/avalon"
)

const (
	avalonUsernameEnv = "AVALON_USERNAME"
	avalonPasswordEnv = "AVALON_PASSWORD"
)

func main() {
	testLogin()
}

func testLogin() {
	username := os.Getenv(avalonUsernameEnv)
	password := os.Getenv(avalonPasswordEnv)
	// Create the scraper.
	_, err := avalon.NewClient(username, password)
	if err != nil {
		panic(err)
	}
	fmt.Println("login works!")
}
