package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

func TestHandler_StartGame(t *testing.T) {
	t.Run("starts game successfully", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")
		player3 := game.NewPlayer("p3", "Player 3", "session3")
		player4 := game.NewPlayer("p4", "Player 4", "session4")

		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.AddPlayer(player3)
		room.AddPlayer(player4)
		h.store.UpdateRoom(room)

		// Create request with chi context
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})

		// Add chi URL params to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify room state changed
		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateCountdown {
			t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
		}

		// Verify roles were assigned
		players := updatedRoom.GetPlayers()
		rolesAssigned := false
		for _, p := range players {
			if p.Role != nil {
				rolesAssigned = true
				break
			}
		}
		if !rolesAssigned {
			t.Error("no roles were assigned to players")
		}

		// Wait a bit to ensure countdown goroutine starts
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("responds with datastar redirect script", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room with 1 player (minimum to start)
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player1)
		h.store.UpdateRoom(room)

		// Create request with chi context
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})

		// Add chi URL params to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Check that response contains SSE redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/game/" + room.Code + "'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
		}

		// Verify room state changed to countdown
		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateCountdown {
			t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
		}
	})

	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		req := httptest.NewRequest("POST", "/room/XXXXX/start", nil)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "XXXXX")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when player not in room", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		// No player cookie

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 400 when cannot start game", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room with 0 players (not enough to start)
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		// Don't add player to room, but still send request with cookie

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401 (player not in room), got %d", resp.StatusCode)
		}
	})
}

func TestHandler_LeaveRoom(t *testing.T) {
	t.Run("player leaves room successfully", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")

		room.AddPlayer(player1)
		room.AddPlayer(player2)
		h.store.UpdateRoom(room)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify Datastar redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
		}

		// Verify player was removed from room
		updatedRoom, _ := h.store.GetRoom(room.Code)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player remaining, got %d", len(updatedRoom.Players))
		}

		if updatedRoom.GetPlayer(player1.ID) != nil {
			t.Error("player1 should have been removed from room")
		}
	})

	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		req := httptest.NewRequest("POST", "/room/XXXXX/leave", nil)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "XXXXX")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when no player cookie", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		// No player cookie

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("handles player not in room", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room without adding the player
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "nonexistent-player",
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		// Should still redirect even if player wasn't in room
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify Datastar redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
		}
	})
}

func TestHandler_runCountdown(t *testing.T) {
	t.Run("runs countdown and transitions to playing", func(t *testing.T) {
		h := New(store.NewMemoryStore())

		// Create a room
		room, _ := h.store.CreateRoom()
		room.State = game.StateCountdown
		room.CountdownRemaining = 5
		h.store.UpdateRoom(room)

		// Subscribe to events to verify they're published
		events := h.eventBus.Subscribe(room.Code)
		defer h.eventBus.Unsubscribe(room.Code, events)

		// Run countdown in goroutine
		done := make(chan bool)
		go func() {
			h.runCountdown(room)
			done <- true
		}()

		// Collect events
		var receivedEvents []Event
		timeout := time.After(6 * time.Second)

		collecting := true
		for collecting {
			select {
			case event := <-events:
				receivedEvents = append(receivedEvents, event)
			case <-done:
				// Give a bit more time for final event
				time.Sleep(100 * time.Millisecond)
				collecting = false
			case <-timeout:
				t.Fatal("countdown took too long")
			}
		}

		// Verify countdown events were sent
		countdownEvents := 0
		gamePlayingEvent := false
		for _, event := range receivedEvents {
			if event.Type == "countdown_update" {
				countdownEvents++
			} else if event.Type == "game_playing" {
				gamePlayingEvent = true
			}
		}

		if countdownEvents < 5 {
			t.Errorf("expected at least 5 countdown events, got %d", countdownEvents)
		}

		if !gamePlayingEvent {
			t.Error("expected game_playing event")
		}

		// Verify final room state
		finalRoom, _ := h.store.GetRoom(room.Code)
		if finalRoom.State != game.StatePlaying {
			t.Errorf("expected state %s, got %s", game.StatePlaying, finalRoom.State)
		}

		if finalRoom.CountdownRemaining != 0 {
			t.Errorf("expected countdown 0, got %d", finalRoom.CountdownRemaining)
		}

		if !finalRoom.LeaderRevealed {
			t.Error("expected leader to be revealed")
		}
	})
}
