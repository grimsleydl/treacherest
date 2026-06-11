package layouts

import (
	"testing"
	"treacherest/internal/testhelpers"
)

func TestBaseLayout(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	t.Run("renders with title", func(t *testing.T) {
		component := Base("Test Page Title")

		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("Test Page Title").
			AssertHasElement("html").
			AssertHasElement("head").
			AssertHasElement("body").
			AssertHasElement("title").
			AssertContains("<!doctype html>")
	})

	t.Run("includes viewport meta tag", func(t *testing.T) {
		component := Base("Mobile Test")

		renderer.Render(component).
			AssertContains(`name="viewport"`).
			AssertContains(`content="width=device-width, initial-scale=1.0"`)
	})

	t.Run("includes datastar script", func(t *testing.T) {
		component := Base("Datastar Test")

		renderer.Render(component).
			AssertHasElement("script").
			AssertContains("datastar")
	})

	t.Run("keeps state backup signal local-only", func(t *testing.T) {
		component := Base("Backup Signal Test")

		renderer.Render(component).
			AssertContains(`data-signals:_stateBackup__ifmissing`).
			AssertContains(`$_stateBackup && window.storeBackup($_stateBackup)`).
			AssertNotContains(`data-signals:stateBackup__ifmissing`).
			AssertNotContains(`$stateBackup && window.storeBackup($stateBackup)`)
	})

	t.Run("does not render debug UI by default", func(t *testing.T) {
		component := Base("Non Debug Test")

		renderer.Render(component).
			AssertNotContains(`id="debug-panel"`).
			AssertNotContains(`id="debug-clear"`).
			AssertNotContains("setupDebugPanel")
	})

	t.Run("includes custom CSS", func(t *testing.T) {
		component := Base("Style Test")

		// CSS is now in external stylesheet
		renderer.Render(component).
			AssertContains(`href="/static/css/output.css"`)
	})

	t.Run("has proper structure", func(t *testing.T) {
		component := Base("Structure Test")

		renderer.Render(component).
			AssertMatches(`(?s)<!doctype html>.*<html.*>.*<head>.*</head>.*<body.*>.*</body>.*</html>`).
			AssertElementCount("html", 1).
			AssertElementCount("head", 1).
			AssertElementCount("body", 1)

		// Ensure title is in head
		renderer.AssertMatches(`(?s)<head>.*<title>.*Structure Test.*</title>.*</head>`)
	})

	t.Run("escapes HTML in title", func(t *testing.T) {
		component := Base("<script>alert('xss')</script>")

		renderer.Render(component).
			AssertNotContains("<script>alert('xss')</script>").
			AssertContains("&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;")
	})
}
