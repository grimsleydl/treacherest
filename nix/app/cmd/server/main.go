package main

import (
	"log"
	"net/http"
	"os"
	"treacherest/internal/game"
	"treacherest/internal/handlers"
	"treacherest/internal/store"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Create CardService with fail-fast initialization
	cardService, err := game.NewCardService()
	if err != nil {
		log.Fatal("Failed to initialize card service: ", err)
	}
	
	// Create store and handler
	s := store.NewMemoryStore()
	h := handlers.New(s, cardService)
	
	// Set up routes
	r := chi.NewRouter()
	
	// Pages
	r.Get("/", h.Home)
	r.Post("/room/new", h.CreateRoom)  // Changed from /room/create to match form action
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Post("/room/{code}/start", h.StartGame)
	r.Get("/game/{code}", h.GamePage)
	
	// SSE routes
	r.Get("/sse/lobby/{code}", h.StreamLobby)
	r.Get("/sse/game/{code}", h.StreamGame)
	r.Get("/sse/host/{code}", h.StreamHost)
	
	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Create custom server with no idle timeout for SSE connections
	server := &http.Server{
		Addr:        addr,
		Handler:     r,
		IdleTimeout: 0, // Disable idle timeout to keep SSE connections alive
	}

	log.Printf("Starting server on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
