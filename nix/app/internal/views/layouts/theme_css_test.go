package layouts

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"treacherest/internal/testhelpers"
)

func TestSelectableThemesHaveGeneratedColorTokens(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	html := renderer.Render(Base("Theme Token Test")).GetHTML()
	css := readGeneratedCSS(t)

	selectableThemes := []string{
		"treacherest",
		"treacherest-day",
		"light",
		"dark",
		"cupcake",
		"bumblebee",
		"emerald",
		"corporate",
		"synthwave",
		"retro",
		"cyberpunk",
		"valentine",
		"halloween",
		"garden",
		"forest",
		"aqua",
		"lofi",
		"pastel",
		"fantasy",
		"wireframe",
		"black",
		"luxury",
		"dracula",
		"cmyk",
		"autumn",
		"business",
		"acid",
		"lemonade",
		"night",
		"coffee",
		"winter",
		"dim",
		"nord",
		"sunset",
		"caramellatte",
		"abyss",
		"silk",
	}

	for _, theme := range selectableThemes {
		if !strings.Contains(html, `value="`+theme+`"`) {
			t.Fatalf("theme switcher exposes no option for %q", theme)
		}
		if !themeBlockHasColorTokens(css, theme) {
			t.Fatalf("selectable theme %q has no generated DaisyUI color token block", theme)
		}
	}
}

func TestLegacyThemesDoNotInheritTreacherestPalette(t *testing.T) {
	css := readGeneratedCSS(t)
	treacherestPrimary := themeToken(t, css, "treacherest", "--color-primary")

	for _, theme := range []string{"dracula", "night", "cupcake", "winter"} {
		primary := themeToken(t, css, theme, "--color-primary")
		if primary == treacherestPrimary {
			t.Fatalf("legacy theme %q primary token unexpectedly matches treacherest (%s)", theme, primary)
		}
	}
}

func TestInteractionCSSIncludesPrivyPeekAndDebugMinimizeRules(t *testing.T) {
	css := readGeneratedCSS(t)

	for _, expected := range []string{
		`.privy:has([data-privy-peek-button]:active) .privy-content`,
		`.privy:has([data-privy-peek-button]:active) .privy-veil`,
		`#debug-control-surface.debug-panel-minimized`,
		`inset-block-start: auto;`,
		`inset-block-end: 1rem;`,
		`width: min(18rem, calc(100vw - 2rem));`,
		`top: auto;`,
		`height: auto;`,
	} {
		if !strings.Contains(css, expected) {
			t.Fatalf("expected generated CSS to contain %q", expected)
		}
	}
}

func readGeneratedCSS(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	path := filepath.Join(filepath.Dir(filename), "..", "..", "..", "static", "css", "output.css")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated CSS %s: %v", path, err)
	}
	return string(content)
}

func themeBlockHasColorTokens(css string, theme string) bool {
	block, ok := themeBlock(css, theme)
	return ok &&
		strings.Contains(block, "--color-base-100:") &&
		strings.Contains(block, "--color-base-content:") &&
		strings.Contains(block, "--color-primary:")
}

func themeToken(t *testing.T, css string, theme string, token string) string {
	t.Helper()

	block, ok := themeBlock(css, theme)
	if !ok {
		t.Fatalf("missing generated CSS block for theme %q", theme)
	}
	re := regexp.MustCompile(regexp.QuoteMeta(token) + `:\s*([^;]+);`)
	match := re.FindStringSubmatch(block)
	if len(match) != 2 {
		t.Fatalf("missing token %s for theme %q in block %q", token, theme, block)
	}
	return strings.TrimSpace(match[1])
}

func themeBlock(css string, theme string) (string, bool) {
	selectorPattern := `\[data-theme="?` + regexp.QuoteMeta(theme) + `"?\]\s*\{`
	re := regexp.MustCompile(selectorPattern)
	for _, loc := range re.FindAllStringIndex(css, -1) {
		blockStart := loc[1]
		blockEnd := strings.Index(css[blockStart:], "}")
		if blockEnd < 0 {
			continue
		}
		block := css[blockStart : blockStart+blockEnd]
		if strings.Contains(block, "--color-primary:") {
			return block, true
		}
	}
	return "", false
}
