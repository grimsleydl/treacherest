package pages

import (
	"strings"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestPuppetMasterFaceDownRoleVisibilityIsPrivate(t *testing.T) {
	room := &game.Room{
		Code:       "PEEK1",
		State:      game.StatePlaying,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 8,
	}

	host := game.NewPlayer("host", "Room Operator", "host-session")
	host.IsHost = true
	puppetMaster := game.NewPlayer("puppet-master", "Puppet Master Player", "puppet-session")
	puppetMaster.Role = &game.Card{
		ID:    27,
		Name:  "The Puppet Master",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	puppetMaster.FaceUp = true
	puppetMaster.RoleRevealed = true
	puppetMaster.AbilityState.GrantViewOthersFaceDown()
	otherViewer := game.NewPlayer("other-viewer", "Other Viewer", "other-session")
	otherViewer.Role = &game.Card{
		ID:    10,
		Name:  "The Bodyguard",
		Types: game.CardTypes{Subtype: "Guardian"},
	}
	otherViewer.FaceUp = false
	hiddenPlayer := game.NewPlayer("hidden-player", "Hidden Player", "hidden-session")
	hiddenPlayer.Role = &game.Card{
		ID:    18,
		Name:  "The Hidden Oracle",
		Text:  "This text is private while the identity is face down.",
		Types: game.CardTypes{Subtype: "Assassin"},
	}
	hiddenPlayer.FaceUp = false
	hiddenPlayer.RoleRevealed = false

	for _, player := range []*game.Player{host, puppetMaster, otherViewer, hiddenPlayer} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("AddPlayer(%q) error = %v", player.ID, err)
		}
	}

	t.Run("face-up Puppet Master player and debug views show the hidden identity", func(t *testing.T) {
		playerHTML := testhelpers.NewTemplateRenderer(t).Render(GameContent(room, puppetMaster)).GetHTML()
		assertRenderedID(t, playerHTML, "zone-roster")
		assertRenderedID(t, playerHTML, "player-row-"+hiddenPlayer.ID)
		assertRoleContentsRendered(t, playerHTML, hiddenPlayer.Role)

		debugHTML := testhelpers.NewTemplateRenderer(t).Render(GamePageWithDebug(room, puppetMaster, true)).GetHTML()
		assertRenderedID(t, debugHTML, "debug-control-surface")
		assertRenderedID(t, debugHTML, "player-row-"+hiddenPlayer.ID)
		assertRoleContentsRendered(t, debugHTML, hiddenPlayer.Role)
	})

	t.Run("face-down Puppet Master does not show the hidden identity", func(t *testing.T) {
		puppetMaster.FaceUp = false
		t.Cleanup(func() { puppetMaster.FaceUp = true })

		html := testhelpers.NewTemplateRenderer(t).Render(GameContent(room, puppetMaster)).GetHTML()
		assertRenderedID(t, html, "zone-roster")
		assertRenderedID(t, html, "player-row-"+hiddenPlayer.ID)
		assertRoleContentsNotRendered(t, html, hiddenPlayer.Role)
	})

	t.Run("face-up Puppet Master without the grant does not show the hidden identity", func(t *testing.T) {
		puppetMaster.AbilityState.CanViewOthersFaceDown = false
		t.Cleanup(func() { puppetMaster.AbilityState.CanViewOthersFaceDown = true })

		html := testhelpers.NewTemplateRenderer(t).Render(GameContent(room, puppetMaster)).GetHTML()
		assertRenderedID(t, html, "zone-roster")
		assertRenderedID(t, html, "player-row-"+hiddenPlayer.ID)
		assertRoleContentsNotRendered(t, html, hiddenPlayer.Role)
	})

	t.Run("another player view does not show the hidden identity", func(t *testing.T) {
		html := testhelpers.NewTemplateRenderer(t).Render(GameContent(room, otherViewer)).GetHTML()
		assertRenderedID(t, html, "zone-roster")
		assertRenderedID(t, html, "player-row-"+hiddenPlayer.ID)
		assertRoleContentsNotRendered(t, html, hiddenPlayer.Role)
	})

	t.Run("non-debug operator dashboard does not show the hidden identity", func(t *testing.T) {
		html := testhelpers.NewTemplateRenderer(t).Render(HostDashboardPlaying(room, host)).GetHTML()
		assertRenderedID(t, html, "operator-live-board")
		assertRenderedID(t, html, "operator-tile-"+hiddenPlayer.ID)
		assertRoleContentsNotRendered(t, html, hiddenPlayer.Role)
	})

	t.Run("public roster does not show the hidden identity", func(t *testing.T) {
		html := testhelpers.NewTemplateRenderer(t).Render(GameRosterZone(room, nil)).GetHTML()
		assertRenderedID(t, html, "zone-roster")
		assertRenderedID(t, html, "player-row-"+hiddenPlayer.ID)
		assertRoleContentsNotRendered(t, html, hiddenPlayer.Role)
	})
}

func assertRenderedID(t *testing.T, html, id string) {
	t.Helper()
	if !strings.Contains(html, `id="`+id+`"`) {
		t.Fatalf("rendered HTML missing id %q", id)
	}
}

func assertRoleContentsRendered(t *testing.T, html string, role *game.Card) {
	t.Helper()
	for _, expected := range []string{role.Name, role.Text} {
		if !strings.Contains(html, expected) {
			t.Fatalf("rendered view missing face-down identity content %q", expected)
		}
	}
}

func assertRoleContentsNotRendered(t *testing.T, html string, role *game.Card) {
	t.Helper()
	for _, privateContent := range []string{role.Name, role.Text} {
		if strings.Contains(html, privateContent) {
			t.Fatalf("rendered view leaked face-down identity content %q", privateContent)
		}
	}
}
