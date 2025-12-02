package game

import (
	"testing"
	"treacherest/internal/game/ability"
)

// mockGameState implements ability.GameStateProvider for testing
type mockGameState struct {
	availableCards []ability.CardLike
	roleOptions    map[int]map[string]interface{}
	cardsByID      map[int]ability.CardLike
}

func (m *mockGameState) GetAvailableCards(filters []ability.Filter) []ability.CardLike {
	// Simple implementation: return all available cards
	// In real implementation, would apply filters
	return m.availableCards
}

func (m *mockGameState) GetRoleOption(cardID int, key string) (interface{}, bool) {
	if opts, ok := m.roleOptions[cardID]; ok {
		if val, ok := opts[key]; ok {
			return val, true
		}
	}
	return nil, false
}

func (m *mockGameState) GetCardByID(cardID int) ability.CardLike {
	return m.cardsByID[cardID]
}

// mockCard implements ability.CardLike for testing
type mockCard struct {
	id   int
	text string
	name string
}

func (m *mockCard) GetID() int {
	return m.id
}

func (m *mockCard) GetText() string {
	return m.text
}

// TestWearerOfMasksResolver_OnTrigger tests the OnTrigger method
func TestWearerOfMasksResolver_OnTrigger(t *testing.T) {
	resolver := NewWearerOfMasksResolver()

	t.Run("Basic trigger with X=3", func(t *testing.T) {
		// Setup mock cards
		availableCards := []ability.CardLike{
			&mockCard{id: 10, text: "Guardian card", name: "Test Guardian"},
			&mockCard{id: 11, text: "Assassin card", name: "Test Assassin"},
			&mockCard{id: 12, text: "Traitor card", name: "Test Traitor"},
			&mockCard{id: 13, text: "Another Guardian", name: "Guardian 2"},
		}

		mockState := &mockGameState{
			availableCards: availableCards,
			roleOptions:    make(map[int]map[string]interface{}),
		}

		ctx := &ability.AbilityContext{
			RoomCode:  "TEST1",
			PlayerID:  "player1",
			CardID:    31, // Wearer of Masks
			TempData:  map[string]interface{}{"X": 3},
			GameState: mockState,
		}

		pending, err := resolver.OnTrigger(ctx)
		if err != nil {
			t.Fatalf("OnTrigger failed: %v", err)
		}

		if pending == nil {
			t.Fatal("Expected pending ability, got nil")
		}

		if pending.AbilityType != "wearer_transform" {
			t.Errorf("Expected ability type 'wearer_transform', got %s", pending.AbilityType)
		}

		if pending.PlayerID != "player1" {
			t.Errorf("Expected player ID 'player1', got %s", pending.PlayerID)
		}

		// Check revealed cards in data
		revealedCards, ok := pending.Data["revealed_cards"]
		if !ok {
			t.Fatal("Expected revealed_cards in pending ability data")
		}

		revealedSlice, ok := revealedCards.([]map[string]interface{})
		if !ok {
			t.Fatal("revealed_cards should be []map[string]interface{}")
		}

		// Should reveal up to 3 cards (X=3)
		if len(revealedSlice) > 3 {
			t.Errorf("Expected at most 3 revealed cards, got %d", len(revealedSlice))
		}
	})

	t.Run("No available cards", func(t *testing.T) {
		mockState := &mockGameState{
			availableCards: []ability.CardLike{}, // No cards available
			roleOptions:    make(map[int]map[string]interface{}),
		}

		ctx := &ability.AbilityContext{
			RoomCode:  "TEST1",
			PlayerID:  "player1",
			CardID:    31,
			TempData:  map[string]interface{}{"X": 5},
			GameState: mockState,
		}

		_, err := resolver.OnTrigger(ctx)
		if err == nil {
			t.Error("Expected error when no cards available")
		}
	})

	t.Run("Respects max_reveal option", func(t *testing.T) {
		availableCards := []ability.CardLike{
			&mockCard{id: 10, text: "Card 1"},
			&mockCard{id: 11, text: "Card 2"},
			&mockCard{id: 12, text: "Card 3"},
			&mockCard{id: 13, text: "Card 4"},
			&mockCard{id: 14, text: "Card 5"},
		}

		mockState := &mockGameState{
			availableCards: availableCards,
			roleOptions: map[int]map[string]interface{}{
				31: { // Wearer of Masks
					"max_reveal": 2, // Limit to 2 cards
				},
			},
		}

		ctx := &ability.AbilityContext{
			RoomCode:  "TEST1",
			PlayerID:  "player1",
			CardID:    31,
			TempData:  map[string]interface{}{"X": 5}, // X=5, but max_reveal=2
			GameState: mockState,
		}

		pending, err := resolver.OnTrigger(ctx)
		if err != nil {
			t.Fatalf("OnTrigger failed: %v", err)
		}

		revealedCards := pending.Data["revealed_cards"].([]map[string]interface{})
		if len(revealedCards) != 2 {
			t.Errorf("Expected exactly 2 revealed cards (max_reveal), got %d", len(revealedCards))
		}
	})

	t.Run("No game state", func(t *testing.T) {
		ctx := &ability.AbilityContext{
			RoomCode:  "TEST1",
			PlayerID:  "player1",
			CardID:    31,
			TempData:  map[string]interface{}{"X": 3},
			GameState: nil, // No game state
		}

		_, err := resolver.OnTrigger(ctx)
		if err == nil {
			t.Error("Expected error when game state is nil")
		}
	})
}

// TestWearerOfMasksResolver_OnChoice tests the OnChoice method
func TestWearerOfMasksResolver_OnChoice(t *testing.T) {
	resolver := NewWearerOfMasksResolver()

	t.Run("Valid choice as int", func(t *testing.T) {
		ctx := &ability.AbilityContext{
			RoomCode: "TEST1",
			PlayerID: "player1",
			CardID:   31,
			TempData: make(map[string]interface{}),
		}

		err := resolver.OnChoice(ctx, 15) // Choose card ID 15
		if err != nil {
			t.Errorf("OnChoice failed: %v", err)
		}

		// Check that transformation data is stored
		if transformID, ok := ctx.TempData["transform_to_card_id"]; !ok || transformID != 15 {
			t.Errorf("Expected transform_to_card_id=15, got %v", transformID)
		}

		if keepTypes, ok := ctx.TempData["keep_types"]; !ok {
			t.Error("Expected keep_types in TempData")
		} else {
			types, ok := keepTypes.([]string)
			if !ok || len(types) != 1 || types[0] != "Traitor" {
				t.Errorf("Expected keep_types=[Traitor], got %v", keepTypes)
			}
		}

		if endCond, ok := ctx.TempData["end_condition"]; !ok || endCond != "face_down" {
			t.Errorf("Expected end_condition=face_down, got %v", endCond)
		}
	})

	t.Run("Valid choice as float64 (JSON unmarshal)", func(t *testing.T) {
		ctx := &ability.AbilityContext{
			RoomCode: "TEST1",
			PlayerID: "player1",
			CardID:   31,
			TempData: make(map[string]interface{}),
		}

		err := resolver.OnChoice(ctx, float64(20))
		if err != nil {
			t.Errorf("OnChoice failed: %v", err)
		}

		if transformID, ok := ctx.TempData["transform_to_card_id"]; !ok || transformID != 20 {
			t.Errorf("Expected transform_to_card_id=20, got %v", transformID)
		}
	})

	t.Run("Valid choice as map", func(t *testing.T) {
		ctx := &ability.AbilityContext{
			RoomCode: "TEST1",
			PlayerID: "player1",
			CardID:   31,
			TempData: make(map[string]interface{}),
		}

		choice := map[string]interface{}{
			"card_id": 25,
		}

		err := resolver.OnChoice(ctx, choice)
		if err != nil {
			t.Errorf("OnChoice failed: %v", err)
		}

		if transformID, ok := ctx.TempData["transform_to_card_id"]; !ok || transformID != 25 {
			t.Errorf("Expected transform_to_card_id=25, got %v", transformID)
		}
	})

	t.Run("Invalid choice - no card ID", func(t *testing.T) {
		ctx := &ability.AbilityContext{
			RoomCode: "TEST1",
			PlayerID: "player1",
			CardID:   31,
			TempData: make(map[string]interface{}),
		}

		err := resolver.OnChoice(ctx, 0) // Invalid: card ID 0
		if err == nil {
			t.Error("Expected error for card ID 0")
		}
	})

	t.Run("Invalid choice - wrong type", func(t *testing.T) {
		ctx := &ability.AbilityContext{
			RoomCode: "TEST1",
			PlayerID: "player1",
			CardID:   31,
			TempData: make(map[string]interface{}),
		}

		err := resolver.OnChoice(ctx, "not a card ID")
		if err == nil {
			t.Error("Expected error for invalid choice type")
		}
	})
}

// TestWearerOfMasksResolver_CanActivate tests the CanActivate method
func TestWearerOfMasksResolver_CanActivate(t *testing.T) {
	resolver := NewWearerOfMasksResolver()

	ctx := &ability.AbilityContext{
		RoomCode: "TEST1",
		PlayerID: "player1",
		CardID:   31,
		TempData: make(map[string]interface{}),
	}

	// This is a triggered ability, should always return true
	if !resolver.CanActivate(ctx) {
		t.Error("Expected CanActivate to return true")
	}
}

// TestWearerOfMasksResolver_ApplyEffect tests the ApplyEffect method
func TestWearerOfMasksResolver_ApplyEffect(t *testing.T) {
	resolver := NewWearerOfMasksResolver()

	ctx := &ability.AbilityContext{
		RoomCode: "TEST1",
		PlayerID: "player1",
		CardID:   31,
		TempData: make(map[string]interface{}),
	}

	// ApplyEffect is handled by game logic, should not error
	err := resolver.ApplyEffect(ctx)
	if err != nil {
		t.Errorf("ApplyEffect failed: %v", err)
	}
}

// TestWearerOfMasksResolver_RemoveEffect tests the RemoveEffect method
func TestWearerOfMasksResolver_RemoveEffect(t *testing.T) {
	resolver := NewWearerOfMasksResolver()

	ctx := &ability.AbilityContext{
		RoomCode: "TEST1",
		PlayerID: "player1",
		CardID:   31,
		TempData: make(map[string]interface{}),
	}

	// RemoveEffect is handled by game logic, should not error
	err := resolver.RemoveEffect(ctx)
	if err != nil {
		t.Errorf("RemoveEffect failed: %v", err)
	}
}
