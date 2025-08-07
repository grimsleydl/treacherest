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
		Text: "Test Leader Card",
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
		Text: "Test Guardian Card",
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
		Text: "Test Assassin Card",
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
		Text: "Test Traitor Card",
	}
}
