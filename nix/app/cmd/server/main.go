package main

import (
	"context"
	"github.com/go-chi/chi/v5"
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

	// Set up routes
	r := chi.NewRouter()

	// Pages
	r.Get("/", h.Home)
	r.Post("/room/new", h.CreateRoom) // Changed from /room/create to match form action
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/join-room", h.JoinRoomPost) // New POST endpoint for joining rooms
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Post("/room/{code}/start", h.StartGame)
	r.Get("/game/{code}", h.GamePage)

	// Role configuration endpoints
	r.Post("/room/{code}/config/preset", h.UpdateRolePreset)
	r.Post("/room/{code}/config/toggle", h.ToggleRole)
	r.Post("/room/{code}/config/count", h.UpdateRoleCount)
	r.Post("/room/{code}/config/leaderless", h.UpdateLeaderlessGame)
	r.Post("/room/{code}/config/hide-distribution", h.UpdateHideDistribution)
	r.Post("/room/{code}/config/fully-random", h.UpdateFullyRandom)
	r.Post("/room/{code}/config/role-type/{roleType}/increment", h.IncrementRoleTypeCount)
	r.Post("/room/{code}/config/role-type/{roleType}/decrement", h.DecrementRoleTypeCount)
	r.Post("/room/{code}/config/player-count/increment", h.IncrementPlayerCount)
	r.Post("/room/{code}/config/player-count/decrement", h.DecrementPlayerCount)

	// New role configuration endpoints
	r.Post("/room/{code}/config/card-toggle", h.ToggleRoleCard)
	r.Post("/room/{code}/config/card-toggle-fast", h.ToggleRoleCardFast)
	r.Post("/room/{code}/config/card-toggle-optimistic", h.ToggleRoleCardOptimistic)

	// SSE routes with validation middleware
	r.Get("/sse/lobby/{code}", handlers.ValidateSSERequest(h.StreamLobby))
	r.Get("/sse/game/{code}", handlers.ValidateSSERequest(h.StreamGame))
	r.Get("/sse/host/{code}", handlers.ValidateSSERequest(h.StreamHost))

	// Health check endpoints (no auth required)
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check if store is available
		if s == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Store not ready"))
			return
		}
		
		// In production, you might check:
		// - Database connections
		// - External service availability
		// - Cache connections
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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
