package pages

import (
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

// Helper functions to create mock cards for testing
func mockLeaderCard() *game.Card {
	return &game.Card{
		ID:   1,
		Name: "Test Leader",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Leader",
		},
		Text:        "Test Leader Card",
		Type:        "Creature — Leader",
		Rarity:      "Rare",
		Artist:      "Test Artist",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func mockGuardianCard() *game.Card {
	return &game.Card{
		ID:   2,
		Name: "Test Guardian",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Guardian",
		},
		Text:        "Test Guardian Card",
		Type:        "Creature — Guardian",
		Rarity:      "Common",
		Artist:      "Test Artist",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func mockAssassinCard() *game.Card {
	return &game.Card{
		ID:   3,
		Name: "Test Assassin",
		Types: game.CardTypes{
			Supertype: "Creature",
			Subtype:   "Assassin",
		},
		Text:        "Test Assassin Card",
		Type:        "Creature — Assassin",
		Rarity:      "Uncommon",
		Artist:      "Test Artist",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func mockCoupCard(id int, name string) *game.Card {
	return &game.Card{
		ID:   id,
		Name: name,
		Types: game.CardTypes{
			Supertype: "Coup",
			Subtype:   name,
		},
		Text:        "Test " + name + " Card",
		Type:        "Coup Role",
		Rarity:      "Coup",
		Artist:      "Test Artist",
		Base64Image: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBD",
	}
}

func TestGamePage(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	// Create test data
	room := &game.Room{
		Code:       "GAME1",
		State:      game.StatePlaying,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}

	player := &game.Player{
		ID:   "p1",
		Name: "Test Player",
		Role: mockGuardianCard(),
	}

	room.Players[player.ID] = player

	t.Run("renders game page structure", func(t *testing.T) {
		component := GamePage(room, player)

		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("Test Guardian").
			AssertHasElementWithID("game-container").
			AssertContains(`data-init="@get(&#39;/sse/game/GAME1&#39;)"`)
	})

	t.Run("shows player role", func(t *testing.T) {
		component := GamePage(room, player)

		renderer.Render(component).
			AssertContains("Test Guardian").
			AssertContains("The Guardians help the Leader, they win or lose with them.")
	})

	t.Run("shows countdown state", func(t *testing.T) {
		room.State = game.StateCountdown
		room.CountdownRemaining = 3

		component := GamePage(room, player)

		renderer.Render(component).
			AssertContains("Revealing roles in...").
			AssertContains(`data-attr:style="'--value:' + $countdown"`)
	})

	t.Run("shows player list", func(t *testing.T) {
		// Reset room state
		room.State = game.StatePlaying

		// Add more players
		revealedPlayer := &game.Player{
			ID:           "p2",
			Name:         "Revealed Player",
			RoleRevealed: true,
			Role:         mockGuardianCard(),
		}
		room.Players[revealedPlayer.ID] = revealedPlayer

		component := GameBody(room, player)

		renderer.Render(component).
			AssertContains("Test Player").
			AssertContains("Revealed Player").
			AssertContains("Guardian")
	})

	t.Run("renders stable player view zones in order", func(t *testing.T) {
		component := GameBody(room, player)
		html := renderer.Render(component).GetHTML()

		zoneIDs := []string{
			`id="zone-status"`,
			`id="zone-privy"`,
			`id="zone-notices"`,
			`id="zone-actions"`,
			`id="zone-roster"`,
		}

		lastIndex := -1
		for _, zoneID := range zoneIDs {
			index := strings.Index(html, zoneID)
			if index < 0 {
				t.Fatalf("expected %s in rendered game body", zoneID)
			}
			if index < lastIndex {
				t.Fatalf("expected %s to render after previous stable zone", zoneID)
			}
			lastIndex = index
		}

		if !strings.Contains(html, `id="zone-notices" aria-live="polite"`) &&
			!strings.Contains(html, `aria-live="polite" id="zone-notices"`) {
			t.Fatalf("expected zone-notices to be an aria-live polite region")
		}
	})

	t.Run("renders sync pill in status zone", func(t *testing.T) {
		component := GameBody(room, player)
		html := renderer.Render(component).GetHTML()

		statusIndex := strings.Index(html, `id="zone-status"`)
		syncIndex := strings.Index(html, `id="sync-pill"`)
		if statusIndex < 0 || syncIndex < 0 {
			t.Fatalf("expected status zone and sync pill in rendered game body")
		}
		if syncIndex < statusIndex {
			t.Fatalf("expected sync pill to render inside or after the status zone starts")
		}
		if !strings.Contains(html, `role="status"`) {
			t.Fatalf("expected sync pill to use role=status")
		}
		if !strings.Contains(html, "live") {
			t.Fatalf("expected default live sync state")
		}
	})

	t.Run("playing room operator gets low emphasis dashboard navigation only", func(t *testing.T) {
		operatorRoom := &game.Room{
			Code:              "OPNAV",
			State:             game.StatePlaying,
			Players:           make(map[string]*game.Player),
			MaxPlayers:        5,
			OperatorSessionID: "session-operator",
		}
		operator := &game.Player{
			ID:        "operator",
			Name:      "Playing Operator",
			Role:      mockCoupCard(1001, "King"),
			SessionID: "session-operator",
			IsHost:    false,
		}
		nonOperator := &game.Player{
			ID:        "player",
			Name:      "Ordinary Player",
			Role:      mockCoupCard(1002, "Blue Knight"),
			SessionID: "session-player",
		}
		operatorRoom.Players[operator.ID] = operator
		operatorRoom.Players[nonOperator.ID] = nonOperator

		operatorHTML := renderer.Render(GameBody(operatorRoom, operator)).GetHTML()
		if !strings.Contains(operatorHTML, `id="operator-dashboard-link"`) {
			t.Fatalf("expected playing room operator to receive dashboard navigation, got %q", operatorHTML)
		}
		if !strings.Contains(operatorHTML, `href="/room/OPNAV/operator"`) {
			t.Fatalf("expected dashboard navigation to use explicit operator route, got %q", operatorHTML)
		}
		for _, forbidden := range []string{
			`id="host-dashboard-container"`,
			`id="operator-start-controls"`,
			`id="coup-role-counts-form"`,
		} {
			if strings.Contains(operatorHTML, forbidden) {
				t.Fatalf("ordinary player view should not inline management control %q", forbidden)
			}
		}

		nonOperatorHTML := renderer.Render(GameBody(operatorRoom, nonOperator)).GetHTML()
		if strings.Contains(nonOperatorHTML, `id="operator-dashboard-link"`) {
			t.Fatalf("non-operator player should not receive dashboard navigation, got %q", nonOperatorHTML)
		}
	})
}

func TestGamePageWithDebug_OperatorDebugSurface(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "DBGPG",
		State:      game.StatePlaying,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}
	player := &game.Player{
		ID:   "p1",
		Name: "Test Player",
		Role: mockGuardianCard(),
	}
	room.Players[player.ID] = player

	renderer.Render(GamePageWithDebug(room, player, true)).
		AssertContains(`id="debug-control-surface"`).
		AssertContains("Debug Control Surface").
		AssertContains(`id="debug-panel-toggle"`).
		AssertContains(`id="debug-clear"`).
		AssertContains("Debug Insights").
		AssertContains("Start with Debug Players").
		AssertContains("View As Player").
		AssertNotContains("Player Perspective: Test Player")

	renderer.Render(GamePageWithDebug(room, player, false)).
		AssertNotContains(`id="debug-control-surface"`).
		AssertNotContains("Debug Control Surface").
		AssertNotContains("Player Perspective: Test Player")
}

func TestGameBody(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	room := &game.Room{
		Code:       "BODY1",
		State:      game.StatePlaying,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}

	player := &game.Player{
		ID:   "p1",
		Name: "Assassin Player",
		Role: mockAssassinCard(),
	}

	room.Players[player.ID] = player

	t.Run("renders game body fragment", func(t *testing.T) {
		component := GameBody(room, player)

		renderer.Render(component).
			AssertNotEmpty().
			AssertHasElementWithID("game-container").
			AssertContains("Test Assassin")
	})

	t.Run("shows win condition", func(t *testing.T) {
		component := GameBody(room, player)

		renderer.Render(component).
			AssertContains("Win Condition:").
			AssertContains("The Assassins win if the Leader is eliminated.")
	})

	t.Run("shows leader when revealed", func(t *testing.T) {
		leaderPlayer := &game.Player{
			ID:           "p2",
			Name:         "Leader Player",
			Role:         mockLeaderCard(),
			RoleRevealed: true,
		}
		room.Players[leaderPlayer.ID] = leaderPlayer
		room.LeaderRevealed = true

		component := GameBody(room, player)

		renderer.Render(component).
			AssertContains("Leader: Leader Player")
	})

	t.Run("hides roles when not revealed", func(t *testing.T) {
		hiddenPlayer := &game.Player{
			ID:           "p3",
			Name:         "Hidden Player",
			Role:         mockGuardianCard(),
			RoleRevealed: false,
		}
		room.Players[hiddenPlayer.ID] = hiddenPlayer

		component := GameBody(room, player)

		renderer.Render(component).
			AssertContains("Hidden Player").
			// Only check that the specific player's role isn't shown
			// since other guardians might be revealed
			AssertNotContains("<span>Hidden Player</span> <span class=\"badge badge-sm\">Guardian</span>")
	})

	t.Run("shows role class styling", func(t *testing.T) {
		component := GameBody(room, player)

		// Role card styling is done with border colors now
		renderer.Render(component).
			AssertContains("card").
			AssertContains("border-error") // Assassin border
	})

	t.Run("renders private role in concealed privy panel with local controls", func(t *testing.T) {
		component := GameBody(room, player)
		html := renderer.Render(component).GetHTML()
		privyHTML := extractBetween(t, html, `id="zone-privy"`, `id="zone-notices"`)

		for _, expected := range []string{
			`class="privy`,
			"PRIVATE",
			"ONLY YOU",
			"CONCEALED",
			"Hold to peek",
			"Open until concealed",
			"Conceal now",
			`data-on:pointerdown`,
			`data-on:pointerup`,
			`data-on:pointerleave`,
			`data-on:interval__duration.30s`,
			`data-on:visibilitychange__window`,
			`_peek`,
			`_privyOpen`,
		} {
			if !strings.Contains(privyHTML, expected) {
				t.Fatalf("expected privy panel HTML to contain %q in %s", expected, privyHTML)
			}
		}

		for _, forbidden := range []string{
			`@get(`,
			`@post(`,
			`mousemove`,
			`requestAnimationFrame`,
		} {
			if strings.Contains(privyHTML, forbidden) {
				t.Fatalf("privy peek/open/conceal controls must not contain %q in %s", forbidden, privyHTML)
			}
		}
	})

	t.Run("selects hero role card until own role is public", func(t *testing.T) {
		player.FaceUp = false
		player.RoleRevealed = false

		hiddenHTML := renderer.Render(GameBody(room, player)).GetHTML()
		hiddenPrivy := extractBetween(t, hiddenHTML, `id="zone-privy"`, `id="zone-notices"`)
		if !strings.Contains(hiddenPrivy, "role-card-hero") {
			t.Fatalf("expected hero role card for hidden own role: %s", hiddenPrivy)
		}
		if strings.Contains(hiddenPrivy, "role-card-compact") {
			t.Fatalf("did not expect compact role card for hidden own role: %s", hiddenPrivy)
		}

		player.FaceUp = true
		publicHTML := renderer.Render(GameBody(room, player)).GetHTML()
		publicPrivy := extractBetween(t, publicHTML, `id="zone-privy"`, `id="zone-notices"`)
		if !strings.Contains(publicPrivy, "role-card-compact") {
			t.Fatalf("expected compact role card after own role is public/face-up: %s", publicPrivy)
		}
	})
}

func extractBetween(t *testing.T, html, startMarker, endMarker string) string {
	t.Helper()
	start := strings.Index(html, startMarker)
	if start < 0 {
		t.Fatalf("expected start marker %q in HTML", startMarker)
	}
	end := strings.Index(html[start:], endMarker)
	if end < 0 {
		t.Fatalf("expected end marker %q after %q in HTML", endMarker, startMarker)
	}
	return html[start : start+end]
}

func extractAfter(t *testing.T, html, startMarker string) string {
	t.Helper()
	start := strings.Index(html, startMarker)
	if start < 0 {
		t.Fatalf("expected start marker %q in HTML", startMarker)
	}
	return html[start:]
}

func TestGameBody_CoupPrivacy(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	room := &game.Room{
		Code:       "COUP1",
		State:      game.StatePlaying,
		RulesMode:  game.RulesModeCoup,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 5,
	}

	king := &game.Player{
		ID:           "p1",
		Name:         "King Player",
		Role:         mockCoupCard(1001, "King"),
		RoleRevealed: true,
		FaceUp:       true,
	}
	blue := &game.Player{
		ID:     "p2",
		Name:   "Blue Player",
		Role:   mockCoupCard(1002, "Blue Knight"),
		FaceUp: false,
	}
	black := &game.Player{
		ID:     "p3",
		Name:   "Black Player",
		Role:   mockCoupCard(1003, "Black Knight"),
		FaceUp: false,
	}
	red := &game.Player{
		ID:     "p4",
		Name:   "Red Player",
		Role:   mockCoupCard(1004, "Red Knight"),
		FaceUp: false,
	}
	green := &game.Player{
		ID:     "p5",
		Name:   "Green Player",
		Role:   mockCoupCard(1005, "Green Knight"),
		FaceUp: false,
	}
	for _, player := range []*game.Player{king, blue, black, red, green} {
		room.Players[player.ID] = player
	}

	component := GameBody(room, blue)

	renderer.Render(component).
		AssertContains("Blue Knight").
		AssertContains("Royal Guard").
		AssertContains("any number of untapped creatures").
		AssertContains("creatures attacking the King player").
		AssertContains(`@post(&#39;/room/COUP1/coup/royal-guard/p2&#39;)`).
		AssertContains("Publicly Reveal Role").
		AssertContains(`@post(&#39;/room/COUP1/reveal/p2&#39;)`).
		AssertContains("King Player").
		AssertContains("Revealed: King").
		AssertNotContains("Black Knight").
		AssertNotContains("Red Knight").
		AssertNotContains("Green Knight")
}

func TestGameRosterZone_UsesStablePublicPlayerRows(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "ROWS2",
		State:      game.StatePlaying,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}
	viewer := &game.Player{
		ID:       "viewer",
		Name:     "Viewer",
		Role:     mockGuardianCard(),
		JoinedAt: time.Unix(1, 0),
	}
	hidden := &game.Player{
		ID:       "hidden",
		Name:     "Hidden Player",
		Role:     playerRosterHiddenCard(),
		FaceUp:   false,
		JoinedAt: time.Unix(2, 0),
	}
	revealed := &game.Player{
		ID:           "revealed",
		Name:         "Revealed Player",
		Role:         mockGuardianCard(),
		RoleRevealed: true,
		JoinedAt:     time.Unix(3, 0),
	}
	eliminated := &game.Player{
		ID:           "eliminated",
		Name:         "Eliminated Player",
		Role:         mockAssassinCard(),
		RoleRevealed: true,
		IsEliminated: true,
		JoinedAt:     time.Unix(4, 0),
	}
	for _, player := range []*game.Player{revealed, eliminated, hidden, viewer} {
		room.Players[player.ID] = player
	}

	html := renderer.Render(GameBody(room, viewer)).GetHTML()
	rosterHTML := extractAfter(t, html, `id="zone-roster"`)

	rowIDs := []string{
		`id="player-row-viewer"`,
		`id="player-row-hidden"`,
		`id="player-row-revealed"`,
		`id="player-row-eliminated"`,
	}
	lastIndex := -1
	for _, rowID := range rowIDs {
		index := strings.Index(rosterHTML, rowID)
		if index < 0 {
			t.Fatalf("expected stable row id %s in %s", rowID, rosterHTML)
		}
		if index < lastIndex {
			t.Fatalf("expected roster rows to follow join order in %s", rosterHTML)
		}
		lastIndex = index
	}

	for _, expected := range []string{
		"You",
		"Face Down",
		"Card is face down.",
		"Revealed: Test Guardian",
		"Eliminated",
		"opacity-60",
	} {
		if !strings.Contains(rosterHTML, expected) {
			t.Fatalf("expected %q in roster HTML: %s", expected, rosterHTML)
		}
	}

	for _, forbidden := range []string{
		"collapse collapse-arrow",
		`type="checkbox"`,
		"line-through",
		"Hidden Assassin",
		"Hidden Assassin Secret Text",
	} {
		if strings.Contains(rosterHTML, forbidden) {
			t.Fatalf("did not expect %q in roster HTML: %s", forbidden, rosterHTML)
		}
	}
}

func playerRosterHiddenCard() *game.Card {
	card := mockAssassinCard()
	card.Name = "Hidden Assassin"
	card.Text = "Hidden Assassin Secret Text"
	return card
}

func TestGameActions_UseConfirmTwiceAndEliminatedNotice(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "ACTS1",
		State:      game.StatePlaying,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}
	player := &game.Player{
		ID:     "p1",
		Name:   "Action Player",
		Role:   mockGuardianCard(),
		FaceUp: true,
	}
	room.Players[player.ID] = player

	html := renderer.Render(GameBody(room, player)).GetHTML()
	actionsHTML := extractBetween(t, html, `id="zone-actions"`, `id="zone-roster"`)
	for _, expected := range []string{
		"_confirmReveal",
		"Publicly Reveal Role",
		"Confirm Reveal Role",
		`@post(&#39;/room/ACTS1/reveal/p1&#39;)`,
		"_confirmEliminated",
		"I&#39;ve Been Eliminated",
		"Confirm Eliminated",
		`@post(&#39;/room/ACTS1/player/p1/eliminate&#39;)`,
	} {
		if !strings.Contains(actionsHTML, expected) {
			t.Fatalf("expected %q in actions HTML: %s", expected, actionsHTML)
		}
	}
	for _, forbidden := range []string{"onclick=", "confirm("} {
		if strings.Contains(actionsHTML, forbidden) {
			t.Fatalf("did not expect browser confirm %q in actions HTML: %s", forbidden, actionsHTML)
		}
	}

	player.RoleRevealed = true
	revealedHTML := renderer.Render(GameBody(room, player)).GetHTML()
	revealedActions := extractBetween(t, revealedHTML, `id="zone-actions"`, `id="zone-roster"`)
	if strings.Contains(revealedActions, "Publicly Reveal Role") {
		t.Fatalf("revealed player should not receive reveal control: %s", revealedActions)
	}

	player.IsEliminated = true
	eliminatedHTML := renderer.Render(GameBody(room, player)).GetHTML()
	eliminatedActions := extractBetween(t, eliminatedHTML, `id="zone-actions"`, `id="zone-roster"`)
	eliminatedNotices := extractBetween(t, eliminatedHTML, `id="zone-notices"`, `id="zone-actions"`)
	for _, forbidden := range []string{"Publicly Reveal Role", "I've Been Eliminated", "_confirmReveal", "_confirmEliminated"} {
		if strings.Contains(eliminatedActions, forbidden) {
			t.Fatalf("eliminated player action bar should be removed server-side: %s", eliminatedActions)
		}
	}
	for _, expected := range []string{
		"notice-card",
		"alert-error",
		"You are out of the game",
		"Room Operator can undo",
		`id="zone-privy"`,
	} {
		if !strings.Contains(eliminatedHTML, expected) && !strings.Contains(eliminatedNotices, expected) {
			t.Fatalf("expected eliminated state to contain %q in notice/game HTML", expected)
		}
	}
}

func TestGameBody_CoupRoyalGuardBlockerLimit(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	room := &game.Room{
		Code:                       "COUPRG",
		State:                      game.StatePlaying,
		RulesMode:                  game.RulesModeCoup,
		CoupRoyalGuardBlockerLimit: 1,
		Players:                    make(map[string]*game.Player),
		MaxPlayers:                 5,
	}
	blue := &game.Player{
		ID:     "p2",
		Name:   "Blue Player",
		Role:   mockCoupCard(1002, "Blue Knight"),
		FaceUp: false,
	}
	room.Players[blue.ID] = blue

	renderer.Render(GameBody(room, blue)).
		AssertContains("Royal Guard").
		AssertContains("one untapped creature").
		AssertNotContains("any number of untapped creatures")
}

func TestGameBody_CoupInquisitionCallUI(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, _, green := makeCoupInquisitionViewRoom()

	html := renderer.Render(GameBody(room, blue)).GetHTML()
	privyHTML := extractBetween(t, html, `id="zone-privy"`, `id="zone-notices"`)
	noticesHTML := extractBetween(t, html, `id="zone-notices"`, `id="zone-actions"`)
	for _, expected := range []string{
		"Inquisition",
		"coup-inquisition-form",
		`@post(&#39;/room/COUPINQ/coup/inquisition/blue&#39;, {contentType: &#39;form&#39;})`,
		`name="targetID"`,
		"Red Player",
		`name="currentLife"`,
		"Wrong guess penalty at 40 life: 20 life",
	} {
		if !strings.Contains(privyHTML, expected) {
			t.Fatalf("expected Blue-only Inquisition form detail %q inside privy HTML: %s", expected, privyHTML)
		}
	}
	if strings.Contains(noticesHTML, "coup-inquisition-form") || strings.Contains(noticesHTML, "Call Inquisition") {
		t.Fatalf("Blue-only Inquisition form should not render in public notices: %s", noticesHTML)
	}

	greenHTML := renderer.Render(GameBody(room, green)).GetHTML()
	if strings.Contains(greenHTML, "coup-inquisition-form") || strings.Contains(greenHTML, "Call Inquisition") {
		t.Fatalf("non-Blue player should not receive Inquisition caller form: %s", greenHTML)
	}
}

func TestCoupInquisitionPanel_ExcludesKingFromTargets(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, _, _ := makeCoupInquisitionViewRoom()
	king := &game.Player{
		ID:           "king",
		Name:         "King Player",
		Role:         mockCoupCard(1001, "King"),
		RoleRevealed: true,
		FaceUp:       true,
	}
	room.Players[king.ID] = king

	renderer.Render(CoupInquisitionPrivatePanel(room, blue)).
		AssertContains(`name="targetID"`).
		AssertContains("Red Player").
		AssertContains("Green Player").
		AssertNotContains("King Player")
}

func TestGameBody_CoupInquisitionPendingNoticeHidesResultUntilConfirmed(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, red, green := makeCoupInquisitionViewRoom()
	blue.RoleRevealed = true
	blue.FaceUp = true
	room.CoupInquisition = &game.CoupInquisitionState{
		Attempts: map[string]game.CoupInquisitionAttempt{
			blue.ID: {
				InquisitorID: blue.ID,
				TargetID:     red.ID,
				CurrentLife:  39,
				PenaltyLife:  20,
			},
		},
		Pending: &game.CoupInquisitionAttempt{
			InquisitorID: blue.ID,
			TargetID:     red.ID,
			CurrentLife:  39,
			PenaltyLife:  20,
		},
	}

	renderer.Render(GameBody(room, green)).
		AssertContains("notice-card").
		AssertContains("Inquisition Notice").
		AssertContains("Blue Player named Red Player").
		AssertContains(`@post(&#39;/room/COUPINQ/coup/inquisition/confirm&#39;)`).
		AssertNotContains("Red Knight")
}

func TestGameBody_CoupInquisitionFailureResultShowsPenalty(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, _, green := makeCoupInquisitionViewRoom()
	room.CoupInquisition = &game.CoupInquisitionState{
		Attempts: map[string]game.CoupInquisitionAttempt{
			blue.ID: {
				InquisitorID: blue.ID,
				TargetID:     green.ID,
				CurrentLife:  39,
				PenaltyLife:  20,
				Resolved:     true,
			},
		},
		Last: &game.CoupInquisitionAttempt{
			InquisitorID: blue.ID,
			TargetID:     green.ID,
			CurrentLife:  39,
			PenaltyLife:  20,
			Resolved:     true,
			Success:      false,
		},
	}

	renderer.Render(GameBody(room, blue)).
		AssertContains("Inquisition failed").
		AssertContains("lose 20 life")
}

func TestGameBody_CoupPrivateInquisitionResultOnlyInformsInquisitor(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, red, green := makeCoupInquisitionViewRoom()
	room.CoupInquisitionResultPolicy = game.CoupInquisitionResultPrivate
	room.CoupInquisition = &game.CoupInquisitionState{
		Attempts: map[string]game.CoupInquisitionAttempt{
			blue.ID: {
				InquisitorID: blue.ID,
				TargetID:     red.ID,
				CurrentLife:  39,
				PenaltyLife:  20,
				Resolved:     true,
				Success:      true,
			},
		},
		Last: &game.CoupInquisitionAttempt{
			InquisitorID: blue.ID,
			TargetID:     red.ID,
			CurrentLife:  39,
			PenaltyLife:  20,
			Resolved:     true,
			Success:      true,
		},
		Succeeded: true,
	}

	blueHTML := renderer.Render(GameBody(room, blue)).GetHTML()
	bluePrivy := extractBetween(t, blueHTML, `id="zone-privy"`, `id="zone-notices"`)
	blueNotices := extractBetween(t, blueHTML, `id="zone-notices"`, `id="zone-actions"`)
	for _, expected := range []string{
		"Private Inquisition result",
		"Inquisition succeeded",
		"Red Player was Red",
	} {
		if !strings.Contains(bluePrivy, expected) {
			t.Fatalf("expected private result detail %q inside inquisitor privy HTML: %s", expected, bluePrivy)
		}
	}
	if strings.Contains(blueNotices, "Red Player was Red") {
		t.Fatalf("private result detail should not render in public notices: %s", blueNotices)
	}
	if !strings.Contains(blueNotices, "Result delivered privately") {
		t.Fatalf("expected neutral private-result public notice for inquisitor too: %s", blueNotices)
	}

	greenHTML := renderer.Render(GameBody(room, green)).GetHTML()
	greenNotices := extractBetween(t, greenHTML, `id="zone-notices"`, `id="zone-actions"`)
	if !strings.Contains(greenNotices, "Result delivered privately") {
		t.Fatalf("expected neutral private-result notice for other players: %s", greenNotices)
	}
	if strings.Contains(greenHTML, "Red Player was Red") {
		t.Fatalf("non-inquisitor should not receive private result detail: %s", greenHTML)
	}
}

func TestGameBody_CoupPublicInquisitionResultUsesNoticeCard(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, blue, red, green := makeCoupInquisitionViewRoom()
	room.CoupInquisitionResultPolicy = game.CoupInquisitionResultPublic
	room.CoupInquisition = &game.CoupInquisitionState{
		Attempts: map[string]game.CoupInquisitionAttempt{
			blue.ID: {
				InquisitorID: blue.ID,
				TargetID:     red.ID,
				CurrentLife:  39,
				PenaltyLife:  20,
				Resolved:     true,
				Success:      true,
			},
		},
		Last: &game.CoupInquisitionAttempt{
			InquisitorID: blue.ID,
			TargetID:     red.ID,
			CurrentLife:  39,
			PenaltyLife:  20,
			Resolved:     true,
			Success:      true,
		},
		Succeeded: true,
	}

	renderer.Render(GameBody(room, green)).
		AssertContains("notice-card").
		AssertContains("Inquisition succeeded").
		AssertContains("Red Player was Red").
		AssertNotContains("Red Knight")
}

func TestGameBody_CoupAdvisoryWinPromptRequiresConfirmation(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, actor := makeCoupWinViewRoom()

	renderer.Render(GameBody(room, actor)).
		AssertContains("Looks like Black might have just won???").
		AssertContains("King has fallen").
		AssertContains("Confirm Win").
		AssertContains("Reject Prompt").
		AssertContains(`@post(&#39;/room/COUPWIN/coup/win/confirm&#39;)`).
		AssertContains(`@post(&#39;/room/COUPWIN/coup/win/reject&#39;)`)
}

func TestGameBody_CoupConfirmedWinShowsOutcome(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, actor := makeCoupWinViewRoom()
	prompt := game.DetectCoupAdvisoryWin(room)
	game.ConfirmCoupWin(room, prompt)
	room.State = game.StateEnded

	renderer.Render(GameBody(room, actor)).
		AssertContains("Confirmed Coup Win").
		AssertContains("Black win confirmed").
		AssertContains("King has fallen")
}

func makeCoupInquisitionViewRoom() (*game.Room, *game.Player, *game.Player, *game.Player) {
	room := &game.Room{
		Code:       "COUPINQ",
		State:      game.StatePlaying,
		RulesMode:  game.RulesModeCoup,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 5,
	}
	blue := &game.Player{
		ID:     "blue",
		Name:   "Blue Player",
		Role:   mockCoupCard(1002, "Blue Knight"),
		FaceUp: false,
	}
	red := &game.Player{
		ID:     "red",
		Name:   "Red Player",
		Role:   mockCoupCard(1004, "Red Knight"),
		FaceUp: false,
	}
	green := &game.Player{
		ID:     "green",
		Name:   "Green Player",
		Role:   mockCoupCard(1005, "Green Knight"),
		FaceUp: false,
	}
	for _, player := range []*game.Player{blue, red, green} {
		room.Players[player.ID] = player
	}
	return room, blue, red, green
}

func makeCoupWinViewRoom() (*game.Room, *game.Player) {
	room := &game.Room{
		Code:      "COUPWIN",
		State:     game.StatePlaying,
		RulesMode: game.RulesModeCoup,
		Players:   make(map[string]*game.Player),
	}
	king := &game.Player{
		ID:           "king",
		Name:         "King Player",
		Role:         mockCoupCard(1001, "King"),
		RoleRevealed: true,
		FaceUp:       true,
		IsEliminated: true,
	}
	black := &game.Player{
		ID:     "black",
		Name:   "Black Player",
		Role:   mockCoupCard(1003, "Black Knight"),
		FaceUp: false,
	}
	red := &game.Player{
		ID:           "red",
		Name:         "Red Player",
		Role:         mockCoupCard(1004, "Red Knight"),
		RoleRevealed: true,
		FaceUp:       true,
		IsEliminated: true,
	}
	for _, player := range []*game.Player{king, black, red} {
		room.Players[player.ID] = player
	}
	return room, black
}

func TestGameBody_CoupPrivateInformationScopedToRecipient(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	room := &game.Room{
		Code:       "COUP2",
		State:      game.StatePlaying,
		RulesMode:  game.RulesModeCoup,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 5,
	}

	kingCard := mockCoupCard(1001, "King")
	kingCard.Rulings = []string{"Private information: Blue Knights: Blue Player"}

	king := &game.Player{
		ID:           "p1",
		Name:         "King Player",
		Role:         kingCard,
		RoleRevealed: true,
		FaceUp:       true,
	}
	blue := &game.Player{
		ID:     "p2",
		Name:   "Blue Player",
		Role:   mockCoupCard(1002, "Blue Knight"),
		FaceUp: false,
	}
	black := &game.Player{
		ID:     "p3",
		Name:   "Black Player",
		Role:   mockCoupCard(1003, "Black Knight"),
		FaceUp: false,
	}
	for _, player := range []*game.Player{king, blue, black} {
		room.Players[player.ID] = player
	}

	renderer.Render(GameBody(room, king)).
		AssertContains("Private information: Blue Knights: Blue Player")

	renderer.Render(GameBody(room, black)).
		AssertNotContains("Private information:").
		AssertNotContains("Blue Knights:")
}
