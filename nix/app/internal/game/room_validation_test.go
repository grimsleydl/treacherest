package game

import (
	"testing"
	"time"
	"treacherest/internal/config"
)

func TestGetValidationState(t *testing.T) {
	// Create test config
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MinPlayersPerRoom: 1,
			MaxPlayersPerRoom: 10,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName:    "Leader",
					Category:       "Leader",
					AlwaysRevealed: true,
				},
				"guardian": {
					DisplayName: "Guardian",
					Category:    "Guardian",
				},
				"assassin": {
					DisplayName: "Assassin",
					Category:    "Assassin",
				},
				"traitor": {
					DisplayName: "Traitor",
					Category:    "Traitor",
				},
			},
			Presets: map[string]config.Preset{
				"standard": {
					Name:        "standard",
					Description: "Standard game",
					Distributions: map[int]map[string]int{
						4: {"leader": 1, "guardian": 2, "assassin": 1},
						5: {"leader": 1, "guardian": 2, "assassin": 1, "traitor": 1},
						6: {"leader": 1, "guardian": 3, "assassin": 2},
					},
				},
			},
		},
	}

	roleService := NewRoleConfigService(cfg)

	tests := []struct {
		name                    string
		setupRoom               func() *Room
		expectedCanStart        bool
		expectedCanAutoScale    bool
		expectedMessageContains string
	}{
		{
			name: "Valid configuration with matching players",
			setupRoom: func() *Room {
				room := &Room{
					Code:    "TEST1",
					State:   StateLobby,
					Players: make(map[string]*Player),
					RoleConfig: &RoleConfiguration{
						PresetName: "standard",
						MinPlayers: 4,
						MaxPlayers: 6,
						RoleTypes: map[string]*RoleTypeConfig{
							"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader": true}},
							"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian": true}},
							"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin": true}},
						},
					},
				}
				// Add 4 players
				for i := 0; i < 4; i++ {
					room.Players[string(rune('A'+i))] = &Player{ID: string(rune('A' + i)), IsHost: false}
				}
				return room
			},
			expectedCanStart:        true,
			expectedCanAutoScale:    false,
			expectedMessageContains: "",
		},
		{
			name: "Too many players, can auto-scale",
			setupRoom: func() *Room {
				room := &Room{
					Code:    "TEST2",
					State:   StateLobby,
					Players: make(map[string]*Player),
					RoleConfig: &RoleConfiguration{
						PresetName: "standard",
						MinPlayers: 4,
						MaxPlayers: 6,
						RoleTypes: map[string]*RoleTypeConfig{
							"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader": true}},
							"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian": true}},
							"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin": true}},
						},
					},
				}
				// Add 5 players (more than configured 4)
				for i := 0; i < 5; i++ {
					room.Players[string(rune('A'+i))] = &Player{ID: string(rune('A' + i)), IsHost: false}
				}
				return room
			},
			expectedCanStart:        true,
			expectedCanAutoScale:    true,
			expectedMessageContains: "Will auto-scale roles from 4 to 5 players",
		},
		{
			name: "Custom configuration cannot auto-scale",
			setupRoom: func() *Room {
				room := &Room{
					Code:    "TEST3",
					State:   StateLobby,
					Players: make(map[string]*Player),
					RoleConfig: &RoleConfiguration{
						PresetName: "custom",
						MinPlayers: 4,
						MaxPlayers: 6,
						RoleTypes: map[string]*RoleTypeConfig{
							"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader": true}},
							"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian": true}},
							"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin": true}},
						},
					},
				}
				// Add 5 players (more than configured 4)
				for i := 0; i < 5; i++ {
					room.Players[string(rune('A'+i))] = &Player{ID: string(rune('A' + i)), IsHost: false}
				}
				return room
			},
			expectedCanStart:        false,
			expectedCanAutoScale:    false,
			expectedMessageContains: "Not enough roles configured (4) for 5 players",
		},
		{
			name: "No players cannot start",
			setupRoom: func() *Room {
				return &Room{
					Code:    "TEST4",
					State:   StateLobby,
					Players: make(map[string]*Player),
					RoleConfig: &RoleConfiguration{
						PresetName: "standard",
						MinPlayers: 4,
						MaxPlayers: 6,
						RoleTypes: map[string]*RoleTypeConfig{
							"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader": true}},
							"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian": true}},
							"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin": true}},
						},
					},
				}
			},
			expectedCanStart:        false,
			expectedCanAutoScale:    false,
			expectedMessageContains: "Need at least 1 player to start",
		},
		{
			name: "Not in lobby state",
			setupRoom: func() *Room {
				room := &Room{
					Code:    "TEST5",
					State:   StatePlaying,
					Players: make(map[string]*Player),
					RoleConfig: &RoleConfiguration{
						PresetName: "standard",
						MinPlayers: 4,
						MaxPlayers: 6,
						RoleTypes: map[string]*RoleTypeConfig{
							"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader": true}},
							"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian": true}},
							"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin": true}},
						},
					},
				}
				// Add 4 players
				for i := 0; i < 4; i++ {
					room.Players[string(rune('A'+i))] = &Player{ID: string(rune('A' + i)), IsHost: false}
				}
				return room
			},
			expectedCanStart:        false,
			expectedCanAutoScale:    false,
			expectedMessageContains: "Game is not in lobby state",
		},
		{
			name: "Missing leader role",
			setupRoom: func() *Room {
				room := &Room{
					Code:    "TEST6",
					State:   StateLobby,
					Players: make(map[string]*Player),
					RoleConfig: &RoleConfiguration{
						PresetName:          "custom",
						MinPlayers:          2,
						MaxPlayers:          6,
						AllowLeaderlessGame: false,
						RoleTypes: map[string]*RoleTypeConfig{
							"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian": true}},
						},
					},
				}
				// Add 2 players
				for i := 0; i < 2; i++ {
					room.Players[string(rune('A'+i))] = &Player{ID: string(rune('A' + i)), IsHost: false}
				}
				return room
			},
			expectedCanStart:        false,
			expectedCanAutoScale:    false,
			expectedMessageContains: "Leader role is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := tt.setupRoom()
			state := room.GetValidationState(roleService)

			// Check version and timestamp
			if state.Version <= 0 {
				t.Errorf("Expected version > 0, got %d", state.Version)
			}
			if state.Timestamp.IsZero() {
				t.Error("Expected non-zero timestamp")
			}
			if state.Timestamp.After(time.Now()) {
				t.Error("Timestamp should not be in the future")
			}

			// Check CanStart
			if state.CanStart != tt.expectedCanStart {
				t.Errorf("Expected CanStart=%v, got %v", tt.expectedCanStart, state.CanStart)
			}

			// Check CanAutoScale
			if state.CanAutoScale != tt.expectedCanAutoScale {
				t.Errorf("Expected CanAutoScale=%v, got %v", tt.expectedCanAutoScale, state.CanAutoScale)
			}

			// Check validation message
			if tt.expectedMessageContains != "" && state.ValidationMessage == "" {
				t.Errorf("Expected validation message containing '%s', but got empty", tt.expectedMessageContains)
			} else if tt.expectedMessageContains != "" && !contains(state.ValidationMessage, tt.expectedMessageContains) {
				t.Errorf("Expected validation message containing '%s', got '%s'", tt.expectedMessageContains, state.ValidationMessage)
			} else if tt.expectedMessageContains == "" && state.ValidationMessage != "" {
				t.Errorf("Expected empty validation message, got '%s'", state.ValidationMessage)
			}

			// Verify version increments on subsequent calls
			state2 := room.GetValidationState(roleService)
			if state2.Version <= state.Version {
				t.Errorf("Expected version to increment, got %d -> %d", state.Version, state2.Version)
			}
		})
	}
}

func TestCanAutoScale(t *testing.T) {
	cfg := &config.ServerConfig{
		Roles: config.RolesConfig{
			Presets: map[string]config.Preset{
				"standard": {
					Name: "standard",
					Distributions: map[int]map[string]int{
						4: {"leader": 1, "guardian": 2, "assassin": 1},
						5: {"leader": 1, "guardian": 2, "assassin": 1, "traitor": 1},
						6: {"leader": 1, "guardian": 3, "assassin": 2},
						8: {"leader": 1, "guardian": 4, "assassin": 2, "traitor": 1},
					},
				},
			},
		},
	}

	roleService := NewRoleConfigService(cfg)

	tests := []struct {
		name             string
		config           *RoleConfiguration
		targetPlayers    int
		expectedCanScale bool
		expectedDetails  string
	}{
		{
			name: "Custom config cannot scale",
			config: &RoleConfiguration{
				PresetName: "custom",
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader": {Count: 1},
				},
			},
			targetPlayers:    5,
			expectedCanScale: false,
			expectedDetails:  "Custom configurations do not support auto-scaling",
		},
		{
			name: "Exact distribution match",
			config: &RoleConfiguration{
				PresetName: "standard",
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 1},
					"Guardian": {Count: 2},
					"Assassin": {Count: 1},
				},
			},
			targetPlayers:    5,
			expectedCanScale: true,
			expectedDetails:  "Can scale from 4 to 5 players using standard preset",
		},
		{
			name: "Scale to nearby distribution",
			config: &RoleConfiguration{
				PresetName: "standard",
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 1},
					"Guardian": {Count: 2},
					"Assassin": {Count: 1},
				},
			},
			targetPlayers:    7,
			expectedCanScale: true,
			expectedDetails:  "Can scale to 7 players by adapting 8-player standard preset",
		},
		{
			name: "Scale by adding guardians",
			config: &RoleConfiguration{
				PresetName: "standard",
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 1},
					"Guardian": {Count: 4},
					"Assassin": {Count: 2},
					"Traitor":  {Count: 1},
				},
			},
			targetPlayers:    10,
			expectedCanScale: true,
			expectedDetails:  "Can scale to 10 players by adding 2 guardian role(s) to 8-player standard preset",
		},
		{
			name: "Cannot scale - no suitable distribution",
			config: &RoleConfiguration{
				PresetName: "minimal",
				RoleTypes:  map[string]*RoleTypeConfig{},
			},
			targetPlayers:    5,
			expectedCanScale: false,
			expectedDetails:  "Preset 'minimal' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canScale, details := roleService.CanAutoScale(tt.config, tt.targetPlayers)

			if canScale != tt.expectedCanScale {
				t.Errorf("Expected canScale=%v, got %v", tt.expectedCanScale, canScale)
			}

			if details != tt.expectedDetails {
				t.Errorf("Expected details='%s', got '%s'", tt.expectedDetails, details)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
