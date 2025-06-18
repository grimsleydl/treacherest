package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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
	// This test would require mocking chi.URLParam, which is complex
	// For now, we'll focus on unit testing the handler logic
	// Integration tests will cover the full flow
	t.Skip("JoinRoom requires chi router context - will be covered in integration tests")
}

func TestHandler_GamePage(t *testing.T) {
	// Similar to JoinRoom, requires chi router context
	t.Skip("GamePage requires chi router context - will be covered in integration tests")
}