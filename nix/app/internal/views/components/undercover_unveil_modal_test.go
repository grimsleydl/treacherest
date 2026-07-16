package components

import (
	"strings"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestUnveilButton_UndercoverOpensOwnerAdvisoryFlow(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	for _, cardID := range []int{17, 18} {
		room := &game.Room{Code: "UNDERCOVER"}
		player := &game.Player{
			ID:     "owner",
			Role:   &game.Card{ID: cardID, Name: "Undercover Guardian"},
			FaceUp: false,
		}

		html := renderer.Render(UnveilButton(room, player)).GetHTML()
		if !strings.Contains(html, "/room/UNDERCOVER/unveil-modal/owner") {
			t.Fatalf("card %d should open the owner unveil flow: %s", cardID, html)
		}
		if strings.Contains(html, "undercover-unveil-advisory") || strings.Contains(html, "Undercover condition") {
			t.Fatalf("card %d should show the advisory only after its owner opens the unveil flow: %s", cardID, html)
		}
		if strings.Contains(html, "/room/UNDERCOVER/unveil/owner") {
			t.Fatalf("card %d should not bypass its advisory flow with a direct unveil action: %s", cardID, html)
		}
	}
}
