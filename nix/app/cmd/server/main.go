package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	// Debug mode - dump config and enable verbose logging
	if os.Getenv("DEBUG") != "" {
		configJSON, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal config for dumping: %v", err)
		} else {
			log.Printf("DEBUG: Server configuration:\n%s", string(configJSON))
		}
		log.Printf("DEBUG: Debug mode enabled - verbose logging active")
	}

	// Create CardService with fail-fast initialization using embedded resources
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		log.Fatal("Failed to initialize card service: ", err)
	}
	if err := game.LoadCoupRoleImages(treacherest.CoupRoleImagesFS); err != nil {
		log.Fatal("Failed to initialize Coup role images: ", err)
	}

	// Create BackupService for game state backup/restore
	backupService, err := game.NewBackupService(
		cfg.Server.BackupEncryptionKey,
		cfg.Server.BackupEncryptionEnabled,
	)
	if err != nil {
		log.Fatal("Failed to initialize backup service: ", err)
	}
	if backupService.IsEnabled() {
		log.Printf("Backup service initialized with encryption enabled")
	} else {
		log.Printf("Backup service initialized in DEBUG mode (encryption disabled)")
	}

	// Create store and handler with configuration
	s := store.NewMemoryStore(cfg)
	s.SetCardService(cardService)
	h := handlers.New(s, cardService, cfg, backupService)

	// Use the unified router setup
	r := handlers.SetupRouter(h, cfg, nil)

	// Start server with production configuration
	addr := cfg.Server.Host + ":" + cfg.Server.Port

	serverCtx, stopServer := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopServer()

	// Create custom server with production settings
	server := newHTTPServer(addr, r, cfg, serverCtx)

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server.
	<-serverCtx.Done()
	stopServer()

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownTimeout := cfg.Server.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown timed out: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Fatal("Server forced shutdown failed:", closeErr)
		}
		log.Println("Server forced to stop")
		return
	}

	log.Println("Server gracefully stopped")
}

func newHTTPServer(addr string, handler http.Handler, cfg *config.ServerConfig, baseCtx context.Context) *http.Server {
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout, // 0 for SSE support
		BaseContext: func(net.Listener) context.Context {
			return baseCtx
		},
	}
}
