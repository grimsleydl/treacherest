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
	localMiddleware "treacherest/internal/middleware"
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

	// Chi's built-in middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	
	// Our custom middleware
	r.Use(localMiddleware.RequestSizeLimiter(cfg.Server.MaxRequestSize))
	r.Use(localMiddleware.SecurityHeaders())
	
	// Rate limiting
	rateLimiter := localMiddleware.NewRateLimiter(cfg.Server.RateLimit, cfg.Server.RateLimitBurst)
	r.Use(rateLimiter.Middleware())

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	r.Get("/", h.Home)
	r.Post("/room/new", h.CreateRoom)
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/room/{code}/start", h.StartGame)
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Get("/game/{code}", h.GamePage)

	// Role configuration endpoints
	r.Post("/room/{code}/config/preset", h.UpdateRolePreset)
	r.Post("/room/{code}/config/toggle", h.ToggleRole)
	r.Post("/room/{code}/config/count", h.UpdateRoleCount)

	// SSE endpoints
	r.Get("/sse/lobby/{code}", h.StreamLobby)
	r.Get("/sse/game/{code}", h.StreamGame)
	r.Get("/sse/host/{code}", h.StreamHost)

	return r
}
