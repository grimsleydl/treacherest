package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// TestJoinFlowE2E tests the complete join flow end-to-end
func TestJoinFlowE2E(t *testing.T) {
	t.Run("direct URL joining with name parameter", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room first
		room, err := h.store.CreateRoom()
		if err != nil {
			t.Fatal("failed to create room:", err)
		}
		roomCode := room.Code

		// Join directly with name in URL
		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name=Alice", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify player was added to room
		updatedRoom, _ := h.store.GetRoom(roomCode)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player in room, got %d", len(updatedRoom.Players))
		}

		// Verify player name
		var player *game.Player
		for _, p := range updatedRoom.Players {
			player = p
			break
		}
		if player == nil || player.Name != "Alice" {
			t.Error("player not found or name mismatch")
		}

		// Verify cookies were set
		cookies := resp.Cookies()
		var sessionCookie, playerCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "session" {
				sessionCookie = c
			} else if c.Name == "player_"+roomCode {
				playerCookie = c
			}
		}

		if sessionCookie == nil {
			t.Error("session cookie not set")
		}
		if playerCookie == nil {
			t.Error("player cookie not set")
		}

		// Verify response shows lobby page
		body := w.Body.String()
		if !strings.Contains(body, "Game Lobby") {
			t.Error("expected lobby page content")
		}
	})

	t.Run("join form submission flow", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// First, GET the join form
		req := httptest.NewRequest("GET", "/room/"+roomCode, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify join form is shown
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Join Room "+roomCode) {
			t.Error("expected join form title")
		}
		if !strings.Contains(body, "Enter your name") {
			t.Error("expected name input placeholder")
		}

		// Now submit the form
		req2 := httptest.NewRequest("GET", "/room/"+roomCode+"?name=Bob", nil)
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)

		// Verify successful join
		if w2.Code != http.StatusOK {
			t.Errorf("expected status 200 after join, got %d", w2.Code)
		}

		// Verify player was added
		updatedRoom, _ := h.store.GetRoom(roomCode)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player after join, got %d", len(updatedRoom.Players))
		}
	})

	t.Run("join invalid room shows 404", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Try to join non-existent room
		req := httptest.NewRequest("GET", "/room/INVALID", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify 404 response
		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Room not found") {
			t.Error("expected room not found message")
		}
	})

	t.Run("join full room shows error", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Fill the room to max capacity (8 players)
		for i := 1; i <= 8; i++ {
			player := game.NewPlayer(generatePlayerID(), "Player"+string(rune('0'+i)), "session"+string(rune('0'+i)))
			room.AddPlayer(player)
		}
		h.store.UpdateRoom(room)

		// Try to join the full room
		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name=ExtraPlayer", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify error response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "room is full") {
			t.Error("expected room full message")
		}
	})

	t.Run("join started game shows error", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room and start the game
		room, _ := h.store.CreateRoom()
		roomCode := room.Code
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Try to join the started game
		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name=LatePlayer", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify error response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Game already started") {
			t.Error("expected game already started message")
		}
	})

	t.Run("rejoin with existing session", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room and join
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// First join
		req1 := httptest.NewRequest("GET", "/room/"+roomCode+"?name=Charlie", nil)
		w1 := httptest.NewRecorder()

		router.ServeHTTP(w1, req1)

		// Get the player cookie
		var playerCookie *http.Cookie
		for _, c := range w1.Result().Cookies() {
			if c.Name == "player_"+roomCode {
				playerCookie = c
				break
			}
		}

		if playerCookie == nil {
			t.Fatal("player cookie not set on first join")
		}

		// Second request with the cookie (simulating page refresh)
		req2 := httptest.NewRequest("GET", "/room/"+roomCode, nil)
		req2.AddCookie(playerCookie)
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)

		// Should show lobby without creating new player
		if w2.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w2.Code)
		}

		// Verify still only 1 player
		updatedRoom, _ := h.store.GetRoom(roomCode)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected still 1 player, got %d", len(updatedRoom.Players))
		}

		// Verify shows lobby
		body := w2.Body.String()
		if !strings.Contains(body, "Game Lobby") {
			t.Error("expected lobby page on rejoin")
		}
	})

	t.Run("form submission with empty name", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Try to join with empty name
		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name=", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should show join form again (not join)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Join Room "+roomCode) {
			t.Error("expected join form when name is empty")
		}

		// Verify no player was added
		updatedRoom, _ := h.store.GetRoom(roomCode)
		if len(updatedRoom.Players) != 0 {
			t.Error("player should not be added with empty name")
		}
	})

	t.Run("URL encoding in player names", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Join with special characters in name
		specialName := "Player One & Two"
		encodedName := url.QueryEscape(specialName)

		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name="+encodedName, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify successful join
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify player name is correctly decoded
		updatedRoom, _ := h.store.GetRoom(roomCode)
		var player *game.Player
		for _, p := range updatedRoom.Players {
			player = p
			break
		}

		if player == nil || player.Name != specialName {
			t.Errorf("expected player name %q, got %q", specialName, player.Name)
		}
	})

	t.Run("concurrent joins to same room", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore())
		router := setupTestRouter(h)

		// Create a room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Simulate 5 concurrent joins
		var wg sync.WaitGroup
		joinCount := 5
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < joinCount; i++ {
			wg.Add(1)
			go func(playerNum int) {
				defer wg.Done()

				playerName := fmt.Sprintf("Player%d", playerNum)
				req := httptest.NewRequest("GET", "/room/"+roomCode+"?name="+playerName, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// All joins should succeed
		if successCount != joinCount {
			t.Errorf("expected %d successful joins, got %d", joinCount, successCount)
		}

		// Verify all players are in the room
		updatedRoom, _ := h.store.GetRoom(roomCode)
		if len(updatedRoom.Players) != joinCount {
			t.Errorf("expected %d players in room, got %d", joinCount, len(updatedRoom.Players))
		}
	})
}

// setupTestRouter creates a test router with all necessary routes
func setupTestRouter(h *Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/room/{code}", h.JoinRoom)
	router.Get("/game/{code}", h.GamePage)
	return router
}
