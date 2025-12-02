package game

import (
	"fmt"
	"math/rand"
	"sync"
)

// CardPool manages the pool of all cards and tracks which are assigned
type CardPool struct {
	AllCards      []*Card
	AssignedCards map[int]*Card // Card ID -> Card
	mu            sync.RWMutex  // Thread-safe access
}

// NewCardPool creates a new card pool with all available cards
func NewCardPool(allCards []*Card) *CardPool {
	return &CardPool{
		AllCards:      allCards,
		AssignedCards: make(map[int]*Card),
	}
}

// MarkCardAssigned marks a card as assigned (in use by a player)
func (cp *CardPool) MarkCardAssigned(cardID int) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Check if card exists
	card := cp.findCardByID(cardID)
	if card == nil {
		return fmt.Errorf("card with ID %d not found", cardID)
	}

	// Check if already assigned
	if _, exists := cp.AssignedCards[cardID]; exists {
		return fmt.Errorf("card with ID %d is already assigned", cardID)
	}

	cp.AssignedCards[cardID] = card
	return nil
}

// UnmarkCardAssigned removes a card from the assigned list
func (cp *CardPool) UnmarkCardAssigned(cardID int) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if _, exists := cp.AssignedCards[cardID]; !exists {
		return fmt.Errorf("card with ID %d is not assigned", cardID)
	}

	delete(cp.AssignedCards, cardID)
	return nil
}

// GetAvailableCards returns cards that are not currently assigned
func (cp *CardPool) GetAvailableCards() []*Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	available := make([]*Card, 0)
	for _, card := range cp.AllCards {
		if _, assigned := cp.AssignedCards[card.ID]; !assigned {
			available = append(available, card)
		}
	}
	return available
}

// FilterByRoleType filters cards by role type
// exclude: if true, returns cards NOT matching the role type
func (cp *CardPool) FilterByRoleType(roleType RoleType, exclude bool) []*Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	filtered := make([]*Card, 0)
	for _, card := range cp.AllCards {
		matches := card.GetRoleType() == roleType
		if (matches && !exclude) || (!matches && exclude) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// FilterAvailableByRoleType filters available (unassigned) cards by role type
// exclude: if true, returns cards NOT matching the role type
func (cp *CardPool) FilterAvailableByRoleType(roleType RoleType, exclude bool) []*Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	filtered := make([]*Card, 0)
	for _, card := range cp.AllCards {
		// Skip if assigned
		if _, assigned := cp.AssignedCards[card.ID]; assigned {
			continue
		}

		matches := card.GetRoleType() == roleType
		if (matches && !exclude) || (!matches && exclude) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// GetRandomCards returns random cards from the full pool (up to count)
func (cp *CardPool) GetRandomCards(count int) []*Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if count <= 0 {
		return []*Card{}
	}

	if count >= len(cp.AllCards) {
		// Return all cards in random order
		result := make([]*Card, len(cp.AllCards))
		copy(result, cp.AllCards)
		rand.Shuffle(len(result), func(i, j int) {
			result[i], result[j] = result[j], result[i]
		})
		return result
	}

	// Shuffle indices and take first 'count'
	indices := rand.Perm(len(cp.AllCards))
	result := make([]*Card, count)
	for i := 0; i < count; i++ {
		result[i] = cp.AllCards[indices[i]]
	}
	return result
}

// GetRandomAvailableCards returns random cards from available (unassigned) cards
func (cp *CardPool) GetRandomAvailableCards(count int) []*Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	available := cp.getAvailableCardsUnsafe()

	if count <= 0 {
		return []*Card{}
	}

	if count >= len(available) {
		// Return all available in random order
		result := make([]*Card, len(available))
		copy(result, available)
		rand.Shuffle(len(result), func(i, j int) {
			result[i], result[j] = result[j], result[i]
		})
		return result
	}

	// Shuffle indices and take first 'count'
	indices := rand.Perm(len(available))
	result := make([]*Card, count)
	for i := 0; i < count; i++ {
		result[i] = available[indices[i]]
	}
	return result
}

// GetCardByID retrieves a card by its ID
func (cp *CardPool) GetCardByID(cardID int) *Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return cp.findCardByID(cardID)
}

// IsCardAssigned checks if a card is currently assigned
func (cp *CardPool) IsCardAssigned(cardID int) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	_, assigned := cp.AssignedCards[cardID]
	return assigned
}

// ResetAssignments clears all card assignments
func (cp *CardPool) ResetAssignments() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.AssignedCards = make(map[int]*Card)
}

// findCardByID is an internal helper (must be called with lock held)
func (cp *CardPool) findCardByID(cardID int) *Card {
	for _, card := range cp.AllCards {
		if card.ID == cardID {
			return card
		}
	}
	return nil
}

// getAvailableCardsUnsafe is an internal helper (must be called with lock held)
func (cp *CardPool) getAvailableCardsUnsafe() []*Card {
	available := make([]*Card, 0)
	for _, card := range cp.AllCards {
		if _, assigned := cp.AssignedCards[card.ID]; !assigned {
			available = append(available, card)
		}
	}
	return available
}

// GetCardsByTypes returns available cards matching any of the specified types
// If no types are specified, returns all available cards
func (cp *CardPool) GetCardsByTypes(types ...string) []*Card {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	// If no types specified, return all available cards
	if len(types) == 0 {
		return cp.getAvailableCardsUnsafe()
	}

	// Create a map for faster type lookup
	typeMap := make(map[string]bool)
	for _, t := range types {
		typeMap[t] = true
	}

	filtered := make([]*Card, 0)
	for _, card := range cp.AllCards {
		// Skip if assigned
		if _, assigned := cp.AssignedCards[card.ID]; assigned {
			continue
		}

		// Check if card type matches any of the specified types
		cardType := string(card.GetRoleType())
		if typeMap[cardType] {
			filtered = append(filtered, card)
		}
	}
	return filtered
}
