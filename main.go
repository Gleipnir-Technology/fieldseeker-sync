package main

import (
	"fmt"
	"os"
)

func main() {
	db, err := connectDB()
	if err != nil {
		fmt.Println("Failed to open database connection: ", err)
	}
	var greeting string
	err = db.QueryRow("select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
	os.Exit(0)
}
