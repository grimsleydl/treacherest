package main

import (
	"log"
	"net/http"

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

	// Use the unified router setup
	return handlers.SetupRouter(h, cfg, nil)
}
