package pages

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestHomePage(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	t.Run("renders home page structure", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("Treacherest").
			AssertContains("A game of deception")
	})

	t.Run("has create room form", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertHasElement("form").
			AssertFormAction("/room/new").
			AssertContains(`method="POST"`).
			AssertHasElement("input").
			AssertContains(`name="playerName"`).
			AssertContains(`placeholder="Enter your name (optional)"`)
	})

	t.Run("has room code input", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertHasElementWithID("roomCode").
			AssertContains(`pattern="[A-Za-z0-9]{5}"`).
			AssertContains(`maxlength="5"`)
	})

	t.Run("has join room section", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains("Join Room").
			AssertContains("Join Existing Game").
			AssertHasElement("button")
	})

	t.Run("has proper styling", func(t *testing.T) {
		component := Home()

		// Style element and container class were removed
		renderer.Render(component).
			AssertContains("card").
			AssertContains("bg-base-100")
	})

	t.Run("has form submit handler", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains(`data-on:submit`).
			AssertContains(`evt.preventDefault()`)
	})

	t.Run("has two forms", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertElementCount("form", 2).
			AssertContains("Create New Game").
			AssertContains("Join Existing Game")
	})

	t.Run("has non-playing operator checkbox", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains("Run the table without playing").
			AssertContains(`name="hostOnly"`).
			AssertContains(`type="checkbox"`).
			AssertContains(`value="true"`).
			AssertNotContains("don't play, just manage")
	})

	t.Run("has rules mode choices", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains("Rules Mode").
			AssertContains(`type="radio"`).
			AssertContains(`name="rulesMode"`).
			AssertContains(`value="treachery"`).
			AssertContains(`value="coup"`).
			AssertContains("Treachery").
			AssertContains("Coup").
			AssertNotContains(`<select id="rulesMode"`).
			AssertNotContains(`<select name="rulesMode"`)
	})

	t.Run("puts join before create", func(t *testing.T) {
		component := Home()
		body := renderer.Render(component).GetHTML()

		joinIndex := strings.Index(body, "Join Existing Game")
		createIndex := strings.Index(body, "Create New Game")
		if joinIndex < 0 || createIndex < 0 {
			t.Fatalf("expected join and create sections in %q", body)
		}
		if joinIndex > createIndex {
			t.Fatalf("expected join section before create section")
		}
	})

	t.Run("uses rules mode radio cards with existing submitted values", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains(`type="radio"`).
			AssertContains(`name="rulesMode"`).
			AssertContains(`value="treachery"`).
			AssertContains(`value="coup"`).
			AssertNotContains(`<select id="rulesMode"`)
	})

	t.Run("renames non-playing creation option without changing submitted value", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains("Run the table without playing").
			AssertContains(`name="hostOnly"`).
			AssertContains(`value="true"`).
			AssertNotContains("Host mode (don't play, just manage)")
	})
}
