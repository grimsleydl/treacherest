package game

import (
	"testing"
	"treacherest/internal/config"
)

func TestRoom_RoleConfiguration(t *testing.T) {
	// Create a test room
	room := &Room{
		Code:       "TEST1",
		MaxPlayers: 8,
		Players:    make(map[string]*Player),
		State:      StateLobby,
		RoleConfig: &RoleConfiguration{
			PresetName: "custom",
			MinPlayers: 4,
			MaxPlayers: 4,
			RoleTypes: map[string]*RoleTypeConfig{
				"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Usurper": true}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
			},
		},
	}

	// Test that room has role configuration
	if room.RoleConfig == nil {
		t.Fatal("Room should have role configuration")
	}

	// Test role counts
	expectedTotal := 4
	actualTotal := 0
	for _, typeConfig := range room.RoleConfig.RoleTypes {
		actualTotal += typeConfig.Count
	}
	if actualTotal != expectedTotal {
		t.Errorf("Expected total roles %d, got %d", expectedTotal, actualTotal)
	}
}

func TestAssignRolesWithConfig(t *testing.T) {
	// Create test configuration
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName: "Leader",
					Category:    "Leader",
					MinCount:    1,
					MaxCount:    1,
				},
				"guardian": {
					DisplayName: "Guardian",
					Category:    "Good",
					MinCount:    0,
					MaxCount:    10,
				},
				"traitor": {
					DisplayName: "Traitor",
					Category:    "Evil",
					MinCount:    0,
					MaxCount:    10,
				},
			},
			Presets: map[string]config.Preset{
				"basic-3p": {
					Name:        "Basic 3 Player",
					Description: "Basic 3 player game",
					Distributions: map[int]map[string]int{
						3: {
							"leader":   1,
							"guardian": 1,
							"traitor":  1,
						},
					},
				},
			},
		},
	}

	// Create card service with test cards
	cardService := &CardService{
		Leaders: []*Card{
			{ID: 1, Name: "Test Leader", Types: CardTypes{Subtype: "Leader"}},
		},
		Guardians: []*Card{
			{ID: 2, Name: "Test Guardian 1", Types: CardTypes{Subtype: "Guardian"}},
			{ID: 3, Name: "Test Guardian 2", Types: CardTypes{Subtype: "Guardian"}},
		},
		Traitors: []*Card{
			{ID: 4, Name: "Test Traitor", Types: CardTypes{Subtype: "Traitor"}},
		},
		Assassins: []*Card{
			{ID: 5, Name: "Test Assassin", Types: CardTypes{Subtype: "Assassin"}},
		},
	}

	// Create role config service
	roleService := NewRoleConfigService(cfg)
	roleService.SetCardService(cardService)

	// Create role configuration from preset
	roleConfig, err := roleService.CreateFromPreset("basic-3p", 10)
	if err != nil {
		t.Fatalf("Failed to create role config: %v", err)
	}

	// Create test players
	players := []*Player{
		{ID: "p1", Name: "Player 1", IsHost: false},
		{ID: "p2", Name: "Player 2", IsHost: false},
		{ID: "p3", Name: "Player 3", IsHost: false},
	}

	// Assign roles
	AssignRolesWithConfig(players, cardService, roleConfig, roleService)

	// Verify all players have roles
	for _, player := range players {
		if player.Role == nil {
			t.Errorf("Player %s has no role assigned", player.Name)
		}
	}

	// Count assigned roles
	roleCounts := make(map[RoleType]int)
	for _, player := range players {
		if player.Role != nil {
			roleType := player.Role.GetRoleType()
			roleCounts[roleType]++
		}
	}

	// Verify role distribution matches preset
	expectedRoles := map[RoleType]int{
		RoleLeader:   1,
		RoleGuardian: 1,
		RoleTraitor:  1,
	}

	for roleType, expectedCount := range expectedRoles {
		if roleCounts[roleType] != expectedCount {
			t.Errorf("Expected %d %s, got %d", expectedCount, roleType, roleCounts[roleType])
		}
	}

	// Verify leader is revealed
	leaderFound := false
	for _, player := range players {
		if player.Role != nil && player.Role.GetRoleType() == RoleLeader {
			if !player.RoleRevealed {
				t.Error("Leader role should be revealed")
			}
			leaderFound = true
		}
	}
	if !leaderFound {
		t.Error("No leader found in players")
	}
}

func TestAssignRolesWithConfig_HostExclusion(t *testing.T) {
	// Create minimal card service
	cardService := &CardService{
		Leaders: []*Card{
			{ID: 1, Name: "The Usurper", Types: CardTypes{Subtype: "Leader"}},
		},
		Guardians: []*Card{
			{ID: 2, Name: "The Bodyguard", Types: CardTypes{Subtype: "Guardian"}},
		},
	}

	// Create role configuration
	roleConfig := &RoleConfiguration{
		PresetName: "custom",
		MinPlayers: 2,
		MaxPlayers: 2,
		RoleTypes: map[string]*RoleTypeConfig{
			"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Usurper": true}},
			"Guardian": {Count: 1, EnabledCards: map[string]bool{"The Bodyguard": true}},
		},
	}

	// Create test players including a host
	players := []*Player{
		{ID: "host", Name: "Host", IsHost: true},
		{ID: "p1", Name: "Player 1", IsHost: false},
		{ID: "p2", Name: "Player 2", IsHost: false},
	}

	// Assign roles
	AssignRolesWithConfig(players, cardService, roleConfig, nil)

	// Verify host has no role
	for _, player := range players {
		if player.IsHost && player.Role != nil {
			t.Errorf("Host player %s should not have a role", player.Name)
		}
		if !player.IsHost && player.Role == nil {
			t.Errorf("Non-host player %s should have a role", player.Name)
		}
	}
}
