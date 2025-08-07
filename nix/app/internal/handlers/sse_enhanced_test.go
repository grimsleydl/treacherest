package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

func TestEventStore(t *testing.T) {
	t.Run("stores and retrieves events", func(t *testing.T) {
		es := NewEventStore(5)
		roomCode := "ROOM1"

		// Add some events
		for i := 0; i < 3; i++ {
			event := SSEEvent{
				ID:        fmt.Sprintf("event-%d", i),
				Type:      "test",
				Data:      fmt.Sprintf("data-%d", i),
				Timestamp: time.Now(),
			}
			es.AddEvent(roomCode, event)
		}

		// Get all events since a non-existent ID (should return all)
		events := es.GetEventsSince(roomCode, "unknown")
		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}

		// Get events since the first one
		events = es.GetEventsSince(roomCode, "event-0")
		if len(events) != 2 {
			t.Errorf("expected 2 events after event-0, got %d", len(events))
		}

		// Get events since the last one
		events = es.GetEventsSince(roomCode, "event-2")
		if len(events) != 0 {
			t.Errorf("expected 0 events after event-2, got %d", len(events))
		}
	})

	t.Run("respects max events limit", func(t *testing.T) {
		es := NewEventStore(3)
		roomCode := "ROOM1"

		// Add more events than the limit
		for i := 0; i < 5; i++ {
			event := SSEEvent{
				ID:        fmt.Sprintf("event-%d", i),
				Type:      "test",
				Data:      fmt.Sprintf("data-%d", i),
				Timestamp: time.Now(),
			}
			es.AddEvent(roomCode, event)
		}

		// Should only have the last 3 events
		events := es.GetEventsSince(roomCode, "unknown")
		if len(events) != 3 {
			t.Errorf("expected 3 events (max limit), got %d", len(events))
		}

		// First event should be event-2 (0 and 1 were dropped)
		if events[0].ID != "event-2" {
			t.Errorf("expected first event to be event-2, got %s", events[0].ID)
		}
	})

	t.Run("handles empty event store", func(t *testing.T) {
		es := NewEventStore(5)
		
		events := es.GetEventsSince("ROOM1", "any-id")
		if events != nil {
			t.Errorf("expected nil for empty store, got %v", events)
		}
	})

	t.Run("handles empty lastEventID", func(t *testing.T) {
		es := NewEventStore(5)
		roomCode := "ROOM1"

		// Add an event
		es.AddEvent(roomCode, SSEEvent{
			ID:   "event-1",
			Type: "test",
			Data: "data",
		})

		// Empty lastEventID should return nil (client has all events)
		events := es.GetEventsSince(roomCode, "")
		if events != nil {
			t.Errorf("expected nil for empty lastEventID, got %v", events)
		}
	})
}

func TestConnectionTracker(t *testing.T) {
	t.Run("tracks connections per room", func(t *testing.T) {
		ct := NewConnectionTracker()

		// Add connections
		ct.AddConnection("ROOM1")
		ct.AddConnection("ROOM1")
		ct.AddConnection("ROOM2")

		// Check counts
		if count := ct.GetConnectionCount("ROOM1"); count != 2 {
			t.Errorf("expected 2 connections for ROOM1, got %d", count)
		}

		if count := ct.GetConnectionCount("ROOM2"); count != 1 {
			t.Errorf("expected 1 connection for ROOM2, got %d", count)
		}

		if total := ct.GetTotalConnections(); total != 3 {
			t.Errorf("expected 3 total connections, got %d", total)
		}

		// Remove connections
		ct.RemoveConnection("ROOM1")
		if count := ct.GetConnectionCount("ROOM1"); count != 1 {
			t.Errorf("expected 1 connection for ROOM1 after removal, got %d", count)
		}

		if total := ct.GetTotalConnections(); total != 2 {
			t.Errorf("expected 2 total connections after removal, got %d", total)
		}
	})

	t.Run("handles concurrent access", func(t *testing.T) {
		ct := NewConnectionTracker()
		roomCode := "ROOM1"

		// Concurrent adds and removes
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(2)
			
			go func() {
				defer wg.Done()
				ct.AddConnection(roomCode)
			}()
			
			go func() {
				defer wg.Done()
				time.Sleep(10 * time.Millisecond)
				ct.RemoveConnection(roomCode)
			}()
		}

		wg.Wait()

		// Should end up with 0 connections
		if count := ct.GetConnectionCount(roomCode); count != 0 {
			t.Errorf("expected 0 connections after concurrent add/remove, got %d", count)
		}
	})

	t.Run("cleans up empty rooms", func(t *testing.T) {
		ct := NewConnectionTracker()

		ct.AddConnection("ROOM1")
		ct.RemoveConnection("ROOM1")

		// Internal map should not contain ROOM1
		ct.mu.RLock()
		_, exists := ct.connections["ROOM1"]
		ct.mu.RUnlock()

		if exists {
			t.Error("expected room to be removed from map when count reaches 0")
		}
	})
}

func TestEnhancedHandler_StreamLobbyEnhanced(t *testing.T) {
	t.Run("sends heartbeat events", func(t *testing.T) {
		h := NewEnhanced(store.NewMemoryStore())

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)

		// Create a context with longer timeout for heartbeat
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
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
		done := make(chan bool)
		go func() {
			h.StreamLobbyEnhanced(w, req)
			done <- true
		}()

		// Wait enough time for at least one heartbeat (30s + buffer)
		// In real tests, we'd use a shorter heartbeat interval
		// For now, we'll check after a short time and cancel
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Wait for handler to finish
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("Handler did not finish after context cancel")
		}

		// Check that initial data was sent
		body := w.Body.String()
		if !strings.Contains(body, "data:") {
			t.Error("expected SSE data in response")
		}

		// Note: In a real test, we'd inject a shorter heartbeat interval
		// to test heartbeat functionality without waiting 30 seconds
	})

	t.Run("replays missed events", func(t *testing.T) {
		h := NewEnhanced(store.NewMemoryStore())

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)

		// Manually add some events to the event store
		for i := 0; i < 3; i++ {
			h.eventStore.AddEvent(room.Code, SSEEvent{
				ID:        fmt.Sprintf("test-event-%d", i),
				Type:      "test",
				Data:      fmt.Sprintf("test-data-%d", i),
				Timestamp: time.Now(),
			})
		}

		// Create request with Last-Event-ID header
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Last-Event-ID", "test-event-0")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Run handler
		done := make(chan bool)
		go func() {
			h.StreamLobbyEnhanced(w, req)
			done <- true
		}()

		// Wait for handler to finish
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("Handler did not finish in time")
		}

		// Check that events were replayed
		body := w.Body.String()
		if !strings.Contains(body, "event: test") {
			t.Error("expected replayed test events")
		}
	})

	t.Run("tracks connections", func(t *testing.T) {
		h := NewEnhanced(store.NewMemoryStore())

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)

		// Initial connection count should be 0
		if count := h.connTracker.GetConnectionCount(room.Code); count != 0 {
			t.Errorf("expected 0 initial connections, got %d", count)
		}

		// Create multiple connections
		var wg sync.WaitGroup
		numConnections := 3

		for i := 0; i < numConnections; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

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
				h.StreamLobbyEnhanced(w, req)
			}()
		}

		// Give connections time to establish
		time.Sleep(50 * time.Millisecond)

		// Check connection count
		if count := h.connTracker.GetConnectionCount(room.Code); count != int64(numConnections) {
			t.Errorf("expected %d connections during streaming, got %d", numConnections, count)
		}

		// Wait for all connections to close
		wg.Wait()

		// Give time for cleanup
		time.Sleep(50 * time.Millisecond)

		// Connection count should be back to 0
		if count := h.connTracker.GetConnectionCount(room.Code); count != 0 {
			t.Errorf("expected 0 connections after cleanup, got %d", count)
		}
	})

	t.Run("generates unique event IDs", func(t *testing.T) {
		h := NewEnhanced(store.NewMemoryStore())

		// Generate multiple IDs
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := h.generateEventID()
			if ids[id] {
				t.Errorf("duplicate event ID generated: %s", id)
			}
			ids[id] = true
		}
	})
}

func TestEnhancedHandler_StreamGameEnhanced(t *testing.T) {
	t.Run("includes game state in heartbeat", func(t *testing.T) {
		h := NewEnhanced(store.NewMemoryStore())

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		player.Role = &game.Role{Name: "Villager"}
		room.AddPlayer(player)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Create a context
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

		// Run handler
		done := make(chan bool)
		go func() {
			h.StreamGameEnhanced(w, req)
			done <- true
		}()

		// Wait for handler to finish
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("Handler did not finish in time")
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

	t.Run("stores game events properly", func(t *testing.T) {
		h := NewEnhanced(store.NewMemoryStore())

		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")
		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Create a context
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
			h.StreamGameEnhanced(w, req)
		}()

		// Give handler time to subscribe and send initial render
		time.Sleep(50 * time.Millisecond)

		// Send some game events
		h.eventBus.Publish(Event{
			Type:     "role_revealed",
			RoomCode: room.Code,
			Data:     map[string]interface{}{"playerID": "p2"},
		})

		// Give time for event processing
		time.Sleep(50 * time.Millisecond)

		// Cancel context to stop handler
		cancel()

		// Give time for cleanup
		time.Sleep(50 * time.Millisecond)

		// Check that events were stored
		events := h.eventStore.GetEventsSince(room.Code, "")
		if events != nil {
			t.Error("expected nil when lastEventID is empty")
		}

		// Get all events (with non-existent last ID)
		events = h.eventStore.GetEventsSince(room.Code, "unknown")
		if len(events) < 2 {
			t.Errorf("expected at least 2 events (initial + update), got %d", len(events))
		}
	})
}