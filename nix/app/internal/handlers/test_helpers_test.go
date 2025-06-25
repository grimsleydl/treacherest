package handlers

import (
	"treacherest/internal/game"
)

// Helper functions to create mock cards for testing
func mockLeaderCard() *game.Card {
	return &game.Card{
		ID:   1,
		Name: "Test Leader",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Leader",
		},
		Text:        "Test Leader Card",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func mockGuardianCard() *game.Card {
	return &game.Card{
		ID:   2,
		Name: "Test Guardian",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Guardian",
		},
		Text:        "Test Guardian Card",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func mockAssassinCard() *game.Card {
	return &game.Card{
		ID:   3,
		Name: "Test Assassin",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Assassin",
		},
		Text:        "Test Assassin Card",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func mockTraitorCard() *game.Card {
	return &game.Card{
		ID:   4,
		Name: "Test Traitor",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Traitor",
		},
		Text:        "Test Traitor Card",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

// createMockCardService creates a CardService with minimal data for testing
func createMockCardService() *game.CardService {
	return &game.CardService{
		Leaders: []*game.Card{
			mockLeaderCard(),
		},
		Guardians: []*game.Card{
			mockGuardianCard(),
		},
		Assassins: []*game.Card{
			mockAssassinCard(),
		},
		Traitors: []*game.Card{
			mockTraitorCard(),
		},
	}
}
