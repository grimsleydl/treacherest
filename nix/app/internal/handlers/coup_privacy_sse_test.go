package handlers

import (
	"strings"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"
)

func TestRenderGameContent_CoupPrivacyIsScopedPerClientLikeSSE(t *testing.T) {
	room := &game.Room{
		Code:      "SSEPR",
		State:     game.StatePlaying,
		RulesMode: game.RulesModeCoup,
		Players:   make(map[string]*game.Player),
		CoupInfoPolicy: game.CoupInformationPolicy{
			KingToBlue: game.CoupKingKnowsAllBlue,
		},
	}

	kingCard := mockHandlerCoupCard(1001, "King")
	kingCard.Rulings = []string{"Private information: Blue Knights: Blue Player"}
	king := game.NewPlayer("king", "King Player", "session-king")
	king.Role = kingCard
	king.RoleRevealed = true
	king.FaceUp = true

	blue := game.NewPlayer("blue", "Blue Player", "session-blue")
	blue.Role = mockHandlerCoupCard(1002, "Blue Knight")
	blue.FaceUp = false

	blackCard := mockHandlerCoupCard(1003, "Black Knight")
	blackCard.Rulings = []string{"Private information: Red Knight: Red Player"}
	black := game.NewPlayer("black", "Black Player", "session-black")
	black.Role = blackCard
	black.FaceUp = false

	red := game.NewPlayer("red", "Red Player", "session-red")
	red.Role = mockHandlerCoupCard(1004, "Red Knight")
	red.FaceUp = false

	for _, player := range []*game.Player{king, blue, black, red} {
		room.Players[player.ID] = player
	}

	kingHTML := renderToString(pages.GameContent(room, king))
	blackHTML := renderToString(pages.GameContent(room, black))
	blueHTML := renderToString(pages.GameContent(room, blue))

	assertContainsText(t, kingHTML, "Known: Blue Knight")
	assertNotContainsText(t, kingHTML, "Private information: Blue Knights: Blue Player")
	assertNotContainsText(t, kingHTML, "Private information: Red Knight: Red Player")
	assertNotContainsText(t, kingHTML, "Black Knight")

	assertContainsText(t, blackHTML, "Private information: Red Knight: Red Player")
	assertNotContainsText(t, blackHTML, "Private information: Blue Knights: Blue Player")

	assertNotContainsText(t, blueHTML, "Private information:")
	assertNotContainsText(t, blueHTML, "Black Knight")
	assertNotContainsText(t, blueHTML, "Red Knight")
}

func assertContainsText(t *testing.T, html string, want string) {
	t.Helper()
	if !strings.Contains(html, want) {
		t.Fatalf("expected rendered HTML to contain %q", want)
	}
}

func assertNotContainsText(t *testing.T, html string, unwanted string) {
	t.Helper()
	if strings.Contains(html, unwanted) {
		t.Fatalf("expected rendered HTML not to contain %q", unwanted)
	}
}
