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

func TestPrivyPanelCSSConstrainsRoleContentHeight(t *testing.T) {
	css := readGeneratedCSS(t)

	assertCSSRuleContains(t, css, ".privy",
		"display: flex;",
		"flex-direction: column;",
	)
	assertCSSRuleContains(t, css, ".privy-body",
		"height: min(70vh, 38rem);",
		"min-height: 22rem;",
		"overflow: hidden;",
	)
	assertCSSRuleContains(t, css, ".privy-content",
		"min-height: 0;",
		"min-width: 0;",
		"overflow-y: auto;",
		"overscroll-behavior: contain;",
	)
	assertCSSRuleContains(t, css, ".privy-veil",
		"min-height: 0;",
		"min-width: 0;",
		"overflow: hidden;",
	)
}

func TestPrivyPanelCSSAllowsShowingRoleSheet(t *testing.T) {
	css := readGeneratedCSS(t)

	assertCSSRuleContains(t, css, ".privy.showing",
		"max-width: min(100%, 42rem);",
		"overflow: visible;",
	)
	assertCSSRuleContains(t, css, ".privy.showing .privy-body",
		"height: auto;",
		"max-height: none;",
		"overflow: visible;",
	)
	assertCSSRuleContains(t, css, ".privy.showing .privy-content",
		"overflow: visible;",
		"overscroll-behavior: auto;",
	)
	assertCSSRuleContains(t, css, ".privy:has([data-privy-peek-button]:active)",
		"max-width: min(100%, 42rem);",
		"overflow: visible;",
	)
	assertCSSRuleContains(t, css, ".privy:has([data-privy-peek-button]:active) .privy-body",
		"height: auto;",
		"max-height: none;",
		"overflow: visible;",
	)
	assertCSSRuleContains(t, css, ".privy:has([data-privy-peek-button]:active) .privy-content",
		"overflow: visible;",
		"overscroll-behavior: auto;",
	)
}

func TestNoticeCardCSSUsesNeutralReadableSurface(t *testing.T) {
	css := readGeneratedCSS(t)

	for _, expected := range []string{
		`.notice-card.alert`,
		`background-color: var(--color-base-100);`,
		`color: var(--color-base-content);`,
		`--notice-card-accent: var(--color-info);`,
		`--notice-card-accent: var(--color-success);`,
		`--notice-card-accent: var(--color-warning);`,
		`--notice-card-accent: var(--color-error);`,
	} {
		if !strings.Contains(css, expected) {
			t.Fatalf("expected generated CSS to contain %q", expected)
		}
	}
}

func TestDebugRoleAccentCSSColorCodesRows(t *testing.T) {
	css := readGeneratedCSS(t)

	assertCSSRuleContains(t, css, ".debug-role-accented",
		"overflow: hidden;",
		"padding-inline-start: 2rem !important;",
		"position: relative;",
	)
	assertCSSRuleContains(t, css, ".debug-role-accented::before",
		"background: var(--debug-role-accent-color, var(--color-base-300));",
		`content: "";`,
		"inset-block: 0;",
		"inset-inline-start: 0;",
		"position: absolute;",
		"width: 0.5rem;",
	)
	assertCSSRuleContains(t, css, ".debug-role-accent-label",
		"color: var(--color-base-content);",
		"background-color: transparent;",
	)

	for _, expected := range []string{
		`--debug-role-accent-color: #d4af37;`,
		`--debug-role-accent-color: #1d4ed8;`,
		`--debug-role-accent-color: #111827;`,
		`--debug-role-accent-color: #b91c1c;`,
		`--debug-role-accent-color: #15803d;`,
		`--debug-role-accent-color: #4b5563;`,
	} {
		if !strings.Contains(css, expected) {
			t.Fatalf("expected generated CSS to contain literal debug role stripe color %q", expected)
		}
	}
}

func TestRoleCountChevronCSSRotatesWhenOpen(t *testing.T) {
	css := readGeneratedCSS(t)

	for _, expected := range []string{
		`.role-count-disclosure > input:checked + .role-count-title .role-count-chevron-icon`,
		`transform: rotate(180deg);`,
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

func assertCSSRuleContains(t *testing.T, css string, selector string, declarations ...string) {
	t.Helper()

	block := cssRuleBlock(t, css, selector)
	for _, declaration := range declarations {
		if !strings.Contains(block, declaration) {
			t.Fatalf("expected CSS rule %q to contain %q in block:\n%s", selector, declaration, block)
		}
	}
}

func cssRuleBlock(t *testing.T, css string, selector string) string {
	t.Helper()

	start := strings.Index(css, selector+" {")
	if start < 0 {
		t.Fatalf("missing CSS rule %q", selector)
	}
	openOffset := strings.Index(css[start:], "{")
	if openOffset < 0 {
		t.Fatalf("missing opening brace for CSS rule %q", selector)
	}

	depth := 0
	for i := start + openOffset; i < len(css); i++ {
		switch css[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return css[start : i+1]
			}
		}
	}
	t.Fatalf("missing closing brace for CSS rule %q", selector)
	return ""
}
