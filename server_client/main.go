package main

import (
	"log"
	"testapp/server"
)

func main() {
	var port uint16 = 8080
	srv := server.NewServer(port)
	
	log.Printf("We are starting on %v", srv.Addr)
	
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}