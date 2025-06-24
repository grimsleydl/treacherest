package game

import (
	"testing"
)

func TestAssignRoles(t *testing.T) {
	tests := []struct {
		name        string
		playerCount int
		wantLeaders int
		wantRoles   int
	}{
		{"1 player", 1, 1, 1},
		{"2 players", 2, 1, 2},
		{"3 players", 3, 1, 3},
		{"4 players", 4, 1, 4},
		{"5 players", 5, 1, 5},
		{"6 players", 6, 1, 6},
		{"7 players", 7, 1, 7},
		{"8 players", 8, 1, 8},
	}

	// Create CardService for testing
	cardService, err := NewCardService()
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create players
			players := make([]*Player, tt.playerCount)
			for i := 0; i < tt.playerCount; i++ {
				players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
			}

			// Assign roles
			AssignRoles(players, cardService)

			// Count roles
			leaderCount := 0
			rolesAssigned := 0

			for _, p := range players {
				if p.Role != nil {
					rolesAssigned++
					if p.Role.GetRoleType() == RoleLeader {
						leaderCount++
						// Leader should be revealed
						if !p.RoleRevealed {
							t.Error("Leader should be revealed")
						}
					}
				}
			}

			if leaderCount != tt.wantLeaders {
				t.Errorf("Expected %d leader, got %d", tt.wantLeaders, leaderCount)
			}

			if rolesAssigned != tt.wantRoles {
				t.Errorf("Expected %d roles assigned, got %d", tt.wantRoles, rolesAssigned)
			}
		})
	}
}

func TestGetRoleDistribution(t *testing.T) {
	tests := []struct {
		playerCount int
		wantRoles   map[RoleType]int
	}{
		{
			1,
			map[RoleType]int{
				RoleLeader: 1,
			},
		},
		{
			2,
			map[RoleType]int{
				RoleLeader:  1,
				RoleTraitor: 1,
			},
		},
		{
			3,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 1,
				RoleTraitor:  1,
			},
		},
		{
			4,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleTraitor:  1,
			},
		},
		{
			5,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleAssassin: 1,
				RoleTraitor:  1,
			},
		},
		{
			6,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleAssassin: 2,
				RoleTraitor:  1,
			},
		},
		{
			7,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 3,
				RoleAssassin: 2,
				RoleTraitor:  1,
			},
		},
		{
			8,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 3,
				RoleAssassin: 2,
				RoleTraitor:  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.playerCount))+" players", func(t *testing.T) {
			roles := getRoleDistribution(tt.playerCount)

			// Check counts match expected
			for roleType, wantCount := range tt.wantRoles {
				if roles[roleType] != wantCount {
					t.Errorf("Role %s: want %d, got %d", roleType, wantCount, roles[roleType])
				}
			}

			// Verify total count
			totalCount := 0
			for _, count := range roles {
				totalCount += count
			}
			if totalCount != tt.playerCount {
				t.Errorf("Total role count %d doesn't match player count %d", totalCount, tt.playerCount)
			}
		})
	}
}

func TestAssignRoles_NoDuplicateCards(t *testing.T) {
	// Create CardService for testing
	cardService, err := NewCardService()
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test with maximum players to increase chance of duplicates
	players := make([]*Player, 8)
	for i := 0; i < 8; i++ {
		players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
	}

	// Run multiple times to test randomness
	for run := 0; run < 10; run++ {
		// Reset players' roles
		for _, p := range players {
			p.Role = nil
			p.RoleRevealed = false
		}

		// Assign roles
		AssignRoles(players, cardService)

		// Track used cards to check for duplicates
		usedCards := make(map[*Card]bool)
		for _, p := range players {
			if p.Role == nil {
				t.Errorf("Player has nil role in run %d", run)
				continue
			}

			if usedCards[p.Role] {
				t.Errorf("Duplicate card assigned in run %d: %s", run, p.Role.Name)
			}
			usedCards[p.Role] = true
		}
	}
}

func TestAssignRoles_CorrectRoleTypes(t *testing.T) {
	// Create CardService for testing
	cardService, err := NewCardService()
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test each player count configuration
	for playerCount := 1; playerCount <= 8; playerCount++ {
		t.Run(string(rune('0'+playerCount))+" players", func(t *testing.T) {
			players := make([]*Player, playerCount)
			for i := 0; i < playerCount; i++ {
				players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
			}

			// Assign roles
			AssignRoles(players, cardService)

			// Get expected distribution
			expectedDist := getRoleDistribution(playerCount)

			// Count actual role types
			actualDist := make(map[RoleType]int)
			for _, p := range players {
				if p.Role != nil {
					roleType := p.Role.GetRoleType()
					actualDist[roleType]++
				}
			}

			// Compare distributions
			for roleType, expectedCount := range expectedDist {
				if actualDist[roleType] != expectedCount {
					t.Errorf("Role %s: expected %d, got %d", roleType, expectedCount, actualDist[roleType])
				}
			}
		})
	}
}
