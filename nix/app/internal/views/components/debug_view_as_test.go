package components

import (
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestDebugViewAsPlayerSelectorLabelsTreacheryRoleColors(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:                "DBG01",
		DebugViewedPlayerID: "traitor",
		Players: map[string]*game.Player{
			"leader": {
				ID:       "leader",
				Name:     "Leader Player",
				Role:     debugViewAsCard("The Queen of Light", game.RoleLeader),
				JoinedAt: time.Unix(1, 0),
			},
			"guardian": {
				ID:       "guardian",
				Name:     "Guardian Player",
				Role:     debugViewAsCard("The Bodyguard", game.RoleGuardian),
				JoinedAt: time.Unix(2, 0),
			},
			"assassin": {
				ID:       "assassin",
				Name:     "Assassin Player",
				Role:     debugViewAsCard("The Assassin", game.RoleAssassin),
				JoinedAt: time.Unix(3, 0),
			},
			"traitor": {
				ID:       "traitor",
				Name:     "Traitor Player",
				Role:     debugViewAsCard("The Villain", game.RoleTraitor),
				JoinedAt: time.Unix(4, 0),
			},
		},
	}

	html := renderer.Render(DebugViewAsPlayerSelector(room, room.Code)).GetHTML()

	for _, expected := range []string{
		"Leader Player - gold - The Queen of Light",
		"Guardian Player - blue - The Bodyguard",
		"Assassin Player - red - The Assassin",
		"Traitor Player - black - The Villain",
		`<option value="traitor" selected>Traitor Player - black - The Villain</option>`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected debug view-as selector to contain %q in %s", expected, html)
		}
	}
}

func debugViewAsCard(name string, role game.RoleType) *game.Card {
	return &game.Card{
		Name: name,
		Types: game.CardTypes{
			Subtype: string(role),
		},
	}
}
