package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/store"
)

func TestHandler_Home(t *testing.T) {
	h := New(store.NewMemoryStore())

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.Home(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected HTML content type, got %s", contentType)
	}

	// Verify some content was rendered
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestHandler_CreateRoom(t *testing.T) {
	t.Run("creates room successfully", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		form := url.Values{}
		form.Add("playerName", "Test Player")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", resp.StatusCode)
		}

		// Verify redirect location
		location := resp.Header.Get("Location")
		if !strings.HasPrefix(location, "/room/") {
			t.Errorf("expected redirect to /room/*, got %s", location)
		}

		// Extract room code from location
		roomCode := strings.TrimPrefix(location, "/room/")
		if len(roomCode) != 5 {
			t.Errorf("expected 5 character room code, got %s", roomCode)
		}

		// Verify room was created in store
		room, err := h.store.GetRoom(roomCode)
		if err != nil {
			t.Fatalf("room not found in store: %v", err)
		}

		// Verify player was added
		if len(room.Players) != 1 {
			t.Errorf("expected 1 player in room, got %d", len(room.Players))
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
	})

	t.Run("returns error when player name is empty", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		form := url.Values{}
		form.Add("playerName", "")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Player name is required") {
			t.Errorf("expected error message about player name, got %s", body)
		}
	})

	t.Run("preserves existing session", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		existingSession := "existing-session-123"

		form := url.Values{}
		form.Add("playerName", "Test Player")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: existingSession,
		})
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		// Should not create new session cookie
		cookies := w.Result().Cookies()
		for _, c := range cookies {
			if c.Name == "session" {
				t.Error("should not create new session cookie when one exists")
			}
		}

		// Verify player has the existing session ID
		location := w.Result().Header.Get("Location")
		roomCode := strings.TrimPrefix(location, "/room/")
		room, _ := h.store.GetRoom(roomCode)

		for _, player := range room.Players {
			if player.SessionID != existingSession {
				t.Errorf("expected player to have session %s, got %s", existingSession, player.SessionID)
			}
		}
	})
}

func TestHandler_JoinRoom(t *testing.T) {
	t.Run("renders join form when no name provided", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room first
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Create a router to properly handle URL params
		router := chi.NewRouter()
		router.Get("/room/{code}", h.JoinRoom)

		req := httptest.NewRequest("GET", "/room/"+roomCode, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		// Verify the Templ template is rendered
		if !strings.Contains(body, "Join Room "+roomCode) {
			t.Error("expected join form with room code")
		}
		if !strings.Contains(body, `name="name"`) {
			t.Error("expected name input field")
		}
		if !strings.Contains(body, "Enter your name") {
			t.Error("expected placeholder text")
		}
		if !strings.Contains(body, `data-store="{}"`) {
			t.Error("expected datastar attributes")
		}
	})
}

func TestHandler_GamePage(t *testing.T) {
	t.Run("shows game page for player in game", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Test Player", "session1")
		player.Role = mockGuardianCard()
		room.AddPlayer(player)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Create a router to handle URL params
		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		// Verify game page content - check for data-star or data- attributes
		if !strings.Contains(body, "data-") {
			t.Error("expected data attributes in game page")
		}
		// Verify it's HTML content
		if len(body) == 0 {
			t.Error("expected non-empty game page body")
		}
	})

	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/XXXXX", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when no player cookie", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room
		room, _ := h.store.CreateRoom()

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when player not in room", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room
		room, _ := h.store.CreateRoom()

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "nonexistent-player",
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})
}
