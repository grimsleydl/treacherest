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
		{"4 players", 4, 1, 4},
		{"5 players", 5, 1, 5},
		{"6 players", 6, 1, 6},
		{"7 players", 7, 1, 7},
		{"8 players", 8, 1, 8},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create players
			players := make([]*Player, tt.playerCount)
			for i := 0; i < tt.playerCount; i++ {
				players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
			}
			
			// Assign roles
			AssignRoles(players)
			
			// Count roles
			leaderCount := 0
			rolesAssigned := 0
			
			for _, p := range players {
				if p.Role != nil {
					rolesAssigned++
					if p.Role.Type == RoleLeader {
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
	}
	
	for _, tt := range tests {
		t.Run(string(rune('0'+tt.playerCount))+" players", func(t *testing.T) {
			roles := getRoleDistribution(tt.playerCount)
			
			// Count roles
			counts := make(map[RoleType]int)
			for _, role := range roles {
				counts[role.Type]++
			}
			
			// Check counts
			for roleType, wantCount := range tt.wantRoles {
				if counts[roleType] != wantCount {
					t.Errorf("Role %s: want %d, got %d", roleType, wantCount, counts[roleType])
				}
			}
		})
	}
}