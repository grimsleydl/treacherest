package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"treacherest/internal/config"
	localMiddleware "treacherest/internal/middleware"
)

// RouterOptions allows customization of router setup for tests
type RouterOptions struct {
	DisableRateLimiting  bool
	DisableRequestLogger bool
	CustomMiddleware     []func(http.Handler) http.Handler
	StaticDir            string // defaults to "static"
}

// SetupRouter creates the application router with all routes and middleware
func SetupRouter(h *Handler, cfg *config.ServerConfig, opts *RouterOptions) *chi.Mux {
	if opts == nil {
		opts = &RouterOptions{}
	}

	// Set default static directory
	if opts.StaticDir == "" {
		opts.StaticDir = "static"
	}

	// Set up router
	r := chi.NewRouter()

	// Chi's built-in middleware (conditionally applied)
	if !opts.DisableRequestLogger {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Our custom middleware
	r.Use(localMiddleware.RequestSizeLimiter(cfg.Server.MaxRequestSize))
	r.Use(localMiddleware.SecurityHeaders())

	// Rate limiting (conditionally applied)
	if !opts.DisableRateLimiting {
		rateLimiter := localMiddleware.NewRateLimiter(cfg.Server.RateLimit, cfg.Server.RateLimitBurst)
		r.Use(rateLimiter.Middleware())
	}

	// Apply custom middleware if provided
	for _, mw := range opts.CustomMiddleware {
		r.Use(mw)
	}

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(opts.StaticDir))))

	// Main pages
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
	r.Get("/sse/lobby/{code}", ValidateSSERequest(h.StreamLobby))
	r.Get("/sse/game/{code}", ValidateSSERequest(h.StreamGame))
	r.Get("/sse/host/{code}", ValidateSSERequest(h.StreamHost))

	// Health check endpoints (no auth required)
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		// In production, you might check:
		// - Database connections
		// - External service availability
		// - Cache connections
		// For now, we assume the service is ready if we can respond
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}