package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// TestTriggerWearerAbility tests triggering The Wearer of Masks ability
func TestTriggerWearerAbility(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with CardPool
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")

	// Give player1 The Wearer of Masks (card ID 31)
	wearerCard := &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	player1.Role = wearerCard

	room.AddPlayer(player1)

	// Initialize CardPool with some cards
	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
		{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
	}
	room.CardPool = game.NewCardPool(availableCards)

	memStore.UpdateRoom(room)

	t.Run("Trigger ability successfully", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify pending ability was created
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")

		if !updatedPlayer.AbilityState.HasPendingAbilities() {
			t.Error("Expected pending ability to be created")
		}

		// Verify ability data
		abilities := updatedPlayer.AbilityState.PendingAbilities
		if len(abilities) != 1 {
			t.Fatalf("Expected 1 pending ability, got %d", len(abilities))
		}

		ability := abilities[0]
		if ability.AbilityType != "unveil" {
			t.Errorf("Expected ability type 'unveil', got %s", ability.AbilityType)
		}

		if ability.CardID != 31 {
			t.Errorf("Expected card ID 31, got %d", ability.CardID)
		}

		availableCardIDs, ok := ability.Data["available_cards"].([]int)
		if !ok {
			t.Fatal("Expected available_cards in ability data")
		}

		if len(availableCardIDs) == 0 {
			t.Error("Expected at least one available card")
		}
	})

	t.Run("Player without Wearer card cannot trigger", func(t *testing.T) {
		// Create player without The Wearer of Masks
		player2 := game.NewPlayer("player2", "Bob", "session2")
		player2.Role = &game.Card{
			ID:    15,
			Name:  "The Bodyguard",
			Type:  "Creature - Guardian",
			Types: game.CardTypes{Subtype: "Guardian"},
		}

		freshRoom, _ := memStore.GetRoom(room.Code)
		freshRoom.AddPlayer(player2)
		memStore.UpdateRoom(freshRoom)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player2/trigger-wearer", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Player not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/nonexistent/trigger-wearer", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/INVALID/player/player1/trigger-wearer", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		rctx.URLParams.Add("playerID", "player1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// TestSelectWearerCard tests selecting a card for transformation
func TestSelectWearerCard(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with CardPool
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")

	// Give player1 The Wearer of Masks
	wearerCard := &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	player1.Role = wearerCard

	room.AddPlayer(player1)

	// Initialize CardPool with some cards
	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
		{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
	}
	room.CardPool = game.NewCardPool(availableCards)

	// Trigger ability first to create pending ability
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.TriggerWearerAbility(w, req)

	// Get the ability ID
	updatedRoom, _ := memStore.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer("player1")
	abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

	t.Run("Select card successfully", func(t *testing.T) {
		// Select card ID 15 (The Bodyguard)
		reqBody := map[string]interface{}{
			"card_id": 15,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify transformation
		finalRoom, _ := memStore.GetRoom(room.Code)
		finalPlayer := finalRoom.GetPlayer("player1")

		if finalPlayer.Role.GetID() != 15 {
			t.Errorf("Expected player role to be card 15, got %d", finalPlayer.Role.GetID())
		}

		if !finalPlayer.AbilityState.IsTransformed() {
			t.Error("Expected player to be transformed")
		}

		if finalPlayer.AbilityState.GetOriginalCardID() != 31 {
			t.Errorf("Expected original card ID 31, got %d", finalPlayer.AbilityState.GetOriginalCardID())
		}

		if finalPlayer.AbilityState.GetTransformedCardID() != 15 {
			t.Errorf("Expected transformed card ID 15, got %d", finalPlayer.AbilityState.GetTransformedCardID())
		}

		// Verify pending ability was resolved
		if finalPlayer.AbilityState.HasPendingAbilities() {
			t.Error("Expected pending ability to be resolved")
		}
	})

	t.Run("Cannot select card not in available list", func(t *testing.T) {
		// Create new room and player
		room2, _ := memStore.CreateRoom()
		player2 := game.NewPlayer("player2", "Bob", "session2")
		player2.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		room2.AddPlayer(player2)
		// Use fresh availableCards copy
		availableCards2 := []*game.Card{
			{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
			{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
			{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
			{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
		}
		room2.CardPool = game.NewCardPool(availableCards2)
		memStore.UpdateRoom(room2)

		// Trigger ability
		req := httptest.NewRequest("POST", "/room/"+room2.Code+"/player/player2/trigger-wearer", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room2.Code)
		rctx.URLParams.Add("playerID", "player2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler.TriggerWearerAbility(w, req)

		// Get ability ID
		updatedRoom2, _ := memStore.GetRoom(room2.Code)
		updatedPlayer2 := updatedRoom2.GetPlayer("player2")
		abilityID2 := updatedPlayer2.AbilityState.PendingAbilities[0].ID

		// Try to select card not in available list (e.g., card 999)
		reqBody := map[string]interface{}{
			"card_id": 999,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req2 := httptest.NewRequest("POST", "/room/"+room2.Code+"/ability/"+abilityID2+"/select-card", bytes.NewReader(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.AddCookie(&http.Cookie{
			Name:  "player_" + room2.Code,
			Value: "player2",
		})

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("code", room2.Code)
		rctx2.URLParams.Add("abilityID", abilityID2)
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

		w2 := httptest.NewRecorder()

		handler.SelectWearerCard(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w2.Code)
		}
	})

	t.Run("Ability not found", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"card_id": 15,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/non-existent/select-card", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "non-existent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("No player cookie", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"card_id": 15,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

// TestWearerAbilityEventPublishing tests that events are published correctly
func TestWearerAbilityEventPublishing(t *testing.T) {
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
	player1.Role = &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	room.AddPlayer(player1)

	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
	}
	room.CardPool = game.NewCardPool(availableCards)
	memStore.UpdateRoom(room)

	// Subscribe to events
	eventChan := eventBus.Subscribe(room.Code)

	// Trigger ability
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.TriggerWearerAbility(w, req)

	// Wait for ability_triggered event
	select {
	case event := <-eventChan:
		if event.Type != "ability_triggered" {
			t.Errorf("Expected event type 'ability_triggered', got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for ability_triggered event")
	}

	// Now select a card
	updatedRoom, _ := memStore.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer("player1")
	abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

	reqBody := map[string]interface{}{
		"card_id": 15,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req2 := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})

	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("code", room.Code)
	rctx2.URLParams.Add("abilityID", abilityID)
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

	w2 := httptest.NewRecorder()

	handler.SelectWearerCard(w2, req2)

	// Wait for transformation_complete event
	select {
	case event := <-eventChan:
		if event.Type != "transformation_complete" {
			t.Errorf("Expected event type 'transformation_complete', got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for transformation_complete event")
	}
}
