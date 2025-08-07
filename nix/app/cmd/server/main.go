package main

import (
	"log"
	"net/http"
)

func main() {
	// Set up server
	handler := SetupServer()

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}
