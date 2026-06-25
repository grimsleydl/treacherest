package handlers

import (
	"net/http"

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

	// Group for regular routes WITH timeout
	r.Group(func(r chi.Router) {
		// Apply timeout middleware to this group
		if cfg.Server.RequestTimeout > 0 {
			r.Use(middleware.Timeout(cfg.Server.RequestTimeout))
		}

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
		r.Get("/room/{code}/qr.png", h.RoomQRCode)
		r.Get("/room/{code}", h.JoinRoom)
		r.Get("/room/{code}/operator", h.OperatorDashboard)
		r.Post("/join-room", h.JoinRoomPost)   // New POST endpoint for joining rooms
		r.Post("/room/restore", h.RestoreRoom) // Restore room from client backup
		r.Post("/room/{code}/leave", h.LeaveRoom)
		r.Post("/room/{code}/start", h.StartGame)
		r.Post("/room/{code}/reveal/{playerID}", h.ToggleReveal)
		r.Post("/room/{code}/facestate/{playerID}", h.ToggleFaceState)
		r.Post("/room/{code}/unveil/{playerID}", h.UnveilPlayer)
		r.Get("/room/{code}/unveil-modal/{playerID}", h.GetUnveilModal)
		r.Get("/game/{code}", h.GamePage)

		// Role configuration endpoints
		r.Post("/room/{code}/config/preset", h.UpdateRolePreset)
		r.Post("/room/{code}/config/coup-preset", h.UpdateCoupPreset)
		r.Post("/room/{code}/config/coup-player-count/increment", h.IncrementCoupPlayerCount)
		r.Post("/room/{code}/config/coup-player-count/decrement", h.DecrementCoupPlayerCount)
		r.Post("/room/{code}/config/coup-role-counts", h.UpdateCoupRoleCounts)
		r.Post("/room/{code}/config/coup-role-count/{role}/increment", h.IncrementCoupRoleCount)
		r.Post("/room/{code}/config/coup-role-count/{role}/decrement", h.DecrementCoupRoleCount)
		r.Post("/room/{code}/config/coup-info", h.UpdateCoupInfoPolicy)
		r.Post("/room/{code}/config/coup-royal-guard", h.UpdateCoupRoyalGuardSettings)
		r.Post("/room/{code}/config/coup-inquisition", h.UpdateCoupInquisitionSettings)
		r.Post("/room/{code}/config/coup-green-hunt", h.UpdateCoupGreenHuntSettings)
		r.Post("/room/{code}/coup/royal-guard/{playerID}", h.UseCoupRoyalGuard)
		r.Post("/room/{code}/coup/inquisition/{playerID}", h.CallCoupInquisition)
		r.Post("/room/{code}/coup/inquisition/confirm", h.ConfirmCoupInquisition)
		r.Post("/room/{code}/coup/win/confirm", h.ConfirmCoupWinPrompt)
		r.Post("/room/{code}/coup/win/reject", h.RejectCoupWinPrompt)
		r.Post("/room/{code}/config/toggle", h.ToggleRole)
		r.Post("/room/{code}/config/count", h.UpdateRoleCount)
		r.Post("/room/{code}/config/leaderless", h.UpdateLeaderlessGame)
		r.Post("/room/{code}/config/hide-distribution", h.UpdateHideDistribution)

		// Role options endpoints (for card-specific configuration)
		r.Get("/room/{code}/options", h.GetRoleOptions)
		r.Post("/room/{code}/options", h.SetRoleOption)

		// Modal dismiss/restore endpoints (for ability modals)
		r.Post("/room/{code}/ability/{abilityID}/dismiss", h.DismissModal)
		r.Post("/room/{code}/ability/{abilityID}/restore", h.RestoreModal)

		// Ability action endpoints
		r.Post("/room/{code}/player/{playerID}/trigger-wearer/{xValue}", h.TriggerWearerAbility)
		r.Post("/room/{code}/ability/{abilityID}/select-card/{cardID}", h.SelectWearerCard)
		r.Post("/room/{code}/ability/{abilityID}/confirm", h.ConfirmAbility)

		// Player elimination
		r.Post("/room/{code}/player/{playerID}/eliminate", h.EliminatePlayer)

		// Metamorph ability endpoints
		r.Post("/room/{code}/player/{playerID}/trigger-metamorph", h.TriggerMetamorphAbility)
		r.Post("/room/{code}/player/{playerID}/steal-role/{targetPlayerID}", h.StealRole)
		r.Post("/room/{code}/player/{playerID}/end-metamorph", h.EndMetamorphEffect)

		// Puppet Master ability endpoints
		r.Post("/room/{code}/player/{playerID}/trigger-puppet-master", h.TriggerPuppetMasterAbility)
		r.Post("/room/{code}/puppet-master/{abilityID}/select-players", h.PuppetMasterSelectPlayers)
		r.Post("/room/{code}/puppet-master/{abilityID}/skip", h.PuppetMasterSkip)
		r.Post("/room/{code}/puppet-master/{abilityID}/back", h.PuppetMasterBack)
		r.Post("/room/{code}/puppet-master/{abilityID}/execute", h.PuppetMasterExecute)

		r.Post("/room/{code}/config/fully-random", h.UpdateFullyRandom)
		r.Post("/room/{code}/config/role-type/{roleType}/increment", h.IncrementRoleTypeCount)
		r.Post("/room/{code}/config/role-type/{roleType}/decrement", h.DecrementRoleTypeCount)
		r.Post("/room/{code}/config/player-count/increment", h.IncrementPlayerCount)
		r.Post("/room/{code}/config/player-count/decrement", h.DecrementPlayerCount)

		// New role configuration endpoints
		r.Post("/room/{code}/config/card-toggle", h.ToggleRoleCard)
		r.Post("/room/{code}/config/card-toggle-fast", h.ToggleRoleCardFast)
		r.Post("/room/{code}/config/card-toggle-optimistic", h.ToggleRoleCardOptimistic)

		if cfg.Server.DebugModeEnabled {
			r.Post("/room/{code}/debug/clear", h.DebugClearRoom)
			r.Post("/room/{code}/debug/start-with-debug-players", h.DebugStartWithDebugPlayers)
			r.Post("/room/{code}/debug/start-as-is", h.DebugStartAsIs)
			r.Get("/room/{code}/debug/operator-view", h.DebugOperatorView)
			r.Get("/room/{code}/debug/view-as/{playerID}", h.DebugViewAsPlayer)
		}
	})

	// Group for SSE routes with NO timeout at all
	r.Group(func(r chi.Router) {
		// SSE routes should have no timeout - they're long-lived connections
		// Don't apply any timeout middleware to this group
		// NOTE: SSE routes should NOT inherit RequestTimeout from regular routes

		// SSE routes with validation middleware
		r.Get("/sse/lobby/{code}", ValidateSSERequest(h.StreamLobby))
		r.Get("/sse/game/{code}", ValidateSSERequest(h.StreamGame))
		r.Get("/sse/host/{code}", ValidateSSERequest(h.StreamHost))
	})

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
