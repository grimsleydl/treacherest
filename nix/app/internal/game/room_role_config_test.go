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
		State:      StateWaiting,
		RoleConfig: &RoleConfiguration{
			PresetName: "custom",
			EnabledRoles: map[string]bool{
				"leader":   true,
				"guardian": true,
				"traitor":  true,
			},
			RoleCounts: map[string]int{
				"leader":   1,
				"guardian": 2,
				"traitor":  1,
			},
			MinPlayers: 4,
			MaxPlayers: 4,
		},
	}

	// Test that room has role configuration
	if room.RoleConfig == nil {
		t.Fatal("Room should have role configuration")
	}

	// Test role counts
	expectedTotal := 4
	actualTotal := 0
	for _, count := range room.RoleConfig.RoleCounts {
		actualTotal += count
	}
	if actualTotal != expectedTotal {
		t.Errorf("Expected total roles %d, got %d", expectedTotal, actualTotal)
	}
}

func TestAssignRolesWithConfig(t *testing.T) {
	// Create test configuration
	cfg := &config.ServerConfig{
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
			Presets: map[string]map[string]config.RoleDistribution{
				"basic-3p": {
					"3": {
						Roles: map[string]int{
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
		cardsByRole: map[RoleType][]*Card{
			RoleLeader: {
				{Name: "Test Leader", RoleType: "Leader"},
			},
			RoleGuardian: {
				{Name: "Test Guardian 1", RoleType: "Guardian"},
				{Name: "Test Guardian 2", RoleType: "Guardian"},
			},
			RoleTraitor: {
				{Name: "Test Traitor", RoleType: "Traitor"},
			},
		},
	}

	// Create role config service
	roleService := NewRoleConfigService(cfg)

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
		cardsByRole: map[RoleType][]*Card{
			RoleLeader: {
				{Name: "Test Leader", RoleType: "Leader"},
			},
			RoleGuardian: {
				{Name: "Test Guardian", RoleType: "Guardian"},
			},
		},
	}

	// Create role configuration
	roleConfig := &RoleConfiguration{
		PresetName: "custom",
		EnabledRoles: map[string]bool{
			"leader":   true,
			"guardian": true,
		},
		RoleCounts: map[string]int{
			"leader":   1,
			"guardian": 1,
		},
		MinPlayers: 2,
		MaxPlayers: 2,
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