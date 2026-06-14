package components

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestNoticeCard(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	html := renderer.Render(NoticeCard("error", "You are out of the game")).GetHTML()

	for _, expected := range []string{
		"notice-card",
		"alert-error",
		"You are out of the game",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected %q in NoticeCard HTML: %s", expected, html)
		}
	}
}
