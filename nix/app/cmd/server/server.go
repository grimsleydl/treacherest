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

	// Create BackupService for game state backup/restore
	backupService, err := game.NewBackupService(
		cfg.Server.BackupEncryptionKey,
		cfg.Server.BackupEncryptionEnabled,
	)
	if err != nil {
		log.Fatal("Failed to initialize backup service: ", err)
	}

	// Initialize in-memory store
	gameStore := store.NewMemoryStore(cfg)
	gameStore.SetCardService(cardService)

	// Initialize handlers
	h := handlers.New(gameStore, cardService, cfg, backupService)

	// Use the unified router setup
	return handlers.SetupRouter(h, cfg, nil)
}
