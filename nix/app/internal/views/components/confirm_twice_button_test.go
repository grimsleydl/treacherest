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
		`data-text="$_confirmReveal ? &#34;Confirm reveal&#34; : &#34;Reveal role&#34;"`,
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

func TestConfirmTwiceButtonQuotesLabelsForDatastarExpressions(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	html := renderer.Render(ConfirmTwiceButton("_confirmEliminated", "I've Been Eliminated", "Confirm Eliminated", "@post('/room/ROOM1/player/p1/eliminate')", "error")).GetHTML()

	for _, expected := range []string{
		`data-text="$_confirmEliminated ? &#34;Confirm Eliminated&#34; : &#34;I&#39;ve Been Eliminated&#34;"`,
		`data-attr:aria-label="$_confirmEliminated ? &#34;Confirm Eliminated&#34; : &#34;I&#39;ve Been Eliminated&#34;"`,
		`aria-label="I&#39;ve Been Eliminated"`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected %q in ConfirmTwiceButton HTML: %s", expected, html)
		}
	}

	for _, forbidden := range []string{
		`: &#39;I&#39;ve Been Eliminated&#39;`,
		`: 'I've Been Eliminated'`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("ConfirmTwiceButton should not emit malformed single-quoted Datastar labels %q: %s", forbidden, html)
		}
	}
}
