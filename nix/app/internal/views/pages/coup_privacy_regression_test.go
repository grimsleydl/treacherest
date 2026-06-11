package pages

import (
	"sort"
	"strings"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestGameBody_CoupInformationPolicyPrivacyBoundaries(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session-1"),
		game.NewPlayer("p2", "Player 2", "session-2"),
		game.NewPlayer("p3", "Player 3", "session-3"),
		game.NewPlayer("p4", "Player 4", "session-4"),
		game.NewPlayer("p5", "Player 5", "session-5"),
		game.NewPlayer("p6", "Player 6", "session-6"),
	}
	err := game.AssignCoupRolesWithInformation(players, game.CoupPresetSix, game.CoupInformationPolicy{
		KingToBlue:   game.CoupKingKnowsAllBlue,
		RedToBlack:   game.CoupRedKnowsAllBlack,
		BlackToRed:   game.CoupBlackToRedAll,
		BlackNetwork: game.CoupBlackNetworkAll,
	})
	if err != nil {
		t.Fatalf("assign Coup roles: %v", err)
	}

	room := coupPrivacyRoom(players...)
	king := coupPrivacyOneByRole(t, players, game.RoleKing)
	red := coupPrivacyOneByRole(t, players, game.RoleRedKnight)
	blue := coupPrivacyOneByRole(t, players, game.RoleBlueKnight)
	blacks := coupPrivacyByRole(players, game.RoleBlackKnight)
	if len(blacks) != 2 {
		t.Fatalf("expected 2 Black Knights, got %d", len(blacks))
	}

	kingInfo := "Private information: Blue Knights: " + blue.Name
	redInfo := "Private information: Black Knights: " + coupPrivacyPlayerNames(blacks)
	blackRedInfo := "Private information: Red Knight: " + red.Name
	blackNetworkInfo := "Private information: Black Knights: " + coupPrivacyPlayerNames(blacks)

	renderer.Render(GameBody(room, king)).
		AssertContains(kingInfo).
		AssertNotContains(redInfo).
		AssertNotContains(blackRedInfo)

	renderer.Render(GameBody(room, red)).
		AssertContains(redInfo).
		AssertNotContains(kingInfo).
		AssertNotContains(blackRedInfo)

	renderer.Render(GameBody(room, blacks[0])).
		AssertContains(blackRedInfo).
		AssertContains(blackNetworkInfo).
		AssertNotContains(kingInfo)

	renderer.Render(GameBody(room, blue)).
		AssertNotContains("Private information:")
}

func TestGameBody_CoupPublicInquisitionResultVisibleToAllClients(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, red, green := makeCoupInquisitionViewRoom()
	red.RoleRevealed = true
	red.FaceUp = true
	room.CoupInquisitionResultPolicy = game.CoupInquisitionResultPublic
	room.CoupInquisition = &game.CoupInquisitionState{
		Attempts: map[string]game.CoupInquisitionAttempt{
			blue.ID: {
				InquisitorID: blue.ID,
				TargetID:     red.ID,
				CurrentLife:  39,
				PenaltyLife:  20,
				Resolved:     true,
				Success:      true,
			},
		},
		Last: &game.CoupInquisitionAttempt{
			InquisitorID: blue.ID,
			TargetID:     red.ID,
			CurrentLife:  39,
			PenaltyLife:  20,
			Resolved:     true,
			Success:      true,
		},
		Succeeded: true,
	}

	renderer.Render(GameBody(room, blue)).
		AssertContains("Inquisition succeeded. Red Player was Red and has been revealed.")

	renderer.Render(GameBody(room, green)).
		AssertContains("Inquisition succeeded. Red Player was Red and has been revealed.")
}

func coupPrivacyRoom(players ...*game.Player) *game.Room {
	room := &game.Room{
		Code:      "CPRIV",
		State:     game.StatePlaying,
		RulesMode: game.RulesModeCoup,
		Players:   make(map[string]*game.Player),
	}
	for _, player := range players {
		room.Players[player.ID] = player
	}
	return room
}

func coupPrivacyOneByRole(t *testing.T, players []*game.Player, role game.RoleType) *game.Player {
	t.Helper()
	matches := coupPrivacyByRole(players, role)
	if len(matches) != 1 {
		t.Fatalf("expected 1 %s, got %d", role, len(matches))
	}
	return matches[0]
}

func coupPrivacyByRole(players []*game.Player, role game.RoleType) []*game.Player {
	matches := make([]*game.Player, 0)
	for _, player := range players {
		if player.Role != nil && player.Role.GetRoleType() == role {
			matches = append(matches, player)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})
	return matches
}

func coupPrivacyPlayerNames(players []*game.Player) string {
	names := make([]string, 0, len(players))
	for _, player := range players {
		names = append(names, player.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
