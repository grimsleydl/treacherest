package handlers

import (
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"
)

// newTestHandler creates a handler with default test configuration
func newTestHandler() *Handler {
	cfg := config.DefaultConfig()
	s := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	s.SetCardService(cardService)
	return New(s, cardService, cfg)
}

// newTestHandlerWithStore creates a handler with a specific store
func newTestHandlerWithStore(s *store.MemoryStore) *Handler {
	cfg := config.DefaultConfig()
	cardService := createMockCardService()
	s.SetCardService(cardService)
	return New(s, cardService, cfg)
}

// createMockCardService creates a CardService with minimal data for testing
func createMockCardService() *game.CardService {
	return &game.CardService{
		Leaders: []*game.Card{
			{ID: 1, Name: "Test Leader", Types: game.CardTypes{Subtype: "Leader"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 5, Name: "Test Leader 2", Types: game.CardTypes{Subtype: "Leader"}, Base64Image: "data:image/jpeg;base64,test"},
		},
		Guardians: []*game.Card{
			{ID: 2, Name: "Test Guardian", Types: game.CardTypes{Subtype: "Guardian"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 6, Name: "Test Guardian 2", Types: game.CardTypes{Subtype: "Guardian"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 7, Name: "Test Guardian 3", Types: game.CardTypes{Subtype: "Guardian"}, Base64Image: "data:image/jpeg;base64,test"},
		},
		Assassins: []*game.Card{
			{ID: 3, Name: "Test Assassin", Types: game.CardTypes{Subtype: "Assassin"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 8, Name: "Test Assassin 2", Types: game.CardTypes{Subtype: "Assassin"}, Base64Image: "data:image/jpeg;base64,test"},
		},
		Traitors: []*game.Card{
			{ID: 4, Name: "Test Traitor", Types: game.CardTypes{Subtype: "Traitor"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 9, Name: "Test Traitor 2", Types: game.CardTypes{Subtype: "Traitor"}, Base64Image: "data:image/jpeg;base64,test"},
		},
	}
}

// mockLeaderCard returns a mock leader card for testing
func mockLeaderCard() *game.Card {
	return &game.Card{
		ID:          1,
		Name:        "Test Leader",
		Types:       game.CardTypes{Subtype: "Leader"},
		Base64Image: "data:image/jpeg;base64,test",
	}
}

// mockGuardianCard returns a mock guardian card for testing
func mockGuardianCard() *game.Card {
	return &game.Card{
		ID:          2,
		Name:        "Test Guardian",
		Types:       game.CardTypes{Subtype: "Guardian"},
		Base64Image: "data:image/jpeg;base64,test",
	}
}
