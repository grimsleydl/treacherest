package components

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestCountdownDisplayWithMessage_RedesignTimerSemantics(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	html := renderer.Render(CountdownDisplayWithMessage(3, "Revealing roles in...")).GetHTML()

	for _, expected := range []string{
		`role="timer"`,
		"font-display",
		"Keep your screen to yourself",
		`data-attr:style="'--value:' + $countdown"`,
		"Revealing roles in...",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected countdown HTML to contain %q in %s", expected, html)
		}
	}
}
