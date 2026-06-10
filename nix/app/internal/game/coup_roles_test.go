package game

import "testing"

func TestAssignCoupRoles_DocumentedPresetMatrix(t *testing.T) {
	tests := []struct {
		name          string
		preset        CoupPreset
		playerCount   int
		expectedRoles map[RoleType]int
	}{
		{
			name:        "five players",
			preset:      CoupPresetFive,
			playerCount: 5,
			expectedRoles: map[RoleType]int{
				RoleKing:        1,
				RoleBlueKnight:  1,
				RoleBlackKnight: 1,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
		},
		{
			name:        "six players",
			preset:      CoupPresetSix,
			playerCount: 6,
			expectedRoles: map[RoleType]int{
				RoleKing:        1,
				RoleBlueKnight:  1,
				RoleBlackKnight: 2,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
		},
		{
			name:        "seven players",
			preset:      CoupPresetSeven,
			playerCount: 7,
			expectedRoles: map[RoleType]int{
				RoleKing:        1,
				RoleBlueKnight:  2,
				RoleBlackKnight: 2,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
		},
		{
			name:        "eight players",
			preset:      CoupPresetEight,
			playerCount: 8,
			expectedRoles: map[RoleType]int{
				RoleKing:        1,
				RoleBlueKnight:  2,
				RoleBlackKnight: 3,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
		},
		{
			name:        "eight player chaos",
			preset:      CoupPresetEightChaos,
			playerCount: 8,
			expectedRoles: map[RoleType]int{
				RoleKing:        1,
				RoleBlueKnight:  2,
				RoleBlackKnight: 2,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
				RoleWasteland:   1,
			},
		},
		{
			name:        "nine players",
			preset:      CoupPresetNine,
			playerCount: 9,
			expectedRoles: map[RoleType]int{
				RoleKing:        1,
				RoleBlueKnight:  2,
				RoleBlackKnight: 3,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
				RoleWasteland:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			players := makeCoupTestPlayers(tt.playerCount)

			if err := AssignCoupRoles(players, tt.preset); err != nil {
				t.Fatalf("AssignCoupRoles returned error: %v", err)
			}

			gotRoles := map[RoleType]int{}
			for _, player := range players {
				if player.Role == nil {
					t.Fatalf("expected player %s to have a role", player.Name)
				}
				gotRoles[player.Role.GetRoleType()]++
			}

			for roleType, expectedCount := range tt.expectedRoles {
				if gotRoles[roleType] != expectedCount {
					t.Errorf("expected %d %s role(s), got %d in %v", expectedCount, roleType, gotRoles[roleType], gotRoles)
				}
			}
			if len(gotRoles) != len(tt.expectedRoles) {
				t.Errorf("expected only roles %v, got %v", tt.expectedRoles, gotRoles)
			}
		})
	}
}

func makeCoupTestPlayers(count int) []*Player {
	players := make([]*Player, 0, count)
	for i := 0; i < count; i++ {
		id := string(rune('a' + i))
		players = append(players, NewPlayer(id, "Player "+id, "session-"+id))
	}
	return players
}
