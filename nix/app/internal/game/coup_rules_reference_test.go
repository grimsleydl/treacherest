package game

import (
	"strings"
	"testing"
)

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

func TestCoupRoyalGuardRuleText(t *testing.T) {
	defaultText := CoupRoyalGuardRuleText(0)
	if !strings.Contains(defaultText, "any number of untapped creatures") {
		t.Fatalf("expected default rule to allow any number of blockers, got %q", defaultText)
	}
	if !strings.Contains(defaultText, "creatures attacking the King player") {
		t.Fatalf("expected default rule to protect only the King player, got %q", defaultText)
	}
	if !strings.Contains(defaultText, "Normal blocking restrictions apply") {
		t.Fatalf("expected default rule to keep normal blocking restrictions, got %q", defaultText)
	}

	oneText := CoupRoyalGuardRuleText(1)
	if !strings.Contains(oneText, "one untapped creature") {
		t.Fatalf("expected one-blocker rule text, got %q", oneText)
	}

	limitedText := CoupRoyalGuardRuleText(3)
	if !strings.Contains(limitedText, "up to 3 untapped creatures") {
		t.Fatalf("expected numeric blocker-limit rule text, got %q", limitedText)
	}
}
