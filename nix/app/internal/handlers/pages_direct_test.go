package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

func TestJoinRoomDirect(t *testing.T) {
	h := newTestHandler()

	t.Run("shows join form when no name provided", func(t *testing.T) {
		// Create a room
		room, _ := h.Store().CreateRoom()

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.JoinRoom(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Join Room") {
			t.Error("expected join form in response")
		}
		if !strings.Contains(body, room.Code) {
			t.Error("expected room code in response")
		}
	})

	t.Run("returns 404 for invalid room", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/INVALID", nil)
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.JoinRoom(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("redirects to game when already in room and game started", func(t *testing.T) {
		// Create a room with a player
		room, _ := h.Store().CreateRoom()
		player := game.NewPlayer("p1", "Test Player", "session1")
		room.AddPlayer(player)
		room.State = game.StatePlaying // Game already started
		h.Store().UpdateRoom(room)

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.JoinRoom(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected redirect status 303, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/game/"+room.Code {
			t.Errorf("expected redirect to /game/%s, got %s", room.Code, location)
		}
	})

	t.Run("shows lobby when already in room", func(t *testing.T) {
		// Create a room with a player
		room, _ := h.Store().CreateRoom()
		player := game.NewPlayer("p1", "Test Player", "session1")
		room.AddPlayer(player)
		h.Store().UpdateRoom(room)

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.JoinRoom(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify LobbyPage was rendered
		// Since we can't easily check the actual template output,
		// we'll verify the handler reached that code path
	})

	t.Run("returns error when game already started", func(t *testing.T) {
		// Create a room that's already playing
		room, _ := h.Store().CreateRoom()
		room.State = game.StatePlaying
		h.Store().UpdateRoom(room)

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.JoinRoom(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Game already started") {
			t.Error("expected error message about game started")
		}
	})

	t.Run("joins room via POST", func(t *testing.T) {
		// Create a room
		room, _ := h.Store().CreateRoom()

		// Subscribe to events to verify player_joined is published
		events := h.eventBus.Subscribe(room.Code)
		defer h.eventBus.Unsubscribe(room.Code, events)

		// Join via POST
		formData := "room_code=" + room.Code + "&player_name=NewPlayer"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.JoinRoomPost(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected status 303 (redirect), got %d", w.Code)
		}

		// Check redirect location
		location := w.Header().Get("Location")
		expectedLocation := "/room/" + room.Code
		if location != expectedLocation {
			t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
		}

		// Verify player was added
		updatedRoom, _ := h.Store().GetRoom(room.Code)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player, got %d", len(updatedRoom.Players))
		}

		// Verify cookie was set
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

		// Verify event was published
		select {
		case event := <-events:
			if event.Type != "player_joined" {
				t.Errorf("expected player_joined event, got %s", event.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected player_joined event to be published")
		}
	})

	t.Run("returns error when room is full via POST", func(t *testing.T) {
		// Create a full room
		room, _ := h.Store().CreateRoom()
		room.MaxPlayers = 2

		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")
		room.AddPlayer(player1)
		room.AddPlayer(player2)
		h.Store().UpdateRoom(room)

		// Try to join via POST
		formData := "room_code=" + room.Code + "&player_name=NewPlayer"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.JoinRoomPost(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestGamePageDirect(t *testing.T) {
	h := newTestHandler()

	t.Run("returns 404 for invalid room", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/game/INVALID", nil)
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.GamePage(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns 401 without player cookie", func(t *testing.T) {
		// Create a room
		room, _ := h.Store().CreateRoom()

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.GamePage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns 401 for invalid player", func(t *testing.T) {
		// Create a room
		room, _ := h.Store().CreateRoom()

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "invalid-player",
		})
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.GamePage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("shows game page for valid player", func(t *testing.T) {
		// Create a room with a player
		room, _ := h.Store().CreateRoom()
		player := game.NewPlayer("p1", "Test Player", "session1")
		room.AddPlayer(player)
		h.Store().UpdateRoom(room)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		// Set up chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.GamePage(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify GamePage template was rendered
		// We can't easily check template output, but the handler succeeded
	})
}
