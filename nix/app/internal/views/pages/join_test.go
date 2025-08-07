package pages

import (
	"testing"
	"treacherest/internal/testhelpers"
)

func TestJoinPage(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	t.Run("renders join page structure", func(t *testing.T) {
		roomCode := "ABC12"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("Join Room " + roomCode).
			AssertHasClass("container")
	})

	t.Run("has join form with correct structure", func(t *testing.T) {
		roomCode := "XYZ99"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertHasElement("form").
			AssertContains(`method="GET"`).
			AssertHasElement("input").
			AssertContains(`name="name"`).
			AssertContains(`placeholder="Enter your name"`).
			AssertContains(`required`)
	})

	t.Run("displays error message when provided", func(t *testing.T) {
		roomCode := "ABC12"
		errorMsg := "Room is full"
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertContains(errorMsg).
			AssertHasClass("error-message")
	})

	t.Run("does not show error section when no error", func(t *testing.T) {
		roomCode := "ABC12"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertNotContains("error-message")
	})

	t.Run("has submit button", func(t *testing.T) {
		roomCode := "TEST1"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertHasElement("button").
			AssertContains(`type="submit"`).
			AssertContains("Join")
	})

	t.Run("input field has proper attributes", func(t *testing.T) {
		roomCode := "ROOM1"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertContains(`type="text"`).
			AssertContains(`autofocus`)
	})

	t.Run("has datastar attributes for real-time updates", func(t *testing.T) {
		roomCode := "LIVE1"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertContains(`data-store="{}"`)
	})

	t.Run("room code is properly displayed", func(t *testing.T) {
		roomCode := "GAME7"
		errorMsg := ""
		component := Join(roomCode, errorMsg)
		
		renderer.Render(component).
			AssertContains("<h1>Join Room " + roomCode + "</h1>")
	})
}