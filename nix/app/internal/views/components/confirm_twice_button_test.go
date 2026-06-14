package components

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestConfirmTwiceButtonUsesLocalConfirmationBeforeSubmit(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	html := renderer.Render(ConfirmTwiceButton("_confirmReveal", "Reveal role", "Confirm reveal", "@post('/room/ROOM1/reveal/p1')", "warning")).GetHTML()

	for _, expected := range []string{
		`data-signals="{_confirmReveal: false}"`,
		`$_confirmReveal ? @post(&#39;/room/ROOM1/reveal/p1&#39;) : ($_confirmReveal = true)`,
		`data-on:click__outside="$_confirmReveal = false"`,
		`data-on:interval__duration.3s="$_confirmReveal = false"`,
		`data-text="$_confirmReveal ? &#39;Confirm reveal&#39; : &#39;Reveal role&#39;"`,
		`aria-label="Reveal role"`,
		"Reveal role",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected %q in ConfirmTwiceButton HTML: %s", expected, html)
		}
	}

	if strings.Contains(html, `onclick=`) {
		t.Fatalf("ConfirmTwiceButton must not use browser onclick confirm handlers: %s", html)
	}
}
