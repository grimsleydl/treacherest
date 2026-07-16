package pages

import (
	"strings"
	"testing"

	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestDebtCounterCountIsPublicOnHiddenCardWithoutLeakingContents(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, collector, target, observer, host := debtCounterRenderingRoom()
	for i := 0; i < 2; i++ {
		if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
			t.Fatal(err)
		}
	}

	surfaces := []struct {
		name string
		html string
		id   string
		end  string
	}{
		{"player game view", renderer.Render(GameContent(room, observer)).GetHTML(), "player-row-target", "</details></div>"},
		{"operator dashboard", renderer.Render(HostDashboardPlaying(room, host)).GetHTML(), "operator-tile-target", "</article>"},
		{"debug view-as-player", renderer.Render(GamePageWithDebug(room, observer, true)).GetHTML(), "player-row-target", "</details></div>"},
		{"public host table state", renderer.Render(GameRosterZone(room, nil)).GetHTML(), "player-row-target", "</details></div>"},
	}

	for _, surface := range surfaces {
		t.Run(surface.name, func(t *testing.T) {
			segment := debtCounterElementByID(t, surface.html, surface.id, surface.end)
			for _, expected := range []string{
				`id="` + surface.id + `"`,
				`data-debt-counter-count="2"`,
				"Debt counters: 2",
				"Alice placed 2 debt counters on Bob&#39;s identity card (0 → 2).",
				"Bob should draw a card for each at the table.",
			} {
				if !strings.Contains(segment, expected) {
					t.Fatalf("%s missing non-vacuous public debt rendering %q in %s", surface.name, expected, segment)
				}
			}
			for _, secret := range []string{"SECRET DEBT ROLE", "Identity — Secret Sentinel", "SECRET ORACLE TEXT"} {
				if strings.Contains(segment, secret) {
					t.Fatalf("%s leaked hidden card contents %q in %s", surface.name, secret, segment)
				}
			}
		})
	}

	affectedHTML := renderer.Render(GameContent(room, target)).GetHTML()
	for _, expected := range []string{"SECRET DEBT ROLE", `data-debt-counter-count="2"`, "Debt counters: 2"} {
		if !strings.Contains(affectedHTML, expected) {
			t.Fatalf("affected player did not immediately see %q in own view: %s", expected, affectedHTML)
		}
	}
}

func TestDebtCounterRenderingPersistsThroughUnveilRehideAndTransfer(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, collector, target, observer, _ := debtCounterRenderingRoom()
	if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
		t.Fatal(err)
	}
	card := target.Role
	state := card.DebtCounters

	target.FaceUp = true
	target.RoleRevealed = true
	unveiled := debtCounterElementByID(t, renderer.Render(GameRosterZone(room, observer)).GetHTML(), "player-row-target", "</details></div>")
	for _, expected := range []string{"SECRET DEBT ROLE", `data-debt-counter-count="1"`, "Debt counters: 1"} {
		if !strings.Contains(unveiled, expected) {
			t.Fatalf("unveiled card missing %q in %s", expected, unveiled)
		}
	}

	target.FaceUp = false
	rehidden := debtCounterElementByID(t, renderer.Render(GameRosterZone(room, observer)).GetHTML(), "player-row-target", "</details></div>")
	if !strings.Contains(rehidden, `data-debt-counter-count="1"`) || target.Role != card || target.Role.DebtCounters != state {
		t.Fatalf("re-hidden card lost debt state: %s", rehidden)
	}

	if err := room.TransferRole(target, observer, true); err != nil {
		t.Fatal(err)
	}
	transferredRoster := renderer.Render(GameRosterZone(room, nil)).GetHTML()
	destination := debtCounterElementByID(t, transferredRoster, "player-row-observer", "</details></div>")
	if observer.Role != card || observer.Role.DebtCounters != state || !strings.Contains(destination, `data-debt-counter-count="1"`) {
		t.Fatalf("transferred card did not carry rendered debt state: %s", destination)
	}
	if strings.Contains(destination, "SECRET DEBT ROLE") {
		t.Fatalf("face-down transferred card leaked its contents: %s", destination)
	}
	source := debtCounterElementByID(t, transferredRoster, "player-row-target", "</details></div>")
	if strings.Contains(source, "debt-counter-public-state") {
		t.Fatalf("counter stayed behind on the old controller: %s", source)
	}
}

func debtCounterRenderingRoom() (*game.Room, *game.Player, *game.Player, *game.Player, *game.Player) {
	collector := game.NewPlayer("collector", "Alice", "collector-session")
	collector.Role = game.NewDealtRoleCard(&game.Card{
		ID: game.TheDebtCollectorCardID, Name: "The Debt Collector", Type: "Identity — Leader", Types: game.CardTypes{Subtype: "Leader"},
	}, 3)
	collector.FaceUp = true
	collector.RoleRevealed = true
	target := game.NewPlayer("target", "Bob", "target-session")
	target.Role = game.NewDealtRoleCard(&game.Card{
		ID: 777, Name: "SECRET DEBT ROLE", Type: "Identity — Secret Sentinel", Text: "SECRET ORACLE TEXT", Types: game.CardTypes{Subtype: "Guardian"},
	}, 3)
	target.FaceUp = false
	target.RoleRevealed = false
	observer := game.NewPlayer("observer", "Carol", "observer-session")
	observer.Role = game.NewDealtRoleCard(&game.Card{
		ID: 8, Name: "Observer Role", Type: "Identity — Assassin", Types: game.CardTypes{Subtype: "Assassin"},
	}, 3)
	observer.FaceUp = false
	observer.RoleRevealed = false
	host := game.NewPlayer("host", "Host", "host-session")
	host.IsHost = true
	room := &game.Room{
		Code:  "DEBTS",
		State: game.StatePlaying,
		Players: map[string]*game.Player{
			collector.ID: collector,
			target.ID:    target,
			observer.ID:  observer,
			host.ID:      host,
		},
	}
	return room, collector, target, observer, host
}

func debtCounterElementByID(t *testing.T, html, id, end string) string {
	t.Helper()
	start := strings.Index(html, `id="`+id+`"`)
	if start < 0 {
		t.Fatalf("render missing anchored element %q in %s", id, html)
	}
	rest := html[start:]
	stop := strings.Index(rest, end)
	if stop < 0 {
		t.Fatalf("anchored element %q missing terminator %q in %s", id, end, rest)
	}
	return rest[:stop+len(end)]
}
