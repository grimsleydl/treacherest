package game

import (
	"testing"
)

// TestNewCardPool tests card pool creation
func TestNewCardPool(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1", Types: CardTypes{Subtype: "Leader"}},
		{ID: 2, Name: "Card 2", Types: CardTypes{Subtype: "Guardian"}},
		{ID: 3, Name: "Card 3", Types: CardTypes{Subtype: "Traitor"}},
	}

	pool := NewCardPool(allCards)

	if pool == nil {
		t.Fatal("Expected card pool to be created")
	}

	if len(pool.AllCards) != 3 {
		t.Errorf("Expected 3 cards in pool, got %d", len(pool.AllCards))
	}

	if len(pool.AssignedCards) != 0 {
		t.Errorf("Expected 0 assigned cards initially, got %d", len(pool.AssignedCards))
	}

	if len(pool.GetAvailableCards()) != 3 {
		t.Errorf("Expected 3 available cards, got %d", len(pool.GetAvailableCards()))
	}
}

// TestMarkCardAssigned tests marking cards as assigned
func TestMarkCardAssigned(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
		{ID: 3, Name: "Card 3"},
	}

	pool := NewCardPool(allCards)

	err := pool.MarkCardAssigned(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(pool.AssignedCards) != 1 {
		t.Errorf("Expected 1 assigned card, got %d", len(pool.AssignedCards))
	}

	if len(pool.GetAvailableCards()) != 2 {
		t.Errorf("Expected 2 available cards, got %d", len(pool.GetAvailableCards()))
	}

	// Try to assign non-existent card
	err = pool.MarkCardAssigned(999)
	if err == nil {
		t.Error("Expected error for non-existent card")
	}

	// Try to assign already assigned card
	err = pool.MarkCardAssigned(1)
	if err == nil {
		t.Error("Expected error for already assigned card")
	}
}

// TestUnmarkCardAssigned tests unmarking assigned cards
func TestUnmarkCardAssigned(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
	}

	pool := NewCardPool(allCards)
	pool.MarkCardAssigned(1)

	err := pool.UnmarkCardAssigned(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(pool.AssignedCards) != 0 {
		t.Errorf("Expected 0 assigned cards, got %d", len(pool.AssignedCards))
	}

	if len(pool.GetAvailableCards()) != 2 {
		t.Errorf("Expected 2 available cards, got %d", len(pool.GetAvailableCards()))
	}

	// Try to unmark non-assigned card
	err = pool.UnmarkCardAssigned(2)
	if err == nil {
		t.Error("Expected error for non-assigned card")
	}
}

// TestFilterByRoleType tests filtering cards by role type
func TestFilterByRoleType(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Leader 1", Types: CardTypes{Subtype: "Leader"}},
		{ID: 2, Name: "Leader 2", Types: CardTypes{Subtype: "Leader"}},
		{ID: 3, Name: "Guardian 1", Types: CardTypes{Subtype: "Guardian"}},
		{ID: 4, Name: "Traitor 1", Types: CardTypes{Subtype: "Traitor"}},
		{ID: 5, Name: "Assassin 1", Types: CardTypes{Subtype: "Assassin"}},
	}

	pool := NewCardPool(allCards)

	// Filter leaders
	leaders := pool.FilterByRoleType(RoleLeader, false)
	if len(leaders) != 2 {
		t.Errorf("Expected 2 leaders, got %d", len(leaders))
	}

	// Filter non-leaders (exclude)
	nonLeaders := pool.FilterByRoleType(RoleLeader, true)
	if len(nonLeaders) != 3 {
		t.Errorf("Expected 3 non-leaders, got %d", len(nonLeaders))
	}

	// Filter guardians
	guardians := pool.FilterByRoleType(RoleGuardian, false)
	if len(guardians) != 1 {
		t.Errorf("Expected 1 guardian, got %d", len(guardians))
	}
}

// TestFilterAvailableByRoleType tests filtering only available cards
func TestFilterAvailableByRoleType(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Leader 1", Types: CardTypes{Subtype: "Leader"}},
		{ID: 2, Name: "Leader 2", Types: CardTypes{Subtype: "Leader"}},
		{ID: 3, Name: "Guardian 1", Types: CardTypes{Subtype: "Guardian"}},
	}

	pool := NewCardPool(allCards)
	pool.MarkCardAssigned(1) // Assign Leader 1

	// Filter available leaders
	availableLeaders := pool.FilterAvailableByRoleType(RoleLeader, false)
	if len(availableLeaders) != 1 {
		t.Errorf("Expected 1 available leader, got %d", len(availableLeaders))
	}

	if availableLeaders[0].ID != 2 {
		t.Errorf("Expected Leader 2 (ID=2), got ID=%d", availableLeaders[0].ID)
	}
}

// TestGetRandomCards tests random card selection
func TestGetRandomCards(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
		{ID: 3, Name: "Card 3"},
		{ID: 4, Name: "Card 4"},
		{ID: 5, Name: "Card 5"},
	}

	pool := NewCardPool(allCards)

	// Get 3 random cards
	random := pool.GetRandomCards(3)
	if len(random) != 3 {
		t.Errorf("Expected 3 random cards, got %d", len(random))
	}

	// Verify uniqueness
	seen := make(map[int]bool)
	for _, card := range random {
		if seen[card.ID] {
			t.Errorf("Duplicate card ID %d in random selection", card.ID)
		}
		seen[card.ID] = true
	}

	// Request more cards than available
	random = pool.GetRandomCards(10)
	if len(random) != 5 {
		t.Errorf("Expected 5 cards (all available), got %d", len(random))
	}

	// Request 0 cards
	random = pool.GetRandomCards(0)
	if len(random) != 0 {
		t.Errorf("Expected 0 cards, got %d", len(random))
	}
}

// TestGetRandomAvailableCards tests random selection from available cards only
func TestGetRandomAvailableCards(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
		{ID: 3, Name: "Card 3"},
		{ID: 4, Name: "Card 4"},
	}

	pool := NewCardPool(allCards)
	pool.MarkCardAssigned(1)
	pool.MarkCardAssigned(2)

	// Get 2 random available cards (only 3 and 4 are available)
	random := pool.GetRandomAvailableCards(2)
	if len(random) != 2 {
		t.Errorf("Expected 2 random available cards, got %d", len(random))
	}

	// Verify they're from available set
	for _, card := range random {
		if card.ID == 1 || card.ID == 2 {
			t.Errorf("Got assigned card ID %d in random available selection", card.ID)
		}
	}
}

// TestGetCardByID tests card retrieval by ID
func TestGetCardByID(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
	}

	pool := NewCardPool(allCards)

	card := pool.GetCardByID(1)
	if card == nil {
		t.Fatal("Expected to find card with ID 1")
	}

	if card.Name != "Card 1" {
		t.Errorf("Expected 'Card 1', got '%s'", card.Name)
	}

	// Non-existent card
	card = pool.GetCardByID(999)
	if card != nil {
		t.Error("Expected nil for non-existent card")
	}
}

// TestIsCardAssigned tests checking if a card is assigned
func TestIsCardAssigned(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
	}

	pool := NewCardPool(allCards)

	if pool.IsCardAssigned(1) {
		t.Error("Card 1 should not be assigned initially")
	}

	pool.MarkCardAssigned(1)

	if !pool.IsCardAssigned(1) {
		t.Error("Card 1 should be assigned")
	}

	if pool.IsCardAssigned(2) {
		t.Error("Card 2 should not be assigned")
	}
}

// TestCardPoolThreadSafety tests concurrent access
func TestCardPoolThreadSafety(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
		{ID: 3, Name: "Card 3"},
		{ID: 4, Name: "Card 4"},
		{ID: 5, Name: "Card 5"},
	}

	pool := NewCardPool(allCards)

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			pool.GetAvailableCards()
			pool.FilterByRoleType(RoleGuardian, false)
			pool.GetRandomCards(2)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestResetAssignments tests resetting all assignments
func TestResetAssignments(t *testing.T) {
	allCards := []*Card{
		{ID: 1, Name: "Card 1"},
		{ID: 2, Name: "Card 2"},
		{ID: 3, Name: "Card 3"},
	}

	pool := NewCardPool(allCards)
	pool.MarkCardAssigned(1)
	pool.MarkCardAssigned(2)

	if len(pool.AssignedCards) != 2 {
		t.Errorf("Expected 2 assigned cards, got %d", len(pool.AssignedCards))
	}

	pool.ResetAssignments()

	if len(pool.AssignedCards) != 0 {
		t.Errorf("Expected 0 assigned cards after reset, got %d", len(pool.AssignedCards))
	}

	if len(pool.GetAvailableCards()) != 3 {
		t.Errorf("Expected 3 available cards after reset, got %d", len(pool.GetAvailableCards()))
	}
}
