package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Set up server
	handler := SetupServer()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
