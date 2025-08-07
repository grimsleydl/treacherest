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

	// Create custom server with no idle timeout for SSE connections
	server := &http.Server{
		Addr:        addr,
		Handler:     handler,
		IdleTimeout: 0, // Disable idle timeout to keep SSE connections alive
	}

	log.Printf("Starting server on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
