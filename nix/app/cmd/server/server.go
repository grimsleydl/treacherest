package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/handlers"
	"treacherest/internal/store"
)

// SetupServer creates and configures the server
func SetupServer() http.Handler {
	// Load server configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	// Create CardService with fail-fast initialization
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		log.Fatal("Failed to initialize card service: ", err)
	}
	
	// Initialize in-memory store
	gameStore := store.NewMemoryStore(cfg)

	// Initialize handlers
	h := handlers.New(gameStore, cardService, cfg)

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
	r.Get("/sse/host/{code}", h.StreamHost)

	return r
}
