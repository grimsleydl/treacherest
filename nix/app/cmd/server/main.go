package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"treacherest/internal/handlers"
	"treacherest/internal/store"
)

func main() {
	// Initialize in-memory store
	gameStore := store.NewMemoryStore()

	// Initialize handlers
	h := handlers.New(gameStore)

	// Set up router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	r.Get("/", h.Home)
	r.Post("/room/new", h.CreateRoom)
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/room/{code}/start", h.StartGame)
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Get("/game/{code}", h.GamePage)

	// SSE endpoints
	r.Get("/sse/lobby/{code}", h.StreamLobby)
	r.Get("/sse/game/{code}", h.StreamGame)

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}
