package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"gleipnir.technology/fieldseeker-sync"
)

func main() {
	err := fssync.InitDB()
	if err != nil {
		log.Println("Failed to init:", err)
		os.Exit(1)
	}

	var username string
	scanValue("Please enter your username : ", &username)

	var displayname string
	scanValue("Please enter your displayname : ", &displayname)

	var password string
	scanValue("Please enter your password : ", &password)

	hash, _ := fssync.HashPassword(password)

	fmt.Println("Username:", username)
	fmt.Println("Display name:", displayname)
	fmt.Println("Password:", password)
	fmt.Println("Hash:    ", hash)

	fssync.SaveUser(displayname, hash, username)
}

func scanValue(message string, result *string) {
	fmt.Printf(message)
	scanner := bufio.NewScanner(os.Stdin)
	if ok := scanner.Scan(); !ok {
		log.Fatal(errors.New("Failed to scan input"))
	}
	*result = scanner.Text()
}
