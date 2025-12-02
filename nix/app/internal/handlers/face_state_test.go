package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// TestToggleFaceState tests the face state toggle handler
func TestToggleFaceState(t *testing.T) {
	// Setup
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	player2 := game.NewPlayer("player2", "Bob", "session2")
	room.AddPlayer(player1)
	room.AddPlayer(player2)
	memStore.UpdateRoom(room)

	t.Run("Toggle face state successfully", func(t *testing.T) {
		// Setup request
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/player1", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		// Setup chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Execute
		handler.ToggleFaceState(w, req)

		// Verify
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check that face state was toggled
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		if updatedPlayer.FaceUp {
			t.Error("Expected FaceUp to be false after first toggle")
		}

		// Toggle again
		req2 := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/player1", nil)
		req2.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx))
		w2 := httptest.NewRecorder()

		handler.ToggleFaceState(w2, req2)

		// Verify toggled back
		updatedRoom2, _ := memStore.GetRoom(room.Code)
		updatedPlayer2 := updatedRoom2.GetPlayer("player1")
		if !updatedPlayer2.FaceUp {
			t.Error("Expected FaceUp to be true after second toggle")
		}
	})

	t.Run("Cannot toggle other player's face state", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/player2", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1", // player1 trying to toggle player2
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.ToggleFaceState(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/INVALID/facestate/player1", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_INVALID",
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		rctx.URLParams.Add("playerID", "player1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.ToggleFaceState(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("No player cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/player1", nil)
		// No cookie set

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.ToggleFaceState(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Target player not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/nonexistent", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.ToggleFaceState(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestToggleFaceStateWithTransformation tests face state toggle with Wearer transformation
func TestToggleFaceStateWithTransformation(t *testing.T) {
	// Setup
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with card pool
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	room.AddPlayer(player1)

	// Setup transformation state
	originalCard := &game.Card{ID: 31, Name: "The Wearer of Masks"}
	transformedCard := &game.Card{ID: 25, Name: "The Metamorph"}
	player1.Role = transformedCard
	player1.AbilityState.StartTransform(31, 25, []string{"Traitor"}, "face_down")

	// Add cards to pool - explicitly set to ensure our test cards are available
	room.CardPool = game.NewCardPool([]*game.Card{originalCard, transformedCard})

	memStore.UpdateRoom(room)

	t.Run("Transformation ends when turned face down", func(t *testing.T) {
		// Verify transformation is active
		if !player1.AbilityState.TransformState.IsTransformed {
			t.Fatal("Expected transformation to be active")
		}

		// Turn face down
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/player1", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.ToggleFaceState(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify transformation ended
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")

		if updatedPlayer.FaceUp {
			t.Error("Expected FaceUp to be false after toggle")
		}

		if updatedPlayer.AbilityState.TransformState != nil && updatedPlayer.AbilityState.TransformState.IsTransformed {
			t.Error("Expected transformation to end when turned face down")
		}

		// Verify role restored to original
		if updatedPlayer.Role.GetID() != 31 {
			t.Errorf("Expected role to be restored to original (ID 31), got ID %d", updatedPlayer.Role.GetID())
		}
	})

	t.Run("Transformation persists when turned face up", func(t *testing.T) {
		// Setup fresh transformation
		player2 := game.NewPlayer("player2", "Bob", "session2")
		room2, _ := memStore.CreateRoom()
		room2.AddPlayer(player2)

		transformCard := &game.Card{ID: 20, Name: "Test Card"}
		player2.Role = transformCard
		player2.AbilityState.StartTransform(31, 20, []string{"Traitor"}, "face_down")
		player2.FaceUp = false // Start face down

		room2.CardPool = game.NewCardPool([]*game.Card{originalCard, transformCard})
		memStore.UpdateRoom(room2)

		// Turn face up (should not end transformation)
		req := httptest.NewRequest("POST", "/room/"+room2.Code+"/facestate/player2", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room2.Code,
			Value: "player2",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room2.Code)
		rctx.URLParams.Add("playerID", "player2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.ToggleFaceState(w, req)

		// Verify transformation still active
		updatedRoom, _ := memStore.GetRoom(room2.Code)
		updatedPlayer := updatedRoom.GetPlayer("player2")

		if !updatedPlayer.FaceUp {
			t.Error("Expected FaceUp to be true")
		}

		if !updatedPlayer.AbilityState.TransformState.IsTransformed {
			t.Error("Expected transformation to persist when turned face up")
		}

		if updatedPlayer.Role.GetID() != 20 {
			t.Errorf("Expected role to remain transformed (ID 20), got ID %d", updatedPlayer.Role.GetID())
		}
	})
}

// TestToggleFaceStateEventPublishing tests that events are published correctly
func TestToggleFaceStateEventPublishing(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room and player
	room, _ := memStore.CreateRoom()

	// Subscribe to events
	eventChan := eventBus.Subscribe(room.Code)
	player1 := game.NewPlayer("player1", "Alice", "session1")
	room.AddPlayer(player1)
	memStore.UpdateRoom(room)

	// Execute toggle
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/facestate/player1", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ToggleFaceState(w, req)

	// Wait for event
	select {
	case event := <-eventChan:
		if event.Type != "face_state_changed" {
			t.Errorf("Expected event type 'face_state_changed', got %s", event.Type)
		}
		if event.RoomCode != room.Code {
			t.Errorf("Expected room code %s, got %s", room.Code, event.RoomCode)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for face_state_changed event")
	}
}
