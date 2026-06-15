package components

import (
	"strings"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestRoleCard(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	card := &game.Card{
		ID:          7,
		Name:        "Test Guardian",
		Type:        "Creature - Guardian",
		Rarity:      "Rare",
		Text:        "Long role rules text appears after the goal.|Second rules line.",
		URI:         "https://mtgtreachery.net/rules/oracle/?card=test-guardian",
		Artist:      "Test Artist",
		Base64Image: "data:image/jpeg;base64,test",
		Rulings: []string{
			"Role Goal: Protect the Leader.",
			"Long ruling detail.",
		},
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Guardian",
		},
	}

	t.Run("uses hero while own role is not publicly revealed", func(t *testing.T) {
		html := renderer.Render(RoleCard(card, false, false)).GetHTML()

		if !strings.Contains(html, "role-card-hero") {
			t.Fatalf("expected hero role card while face down and unrevealed: %s", html)
		}
		if strings.Contains(html, "role-card-compact") {
			t.Fatalf("did not expect compact role card while face down and unrevealed: %s", html)
		}
	})

	t.Run("uses compact once own role is public or face up", func(t *testing.T) {
		for _, tc := range []struct {
			name         string
			faceUp       bool
			roleRevealed bool
		}{
			{name: "face up", faceUp: true, roleRevealed: false},
			{name: "role revealed", faceUp: false, roleRevealed: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				html := renderer.Render(RoleCard(card, tc.faceUp, tc.roleRevealed)).GetHTML()
				if !strings.Contains(html, "role-card-compact") {
					t.Fatalf("expected compact role card: %s", html)
				}
			})
		}
	})

	t.Run("prioritizes goal before image and rules in private hero card", func(t *testing.T) {
		html := renderer.Render(RoleCard(card, false, false)).GetHTML()

		nameIndex := strings.Index(html, "Test Guardian")
		winIndex := strings.Index(html, "Win Condition")
		textIndex := strings.Index(html, "Long role rules text")
		imageIndex := strings.Index(html, "<img")
		rulingsDisclosureIndex := strings.Index(html, "Rulings")
		rulingIndex := strings.Index(html, "Long ruling detail")

		for label, index := range map[string]int{
			"role name":          nameIndex,
			"win condition":      winIndex,
			"rules text":         textIndex,
			"image":              imageIndex,
			"rulings disclosure": rulingsDisclosureIndex,
			"ruling detail":      rulingIndex,
		} {
			if index < 0 {
				t.Fatalf("expected %s in role card HTML: %s", label, html)
			}
		}
		if strings.Contains(html, "Full card image") {
			t.Fatalf("private hero role card should render image inline without second disclosure: %s", html)
		}
		if !(nameIndex < winIndex && winIndex < imageIndex && imageIndex < textIndex) {
			t.Fatalf("expected role name, then win condition, then image, then rules text: %s", html)
		}
		if !(rulingsDisclosureIndex < rulingIndex) {
			t.Fatalf("expected rulings to be behind their disclosure: %s", html)
		}
		if !strings.Contains(html, `src="data:image/jpeg;base64,test"`) {
			t.Fatalf("expected role card image to render its data URI source: %s", html)
		}
		if strings.Contains(html, `about:invalid`) {
			t.Fatalf("role card image source should not be sanitized to an invalid URL: %s", html)
		}
	})

	t.Run("keeps full image behind disclosure in compact public card", func(t *testing.T) {
		html := renderer.Render(RoleCard(card, true, true)).GetHTML()

		imageDisclosureIndex := strings.Index(html, "Full card image")
		imageIndex := strings.Index(html, "<img")
		if imageDisclosureIndex < 0 || imageIndex < 0 {
			t.Fatalf("expected compact role card image disclosure and image: %s", html)
		}
		if !(imageDisclosureIndex < imageIndex) {
			t.Fatalf("expected compact role card image to stay behind disclosure: %s", html)
		}
	})

	t.Run("renders public role surface with inline image", func(t *testing.T) {
		html := renderer.Render(RoleCardPublic(card)).GetHTML()

		for _, expected := range []string{
			"role-card-public",
			"Public role",
			"Test Guardian",
			"<img",
			`src="data:image/jpeg;base64,test"`,
			`href="https://mtgtreachery.net/rules/oracle/?card=test-guardian"`,
			`title="View on MTG Treachery Oracle"`,
			`rel="noopener noreferrer"`,
			`<svg`,
			`class="w-5 h-5"`,
			`M13.5 6H5.25`,
		} {
			if !strings.Contains(html, expected) {
				t.Fatalf("expected public role card to contain %q: %s", expected, html)
			}
		}
		if strings.Contains(html, "Full card image") {
			t.Fatalf("public role card should render image inline without second disclosure: %s", html)
		}
	})

	t.Run("omits rulings that only repeat the win condition", func(t *testing.T) {
		duplicateRulingCard := &game.Card{
			Name:        "Blue Knight",
			Text:        "Protects the King.",
			Type:        "Coup Role",
			Rarity:      "Coup",
			Base64Image: "data:image/jpeg;base64,test",
			Types: game.CardTypes{
				Supertype: "Coup",
				Subtype:   "Blue Knight",
			},
			Rulings: []string{
				"Role Goal: Win with the King. Lose when the King loses.",
				"Win Condition: Win with the King. Lose when the King loses.",
			},
		}

		html := renderer.Render(RoleCard(duplicateRulingCard, false, false)).GetHTML()
		if strings.Contains(html, "Rulings") {
			t.Fatalf("expected duplicate-only rulings to be omitted: %s", html)
		}
		if got := strings.Count(html, "Win with the King. Lose when the King loses."); got != 1 {
			t.Fatalf("expected win condition text exactly once, got %d in %s", got, html)
		}
	})

	t.Run("renders green strict eligibility as structured win condition bullets", func(t *testing.T) {
		greenCard := &game.Card{
			Name:   "Green Knight",
			Type:   "Coup Role",
			Rarity: "Coup",
			Text:   "Opportunist with conditional shared victories.",
			Types: game.CardTypes{
				Supertype: "Coup",
				Subtype:   "Green Knight",
			},
			Rulings: []string{
				"Role Goal: " + game.CoupStrictGreenWinCondition,
				"Win Condition: " + game.CoupStrictGreenWinCondition,
			},
		}

		for _, tc := range []struct {
			name string
			html string
		}{
			{name: "private hero", html: renderer.Render(RoleCard(greenCard, false, false)).GetHTML()},
			{name: "compact", html: renderer.Render(RoleCard(greenCard, true, true)).GetHTML()},
			{name: "public", html: renderer.Render(RoleCardPublic(greenCard)).GetHTML()},
		} {
			t.Run(tc.name, func(t *testing.T) {
				for _, expected := range []string{
					"<ul",
					"<li",
					"May share a King-side victory while alive",
					"no Blue Knights are alive or Inquisition has succeeded",
					"May share a Red-side victory while alive",
					"already dead before King Fall",
					"Does not share Black or Wasteland victories",
					"King falls do not make Green eligible",
				} {
					if !strings.Contains(tc.html, expected) {
						t.Fatalf("expected Green role card to contain %q: %s", expected, tc.html)
					}
				}
				if strings.Contains(tc.html, "selected Green rules") {
					t.Fatalf("expected Green role card to omit vague selected-rules copy: %s", tc.html)
				}
				if strings.Contains(tc.html, "Broad Amnesty") || strings.Contains(tc.html, "Simple Green") {
					t.Fatalf("expected Green role card to omit inactive variants: %s", tc.html)
				}
				if strings.Contains(tc.html, "Rulings") {
					t.Fatalf("expected duplicate-only Green rulings to be omitted: %s", tc.html)
				}
			})
		}
	})
}

func TestRoleCardRulingsRenderBulletedRows(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	html := renderer.Render(RoleCardRulings([]string{
		"First ruling.",
		"Second ruling with {X}.",
	})).GetHTML()

	if got := strings.Count(html, ">•</span>"); got != 2 {
		t.Fatalf("expected one bullet marker per ruling, got %d in %s", got, html)
	}
	for _, expected := range []string{
		`class="flex gap-2"`,
		`class="flex-1 break-words"`,
		"First ruling.",
		`class="ms ms-x ms-cost"`,
		"Second ruling with ",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected bulleted ruling markup %q in %s", expected, html)
		}
	}
	if strings.Contains(html, "{X}") {
		t.Fatalf("expected mana notation in rulings to render as icon, got raw text in %s", html)
	}
}
