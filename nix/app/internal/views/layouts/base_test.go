package layouts

import (
	"testing"
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
