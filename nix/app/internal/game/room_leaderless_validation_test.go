package game

import (
	"testing"
	"treacherest/internal/config"
)

func TestRoomValidation_LeaderlessGames(t *testing.T) {
	cfg := config.DefaultConfig()
	roleService := NewRoleConfigService(cfg)

	t.Run("allows starting game with 0 leaders when leaderless enabled", func(t *testing.T) {
		room := &Room{
			Code:  "TEST1",
			State: StateLobby,
			Players: map[string]*Player{
				"p1": {ID: "p1", Name: "Player 1"},
				"p2": {ID: "p2", Name: "Player 2"},
				"p3": {ID: "p3", Name: "Player 3"},
			},
			RoleConfig: &RoleConfiguration{
				AllowLeaderlessGame: true,
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 0, EnabledCards: map[string]bool{}},
					"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true}},
					"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
				},
			},
		}

		state := room.GetValidationState(roleService)

		if !state.CanStart {
			t.Errorf("Should be able to start with 0 leaders when leaderless is enabled, but got: %s", state.ValidationMessage)
		}

		if state.ConfiguredRoles != 3 {
			t.Errorf("Expected 3 configured roles (0 leader + 2 guardian + 1 traitor), got %d", state.ConfiguredRoles)
		}
	})

	t.Run("prevents starting game with 0 leaders when leaderless disabled", func(t *testing.T) {
		room := &Room{
			Code:  "TEST2",
			State: StateLobby,
			Players: map[string]*Player{
				"p1": {ID: "p1", Name: "Player 1"},
				"p2": {ID: "p2", Name: "Player 2"},
				"p3": {ID: "p3", Name: "Player 3"},
			},
			RoleConfig: &RoleConfiguration{
				AllowLeaderlessGame: false, // Disabled
				RoleTypes: map[string]*RoleTypeConfig{
					"Leader":   {Count: 0, EnabledCards: map[string]bool{}},
					"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true}},
					"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
				},
			},
		}

		state := room.GetValidationState(roleService)

		if state.CanStart {
			t.Error("Should NOT be able to start with 0 leaders when leaderless is disabled")
		}

		if state.ValidationMessage != "Leader role is required (or enable leaderless games)" {
			t.Errorf("Expected leader required message, got: %s", state.ValidationMessage)
		}
	})

	t.Run("allows starting game with 1 leader regardless of leaderless setting", func(t *testing.T) {
		testCases := []bool{true, false}

		for _, allowLeaderless := range testCases {
			room := &Room{
				Code:  "TEST3",
				State: StateLobby,
				Players: map[string]*Player{
					"p1": {ID: "p1", Name: "Player 1"},
					"p2": {ID: "p2", Name: "Player 2"},
					"p3": {ID: "p3", Name: "Player 3"},
				},
				RoleConfig: &RoleConfiguration{
					AllowLeaderlessGame: allowLeaderless,
					RoleTypes: map[string]*RoleTypeConfig{
						"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Usurper": true}},
						"Guardian": {Count: 1, EnabledCards: map[string]bool{"The Bodyguard": true}},
						"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
					},
				},
			}

			state := room.GetValidationState(roleService)

			if !state.CanStart {
				t.Errorf("Should be able to start with 1 leader (leaderless=%v), but got: %s", allowLeaderless, state.ValidationMessage)
			}
		}
	})
}
