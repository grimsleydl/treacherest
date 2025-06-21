package handlers

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"treacherest/internal/store"
)

func TestJoinFlowBackNavigation(t *testing.T) {
	h := New(store.NewMemoryStore())

	// Create first room
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=Alice"))
	r1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.CreateRoom(w1, r1)

	// Extract room code from redirect
	location1 := w1.Header().Get("Location")
	roomCode1 := strings.TrimPrefix(location1, "/room/")
	t.Logf("Created room 1: %s", roomCode1)

	// Extract session cookie
	cookies := w1.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("No session cookie found")
	}

	// Create second room
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=Bob"))
	r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.CreateRoom(w2, r2)

	location2 := w2.Header().Get("Location")
	roomCode2 := strings.TrimPrefix(location2, "/room/")
	t.Logf("Created room 2: %s", roomCode2)

	// Scenario 1: Try to join room 1 after being in room 2
	// This simulates user pressing back button or entering URL directly
	t.Run("JoinFirstRoomAfterSecond", func(t *testing.T) {
		// Join room 1
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/room/%s?name=Charlie", roomCode1), nil)
		r.AddCookie(sessionCookie)
		
		// Wrap handler to capture routing params
		testRouter := setupTestRouter(h)
		testRouter.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		// Check that the player was added to room 1
		room1, _ := h.store.GetRoom(roomCode1)
		if len(room1.Players) != 2 { // Alice + Charlie
			t.Errorf("Expected 2 players in room 1, got %d", len(room1.Players))
		}

		// Check cookie was set for room 1
		playerCookie := getPlayerCookie(w.Result().Cookies(), roomCode1)
		if playerCookie == nil {
			t.Error("Player cookie not set for room 1")
		}
	})

	// Scenario 2: Try to join a different room in the same tab
	t.Run("JoinDifferentRoomSameSession", func(t *testing.T) {
		// First join room 1
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest("GET", fmt.Sprintf("/room/%s?name=Dave", roomCode1), nil)
		r1.AddCookie(sessionCookie)
		
		testRouter := setupTestRouter(h)
		testRouter.ServeHTTP(w1, r1)

		if w1.Code != http.StatusOK {
			t.Errorf("Expected status 200 for room 1, got %d", w1.Code)
		}

		// Get the player cookie for room 1
		playerCookie1 := getPlayerCookie(w1.Result().Cookies(), roomCode1)
		if playerCookie1 == nil {
			t.Fatal("No player cookie for room 1")
		}

		// Now try to join room 2
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", fmt.Sprintf("/room/%s?name=Dave", roomCode2), nil)
		r2.AddCookie(sessionCookie)
		r2.AddCookie(playerCookie1) // This shouldn't interfere
		
		testRouter.ServeHTTP(w2, r2)

		if w2.Code != http.StatusOK {
			t.Errorf("Expected status 200 for room 2, got %d", w2.Code)
			t.Logf("Response body: %s", w2.Body.String())
		}

		// Check that player was added to room 2
		room2, _ := h.store.GetRoom(roomCode2)
		found := false
		for _, p := range room2.Players {
			if p.Name == "Dave" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Dave not found in room 2")
		}

		// Check cookie was set for room 2
		playerCookie2 := getPlayerCookie(w2.Result().Cookies(), roomCode2)
		if playerCookie2 == nil {
			t.Error("Player cookie not set for room 2")
		}
	})

	// Scenario 3: Multiple SSE connections
	t.Run("MultipleSSEConnections", func(t *testing.T) {
		// Join room as a player
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/room/%s?name=Eve", roomCode1), nil)
		r.AddCookie(sessionCookie)
		
		testRouter := setupTestRouter(h)
		testRouter.ServeHTTP(w, r)

		playerCookie := getPlayerCookie(w.Result().Cookies(), roomCode1)
		if playerCookie == nil {
			t.Fatal("No player cookie")
		}

		// Start SSE connection for room 1
		sse1Done := make(chan bool)
		go func() {
			defer close(sse1Done)
			
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", fmt.Sprintf("/sse/lobby/%s", roomCode1), nil)
			r.AddCookie(sessionCookie)
			r.AddCookie(playerCookie)
			
			// Create context that we can cancel
			ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
			defer cancel()
			r = r.WithContext(ctx)
			
			testRouter.ServeHTTP(w, r)
		}()

		// Wait a bit for SSE to establish
		time.Sleep(50 * time.Millisecond)

		// Now try to join room 2 (simulating navigation without proper cleanup)
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", fmt.Sprintf("/room/%s?name=Eve", roomCode2), nil)
		r2.AddCookie(sessionCookie)
		
		testRouter.ServeHTTP(w2, r2)

		if w2.Code != http.StatusOK {
			t.Errorf("Expected status 200 for room 2, got %d", w2.Code)
		}

		// Wait for SSE to finish
		select {
		case <-sse1Done:
			// Good, SSE connection closed
		case <-time.After(200 * time.Millisecond):
			// SSE should have timed out by now
		}
	})
}

func getPlayerCookie(cookies []*http.Cookie, roomCode string) *http.Cookie {
	cookieName := "player_" + roomCode
	for _, c := range cookies {
		if c.Name == cookieName {
			return c
		}
	}
	return nil
}

func setupTestRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	
	// Room routes
	r.Post("/room/create", h.CreateRoom)
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Post("/room/{code}/start", h.StartGame)
	
	// SSE routes
	r.Get("/sse/lobby/{code}", h.StreamLobby)
	r.Get("/sse/game/{code}", h.StreamGame)
	
	// Game routes
	r.Get("/game/{code}", h.GamePage)
	
	return r
}

func TestJoinFlowBrowserBackButton(t *testing.T) {
	h := New(store.NewMemoryStore())
	testRouter := setupTestRouter(h)
	
	// Create a room
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=Alice"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.CreateRoom(w, r)
	
	location := w.Header().Get("Location")
	roomCode := strings.TrimPrefix(location, "/room/")
	
	// Get cookies
	cookies := w.Result().Cookies()
	var sessionCookie, playerCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
		} else if c.Name == "player_"+roomCode {
			playerCookie = c
		}
	}
	
	t.Run("RejoinAfterLeavingRoom", func(t *testing.T) {
		// Leave the room
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", fmt.Sprintf("/room/%s/leave", roomCode), nil)
		r.AddCookie(sessionCookie)
		r.AddCookie(playerCookie)
		
		testRouter.ServeHTTP(w, r)
		
		// Check for Datastar redirect script
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		// Verify Datastar redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("Expected response to contain redirect script %q, got: %s", expectedScript, body)
		}
		
		// Check cookie was cleared
		clearedCookie := false
		for _, c := range w.Result().Cookies() {
			if c.Name == "player_"+roomCode && c.MaxAge < 0 {
				clearedCookie = true
				break
			}
		}
		if !clearedCookie {
			t.Error("Player cookie was not cleared")
		}
		
		// Now try to join the same room again (simulating back button)
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", fmt.Sprintf("/room/%s?name=Alice", roomCode), nil)
		r2.AddCookie(sessionCookie)
		
		testRouter.ServeHTTP(w2, r2)
		
		if w2.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w2.Code)
			t.Logf("Response: %s", w2.Body.String())
		}
		
		// Verify player was re-added to room
		room, _ := h.store.GetRoom(roomCode)
		found := false
		for _, p := range room.Players {
			if p.Name == "Alice" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Alice was not re-added to the room")
		}
	})
	
	t.Run("StalePlayerCookieHandling", func(t *testing.T) {
		// Create a fake player cookie for a non-existent player
		staleCookie := &http.Cookie{
			Name:  "player_" + roomCode,
			Value: "nonexistentplayer",
			Path:  "/",
		}
		
		// Try to access room with stale cookie
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/room/%s", roomCode), nil)
		r.AddCookie(sessionCookie)
		r.AddCookie(staleCookie)
		
		testRouter.ServeHTTP(w, r)
		
		// Should show join form since player doesn't exist
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		// Check that stale cookie was cleared
		cookieCleared := false
		for _, c := range w.Result().Cookies() {
			if c.Name == "player_"+roomCode && c.MaxAge < 0 {
				cookieCleared = true
				break
			}
		}
		if !cookieCleared {
			t.Error("Stale player cookie was not cleared")
		}
		
		// Verify join form is shown
		body := w.Body.String()
		if !strings.Contains(body, "Enter your name") {
			t.Error("Join form not shown for stale cookie")
		}
	})
}