package layouts

import (
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestBaseLayout(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	t.Run("renders with title", func(t *testing.T) {
		component := Base("Test Page Title")

		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("Test Page Title").
			AssertHasElement("html").
			AssertHasElement("head").
			AssertHasElement("body").
			AssertHasElement("title").
			AssertContains("<!doctype html>")
	})

	t.Run("includes viewport meta tag", func(t *testing.T) {
		component := Base("Mobile Test")

		renderer.Render(component).
			AssertContains(`name="viewport"`).
			AssertContains(`content="width=device-width, initial-scale=1.0"`)
	})

	t.Run("includes datastar script", func(t *testing.T) {
		component := Base("Datastar Test")

		renderer.Render(component).
			AssertHasElement("script").
			AssertContains("datastar")
	})

	t.Run("keeps state backup signal local-only", func(t *testing.T) {
		component := Base("Backup Signal Test")

		renderer.Render(component).
			AssertContains(`data-signals:_stateBackup__ifmissing`).
			AssertContains(`$_stateBackup && window.storeBackup($_stateBackup)`).
			AssertNotContains(`data-signals:stateBackup__ifmissing`).
			AssertNotContains(`$stateBackup && window.storeBackup($stateBackup)`)
	})

	t.Run("does not render debug UI by default", func(t *testing.T) {
		component := Base("Non Debug Test")

		renderer.Render(component).
			AssertNotContains(`id="debug-panel"`).
			AssertNotContains(`id="debug-clear"`).
			AssertNotContains("setupDebugPanel")
	})

	t.Run("includes custom CSS", func(t *testing.T) {
		component := Base("Style Test")

		// CSS is now in external stylesheet
		renderer.Render(component).
			AssertContains(`href="/static/css/output.css"`)
	})

	t.Run("uses treacherest as the default theme", func(t *testing.T) {
		component := Base("Theme Test")

		renderer.Render(component).
			AssertContains(`data-theme="treacherest"`).
			AssertContains(`localStorage.getItem('theme') || 'treacherest'`)
	})

	t.Run("loads display font and keeps mana font", func(t *testing.T) {
		component := Base("Font Test")

		renderer.Render(component).
			AssertContains(`https://fonts.googleapis.com`).
			AssertContains(`https://fonts.gstatic.com`).
			AssertContains(`Cormorant+Garamond:wght@500;600;700`).
			AssertContains(`mana-font@latest/css/mana.css`)
	})

	t.Run("keeps bespoke and legacy themes selectable", func(t *testing.T) {
		component := Base("Theme Switcher Test")

		renderer.Render(component).
			AssertContains(`value="treacherest"`).
			AssertContains(`value="treacherest-day"`).
			AssertContains(`value="dracula"`).
			AssertContains(`value="night"`).
			AssertContains(`value="nord"`)
	})

	t.Run("does not show room or operator chips on non-room pages", func(t *testing.T) {
		component := Base("Plain Page Test")

		renderer.Render(component).
			AssertNotContains(`id="app-room-code-chip"`).
			AssertNotContains(`id="app-operator-chip"`)
	})

	t.Run("shows room code chip on player room surfaces", func(t *testing.T) {
		room := &game.Room{Code: "ABCD1"}
		player := &game.Player{ID: "p1", Name: "Player One"}
		component := BaseWithDebugForPlayer("Lobby", false, room.Code, room, player)

		renderer.Render(component).
			AssertContains(`id="app-room-code-chip"`).
			AssertContains(`Room`).
			AssertContains(`ABCD1`).
			AssertNotContains(`id="app-operator-chip"`)
	})

	t.Run("shows operator chip on operator dashboard surfaces", func(t *testing.T) {
		room := &game.Room{Code: "WXYZ9"}
		component := BaseWithDebug("Host Dashboard", false, room.Code, room)

		renderer.Render(component).
			AssertContains(`id="app-room-code-chip"`).
			AssertContains(`WXYZ9`).
			AssertContains(`id="app-operator-chip"`).
			AssertContains(`OPERATOR`)
	})

	t.Run("has proper structure", func(t *testing.T) {
		component := Base("Structure Test")

		renderer.Render(component).
			AssertMatches(`(?s)<!doctype html>.*<html.*>.*<head>.*</head>.*<body.*>.*</body>.*</html>`).
			AssertElementCount("html", 1).
			AssertElementCount("head", 1).
			AssertElementCount("body", 1)

		// Ensure title is in head
		renderer.AssertMatches(`(?s)<head>.*<title>.*Structure Test.*</title>.*</head>`)
	})

	t.Run("escapes HTML in title", func(t *testing.T) {
		component := Base("<script>alert('xss')</script>")

		renderer.Render(component).
			AssertNotContains("<script>alert('xss')</script>").
			AssertContains("&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;")
	})
}

func TestDebugInsightsShowsTreacheryRoleAccents(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:      "DBG02",
		RulesMode: game.RulesModeTreachery,
		Players: map[string]*game.Player{
			"leader": {
				ID:       "leader",
				Name:     "Leader Player",
				Role:     debugInsightCard("The Queen of Light", game.RoleLeader),
				JoinedAt: time.Unix(1, 0),
			},
			"guardian": {
				ID:       "guardian",
				Name:     "Guardian Player",
				Role:     debugInsightCard("The Bodyguard", game.RoleGuardian),
				JoinedAt: time.Unix(2, 0),
			},
			"assassin": {
				ID:       "assassin",
				Name:     "Assassin Player",
				Role:     debugInsightCard("The Assassin", game.RoleAssassin),
				JoinedAt: time.Unix(3, 0),
			},
			"traitor": {
				ID:       "traitor",
				Name:     "Traitor Player",
				Role:     debugInsightCard("The Villain", game.RoleTraitor),
				JoinedAt: time.Unix(4, 0),
			},
		},
	}

	html := renderer.Render(DebugInsights(room, room.Code)).GetHTML()

	for _, expected := range []string{
		`id="debug-insight-player-leader"`,
		`data-debug-role-accent="gold"`,
		"debug-role-accent-badge",
		">gold</span>",
		`id="debug-insight-player-guardian"`,
		`data-debug-role-accent="blue"`,
		">blue</span>",
		`id="debug-insight-player-assassin"`,
		`data-debug-role-accent="black"`,
		">black</span>",
		`id="debug-insight-player-traitor"`,
		`data-debug-role-accent="red"`,
		">red</span>",
		"debug-role-accented",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected Treachery debug insights to contain %q in %s", expected, html)
		}
	}

	for _, forbidden := range []string{
		"badge-warning",
		"badge-info",
		"badge-neutral",
		"badge-error",
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("expected debug insights to use literal role accent styling, not %q in %s", forbidden, html)
		}
	}
}

func debugInsightCard(name string, role game.RoleType) *game.Card {
	return &game.Card{
		Name: name,
		Types: game.CardTypes{
			Subtype: string(role),
		},
	}
}
