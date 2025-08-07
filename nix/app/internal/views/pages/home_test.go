package pages

import (
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
			AssertContains("MTG Treacherest").
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
			AssertContains(`data-on-submit`).
			AssertContains(`evt.preventDefault()`)
	})

	t.Run("has two forms", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertElementCount("form", 2).
			AssertContains("Create New Game").
			AssertContains("Join Existing Game")
	})

	t.Run("has host mode checkbox", func(t *testing.T) {
		component := Home()

		renderer.Render(component).
			AssertContains("Host mode").
			AssertContains(`name="hostOnly"`).
			AssertContains(`type="checkbox"`).
			AssertContains("don't play, just manage")
	})
}
