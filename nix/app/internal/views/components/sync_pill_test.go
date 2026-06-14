package components

import (
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestSyncPill(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	for _, state := range []string{"live", "reconnecting", "stale"} {
		t.Run(state, func(t *testing.T) {
			html := renderer.Render(SyncPill(state)).GetHTML()

			if !strings.Contains(html, `id="sync-pill"`) {
				t.Fatalf("expected sync pill id in %s", html)
			}
			if !strings.Contains(html, `role="status"`) {
				t.Fatalf("expected sync pill to use role=status in %s", html)
			}
			if !strings.Contains(html, state) {
				t.Fatalf("expected sync pill state %q in %s", state, html)
			}
		})
	}
}
