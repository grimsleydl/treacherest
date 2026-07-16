package pages

import (
	"fmt"
	"strings"
	"testing"

	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestLeaderCounterPublicRenderingOnEverySurface(t *testing.T) {
	tests := []struct {
		id          int
		name        string
		counterName string
		countText   string
		oncePerTurn bool
	}{
		{game.HerSeedbornHighnessCardID, "Her Seedborn Highness", "seed", "Seed counters: 2 of 3", true},
		{game.TheLichQueenCardID, "The Lich Queen", "grave", "Grave counters: 2 of 3", true},
		{game.TheQueenOfLightCardID, "The Queen of Light", "purification", "Purification counters: 0 of 1", false},
		{game.TheVoidTyrantCardID, "The Void Tyrant", "void", "Void counters: 0 of 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := testhelpers.NewTemplateRenderer(t)
			owner := game.NewPlayer("owner", "Alice", "owner-session")
			owner.Role = game.NewDealtRoleCard(&game.Card{
				ID: tt.id, Name: tt.name, Types: game.CardTypes{Subtype: "Leader"},
			}, 3)
			owner.RoleRevealed = true
			owner.FaceUp = true
			observer := game.NewPlayer("observer", "Bob", "observer-session")
			host := game.NewPlayer("host", "Host", "host-session")
			host.IsHost = true
			room := &game.Room{
				Code:  "PUBLIC",
				State: game.StatePlaying,
				Players: map[string]*game.Player{
					owner.ID: owner, observer.ID: observer, host.ID: host,
				},
			}
			if _, err := room.ActivateLeaderCounter(owner.ID); err != nil {
				t.Fatal(err)
			}

			eventText := fmt.Sprintf("Activation #1: Alice removed a %s counter", tt.counterName)
			surfaces := []struct {
				name   string
				html   string
				anchor string
			}{
				{"player view", renderer.Render(GameContent(room, owner)).GetHTML(), `id="game-container"`},
				{"operator dashboard", renderer.Render(HostDashboardPlaying(room, host)).GetHTML(), `id="operator-tile-owner"`},
				{"debug view-as-player", renderer.Render(GamePageWithDebug(room, owner, true)).GetHTML(), `id="debug-control-surface"`},
				{"public host table state", renderer.Render(GameRosterZone(room, nil)).GetHTML(), `id="zone-roster"`},
			}
			for _, surface := range surfaces {
				t.Run(surface.name, func(t *testing.T) {
					for _, expected := range []string{
						surface.anchor,
						tt.name,
						fmt.Sprintf(`data-leader-counter-card="%d"`, tt.id),
						tt.countText,
						eventText,
					} {
						if !strings.Contains(surface.html, expected) {
							t.Fatalf("%s missing non-vacuous public rendering %q in %s", surface.name, expected, surface.html)
						}
					}
				})
			}

			playerHTML := surfaces[0].html
			if !strings.Contains(playerHTML, fmt.Sprintf("/room/PUBLIC/player/owner/leader-counter/activate")) {
				t.Fatalf("owner view lacks activation action in %s", playerHTML)
			}
			if tt.oncePerTurn {
				for _, expected := range []string{"Once per turn · Used this turn", "Reset for next turn", "Confirm next turn"} {
					if !strings.Contains(playerHTML, expected) {
						t.Fatalf("once-per-turn owner state missing %q in %s", expected, playerHTML)
					}
				}
			} else if !strings.Contains(playerHTML, "No per-turn limit") {
				t.Fatalf("unlimited role did not surface its activation limit in %s", playerHTML)
			}
		})
	}
}
