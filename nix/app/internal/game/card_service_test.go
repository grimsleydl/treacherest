package game

import (
	"strings"
	"testing"
	"treacherest"
)

func TestNewCardService(t *testing.T) {
	service, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test that cards were loaded and categorized
	tests := []struct {
		name  string
		cards []*Card
	}{
		{"Leaders", service.Leaders},
		{"Guardians", service.Guardians},
		{"Assassins", service.Assassins},
		{"Traitors", service.Traitors},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.cards) == 0 {
				t.Errorf("No %s cards were loaded", tt.name)
			}
			// Log the count for visibility
			t.Logf("Loaded %d %s cards", len(tt.cards), tt.name)
		})
	}

	// Verify total card count matches allCards
	totalCards := len(service.Leaders) + len(service.Guardians) + len(service.Assassins) + len(service.Traitors)
	if totalCards != len(service.allCards) {
		t.Errorf("Total categorized cards (%d) doesn't match allCards count (%d)", totalCards, len(service.allCards))
	}

	// Verify we have a reasonable distribution for gameplay
	if len(service.Leaders) < 1 {
		t.Error("Need at least 1 Leader card for gameplay")
	}
	if len(service.Guardians) < 3 {
		t.Error("Need at least 3 Guardian cards for max player games")
	}
	if len(service.Assassins) < 2 {
		t.Error("Need at least 2 Assassin cards for max player games")
	}
	if len(service.Traitors) < 2 {
		t.Error("Need at least 2 Traitor cards for max player games")
	}
}

func TestCardService_GetRandomCards(t *testing.T) {
	service, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	tests := []struct {
		name     string
		cardType RoleType
		count    int
		poolSize int
	}{
		{"Get 3 Leaders", RoleLeader, 3, len(service.Leaders)},
		{"Get 5 Guardians", RoleGuardian, 5, len(service.Guardians)},
		{"Get 2 Assassins", RoleAssassin, 2, len(service.Assassins)},
		{"Get 1 Traitor", RoleTraitor, 1, len(service.Traitors)},
		{"Request more than available", RoleTraitor, 20, len(service.Traitors)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := service.GetRandomCards(tt.cardType, tt.count)

			expectedCount := tt.count
			if tt.count > tt.poolSize {
				expectedCount = tt.poolSize
			}

			if len(cards) != expectedCount {
				t.Errorf("Expected %d cards, got %d", expectedCount, len(cards))
			}

			// Check for duplicates
			seen := make(map[int]bool)
			for _, card := range cards {
				if seen[card.ID] {
					t.Errorf("Duplicate card found: %s (ID: %d)", card.Name, card.ID)
				}
				seen[card.ID] = true
			}

			// Verify all cards are of the correct type
			for _, card := range cards {
				if card.GetRoleType() != tt.cardType {
					t.Errorf("Expected card type %s, got %s", tt.cardType, card.GetRoleType())
				}
			}
		})
	}
}

func TestCardService_GetRandomSingleCards(t *testing.T) {
	service, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test individual random card getters
	if card := service.GetRandomLeader(); card == nil || card.GetRoleType() != RoleLeader {
		t.Error("GetRandomLeader failed")
	}

	if card := service.GetRandomGuardian(); card == nil || card.GetRoleType() != RoleGuardian {
		t.Error("GetRandomGuardian failed")
	}

	if card := service.GetRandomAssassin(); card == nil || card.GetRoleType() != RoleAssassin {
		t.Error("GetRandomAssassin failed")
	}

	if card := service.GetRandomTraitor(); card == nil || card.GetRoleType() != RoleTraitor {
		t.Error("GetRandomTraitor failed")
	}
}

func TestCard_GetRoleType(t *testing.T) {
	tests := []struct {
		name     string
		subtype  string
		expected RoleType
	}{
		{"Leader card", "Leader", RoleLeader},
		{"Guardian card", "Guardian", RoleGuardian},
		{"Assassin card", "Assassin", RoleAssassin},
		{"Traitor card", "Traitor", RoleTraitor},
		{"Unknown card", "Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{
				Types: CardTypes{Subtype: tt.subtype},
			}
			if got := card.GetRoleType(); got != tt.expected {
				t.Errorf("GetRoleType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCard_GetWinCondition(t *testing.T) {
	tests := []struct {
		name     string
		subtype  string
		expected string
	}{
		{"Leader", "Leader", "The Leader, and their Guardians, win if they are the last players standing."},
		{"Guardian", "Guardian", "The Guardians help the Leader, they win or lose with them."},
		{"Assassin", "Assassin", "The Assassins win if the Leader is eliminated."},
		{"Traitor", "Traitor", "The Traitor wins if they are the last player standing."},
		{"Unknown", "Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{
				Types: CardTypes{Subtype: tt.subtype},
			}
			if got := card.GetWinCondition(); got != tt.expected {
				t.Errorf("GetWinCondition() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCardService_Base64Images(t *testing.T) {
	service, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test that all cards have base64 images
	allCards := append([]*Card{}, service.Leaders...)
	allCards = append(allCards, service.Guardians...)
	allCards = append(allCards, service.Assassins...)
	allCards = append(allCards, service.Traitors...)

	for _, card := range allCards {
		t.Run(card.Name, func(t *testing.T) {
			base64Image := card.GetImageBase64()

			// Verify the image is not empty
			if base64Image == "" {
				t.Errorf("Card %s (ID: %d) has empty base64 image", card.Name, card.ID)
			}

			// Verify it's a valid data URI
			if !strings.HasPrefix(base64Image, "data:image/") {
				t.Errorf("Card %s (ID: %d) base64 image doesn't start with 'data:image/'", card.Name, card.ID)
			}

			// Verify it contains base64 marker
			if !strings.Contains(base64Image, ";base64,") {
				t.Errorf("Card %s (ID: %d) base64 image doesn't contain ';base64,' marker", card.Name, card.ID)
			}

			// Verify MIME type is detected (should be jpeg for our cards)
			if !strings.HasPrefix(base64Image, "data:image/jpeg;base64,") {
				t.Logf("Card %s (ID: %d) has MIME type: %s", card.Name, card.ID, strings.Split(base64Image, ";")[0])
			}
		})
	}
}

func TestCardService_EmbeddedAssets(t *testing.T) {
	// This test verifies that embedded assets are loaded correctly
	service, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService with embedded assets: %v", err)
	}

	// Verify that the service loaded cards with base64 images
	if len(service.Leaders) == 0 {
		t.Error("No leader cards loaded from embedded assets")
	}

	// Check that all cards have base64 images
	allCards := append([]*Card{}, service.Leaders...)
	allCards = append(allCards, service.Guardians...)
	allCards = append(allCards, service.Assassins...)
	allCards = append(allCards, service.Traitors...)

	for _, card := range allCards {
		if card.Base64Image == "" {
			t.Errorf("Card %s (ID: %d) has no base64 image from embedded assets", card.Name, card.ID)
		}
	}

	// Since files are embedded at compile time, they are guaranteed to exist
	t.Logf("CardService successfully loaded %d cards from embedded assets", len(allCards))
}
