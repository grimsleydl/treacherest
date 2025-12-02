package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/game/ability"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// TestDismissModal tests the dismiss modal endpoint
func TestDismissModal(t *testing.T) {
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
	player1 := game.NewPlayer("player1", "Alice", "session1")
	room.AddPlayer(player1)

	// Add a pending ability for the player
	pendingAbility := &ability.PendingAbility{
		ID:             "ability-1",
		PlayerID:       "player1",
		CardID:         31,
		AbilityType:    "unveil",
		Data:           make(map[string]interface{}),
		ModalDismissed: false,
	}
	player1.AbilityState.AddPendingAbility(pendingAbility)

	memStore.UpdateRoom(room)

	t.Run("Dismiss modal successfully", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-1/dismiss", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "ability-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DismissModal(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify modal was dismissed
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		if !updatedPlayer.AbilityState.IsModalDismissed("ability-1") {
			t.Error("Expected modal to be dismissed")
		}
	})

	t.Run("Ability with wrong player ID returns forbidden", func(t *testing.T) {
		// Get fresh room from store
		freshRoom, _ := memStore.GetRoom(room.Code)

		player1Fresh := freshRoom.GetPlayer("player1")

		// Add ability to player1's state but with wrong PlayerID
		// This simulates data corruption or attack scenario
		pendingAbility3 := &ability.PendingAbility{
			ID:             "ability-3",
			PlayerID:       "player2", // Wrong player ID
			CardID:         25,
			AbilityType:    "unveil",
			ModalDismissed: false,
		}
		player1Fresh.AbilityState.AddPendingAbility(pendingAbility3)
		memStore.UpdateRoom(freshRoom)

		// Player1 tries to dismiss ability with wrong PlayerID
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-3/dismiss", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "ability-3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DismissModal(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("Ability not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/non-existent/dismiss", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "non-existent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DismissModal(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/INVALID/ability/ability-1/dismiss", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_INVALID",
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		rctx.URLParams.Add("abilityID", "ability-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DismissModal(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("No player cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-1/dismiss", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "ability-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DismissModal(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

// TestRestoreModal tests the restore modal endpoint
func TestRestoreModal(t *testing.T) {
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
	player1 := game.NewPlayer("player1", "Alice", "session1")
	room.AddPlayer(player1)

	// Add a pending ability that is dismissed
	pendingAbility := &ability.PendingAbility{
		ID:             "ability-1",
		PlayerID:       "player1",
		CardID:         31,
		AbilityType:    "unveil",
		Data:           make(map[string]interface{}),
		ModalDismissed: true, // Start dismissed
	}
	player1.AbilityState.AddPendingAbility(pendingAbility)

	memStore.UpdateRoom(room)

	t.Run("Restore modal successfully", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-1/restore", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "ability-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.RestoreModal(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify modal was restored
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		if updatedPlayer.AbilityState.IsModalDismissed("ability-1") {
			t.Error("Expected modal to not be dismissed after restore")
		}
	})

	t.Run("Ability with wrong player ID returns forbidden", func(t *testing.T) {
		// Get fresh room from store
		freshRoom, _ := memStore.GetRoom(room.Code)

		player1Fresh := freshRoom.GetPlayer("player1")

		// Add ability to player1's state but with wrong PlayerID
		pendingAbility3 := &ability.PendingAbility{
			ID:             "ability-3",
			PlayerID:       "player2", // Wrong player ID
			CardID:         25,
			AbilityType:    "unveil",
			ModalDismissed: true,
		}
		player1Fresh.AbilityState.AddPendingAbility(pendingAbility3)
		memStore.UpdateRoom(freshRoom)

		// Player1 tries to restore ability with wrong PlayerID
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-3/restore", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "ability-3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.RestoreModal(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})
}

// TestModalEventPublishing tests that events are published correctly
func TestModalEventPublishing(t *testing.T) {
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
	player1 := game.NewPlayer("player1", "Alice", "session1")
	room.AddPlayer(player1)

	// Add a pending ability
	pendingAbility := &ability.PendingAbility{
		ID:             "ability-1",
		PlayerID:       "player1",
		CardID:         31,
		AbilityType:    "unveil",
		ModalDismissed: false,
	}
	player1.AbilityState.AddPendingAbility(pendingAbility)

	memStore.UpdateRoom(room)

	// Subscribe to events
	eventChan := eventBus.Subscribe(room.Code)

	// Dismiss modal
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-1/dismiss", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("abilityID", "ability-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.DismissModal(w, req)

	// Wait for event
	select {
	case event := <-eventChan:
		if event.Type != "modal_dismissed" {
			t.Errorf("Expected event type 'modal_dismissed', got %s", event.Type)
		}
		if event.RoomCode != room.Code {
			t.Errorf("Expected room code %s, got %s", room.Code, event.RoomCode)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for modal_dismissed event")
	}

	// Now restore and check event
	req2 := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/ability-1/restore", nil)
	req2.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx))

	w2 := httptest.NewRecorder()

	handler.RestoreModal(w2, req2)

	// Wait for event
	select {
	case event := <-eventChan:
		if event.Type != "modal_restored" {
			t.Errorf("Expected event type 'modal_restored', got %s", event.Type)
		}
		if event.RoomCode != room.Code {
			t.Errorf("Expected room code %s, got %s", room.Code, event.RoomCode)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for modal_restored event")
	}
}
