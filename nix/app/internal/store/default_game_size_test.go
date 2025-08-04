package store

import (
	"testing"
	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
)

func TestDefaultGameSize(t *testing.T) {
	// Set required environment variables
	t.Setenv("HOST", "localhost")
	t.Setenv("PORT", "8080")

	// Load the configuration
	cfg, err := config.LoadConfig("../../config/server.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify DefaultGameSize is set to 5
	if cfg.Server.DefaultGameSize != 5 {
		t.Errorf("Expected DefaultGameSize to be 5, got %d", cfg.Server.DefaultGameSize)
	}

	// Create store
	store := NewMemoryStore(cfg)

	// Create card service and set it
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create card service: %v", err)
	}
	store.SetCardService(cardService)

	// Create a new room
	room, err := store.CreateRoom()
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Verify room is created with the default game size configuration
	if room.RoleConfig == nil {
		t.Fatal("Room RoleConfig is nil")
	}

	// Check that the role config uses the default game size (5 players)
	if room.RoleConfig.MaxPlayers != 5 {
		t.Errorf("Expected RoleConfig.MaxPlayers to be 5 (defaultGameSize), got %d", room.RoleConfig.MaxPlayers)
	}

	// But the room itself should still have the max capacity
	if room.MaxPlayers != cfg.Server.MaxPlayersPerRoom {
		t.Errorf("Expected room.MaxPlayers to be %d, got %d", cfg.Server.MaxPlayersPerRoom, room.MaxPlayers)
	}

	// Verify the role distribution matches the 5-player preset
	expectedRoles := map[string]int{
		"Leader":   1,
		"Guardian": 1,
		"Assassin": 2,
		"Traitor":  1,
	}

	totalRoles := 0
	for roleName, expectedCount := range expectedRoles {
		if roleType, exists := room.RoleConfig.RoleTypes[roleName]; exists {
			if roleType.Count != expectedCount {
				t.Errorf("Expected %d %s(s), got %d", expectedCount, roleName, roleType.Count)
			}
			totalRoles += roleType.Count
		} else {
			t.Errorf("Role type %s not found in configuration", roleName)
		}
	}

	if totalRoles != 5 {
		t.Errorf("Expected total of 5 roles, got %d", totalRoles)
	}
}
