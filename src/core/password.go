package core

import (
	"fmt"
	"log"
	"syscall"

	"golang.org/x/term"
)

// ReadPassword prompts the user for a password without echoing input to the terminal.
func ReadPassword() string {
	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	password := string(bytePassword)
	fmt.Println("🔒")
	return password
}
