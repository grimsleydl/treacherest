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
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/5", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player1")
		rctx.URLParams.Add("xValue", "5")
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

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player2/trigger-wearer/3", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player2")
		rctx.URLParams.Add("xValue", "3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Player not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/nonexistent/trigger-wearer/3", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "nonexistent")
		rctx.URLParams.Add("xValue", "3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/INVALID/player/player1/trigger-wearer/3", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		rctx.URLParams.Add("playerID", "player1")
		rctx.URLParams.Add("xValue", "3")
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
	leader := game.NewPlayer("leader1", "Leader", "session2")

	// Give player1 The Wearer of Masks
	wearerCard := &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	player1.Role = wearerCard

	// Give leader a Leader role (for confirmation)
	leaderCard := &game.Card{
		ID:    1,
		Name:  "Test Leader",
		Type:  "Creature - Leader",
		Types: game.CardTypes{Subtype: "Leader"},
	}
	leader.Role = leaderCard

	room.AddPlayer(player1)
	room.AddPlayer(leader)

	// Initialize CardPool with some cards
	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
		{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
	}
	room.CardPool = game.NewCardPool(availableCards)

	// Trigger ability first to create pending ability
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/5", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	rctx.URLParams.Add("xValue", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.TriggerWearerAbility(w, req)

	// Get the ability ID
	updatedRoom, _ := memStore.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer("player1")
	abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

	// Leader confirms the ability (required before card selection)
	confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/confirm", nil)
	confirmRctx := chi.NewRouteContext()
	confirmRctx.URLParams.Add("code", room.Code)
	confirmRctx.URLParams.Add("abilityID", abilityID)
	confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
	confirmReq.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "leader1",
	})
	confirmW := httptest.NewRecorder()
	handler.ConfirmAbility(confirmW, confirmReq)

	if confirmW.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for confirm, got %d", confirmW.Code)
	}

	t.Run("Select card successfully", func(t *testing.T) {
		// Select card ID 15 (The Bodyguard)
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		rctx.URLParams.Add("cardID", "15")
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
		// Add a Leader to this room for confirmation
		leader2 := game.NewPlayer("leader2", "Leader2", "session3")
		leader2.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room2.AddPlayer(player2)
		room2.AddPlayer(leader2)
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
		req := httptest.NewRequest("POST", "/room/"+room2.Code+"/player/player2/trigger-wearer/5", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room2.Code)
		rctx.URLParams.Add("playerID", "player2")
		rctx.URLParams.Add("xValue", "5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler.TriggerWearerAbility(w, req)

		// Get ability ID
		updatedRoom2, _ := memStore.GetRoom(room2.Code)
		updatedPlayer2 := updatedRoom2.GetPlayer("player2")
		abilityID2 := updatedPlayer2.AbilityState.PendingAbilities[0].ID

		// Leader confirms the ability first
		confirmReq := httptest.NewRequest("POST", "/room/"+room2.Code+"/ability/"+abilityID2+"/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room2.Code)
		confirmRctx.URLParams.Add("abilityID", abilityID2)
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room2.Code,
			Value: "leader2",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		// Try to select card not in available list (e.g., card 999)
		req2 := httptest.NewRequest("POST", "/room/"+room2.Code+"/ability/"+abilityID2+"/select-card/999", nil)
		req2.AddCookie(&http.Cookie{
			Name:  "player_" + room2.Code,
			Value: "player2",
		})

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("code", room2.Code)
		rctx2.URLParams.Add("abilityID", abilityID2)
		rctx2.URLParams.Add("cardID", "999")
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

		w2 := httptest.NewRecorder()

		handler.SelectWearerCard(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w2.Code)
		}
	})

	t.Run("Ability not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/non-existent/select-card/15", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "non-existent")
		rctx.URLParams.Add("cardID", "15")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("No player cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		rctx.URLParams.Add("cardID", "15")
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

	// Create room with player and leader
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	player1.Role = &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	leader := game.NewPlayer("leader1", "Leader", "session2")
	leader.Role = &game.Card{
		ID:    1,
		Name:  "Test Leader",
		Type:  "Creature - Leader",
		Types: game.CardTypes{Subtype: "Leader"},
	}
	room.AddPlayer(player1)
	room.AddPlayer(leader)

	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
	}
	room.CardPool = game.NewCardPool(availableCards)
	memStore.UpdateRoom(room)

	// Subscribe to events
	eventChan := eventBus.Subscribe(room.Code)

	// Trigger ability
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	rctx.URLParams.Add("xValue", "3")
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

	// Get the pending ability ID
	updatedRoom, _ := memStore.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer("player1")
	abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

	// Leader confirms the ability (required before card selection)
	confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/confirm", nil)
	confirmRctx := chi.NewRouteContext()
	confirmRctx.URLParams.Add("code", room.Code)
	confirmRctx.URLParams.Add("abilityID", abilityID)
	confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
	confirmReq.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "leader1",
	})
	confirmW := httptest.NewRecorder()
	handler.ConfirmAbility(confirmW, confirmReq)

	if confirmW.Code != http.StatusOK {
		t.Fatalf("Leader confirmation failed with status %d", confirmW.Code)
	}

	// Wait for ability_confirmed event
	select {
	case event := <-eventChan:
		if event.Type != "ability_confirmed" {
			t.Errorf("Expected event type 'ability_confirmed', got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for ability_confirmed event")
	}

	// Now select a card
	req2 := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)
	req2.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})

	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("code", room.Code)
	rctx2.URLParams.Add("abilityID", abilityID)
	rctx2.URLParams.Add("cardID", "15")
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

// TestConfirmAbility tests the ability confirmation system
func TestConfirmAbility(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	t.Run("Leader can confirm ability successfully", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		player1 := game.NewPlayer("player1", "Alice", "session1")
		player1.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		leader := game.NewPlayer("leader1", "Leader", "session2")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(player1)
		room.AddPlayer(leader)

		availableCards := []*game.Card{
			{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		}
		room.CardPool = game.NewCardPool(availableCards)
		memStore.UpdateRoom(room)

		// Trigger the ability
		triggerReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
		triggerRctx := chi.NewRouteContext()
		triggerRctx.URLParams.Add("code", room.Code)
		triggerRctx.URLParams.Add("playerID", "player1")
		triggerRctx.URLParams.Add("xValue", "3")
		triggerReq = triggerReq.WithContext(context.WithValue(triggerReq.Context(), chi.RouteCtxKey, triggerRctx))
		triggerW := httptest.NewRecorder()
		handler.TriggerWearerAbility(triggerW, triggerReq)

		// Get the ability ID
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

		// Verify ability requires confirmation and is not confirmed
		pendingAbility := updatedPlayer.AbilityState.GetPendingAbility(abilityID)
		if !pendingAbility.RequiresConfirmation {
			t.Error("Expected ability to require confirmation")
		}
		if pendingAbility.IsConfirmed() {
			t.Error("Expected ability to not be confirmed yet")
		}

		// Leader confirms the ability
		confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room.Code)
		confirmRctx.URLParams.Add("abilityID", abilityID)
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "leader1",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		if confirmW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", confirmW.Code)
		}

		// Verify ability is now confirmed
		updatedRoom, _ = memStore.GetRoom(room.Code)
		updatedPlayer = updatedRoom.GetPlayer("player1")
		pendingAbility = updatedPlayer.AbilityState.GetPendingAbility(abilityID)
		if !pendingAbility.IsConfirmed() {
			t.Error("Expected ability to be confirmed after Leader confirmation")
		}
	})

	t.Run("Non-leader cannot confirm leader-only ability", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		player1 := game.NewPlayer("player1", "Alice", "session1")
		player1.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		player2 := game.NewPlayer("player2", "Bob", "session2")
		player2.Role = &game.Card{
			ID:    15,
			Name:  "The Bodyguard",
			Type:  "Creature - Guardian",
			Types: game.CardTypes{Subtype: "Guardian"},
		}
		leader := game.NewPlayer("leader1", "Leader", "session3")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.AddPlayer(leader)

		availableCards := []*game.Card{
			{ID: 16, Name: "Another Card", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		}
		room.CardPool = game.NewCardPool(availableCards)
		memStore.UpdateRoom(room)

		// Trigger the ability
		triggerReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
		triggerRctx := chi.NewRouteContext()
		triggerRctx.URLParams.Add("code", room.Code)
		triggerRctx.URLParams.Add("playerID", "player1")
		triggerRctx.URLParams.Add("xValue", "3")
		triggerReq = triggerReq.WithContext(context.WithValue(triggerReq.Context(), chi.RouteCtxKey, triggerRctx))
		triggerW := httptest.NewRecorder()
		handler.TriggerWearerAbility(triggerW, triggerReq)

		// Get the ability ID
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

		// Non-leader player tries to confirm (should fail)
		confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room.Code)
		confirmRctx.URLParams.Add("abilityID", abilityID)
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player2",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		if confirmW.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", confirmW.Code)
		}
	})

	t.Run("Ability not found", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		leader := game.NewPlayer("leader1", "Leader", "session1")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(leader)
		memStore.UpdateRoom(room)

		confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/nonexistent/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room.Code)
		confirmRctx.URLParams.Add("abilityID", "nonexistent")
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "leader1",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		if confirmW.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", confirmW.Code)
		}
	})

	t.Run("Cannot select card without confirmation", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		player1 := game.NewPlayer("player1", "Alice", "session1")
		player1.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		leader := game.NewPlayer("leader1", "Leader", "session2")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(player1)
		room.AddPlayer(leader)

		availableCards := []*game.Card{
			{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		}
		room.CardPool = game.NewCardPool(availableCards)
		memStore.UpdateRoom(room)

		// Trigger the ability
		triggerReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
		triggerRctx := chi.NewRouteContext()
		triggerRctx.URLParams.Add("code", room.Code)
		triggerRctx.URLParams.Add("playerID", "player1")
		triggerRctx.URLParams.Add("xValue", "3")
		triggerReq = triggerReq.WithContext(context.WithValue(triggerReq.Context(), chi.RouteCtxKey, triggerRctx))
		triggerW := httptest.NewRecorder()
		handler.TriggerWearerAbility(triggerW, triggerReq)

		// Get the ability ID
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

		// Try to select card WITHOUT Leader confirmation (should fail)
		selectReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)
		selectReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})
		selectRctx := chi.NewRouteContext()
		selectRctx.URLParams.Add("code", room.Code)
		selectRctx.URLParams.Add("abilityID", abilityID)
		selectRctx.URLParams.Add("cardID", "15")
		selectReq = selectReq.WithContext(context.WithValue(selectReq.Context(), chi.RouteCtxKey, selectRctx))
		selectW := httptest.NewRecorder()
		handler.SelectWearerCard(selectW, selectReq)

		if selectW.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", selectW.Code)
		}
	})
}
