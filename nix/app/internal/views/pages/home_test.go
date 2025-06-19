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
			AssertContains("A game of deception").
			AssertHasClass("container")
	})

	t.Run("has create room form", func(t *testing.T) {
		component := Home()
		
		renderer.Render(component).
			AssertHasElement("form").
			AssertFormAction("/room/new").
			AssertContains(`method="POST"`).
			AssertHasElement("input").
			AssertContains(`name="playerName"`).
			AssertContains(`placeholder="Enter your name"`)
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
		
		renderer.Render(component).
			AssertHasClass("container").
			AssertHasElement("style")
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
			AssertContains("Create Room").
			AssertContains("Join Room")
	})
}