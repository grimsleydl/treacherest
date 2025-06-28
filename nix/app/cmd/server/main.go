package main

import (
	"log"
	"net/http"
	"os"
	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/handlers"
	"treacherest/internal/store"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Load server configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}
	log.Printf("Loaded configuration: max players per room = %d", cfg.Server.MaxPlayersPerRoom)
	
	// Create CardService with fail-fast initialization using embedded resources
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		log.Fatal("Failed to initialize card service: ", err)
	}
	
	// Create store and handler with configuration
	s := store.NewMemoryStore(cfg)
	h := handlers.New(s, cardService, cfg)
	
	// Set up routes
	r := chi.NewRouter()
	
	// Pages
	r.Get("/", h.Home)
	r.Post("/room/new", h.CreateRoom)  // Changed from /room/create to match form action
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Post("/room/{code}/start", h.StartGame)
	r.Get("/game/{code}", h.GamePage)
	
	// Role configuration endpoints
	r.Post("/room/{code}/config/preset", h.UpdateRolePreset)
	r.Post("/room/{code}/config/toggle", h.ToggleRole)
	r.Post("/room/{code}/config/count", h.UpdateRoleCount)
	
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
