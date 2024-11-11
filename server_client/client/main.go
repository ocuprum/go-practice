package main

import (
	"io"
	"log"
	"net/http"
)


func main() {
	resp, err := http.DefaultClient.Get("http://0.0.0.0:8080/foo")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(string(body))
}