package handlers

import (
	"testing"

	"github.com/go-rod/rod/lib/launcher"
)

// skipIfNoBrowser skips the test if Chrome/Chromium is not available
func skipIfNoBrowser(t *testing.T) {
	t.Helper()

	// Try to find Chrome binary
	path, exists := launcher.LookPath()
	if !exists {
		t.Skip("Skipping browser test: Chrome/Chromium not available")
	}
	t.Logf("Found browser at: %s", path)
}
