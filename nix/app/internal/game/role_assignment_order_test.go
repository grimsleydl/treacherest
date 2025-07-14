package game

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"treacherest"
)

func TestRoleAssignmentOrder(t *testing.T) {
	// Create a card service
	cardService, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	require.NoError(t, err)

	t.Run("Leader assigned first when more roles than players", func(t *testing.T) {
		// Create a scenario where we have more configured roles than players
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			AllowLeaderlessGame: false,
			RoleTypes: map[string]*RoleTypeConfig{
				"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Blood Empress": true}},
				"Guardian": {Count: 3, EnabledCards: map[string]bool{"The Bodyguard": true, "The Knight": true}},
				"Assassin": {Count: 2, EnabledCards: map[string]bool{"The Assassin": true}},
				"Traitor":  {Count: 2, EnabledCards: map[string]bool{"The Banisher": true, "The Cleaner": true}},
			},
		}

		// Test multiple times to ensure consistency
		for i := 0; i < 20; i++ {
			// Create 4 players but configure 8 roles
			players := []*Player{
				{ID: "1", Name: "Player 1"},
				{ID: "2", Name: "Player 2"},
				{ID: "3", Name: "Player 3"},
				{ID: "4", Name: "Player 4"},
			}

			// Assign roles
			AssignRolesWithConfig(players, cardService, roleConfig, nil)

			// Check if leader was assigned
			hasLeader := false
			leaderCount := 0
			for _, p := range players {
				if p.Role != nil && p.Role.GetRoleType() == RoleLeader {
					hasLeader = true
					leaderCount++
				}
			}
			
			// Debug output
			if !hasLeader && i == 0 {
				t.Logf("Debug: Role distribution for iteration 1:")
				for _, p := range players {
					if p.Role != nil {
						t.Logf("  Player %s: %s (type: %s)", p.Name, p.Role.Name, p.Role.GetRoleType())
					} else {
						t.Logf("  Player %s: NO ROLE", p.Name)
					}
				}
			}

			// When AllowLeaderlessGame is false and Leader count > 0, leader must be assigned
			assert.True(t, hasLeader, "Iteration %d: Leader should be assigned when configured", i+1)
			assert.Equal(t, 1, leaderCount, "Iteration %d: Exactly one leader should be assigned", i+1)
		}
	})

	t.Run("Legacy assignment also prioritizes leaders", func(t *testing.T) {
		// Test with various player counts
		playerCounts := []int{2, 3, 4, 5, 6, 7, 8}

		for _, count := range playerCounts {
			// Create players
			players := make([]*Player, count)
			for i := 0; i < count; i++ {
				players[i] = &Player{ID: fmt.Sprintf("%d", i+1), Name: fmt.Sprintf("Player %d", i+1)}
			}

			// Assign roles using legacy method
			AssignRoles(players, cardService)

			// Get expected distribution
			expectedDist := getRoleDistribution(count)
			expectedLeaders := expectedDist[RoleLeader]

			// Count actual leaders
			actualLeaders := 0
			for _, p := range players {
				if p.Role != nil && p.Role.GetRoleType() == RoleLeader {
					actualLeaders++
				}
			}

			assert.Equal(t, expectedLeaders, actualLeaders, 
				"Player count %d: Expected %d leaders, got %d", count, expectedLeaders, actualLeaders)
		}
	})

	t.Run("Custom roles with exact player count", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			AllowLeaderlessGame: false,
			RoleTypes: map[string]*RoleTypeConfig{
				"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Lich Queen": true}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true, "The Cathar": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Banisher": true}},
			},
		}

		// Create exactly 4 players for 4 configured roles
		players := []*Player{
			{ID: "1", Name: "Player 1"},
			{ID: "2", Name: "Player 2"},
			{ID: "3", Name: "Player 3"},
			{ID: "4", Name: "Player 4"},
		}

		// Assign roles
		t.Logf("Debug: Assigning roles with config: %+v", roleConfig)
		AssignRolesWithConfig(players, cardService, roleConfig, nil)

		// Verify role distribution
		roleCounts := make(map[RoleType]int)
		for _, p := range players {
			if p.Role == nil {
				t.Logf("Debug: Player %s has NO ROLE", p.Name)
			} else {
				t.Logf("Debug: Player %s has role %s", p.Name, p.Role.Name)
			}
			require.NotNil(t, p.Role, "Player %s should have a role", p.Name)
			roleCounts[p.Role.GetRoleType()]++
		}

		assert.Equal(t, 1, roleCounts[RoleLeader], "Should have exactly 1 leader")
		assert.Equal(t, 2, roleCounts[RoleGuardian], "Should have exactly 2 guardians")
		assert.Equal(t, 1, roleCounts[RoleTraitor], "Should have exactly 1 traitor")
	})
}