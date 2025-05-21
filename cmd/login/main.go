package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	password := flag.String("password", "", "The password to send")
	username := flag.String("username", "", "The username to send")
	flag.Parse()

	values := url.Values{}
	values.Add("password", *password)
	values.Add("username", *username)

	u := "http://127.0.0.1:3000/login"
	res, err := http.PostForm(u, values)

	if err != nil {
		log.Fatal("Failed", err)
		os.Exit(1)
	}
	log.Println("Status", res.StatusCode)
	for k, v := range res.Header {
		log.Println(k, v)
	}
	defer res.Body.Close()

	var v map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&v)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(v["name"])
}
