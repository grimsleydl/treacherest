package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/handlers"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// setupTestRouter creates a test router with all routes configured
func setupTestRouter() (*chi.Mux, *handlers.Handler) {
	// Initialize in-memory store
	gameStore := store.NewMemoryStore()

	// Initialize handlers
	h := handlers.New(gameStore)

	// Set up router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

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

	return r, h
}

func TestMainRoutes(t *testing.T) {
	router, _ := setupTestRouter()

	t.Run("GET / returns home page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Treacherest") {
			t.Error("expected home page content")
		}
	})

	t.Run("POST /room/new creates room", func(t *testing.T) {
		form := url.Values{}
		form.Add("playerName", "Test Player")

		req := httptest.NewRequest("POST", "/room/new", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected redirect status 303, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !strings.HasPrefix(location, "/room/") {
			t.Errorf("expected redirect to room, got %s", location)
		}
	})
}

func TestJoinRoomHandler(t *testing.T) {
	router, h := setupTestRouter()

	// Create a room first
	room, _ := h.Store().CreateRoom()

	t.Run("GET /room/{code} shows join page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, room.Code) {
			t.Error("expected room code in response")
		}
	})

	t.Run("GET /room/{code} with invalid code returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/INVALID", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("GET /room/{code}?name=Player joins room", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/"+room.Code+"?name=New+Player", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check that player was added
		updatedRoom, _ := h.Store().GetRoom(room.Code)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player in room, got %d", len(updatedRoom.Players))
		}

		// Check cookie was set
		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "player_"+room.Code {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected player cookie to be set")
		}
	})
}

func TestGamePageHandler(t *testing.T) {
	router, h := setupTestRouter()

	// Create a room with a player
	room, _ := h.Store().CreateRoom()
	player := game.NewPlayer("p1", "Test Player", "session1")
	room.AddPlayer(player)
	h.Store().UpdateRoom(room)

	t.Run("GET /game/{code} shows game page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, room.Code) {
			t.Error("expected room code in game page")
		}
	})

	t.Run("GET /game/{code} without player cookie returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("GET /game/{code} with invalid room returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/game/INVALID", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestStartGameIntegration(t *testing.T) {
	router, h := setupTestRouter()

	// Create a room with enough players
	room, _ := h.Store().CreateRoom()
	for i := 1; i <= 4; i++ {
		player := game.NewPlayer(
			strings.Repeat("p", i),
			"Player "+string(rune('0'+i)),
			"session"+string(rune('0'+i)),
		)
		room.AddPlayer(player)
	}
	h.Store().UpdateRoom(room)

	t.Run("POST /room/{code}/start starts game", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "p", // First player
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check room state changed
		updatedRoom, _ := h.Store().GetRoom(room.Code)
		if updatedRoom.State != game.StateCountdown {
			t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
		}
	})
}

func TestLeaveRoomIntegration(t *testing.T) {
	router, h := setupTestRouter()

	// Create a room with players
	room, _ := h.Store().CreateRoom()
	player1 := game.NewPlayer("p1", "Player 1", "session1")
	player2 := game.NewPlayer("p2", "Player 2", "session2")
	room.AddPlayer(player1)
	room.AddPlayer(player2)
	h.Store().UpdateRoom(room)

	t.Run("POST /room/{code}/leave removes player", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected redirect status 303, got %d", w.Code)
		}

		// Check player was removed
		updatedRoom, _ := h.Store().GetRoom(room.Code)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player remaining, got %d", len(updatedRoom.Players))
		}
	})
}

func TestSSEEndpoints(t *testing.T) {
	router, h := setupTestRouter()

	// Create a room with a player
	room, _ := h.Store().CreateRoom()
	player := game.NewPlayer("p1", "Test Player", "session1")
	room.AddPlayer(player)
	h.Store().UpdateRoom(room)

	t.Run("GET /sse/lobby/{code} requires auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sse/lobby/"+room.Code, nil)
		req.Header.Set("Accept", "text/event-stream")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("GET /sse/game/{code} requires auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sse/game/"+room.Code, nil)
		req.Header.Set("Accept", "text/event-stream")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
}

// TestMainFunction tests that main() sets up the server correctly
// This is tricky to test directly, so we verify the setup logic
func TestMainFunction(t *testing.T) {
	// We can't easily test main() directly since it calls log.Fatal
	// Instead, we ensure all the components it uses work correctly
	t.Run("server components initialize without error", func(t *testing.T) {
		// This effectively tests the initialization logic in main()
		gameStore := store.NewMemoryStore()
		if gameStore == nil {
			t.Fatal("failed to create store")
		}

		h := handlers.New(gameStore)
		if h == nil {
			t.Fatal("failed to create handlers")
		}

		r := chi.NewRouter()
		if r == nil {
			t.Fatal("failed to create router")
		}

		// Verify middleware can be added
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		r.Use(middleware.Timeout(60 * time.Second))

		// Verify routes can be added
		r.Get("/", h.Home)

		// Test that the server would start (without actually starting it)
		server := &http.Server{
			Addr:    ":8080",
			Handler: r,
		}

		// Start server in goroutine and immediately shut down
		go func() {
			server.ListenAndServe()
		}()

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// Shut down gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	})
}
