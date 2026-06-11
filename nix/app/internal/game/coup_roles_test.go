package game

import (
	"strings"
	"testing"
)

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

func TestCoupRoleCountsForPreset_SeedsEditableCounts(t *testing.T) {
	counts, ok := CoupRoleCountsForPreset(CoupPresetEightChaos)
	if !ok {
		t.Fatal("expected eight-player chaos preset counts")
	}

	want := CoupRoleCounts{
		RoleKing:        1,
		RoleBlueKnight:  2,
		RoleBlackKnight: 2,
		RoleRedKnight:   1,
		RoleGreenKnight: 1,
		RoleWasteland:   1,
	}
	for role, wantCount := range want {
		if counts[role] != wantCount {
			t.Fatalf("expected %s count %d, got %d in %v", role, wantCount, counts[role], counts)
		}
	}

	counts[RoleWasteland] = 0
	freshCounts, _ := CoupRoleCountsForPreset(CoupPresetEightChaos)
	if freshCounts[RoleWasteland] != 1 {
		t.Fatal("expected preset counts to be returned as a defensive copy")
	}
}

func TestCoupRoleCountsForRoom_IgnoresStalePresetSeedWhenNotCustom(t *testing.T) {
	staleFivePlayerCounts, ok := CoupRoleCountsForPreset(CoupPresetFive)
	if !ok {
		t.Fatal("expected five-player preset counts")
	}

	room := &Room{
		CoupPreset:           CoupPresetSix,
		CoupRoleCounts:       staleFivePlayerCounts,
		CoupRoleCountsCustom: false,
	}

	counts := CoupRoleCountsForRoom(room)
	if counts[RoleBlackKnight] != 2 {
		t.Fatalf("expected non-custom counts to come from selected 6-player preset, got %v", counts)
	}
}

func TestCoupDefaultPresetForPlayerCount(t *testing.T) {
	tests := []struct {
		playerCount int
		want        CoupPreset
	}{
		{playerCount: 5, want: CoupPresetFive},
		{playerCount: 6, want: CoupPresetSix},
		{playerCount: 7, want: CoupPresetSeven},
		{playerCount: 8, want: CoupPresetEight},
		{playerCount: 9, want: CoupPresetNine},
	}

	for _, tt := range tests {
		got, ok := CoupDefaultPresetForPlayerCount(tt.playerCount)
		if !ok {
			t.Fatalf("expected default preset for %d players", tt.playerCount)
		}
		if got != tt.want {
			t.Fatalf("expected %d players to use %q, got %q", tt.playerCount, tt.want, got)
		}
	}
}

func TestValidateCoupRoleCounts_ReportsNormalStartFailures(t *testing.T) {
	tests := []struct {
		name        string
		counts      CoupRoleCounts
		players     int
		wantMessage string
	}{
		{
			name: "total mismatch",
			counts: CoupRoleCounts{
				RoleKing:        1,
				RoleBlueKnight:  1,
				RoleBlackKnight: 1,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
			players:     6,
			wantMessage: "Coup role counts total 5 but there are 6 active players",
		},
		{
			name: "missing King",
			counts: CoupRoleCounts{
				RoleBlueKnight:  2,
				RoleBlackKnight: 1,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
			players:     5,
			wantMessage: "Coup role counts require exactly one King",
		},
		{
			name: "missing Red",
			counts: CoupRoleCounts{
				RoleKing:        1,
				RoleBlueKnight:  2,
				RoleBlackKnight: 1,
				RoleGreenKnight: 1,
			},
			players:     5,
			wantMessage: "Coup role counts require exactly one Red Knight",
		},
		{
			name: "multiple Kings",
			counts: CoupRoleCounts{
				RoleKing:        2,
				RoleBlueKnight:  1,
				RoleRedKnight:   1,
				RoleGreenKnight: 1,
			},
			players:     5,
			wantMessage: "Coup role counts require exactly one King",
		},
		{
			name: "multiple Reds",
			counts: CoupRoleCounts{
				RoleKing:        1,
				RoleBlueKnight:  1,
				RoleRedKnight:   2,
				RoleGreenKnight: 1,
			},
			players:     5,
			wantMessage: "Coup role counts require exactly one Red Knight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoupRoleCounts(tt.counts, tt.players)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if err.Error() != tt.wantMessage {
				t.Fatalf("expected %q, got %q", tt.wantMessage, err.Error())
			}
		})
	}
}

func TestValidateCoupRoleCountsWithUnsafeOverride_SkipsOnlyKingRedCardinality(t *testing.T) {
	missingKingAndRed := CoupRoleCounts{
		RoleBlueKnight:  2,
		RoleBlackKnight: 2,
		RoleGreenKnight: 1,
	}
	if err := ValidateCoupRoleCountsWithUnsafeOverride(missingKingAndRed, 5, true); err != nil {
		t.Fatalf("expected unsafe override to allow missing King and Red, got %v", err)
	}

	tooManyKingsAndReds := CoupRoleCounts{
		RoleKing:       2,
		RoleBlueKnight: 1,
		RoleRedKnight:  2,
	}
	if err := ValidateCoupRoleCountsWithUnsafeOverride(tooManyKingsAndReds, 5, true); err != nil {
		t.Fatalf("expected unsafe override to allow multiple Kings and Reds, got %v", err)
	}

	totalMismatch := CoupRoleCounts{
		RoleBlueKnight:  2,
		RoleBlackKnight: 2,
		RoleGreenKnight: 1,
		RoleWasteland:   1,
	}
	err := ValidateCoupRoleCountsWithUnsafeOverride(totalMismatch, 5, true)
	if err == nil {
		t.Fatal("expected unsafe override not to bypass total count validation")
	}
	if err.Error() != "Coup role counts total 6 but there are 5 active players" {
		t.Fatalf("expected total mismatch error, got %q", err.Error())
	}
}

func TestAssignCoupRolesWithCounts_UsesCustomPoolAndInformationPolicy(t *testing.T) {
	players := makeCoupTestPlayers(6)
	counts := CoupRoleCounts{
		RoleKing:        1,
		RoleBlueKnight:  2,
		RoleBlackKnight: 1,
		RoleRedKnight:   1,
		RoleGreenKnight: 1,
	}
	policy := CoupInformationPolicy{
		KingToBlue: CoupKingKnowsAllBlue,
		RedToBlack: CoupRedKnowsAllBlack,
	}

	if err := AssignCoupRolesWithCountsAndInformation(players, counts, policy); err != nil {
		t.Fatalf("AssignCoupRolesWithCountsAndInformation returned error: %v", err)
	}

	gotRoles := map[RoleType]int{}
	for _, player := range players {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		roleType := player.Role.GetRoleType()
		gotRoles[roleType]++
		if roleType == RoleKing {
			if !player.RoleRevealed || !player.FaceUp {
				t.Fatalf("expected King %s to start revealed", player.Name)
			}
		} else if player.RoleRevealed || player.FaceUp {
			t.Fatalf("expected non-King %s to start hidden", player.Name)
		}
	}

	for role, wantCount := range counts {
		if gotRoles[role] != wantCount {
			t.Fatalf("expected %d %s role(s), got counts %v", wantCount, role, gotRoles)
		}
	}

	king := findCoupTestPlayer(t, players, RoleKing)
	red := findCoupTestPlayer(t, players, RoleRedKnight)
	if !cardHasRulingContaining(king.Role, "Private information: Blue Knights:") {
		t.Fatalf("expected King to receive Blue Knight private info, got %#v", king.Role.Rulings)
	}
	if !cardHasRulingContaining(red.Role, "Private information: Black Knights:") {
		t.Fatalf("expected Red to receive Black Knight private info, got %#v", red.Role.Rulings)
	}
}

func TestAssignCoupRolesWithCountsUnsafe_HandlesMissingKingAndRed(t *testing.T) {
	players := makeCoupTestPlayers(5)
	counts := CoupRoleCounts{
		RoleBlueKnight:  2,
		RoleBlackKnight: 2,
		RoleGreenKnight: 1,
	}

	if err := AssignCoupRolesWithCountsAndInformationUnsafe(players, counts, CoupInformationPolicy{}, true); err != nil {
		t.Fatalf("AssignCoupRolesWithCountsAndInformationUnsafe returned error: %v", err)
	}

	gotRoles := map[RoleType]int{}
	for _, player := range players {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		roleType := player.Role.GetRoleType()
		gotRoles[roleType]++
		if player.RoleRevealed || player.FaceUp {
			t.Fatalf("expected %s to start hidden when unsafe pool has no King", player.Name)
		}
	}

	for role, wantCount := range counts {
		if gotRoles[role] != wantCount {
			t.Fatalf("expected %d %s role(s), got counts %v", wantCount, role, gotRoles)
		}
	}
	if gotRoles[RoleKing] != 0 || gotRoles[RoleRedKnight] != 0 {
		t.Fatalf("expected unsafe pool to omit King and Red, got counts %v", gotRoles)
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

func cardHasRulingContaining(card *Card, needle string) bool {
	if card == nil {
		return false
	}
	for _, ruling := range card.Rulings {
		if strings.Contains(ruling, needle) {
			return true
		}
	}
	return false
}
