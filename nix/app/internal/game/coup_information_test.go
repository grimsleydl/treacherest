package game

import (
	"strings"
	"testing"
)

func TestAssignCoupRoles_DefaultInformationPolicy(t *testing.T) {
	players := makeCoupTestPlayers(5)

	if err := AssignCoupRoles(players, CoupPresetFive); err != nil {
		t.Fatalf("AssignCoupRoles returned error: %v", err)
	}

	king := findCoupTestPlayer(t, players, RoleKing)
	blue := findCoupTestPlayer(t, players, RoleBlueKnight)
	black := findCoupTestPlayer(t, players, RoleBlackKnight)
	red := findCoupTestPlayer(t, players, RoleRedKnight)

	if !roleRulingsContain(king.Role, "Blue Knights: "+blue.Name) {
		t.Fatalf("expected King to know Blue Knight %q, got rulings %v", blue.Name, king.Role.Rulings)
	}
	if !roleRulingsContain(red.Role, "Black Knights: "+black.Name) {
		t.Fatalf("expected Red to know Black Knight %q, got rulings %v", black.Name, red.Role.Rulings)
	}
	if roleRulingsContain(black.Role, "Red Knight: "+red.Name) {
		t.Fatalf("expected Black not to know Red by default, got rulings %v", black.Role.Rulings)
	}
	if roleRulingsContain(black.Role, "Black Knights:") {
		t.Fatalf("expected Black not to know a Black network by default, got rulings %v", black.Role.Rulings)
	}
}

func TestAssignCoupRoles_BlackInformationVariants(t *testing.T) {
	players := makeCoupTestPlayers(6)
	policy := CoupInformationPolicy{
		BlackToRed:   CoupBlackToRedAll,
		BlackNetwork: CoupBlackNetworkAll,
	}

	if err := AssignCoupRolesWithInformation(players, CoupPresetSix, policy); err != nil {
		t.Fatalf("AssignCoupRolesWithInformation returned error: %v", err)
	}

	red := findCoupTestPlayer(t, players, RoleRedKnight)
	blacks := findCoupTestPlayers(t, players, RoleBlackKnight)
	if len(blacks) != 2 {
		t.Fatalf("expected 2 Black Knights, got %d", len(blacks))
	}

	for _, black := range blacks {
		if !roleRulingsContain(black.Role, "Red Knight: "+red.Name) {
			t.Fatalf("expected Black %q to know Red %q, got rulings %v", black.Name, red.Name, black.Role.Rulings)
		}
		for _, otherBlack := range blacks {
			if !roleRulingsContain(black.Role, otherBlack.Name) {
				t.Fatalf("expected Black %q to know Black network member %q, got rulings %v", black.Name, otherBlack.Name, black.Role.Rulings)
			}
		}
	}
}

func TestAssignCoupRoles_KingBlueCandidateVariant(t *testing.T) {
	players := makeCoupTestPlayers(5)
	policy := CoupInformationPolicy{
		KingToBlue: CoupKingGetsBlueCandidates,
	}

	if err := AssignCoupRolesWithInformation(players, CoupPresetFive, policy); err != nil {
		t.Fatalf("AssignCoupRolesWithInformation returned error: %v", err)
	}

	king := findCoupTestPlayer(t, players, RoleKing)
	blue := findCoupTestPlayer(t, players, RoleBlueKnight)

	if !roleRulingsContain(king.Role, "Blue Knight candidates:") {
		t.Fatalf("expected King to receive Blue candidate info, got rulings %v", king.Role.Rulings)
	}
	if !roleRulingsContain(king.Role, blue.Name) {
		t.Fatalf("expected King candidate info to include true Blue %q, got rulings %v", blue.Name, king.Role.Rulings)
	}
	if roleRulingsContain(king.Role, "Blue Knights: "+blue.Name) {
		t.Fatalf("expected candidate variant not to use full Blue knowledge wording, got rulings %v", king.Role.Rulings)
	}
}

func TestAssignCoupRoles_RedInformationVariants(t *testing.T) {
	t.Run("one black", func(t *testing.T) {
		players := makeCoupTestPlayers(6)
		policy := CoupInformationPolicy{
			RedToBlack: CoupRedKnowsOneBlack,
		}

		if err := AssignCoupRolesWithInformation(players, CoupPresetSix, policy); err != nil {
			t.Fatalf("AssignCoupRolesWithInformation returned error: %v", err)
		}

		red := findCoupTestPlayer(t, players, RoleRedKnight)
		blacks := findCoupTestPlayers(t, players, RoleBlackKnight)
		knownBlacks := 0
		for _, black := range blacks {
			if roleRulingsContain(red.Role, black.Name) {
				knownBlacks++
			}
		}
		if knownBlacks != 1 {
			t.Fatalf("expected Red to know exactly 1 Black Knight, knew %d via rulings %v", knownBlacks, red.Role.Rulings)
		}
	})

	t.Run("none", func(t *testing.T) {
		players := makeCoupTestPlayers(6)
		policy := CoupInformationPolicy{
			RedToBlack: CoupRedKnowsNoBlack,
		}

		if err := AssignCoupRolesWithInformation(players, CoupPresetSix, policy); err != nil {
			t.Fatalf("AssignCoupRolesWithInformation returned error: %v", err)
		}

		red := findCoupTestPlayer(t, players, RoleRedKnight)
		if roleRulingsContain(red.Role, "Black Knights:") {
			t.Fatalf("expected Red to know no Black Knights, got rulings %v", red.Role.Rulings)
		}
	})
}

func findCoupTestPlayer(t *testing.T, players []*Player, roleType RoleType) *Player {
	t.Helper()
	for _, player := range players {
		if player.Role != nil && player.Role.GetRoleType() == roleType {
			return player
		}
	}
	t.Fatalf("expected to find %s in assigned players", roleType)
	return nil
}

func findCoupTestPlayers(t *testing.T, players []*Player, roleType RoleType) []*Player {
	t.Helper()
	found := make([]*Player, 0)
	for _, player := range players {
		if player.Role != nil && player.Role.GetRoleType() == roleType {
			found = append(found, player)
		}
	}
	if len(found) == 0 {
		t.Fatalf("expected to find %s in assigned players", roleType)
	}
	return found
}

func roleRulingsContain(card *Card, want string) bool {
	if card == nil {
		return false
	}
	for _, ruling := range card.Rulings {
		if strings.Contains(ruling, want) {
			return true
		}
	}
	return false
}
