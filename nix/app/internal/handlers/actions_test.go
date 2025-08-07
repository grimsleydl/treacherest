package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		
		// Create a room with only 1 player (not enough to start)
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)
		
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
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
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
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", resp.StatusCode)
		}
		
		// Verify redirect to home
		location := resp.Header.Get("Location")
		if location != "/" {
			t.Errorf("expected redirect to /, got %s", location)
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
}