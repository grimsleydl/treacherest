package components

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestStateChipIncludesTextLabel(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	for _, tc := range []struct {
		kind  string
		label string
	}{
		{kind: "you", label: "You"},
		{kind: "operator", label: "Operator"},
		{kind: "leader", label: "Leader"},
		{kind: "revealed", label: "Revealed: Test Guardian"},
		{kind: "facedown", label: "Face Down"},
		{kind: "eliminated", label: "Eliminated"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			html := renderer.Render(StateChip(tc.kind, tc.label)).GetHTML()
			if !strings.Contains(html, tc.label) {
				t.Fatalf("expected chip label %q in %s", tc.label, html)
			}
			if !strings.Contains(html, "state-chip") {
				t.Fatalf("expected state-chip class in %s", html)
			}
		})
	}
}
