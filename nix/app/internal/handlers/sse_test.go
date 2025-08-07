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

func TestHandler_StreamLobby(t *testing.T) {
	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		req := httptest.NewRequest("GET", "/room/XXXXX/lobby", nil)
		req.Header.Set("Accept", "text/event-stream")
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "XXXXX")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		h.StreamLobby(w, req)
		
		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
	
	t.Run("returns 401 when no player cookie", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		// Create a room
		room, _ := h.store.CreateRoom()
		
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req.Header.Set("Accept", "text/event-stream")
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		h.StreamLobby(w, req)
		
		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})
	
	t.Run("returns 401 when player not in room", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		// Create a room
		room, _ := h.store.CreateRoom()
		
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "nonexistent-player",
		})
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		h.StreamLobby(w, req)
		
		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})
	
	t.Run("streams lobby updates successfully", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)
		
		// Create a context with timeout to prevent hanging
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		// Run in goroutine since it blocks
		done := make(chan bool)
		go func() {
			h.StreamLobby(w, req)
			done <- true
		}()
		
		// Wait for handler to finish or timeout
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("StreamLobby did not finish in time")
		}
		
		// Check response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		// Verify SSE headers
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			t.Errorf("expected SSE content type, got %s", contentType)
		}
		
		// Verify some SSE data was sent
		body := w.Body.String()
		if !strings.Contains(body, "data:") {
			t.Error("expected SSE data in response body")
		}
	})
	
	t.Run("sends game_started event", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)
		
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		// Run handler in goroutine
		go func() {
			h.StreamLobby(w, req)
		}()
		
		// Give handler time to subscribe
		time.Sleep(50 * time.Millisecond)
		
		// Send game_started event
		h.eventBus.Publish(Event{
			Type:     "game_started",
			RoomCode: room.Code,
		})
		
		// Wait for context to timeout
		<-ctx.Done()
		
		// Check if redirect script was sent
		body := w.Body.String()
		if !strings.Contains(body, "window.location.href") {
			t.Error("expected redirect script in response")
		}
		if !strings.Contains(body, "/game/"+room.Code) {
			t.Error("expected redirect to game page")
		}
	})
}

func TestHandler_StreamGame(t *testing.T) {
	t.Run("streams game updates successfully", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		player.Role = &game.Role{Name: "Villager"}
		room.AddPlayer(player)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)
		
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/game", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		// Run in goroutine
		done := make(chan bool)
		go func() {
			h.StreamGame(w, req)
			done <- true
		}()
		
		// Wait for handler to finish
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("StreamGame did not finish in time")
		}
		
		// Check response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		// Verify SSE data was sent
		body := w.Body.String()
		if !strings.Contains(body, "data:") {
			t.Error("expected SSE data in response body")
		}
	})
	
	t.Run("updates on event", func(t *testing.T) {
		h := New(store.NewMemoryStore())
		
		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")
		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)
		
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/game", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})
		
		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()
		
		// Run handler in goroutine
		go func() {
			h.StreamGame(w, req)
		}()
		
		// Give handler time to subscribe and send initial render
		time.Sleep(50 * time.Millisecond)
		
		// Record initial response size
		initialSize := w.Body.Len()
		
		// Send an event to trigger update
		h.eventBus.Publish(Event{
			Type:     "role_revealed",
			RoomCode: room.Code,
			Data:     map[string]interface{}{"playerID": "p2"},
		})
		
		// Give time for update to be sent
		time.Sleep(50 * time.Millisecond)
		
		// Verify additional data was sent
		if w.Body.Len() <= initialSize {
			t.Error("expected additional SSE data after event")
		}
		
		// Cancel context to stop handler
		cancel()
	})
}

func TestRenderToString(t *testing.T) {
	// This test would require a real Templ component
	// Since renderToString is a simple wrapper, we'll skip for now
	t.Skip("Requires real Templ component - covered by integration tests")
}

func TestSSEHelpers(t *testing.T) {
	t.Run("renderLobby sends correct selector", func(t *testing.T) {
		// This test would require mocking datastar.ServerSentEventGenerator
		// which is complex. Will be covered in integration tests.
		t.Skip("Requires datastar SSE mock - covered in integration tests")
	})
	
	t.Run("renderGame sends correct selector", func(t *testing.T) {
		// Similar to above
		t.Skip("Requires datastar SSE mock - covered in integration tests")
	})
}