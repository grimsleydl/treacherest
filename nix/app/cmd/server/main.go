package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/handlers"
	"treacherest/internal/store"
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
	s.SetCardService(cardService)
	h := handlers.New(s, cardService, cfg)

	// Use the unified router setup
	r := handlers.SetupRouter(h, cfg, nil)

	// Start server with production configuration
	addr := cfg.Server.Host + ":" + cfg.Server.Port

	// Create custom server with production settings
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout, // 0 for SSE support
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server gracefully stopped")
}
