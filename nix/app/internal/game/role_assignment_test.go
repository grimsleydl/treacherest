package game

import (
	"fmt"
	"testing"
	"treacherest/internal/config"
)

func TestRoleAssignmentRespectsConfiguration(t *testing.T) {
	// Create test card service
	cardService := createMockCardService()

	// Create test config
	testConfig := &config.ServerConfig{
		Server: config.ServerSettings{
			MinPlayersPerRoom: 1,
			MaxPlayersPerRoom: 8,
		},
	}

	// Create role config service
	roleService := NewRoleConfigService(testConfig)
	roleService.SetCardService(cardService)

	tests := []struct {
		name            string
		roleConfig      *RoleConfiguration
		playerCount     int
		expectLeader    bool
		expectGuardian  bool
		expectAssassin  bool
		expectTraitor   bool
	}{
		{
			name: "Zero guardian count should not assign guardians",
			roleConfig: &RoleConfiguration{
				PresetName: "custom",
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Usurper": true}},
					"Guardian": {Count: 0, EnabledCards: map[string]bool{}}, // 0 Guardians
					"Assassin": {Count: 1, EnabledCards: map[string]bool{"The Assassin": true}},
					"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
				},
			},
			playerCount:    3,
			expectLeader:   true,
			expectGuardian: false, // Should NOT have guardians
			expectAssassin: true,
			expectTraitor:  true,
		},
		{
			name: "Leaderless game should not assign leaders",
			roleConfig: &RoleConfiguration{
				PresetName:          "custom",
				AllowLeaderlessGame: true,
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 0, EnabledCards: map[string]bool{}}, // 0 Leaders
					"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true, "The Knight": true}},
					"Assassin": {Count: 1, EnabledCards: map[string]bool{"The Assassin": true}},
					"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
				},
			},
			playerCount:    4,
			expectLeader:   false, // Should NOT have leaders
			expectGuardian: true,
			expectAssassin: true,
			expectTraitor:  true,
		},
		{
			name: "Total roles less than players should fail gracefully",
			roleConfig: &RoleConfiguration{
				PresetName: "custom",
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Usurper": true}},
					"Guardian": {Count: 1, EnabledCards: map[string]bool{"The Bodyguard": true}},
					"Assassin": {Count: 1, EnabledCards: map[string]bool{"The Assassin": true}},
					"Traitor":  {Count: 0, EnabledCards: map[string]bool{}},
				},
			},
			playerCount:    6, // 6 players but only 3 roles configured
			expectLeader:   true,
			expectGuardian: true,
			expectAssassin: true,
			expectTraitor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create players
			players := make([]*Player, tt.playerCount)
			for i := 0; i < tt.playerCount; i++ {
				players[i] = &Player{
					ID:   fmt.Sprintf("player%d", i+1),
					Name: fmt.Sprintf("Player %d", i+1),
				}
			}

			// Initialize EnabledCards if nil
			for _, typeConfig := range tt.roleConfig.RoleTypes {
				if typeConfig.EnabledCards == nil {
					typeConfig.EnabledCards = make(map[string]bool)
				}
			}

			// Assign roles
			AssignRolesWithConfig(players, cardService, tt.roleConfig, roleService)

			// Count assigned roles
			leaderCount := 0
			guardianCount := 0
			assassinCount := 0
			traitorCount := 0
			unassignedCount := 0

			for _, player := range players {
				if player.Role == nil {
					unassignedCount++
					continue
				}

				switch player.Role.GetRoleType() {
				case RoleLeader:
					leaderCount++
				case RoleGuardian:
					guardianCount++
				case RoleAssassin:
					assassinCount++
				case RoleTraitor:
					traitorCount++
				}
			}

			// Verify expectations
			if tt.expectLeader && leaderCount == 0 {
				t.Errorf("Expected at least one Leader, but got none")
			}
			if !tt.expectLeader && leaderCount > 0 {
				t.Errorf("Expected no Leaders, but got %d", leaderCount)
			}

			if tt.expectGuardian && guardianCount == 0 {
				t.Errorf("Expected at least one Guardian, but got none")
			}
			if !tt.expectGuardian && guardianCount > 0 {
				t.Errorf("Expected no Guardians, but got %d", guardianCount)
			}

			if tt.expectAssassin && assassinCount == 0 {
				t.Errorf("Expected at least one Assassin, but got none")
			}
			if !tt.expectAssassin && assassinCount > 0 {
				t.Errorf("Expected no Assassins, but got %d", assassinCount)
			}

			if tt.expectTraitor && traitorCount == 0 {
				t.Errorf("Expected at least one Traitor, but got none")
			}
			if !tt.expectTraitor && traitorCount > 0 {
				t.Errorf("Expected no Traitors, but got %d", traitorCount)
			}

			// Log the actual distribution for debugging
			t.Logf("Role distribution: Leaders=%d, Guardians=%d, Assassins=%d, Traitors=%d, Unassigned=%d",
				leaderCount, guardianCount, assassinCount, traitorCount, unassignedCount)

			// Verify configured counts match assigned counts (when player count allows)
			totalConfigured := 0
			for _, typeConfig := range tt.roleConfig.RoleTypes {
				totalConfigured += typeConfig.Count
			}

			if totalConfigured <= tt.playerCount {
				// When we have enough players, counts should match exactly
				if leaderCount != tt.roleConfig.RoleTypes["Leader"].Count {
					t.Errorf("Leader count mismatch: expected %d, got %d",
						tt.roleConfig.RoleTypes["Leader"].Count, leaderCount)
				}
				if guardianCount != tt.roleConfig.RoleTypes["Guardian"].Count {
					t.Errorf("Guardian count mismatch: expected %d, got %d",
						tt.roleConfig.RoleTypes["Guardian"].Count, guardianCount)
				}
				if assassinCount != tt.roleConfig.RoleTypes["Assassin"].Count {
					t.Errorf("Assassin count mismatch: expected %d, got %d",
						tt.roleConfig.RoleTypes["Assassin"].Count, assassinCount)
				}
				if traitorCount != tt.roleConfig.RoleTypes["Traitor"].Count {
					t.Errorf("Traitor count mismatch: expected %d, got %d",
						tt.roleConfig.RoleTypes["Traitor"].Count, traitorCount)
				}
			}
		})
	}
}

func TestCanStartValidation(t *testing.T) {
	tests := []struct {
		name       string
		room       *Room
		canStart   bool
		wantErr    string
	}{
		{
			name: "Should not allow start with more players than roles (CanStart doesn't support autoscaling)",
			room: &Room{
				State:   StateLobby,
				Players: map[string]*Player{
					"1": {ID: "1", Name: "Player 1"},
					"2": {ID: "2", Name: "Player 2"},
					"3": {ID: "3", Name: "Player 3"},
					"4": {ID: "4", Name: "Player 4"},
					"5": {ID: "5", Name: "Player 5"},
					"6": {ID: "6", Name: "Player 6"},
				},
				RoleConfig: &RoleConfiguration{
					PresetName: "standard", // Presets support autoscaling in GetValidationState, but not CanStart
					MinPlayers: 1,
					MaxPlayers: 6,
					RoleTypes: map[string]*RoleTypeConfig{
						"Leader":   {Count: 1},
						"Guardian": {Count: 1},
						"Assassin": {Count: 1},
						"Traitor":  {Count: 0},
					},
				},
			},
			canStart: false, // CanStart() doesn't handle autoscaling
		},
		{
			name: "Should not require leader if leaderless allowed",
			room: &Room{
				State:   StateLobby,
				Players: map[string]*Player{
					"1": {ID: "1", Name: "Player 1"},
					"2": {ID: "2", Name: "Player 2"},
				},
				RoleConfig: &RoleConfiguration{
					AllowLeaderlessGame: true,
					RoleTypes: map[string]*RoleTypeConfig{
						"Leader":   {Count: 0},
						"Guardian": {Count: 1},
						"Assassin": {Count: 1},
					},
				},
			},
			canStart: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canStart := tt.room.CanStart()
			if canStart != tt.canStart {
				t.Errorf("CanStart() = %v, want %v", canStart, tt.canStart)
			}
		})
	}
}