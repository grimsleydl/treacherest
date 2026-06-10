package game

import "testing"

func TestAssignCoupRoles_AddsPrivateRoleGoalText(t *testing.T) {
	players := makeCoupTestPlayers(5)

	if err := AssignCoupRoles(players, CoupPresetFive); err != nil {
		t.Fatalf("AssignCoupRoles returned error: %v", err)
	}

	for _, player := range players {
		if player.Role == nil {
			t.Fatalf("expected %s to have a role", player.Name)
		}
		if !roleRulingsContain(player.Role, "Role Goal:") {
			t.Fatalf("expected %s to have private role goal text, got rulings %v", player.Role.Name, player.Role.Rulings)
		}
	}
}
