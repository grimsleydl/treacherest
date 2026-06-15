package components

import (
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestPlayerRowPublicState(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "ROWS1",
		State:   game.StatePlaying,
		Players: map[string]*game.Player{},
	}
	viewer := &game.Player{
		ID:       "viewer",
		Name:     "Viewer",
		JoinedAt: time.Unix(1, 0),
	}
	hidden := &game.Player{
		ID:       "hidden",
		Name:     "Hidden Player",
		Role:     playerRowCard("Hidden Assassin", game.RoleAssassin),
		FaceUp:   false,
		JoinedAt: time.Unix(2, 0),
	}
	revealed := &game.Player{
		ID:           "revealed",
		Name:         "Revealed Player",
		Role:         playerRowCard("Revealed Guardian", game.RoleGuardian),
		RoleRevealed: true,
		JoinedAt:     time.Unix(3, 0),
	}
	eliminated := &game.Player{
		ID:           "eliminated",
		Name:         "Eliminated Player",
		Role:         playerRowCard("Revealed Traitor", game.RoleTraitor),
		RoleRevealed: true,
		IsEliminated: true,
		JoinedAt:     time.Unix(4, 0),
	}
	for _, player := range []*game.Player{viewer, hidden, revealed, eliminated} {
		room.Players[player.ID] = player
	}

	t.Run("hidden face-down row is public-safe", func(t *testing.T) {
		html := renderer.Render(PlayerRow(room, hidden, viewer)).GetHTML()

		for _, expected := range []string{
			`id="player-row-hidden"`,
			"Hidden Player",
			"Face Down",
			"Card is face down.",
		} {
			if !strings.Contains(html, expected) {
				t.Fatalf("expected %q in hidden row HTML: %s", expected, html)
			}
		}
		if strings.Contains(html, "Hidden Assassin") {
			t.Fatalf("hidden role name leaked in row HTML: %s", html)
		}
	})

	t.Run("revealed row exposes public role label", func(t *testing.T) {
		html := renderer.Render(PlayerRow(room, revealed, viewer)).GetHTML()
		if !strings.Contains(html, `id="player-row-revealed"`) {
			t.Fatalf("expected stable row id in %s", html)
		}
		if !strings.Contains(html, "Revealed: Revealed Guardian") {
			t.Fatalf("expected revealed role chip in %s", html)
		}
		if !strings.Contains(html, "Revealed Guardian") {
			t.Fatalf("expected public role details in %s", html)
		}
		if !strings.Contains(html, `<img`) || !strings.Contains(html, `src="data:image/jpeg;base64,row-test"`) {
			t.Fatalf("expected revealed row expanded details to include public role image: %s", html)
		}
		if !strings.Contains(html, `href="https://mtgtreachery.net/rules/oracle/?card=revealed-guardian"`) ||
			!strings.Contains(html, `title="View on MTG Treachery Oracle"`) ||
			!strings.Contains(html, `<svg`) ||
			!strings.Contains(html, `class="w-4 h-4"`) {
			t.Fatalf("expected revealed row expanded details to include MTG Treachery role link: %s", html)
		}
		if strings.Contains(html, "Full card image") {
			t.Fatalf("expected revealed row image inline, not behind a second disclosure: %s", html)
		}
	})

	t.Run("revealed green row uses structured strict eligibility win condition", func(t *testing.T) {
		green := &game.Player{
			ID:           "green",
			Name:         "Green Player",
			Role:         playerRowCard("Green Knight", game.RoleGreenKnight),
			RoleRevealed: true,
			JoinedAt:     time.Unix(5, 0),
		}
		room.Players[green.ID] = green

		html := renderer.Render(PlayerRow(room, green, viewer)).GetHTML()

		for _, expected := range []string{
			"<ul",
			"<li",
			"May share a King-side victory while alive",
			"May share a Red-side victory while alive",
			"Does not share Black or Wasteland victories",
			"King falls do not make Green eligible",
		} {
			if !strings.Contains(html, expected) {
				t.Fatalf("expected Green row details to contain %q: %s", expected, html)
			}
		}
		if strings.Contains(html, "selected Green rules") {
			t.Fatalf("expected Green row details to omit vague selected-rules copy: %s", html)
		}
	})

	t.Run("eliminated row remains visible without strikethrough", func(t *testing.T) {
		html := renderer.Render(PlayerRow(room, eliminated, viewer)).GetHTML()
		if !strings.Contains(html, "Eliminated") {
			t.Fatalf("expected eliminated chip in %s", html)
		}
		if !strings.Contains(html, "opacity-60") {
			t.Fatalf("expected reduced emphasis in %s", html)
		}
		if strings.Contains(html, "line-through") {
			t.Fatalf("did not expect harsh strikethrough in %s", html)
		}
	})
}

func playerRowCard(name string, roleType game.RoleType) *game.Card {
	return &game.Card{
		Name:        name,
		Type:        "Role",
		Text:        "Test public role text",
		URI:         "https://mtgtreachery.net/rules/oracle/?card=" + strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		Base64Image: "data:image/jpeg;base64,row-test",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   string(roleType),
		},
	}
}
