package pages

import (
	"strings"
	"testing"

	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestGatheringChecklistIsIdenticalAndNonVacuousOnEverySurface(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, owner, _, host := gatheringRenderingRoom()
	if _, err := room.ChooseGatheringMode(owner.ID, game.GatheringModeWhite); err != nil {
		t.Fatal(err)
	}

	surfaces := []struct {
		name string
		html string
	}{
		{"player game view", renderer.Render(GameContent(room, owner)).GetHTML()},
		{"operator dashboard", renderer.Render(HostDashboardPlaying(room, host)).GetHTML()},
		{"debug view-as-player", renderer.Render(GamePageWithDebug(room, owner, true)).GetHTML()},
		{"public host table state", renderer.Render(GameRosterZone(room, nil)).GetHTML()},
	}

	var canonical string
	for _, surface := range surfaces {
		t.Run(surface.name, func(t *testing.T) {
			checklist := gatheringChecklistSection(t, surface.html)
			for _, expected := range []string{
				`data-gathering-card="54"`,
				`data-gathering-mode="white"`,
				`data-gathering-chosen="true"`,
				`data-gathering-selectable="false"`,
				"White",
				"Create four 1/1 white Soldier creature tokens.",
				"Chosen #1 by Alice",
				`data-gathering-mode="green"`,
				`data-gathering-chosen="false"`,
				`data-gathering-selectable="true"`,
				"Destroy target noncreature permanent.",
				"Available",
			} {
				if !strings.Contains(checklist, expected) {
					t.Fatalf("%s missing non-vacuous checklist rendering %q in %s", surface.name, expected, checklist)
				}
			}
			if canonical == "" {
				canonical = checklist
			} else if checklist != canonical {
				t.Fatalf("%s checklist differs from the public component rendered on other surfaces\nwant: %s\ngot: %s", surface.name, canonical, checklist)
			}
		})
	}

	ownerHTML := surfaces[0].html
	chosenButton := gatheringModeButton(t, ownerHTML, game.GatheringModeWhite)
	for _, expected := range []string{"disabled", "White — already chosen", "/room/GATHR/player/owner/gathering/choose/white"} {
		if !strings.Contains(chosenButton, expected) {
			t.Fatalf("chosen owner control missing %q in %s", expected, chosenButton)
		}
	}
	availableButton := gatheringModeButton(t, ownerHTML, game.GatheringModeBlue)
	if strings.Contains(availableButton, " disabled") || !strings.Contains(availableButton, "Choose Blue") {
		t.Fatalf("unchosen owner control is not selectable: %s", availableButton)
	}
}

func TestGatheringRenderedChecklistTravelsWithRedistributedCard(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, owner, observer, _ := gatheringRenderingRoom()
	if _, err := room.ChooseGatheringMode(owner.ID, game.GatheringModeRed); err != nil {
		t.Fatal(err)
	}
	card := owner.Role
	checklist := card.GatheringChecklist
	otherCard := observer.Role
	if err := room.RedistributeRoles(map[string]*game.Card{
		owner.ID:    otherCard,
		observer.ID: card,
	}); err != nil {
		t.Fatal(err)
	}

	roster := renderer.Render(GameRosterZone(room, nil)).GetHTML()
	destination := gatheringRenderedElementByID(t, roster, "player-row-observer", "</details></div>")
	for _, expected := range []string{"The Gathering", `data-gathering-card="54"`, `data-gathering-mode="red"`, `data-gathering-chosen="true"`, "Chosen #1 by Alice"} {
		if !strings.Contains(destination, expected) {
			t.Fatalf("redistributed destination missing %q in %s", expected, destination)
		}
	}
	if observer.Role != card || observer.Role.GatheringChecklist != checklist {
		t.Fatal("redistribution replaced the card-attached Gathering checklist")
	}
	source := gatheringRenderedElementByID(t, roster, "player-row-owner", "</details></div>")
	if strings.Contains(source, "gathering-checklist-public-state") {
		t.Fatalf("Gathering checklist stayed behind on its former controller: %s", source)
	}

	newOwnerHTML := renderer.Render(GameContent(room, observer)).GetHTML()
	button := gatheringModeButton(t, newOwnerHTML, game.GatheringModeRed)
	if !strings.Contains(button, "disabled") || !strings.Contains(button, "/room/GATHR/player/observer/gathering/choose/red") {
		t.Fatalf("new controller did not receive the disabled chosen-mode control: %s", button)
	}
}

func gatheringRenderingRoom() (*game.Room, *game.Player, *game.Player, *game.Player) {
	owner := game.NewPlayer("owner", "Alice", "owner-session")
	owner.Role = game.NewDealtRoleCard(&game.Card{
		ID: game.TheGatheringCardID, Name: "The Gathering", Type: "Identity — Leader", Text: "Gathering oracle text", Types: game.CardTypes{Subtype: "Leader"},
	}, 3)
	owner.FaceUp = true
	owner.RoleRevealed = true
	observer := game.NewPlayer("observer", "Bob", "observer-session")
	observer.Role = game.NewDealtRoleCard(&game.Card{
		ID: 7, Name: "Observer Role", Types: game.CardTypes{Subtype: "Guardian"},
	}, 3)
	host := game.NewPlayer("host", "Host", "host-session")
	host.IsHost = true
	room := &game.Room{
		Code:  "GATHR",
		State: game.StatePlaying,
		Players: map[string]*game.Player{
			owner.ID: owner, observer.ID: observer, host.ID: host,
		},
	}
	return room, owner, observer, host
}

func gatheringChecklistSection(t *testing.T, html string) string {
	t.Helper()
	anchor := strings.Index(html, `data-gathering-card="54"`)
	if anchor < 0 {
		t.Fatalf("render missing Gathering checklist in %s", html)
	}
	start := strings.LastIndex(html[:anchor], "<section")
	end := strings.Index(html[anchor:], "</section>")
	if start < 0 || end < 0 {
		t.Fatalf("render has malformed Gathering checklist in %s", html)
	}
	return html[start : anchor+end+len("</section>")]
}

func gatheringModeButton(t *testing.T, html string, mode game.GatheringMode) string {
	t.Helper()
	anchorText := `data-gathering-mode-control="` + string(mode) + `"`
	anchor := strings.Index(html, anchorText)
	if anchor < 0 {
		t.Fatalf("render missing Gathering %s control in %s", mode, html)
	}
	start := strings.LastIndex(html[:anchor], "<button")
	end := strings.Index(html[anchor:], "</button>")
	if start < 0 || end < 0 {
		t.Fatalf("render has malformed Gathering %s control in %s", mode, html)
	}
	return html[start : anchor+end+len("</button>")]
}

func gatheringRenderedElementByID(t *testing.T, html, id, terminator string) string {
	t.Helper()
	start := strings.Index(html, `id="`+id+`"`)
	if start < 0 {
		t.Fatalf("render missing anchored element %q in %s", id, html)
	}
	rest := html[start:]
	end := strings.Index(rest, terminator)
	if end < 0 {
		t.Fatalf("anchored element %q missing terminator %q in %s", id, terminator, rest)
	}
	return rest[:end+len(terminator)]
}
