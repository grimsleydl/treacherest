package components

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestPrivyPanelControlsAreKeyboardAccessibleAndLocalOnly(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	html := renderer.Render(PrivyPanel()).GetHTML()

	for _, expected := range []string{
		`id="zone-privy"`,
		`data-signals="{_peek: false, _privyOpen: false}"`,
		`data-on:interval__duration.30s`,
		`data-on:visibilitychange__window`,
		`type="button"`,
		`min-h-11`,
		`Hold to peek`,
		`Open until concealed`,
		`aria-pressed="false"`,
		`data-attr:aria-pressed="$_privyOpen ? 'true' : 'false'"`,
		`Conceal now`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected Privy Panel HTML to contain %q in %s", expected, html)
		}
	}

	for _, forbidden := range []string{
		`@get(`,
		`@post(`,
		`mousemove`,
		`requestAnimationFrame`,
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("Privy Panel should not rely on %q: %s", forbidden, html)
		}
	}
}
