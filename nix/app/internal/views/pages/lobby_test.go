package pages

import (
	"strings"
	"testing"
	"time"
	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestLobbyPage(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	// Create test data
	room := &game.Room{
		Code:       "TEST1",
		State:      game.StateLobby,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}

	player1 := &game.Player{
		ID:        "p1",
		Name:      "Player One",
		SessionID: "session-one",
	}

	player2 := &game.Player{
		ID:        "p2",
		Name:      "Player Two",
		SessionID: "session-two",
	}

	room.Players[player1.ID] = player1
	room.Players[player2.ID] = player2
	room.OperatorSessionID = player1.SessionID

	// Set player1 as the first player (simulating room creator)
	player1.JoinedAt = room.CreatedAt

	// Create config and card service
	cfg := config.DefaultConfig()
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create card service: %v", err)
	}

	t.Run("renders lobby page structure", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("TEST1").
			AssertHasElementWithID("lobby-container").
			AssertHasElementWithID("player-lobby").
			AssertContains("Room Code")
	})

	t.Run("shows SSE connection", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains(`data-init="@get(&#39;/sse/lobby/TEST1&#39;)"`)
	})

	t.Run("renders player list", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Players").
			AssertContains("2 of 4 seats filled").
			AssertContains("Player One").
			AssertContains("Player Two")
	})

	t.Run("shows full debug surface for room operator when debug enabled", func(t *testing.T) {
		debugCfg := config.DefaultConfig()
		debugCfg.Server.DebugModeEnabled = true

		component := LobbyPage(room, player1, debugCfg, cardService)

		renderer.Render(component).
			AssertContains(`id="debug-control-surface"`).
			AssertContains("Debug Control Surface").
			AssertContains(`id="debug-panel-toggle"`).
			AssertContains(`id="debug-clear"`).
			AssertContains("Debug Insights").
			AssertContains("Start with Debug Players").
			AssertContains("View As Player")
	})

	t.Run("hides debug surface from non-operator when debug enabled", func(t *testing.T) {
		debugCfg := config.DefaultConfig()
		debugCfg.Server.DebugModeEnabled = true

		component := LobbyPage(room, player2, debugCfg, cardService)

		renderer.Render(component).
			AssertNotContains(`id="debug-control-surface"`).
			AssertNotContains("Debug Control Surface").
			AssertNotContains("Debug Mode active")
	})

	t.Run("shows selected rules mode", func(t *testing.T) {
		coupRoom := &game.Room{
			Code:       "COUP1",
			State:      game.StateLobby,
			RulesMode:  game.RulesModeCoup,
			Players:    make(map[string]*game.Player),
			MaxPlayers: 4,
		}
		coupRoom.Players[player1.ID] = player1
		coupRoom.Players[player2.ID] = player2
		coupRoom.OperatorSessionID = player1.SessionID

		component := LobbyPage(coupRoom, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Coup")
	})

	t.Run("summarizes Green Blue Hunt settings", func(t *testing.T) {
		coupRoom := &game.Room{
			Code:                     "COUPH",
			State:                    game.StateLobby,
			RulesMode:                game.RulesModeCoup,
			CoupPreset:               game.CoupPresetSeven,
			CoupGreenHuntRequirement: game.CoupGreenHuntAllBlues,
			CoupInquisitionAmnesty:   game.CoupInquisitionAmnestyBroad,
			Players:                  make(map[string]*game.Player),
		}
		coupRoom.Players[player1.ID] = player1
		coupRoom.Players[player2.ID] = player2
		coupRoom.OperatorSessionID = player1.SessionID

		component := LobbyPage(coupRoom, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("All Blue hunt").
			AssertContains("Broad Amnesty")
	})

	t.Run("room operator coup lobby is read-only player surface", func(t *testing.T) {
		coupRoom := &game.Room{
			Code:       "COUP2",
			State:      game.StateLobby,
			RulesMode:  game.RulesModeCoup,
			CoupPreset: game.CoupPresetSix,
			Players:    make(map[string]*game.Player),
			MaxPlayers: 4,
		}
		for i := 1; i <= 6; i++ {
			player := game.NewPlayer(string(rune('a'+i)), "Coup Player", "session"+string(rune('a'+i)))
			player.JoinedAt = time.Unix(int64(i), 0)
			coupRoom.Players[player.ID] = player
		}
		currentPlayer := coupRoom.Players["b"]
		coupRoom.OperatorSessionID = currentPlayer.SessionID

		component := LobbyPage(coupRoom, currentPlayer, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		for _, want := range []string{
			`id="player-lobby"`,
			"Waiting for Room Operator - 6 of 6 seats filled",
			"Coup - 6 players",
			"Coup Rules Reference",
			"Every other player remains an opponent",
			"Do not prove a hidden role",
			"Royal Guard",
			"Inquisition",
			"Green Blue Hunt",
			"Green Hunt Requirement",
			"King-Side Inquisition Amnesty",
			"If Broad Amnesty is on",
			"Wasteland",
			"Advisory Win Prompts",
			"Leave Room",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected room operator player lobby to contain %q, got %q", want, body)
			}
		}
		assertNoLobbyManagementDOM(t, body, coupRoom.Code)
	})

	t.Run("coup rules reference explains green blue hunt without stale strict copy", func(t *testing.T) {
		body := renderer.Render(CoupRulesReference()).GetHTML()

		for _, want := range []string{
			"Coup Rules Reference",
			"Green Blue Hunt",
			"Green hunts Blue Knights",
			"Green Hunt Requirement",
			"one Blue Knight must die before the King falls",
			"all Blue Knights must die before the King falls",
			"King-Side Inquisition Amnesty",
			"If Inquisition succeeds, Green may share a King victory even without a Blue death",
			"Green can share a Red victory only if Green Hunt was satisfied before the King fell",
			"If Broad Amnesty is on, a successful Inquisition before the King falls also lets Green share a Red victory.",
			"Blue reveal, exposure, Royal Guard reveal, and Inquisition reveal do not satisfy the Hunt",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected Coup rules reference to contain %q, got %q", want, body)
			}
		}
		for _, stale := range []string{
			"Green Eligibility",
			"Strict Green",
			"selected Green rules",
			"selected Green eligibility rules",
			"Successful Inquisition can satisfy Green for a King victory",
			"Blue exposure satisfies",
		} {
			if strings.Contains(body, stale) {
				t.Fatalf("expected Coup rules reference to omit stale copy %q, got %q", stale, body)
			}
		}
	})

	t.Run("hides room management controls from first active non-operator", func(t *testing.T) {
		coupRoom := &game.Room{
			Code:       "COUP3",
			State:      game.StateLobby,
			RulesMode:  game.RulesModeCoup,
			CoupPreset: game.CoupPresetSix,
			Players:    make(map[string]*game.Player),
			MaxPlayers: 4,
		}
		firstPlayer := game.NewPlayer("p1", "First Player", "session-first")
		operator := game.NewPlayer("p2", "Operator", "session-operator")
		firstPlayer.JoinedAt = time.Unix(1, 0)
		operator.JoinedAt = time.Unix(2, 0)
		coupRoom.Players[firstPlayer.ID] = firstPlayer
		coupRoom.Players[operator.ID] = operator
		coupRoom.OperatorSessionID = operator.SessionID

		component := LobbyPage(coupRoom, firstPlayer, cfg, cardService)

		renderer.Render(component).
			AssertNotContains("coup-preset-form").
			AssertNotContains(`@post(&#39;/room/COUP3/start&#39;)`).
			AssertContains("Leave Room")
	})

	t.Run("non-operator coup lobby omits all room management dom", func(t *testing.T) {
		coupRoom := &game.Room{
			Code:       "COUP4",
			State:      game.StateLobby,
			RulesMode:  game.RulesModeCoup,
			CoupPreset: game.CoupPresetFive,
			Players:    make(map[string]*game.Player),
			MaxPlayers: 5,
		}
		operator := game.NewPlayer("operator", "Room Operator", "session-operator")
		player := game.NewPlayer("player", "Regular Player", "session-player")
		coupRoom.Players[operator.ID] = operator
		coupRoom.Players[player.ID] = player
		coupRoom.OperatorSessionID = operator.SessionID

		component := LobbyPage(coupRoom, player, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		if !strings.Contains(body, `id="player-lobby"`) || !strings.Contains(body, "Regular Player") || !strings.Contains(body, "Leave Room") {
			t.Fatalf("expected player-safe lobby content, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, coupRoom.Code)
	})

	t.Run("non-operator lobby renders redesigned player-safe surface", func(t *testing.T) {
		coupRoom := &game.Room{
			Code:                        "COUP5",
			State:                       game.StateLobby,
			RulesMode:                   game.RulesModeCoup,
			CoupPreset:                  game.CoupPresetFive,
			CoupInquisitionResultPolicy: game.CoupInquisitionResultPublic,
			CoupInfoPolicy:              game.CoupInformationPolicy{KingToBlue: game.CoupKingKnowsAllBlue},
			Players:                     make(map[string]*game.Player),
			MaxPlayers:                  5,
		}
		operator := game.NewPlayer("operator", "Room Operator", "session-operator")
		player := game.NewPlayer("player", "Regular Player", "session-player")
		coupRoom.Players[operator.ID] = operator
		coupRoom.Players[player.ID] = player
		coupRoom.OperatorSessionID = operator.SessionID

		component := LobbyPage(coupRoom, player, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		for _, want := range []string{
			`id="player-lobby"`,
			`id="player-lobby-hero"`,
			`id="lobby-qr-code"`,
			`aria-label="QR code for room COUP5"`,
			`src="/room/COUP5/qr.png"`,
			"Waiting for Room Operator",
			"2 of 5 seats filled",
			`id="player-row-player"`,
			"Open Seat 3",
			"Open Seat 5",
			"Coup - 5 players - Public inquisition - One Blue hunt - King victory only - Full King knowledge",
			`id="lobby-settings-summary"`,
			`id="rules-reference"`,
			"Rules Reference",
			"Leave Room",
			`@post(&#39;/room/COUP5/leave&#39;)`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected redesigned player lobby to contain %q, got %q", want, body)
			}
		}
		if strings.Contains(body, `lobby-qr-placeholder`) || strings.Contains(body, `data-attr:src="$qrCode"`) {
			t.Fatalf("expected player lobby QR to render a static image endpoint, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, coupRoom.Code)
	})

	t.Run("non-operator treachery lobby omits role configuration dom", func(t *testing.T) {
		treacheryRoom := &game.Room{
			Code:      "TRCH1",
			State:     game.StateLobby,
			RulesMode: game.RulesModeTreachery,
			Players:   make(map[string]*game.Player),
		}
		roleService := game.NewRoleConfigService(cfg)
		roleConfig, err := roleService.CreateFromPreset("standard", 5)
		if err != nil {
			t.Fatalf("create role config: %v", err)
		}
		treacheryRoom.RoleConfig = roleConfig
		operator := game.NewPlayer("operator", "Room Operator", "session-operator")
		player := game.NewPlayer("player", "Regular Player", "session-player")
		treacheryRoom.Players[operator.ID] = operator
		treacheryRoom.Players[player.ID] = player
		treacheryRoom.OperatorSessionID = operator.SessionID

		component := LobbyPage(treacheryRoom, player, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		if !strings.Contains(body, `id="player-lobby"`) || !strings.Contains(body, "Regular Player") || !strings.Contains(body, "Leave Room") {
			t.Fatalf("expected player-safe lobby content, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, treacheryRoom.Code)
	})

	t.Run("room operator treachery lobby omits setup controls", func(t *testing.T) {
		treacheryRoom := &game.Room{
			Code:      "TRCH2",
			State:     game.StateLobby,
			RulesMode: game.RulesModeTreachery,
			Players:   make(map[string]*game.Player),
		}
		roleService := game.NewRoleConfigService(cfg)
		roleConfig, err := roleService.CreateFromPreset("standard", 5)
		if err != nil {
			t.Fatalf("create role config: %v", err)
		}
		treacheryRoom.RoleConfig = roleConfig
		operator := game.NewPlayer("operator", "Room Operator", "session-operator")
		treacheryRoom.Players[operator.ID] = operator
		treacheryRoom.OperatorSessionID = operator.SessionID

		component := LobbyPage(treacheryRoom, operator, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		if !strings.Contains(body, `id="player-lobby"`) || !strings.Contains(body, "Treachery - 5 players") || !strings.Contains(body, "Leave Room") {
			t.Fatalf("expected room operator lobby to render player-safe Treachery lobby content, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, treacheryRoom.Code)
	})

	t.Run("shows empty lobby waiting status", func(t *testing.T) {
		// Create empty room to test minimum message
		emptyRoom := &game.Room{
			Code:       "EMPTY",
			State:      game.StateLobby,
			Players:    make(map[string]*game.Player),
			MaxPlayers: 4,
		}
		emptyRoom.OperatorSessionID = player1.SessionID

		component := LobbyPage(emptyRoom, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Waiting for Room Operator - 0 of 4 seats filled").
			AssertNotContains("Start Game")
	})

	t.Run("shows player lobby instead of start controls when enough players", func(t *testing.T) {
		// Room already has 2 players (player1 and player2), which is >= 1
		component := LobbyPage(room, player1, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		if !strings.Contains(body, `id="player-lobby"`) || !strings.Contains(body, "Waiting for Room Operator - 2 of 4 seats filled") {
			t.Fatalf("expected read-only player lobby, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, room.Code)
	})

	t.Run("shows leave button", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Leave Room").
			AssertContains(`@post(&#39;/room/TEST1/leave&#39;)`)
	})
}

func assertNoLobbyManagementDOM(t *testing.T, body string, roomCode string) {
	t.Helper()

	for _, forbidden := range []string{
		`id="role-config"`,
		`id="preset-form"`,
		`id="coup-preset-form"`,
		`id="coup-role-counts-form"`,
		`id="coup-info-form"`,
		`id="coup-royal-guard-form"`,
		`id="coup-inquisition-settings-form"`,
		`data-signals:is-starting`,
		`data-signals:start-error`,
		`data-signals:can-start-game`,
		`/room/` + roomCode + `/config/`,
		`/room/` + roomCode + `/start`,
		`Unsafe Role Count Override`,
		`Allow Leaderless Games`,
		`Hide Role Distribution`,
		`Fully Random Roles`,
		`Start Game`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("non-operator lobby rendered forbidden management DOM %q in %q", forbidden, body)
		}
	}
}

func TestLobbySettingsSummary(t *testing.T) {
	room := &game.Room{
		Code:                            "SUMRY",
		State:                           game.StateLobby,
		RulesMode:                       game.RulesModeCoup,
		CoupPreset:                      game.CoupPresetFive,
		CoupInquisitionResultPolicy:     game.CoupInquisitionResultPublic,
		CoupInfoPolicy:                  game.CoupInformationPolicy{KingToBlue: game.CoupKingKnowsAllBlue},
		Players:                         make(map[string]*game.Player),
		MaxPlayers:                      5,
		CoupRoyalGuardBlockerLimit:      0,
		CoupAllowUnsafeRoleCounts:       false,
		CoupGreenEligibleBeforeKingFall: false,
	}

	summary := LobbySettingsSummary(room)

	for _, want := range []string{
		"Coup",
		"5 players",
		"Public inquisition",
		"Full King knowledge",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("expected settings summary to contain %q, got %q", want, summary)
		}
	}
}

func TestLobbyBody(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	room := &game.Room{
		Code:       "BODY1",
		State:      game.StateLobby,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 4,
	}

	player := &game.Player{
		ID:        "p1",
		Name:      "Test Player",
		SessionID: "session-test",
	}

	room.Players[player.ID] = player
	room.OperatorSessionID = player.SessionID

	// Create config and card service
	cfg := config.DefaultConfig()
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create card service: %v", err)
	}

	t.Run("renders lobby body fragment", func(t *testing.T) {
		component := LobbyBody(room, player, cfg, cardService)

		renderer.Render(component).
			AssertNotEmpty().
			AssertHasElementWithID("lobby-container").
			AssertHasElementWithID("player-lobby").
			AssertContains("Players").
			AssertContains("1 of 4 seats filled")
	})

	t.Run("renders read-only player surface", func(t *testing.T) {
		component := LobbyBody(room, player, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		if !strings.Contains(body, "Waiting for Room Operator - 1 of 4 seats filled") || !strings.Contains(body, "Leave Room") {
			t.Fatalf("expected read-only lobby body, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, room.Code)
	})

	t.Run("does not render start controls when enough players", func(t *testing.T) {
		// Room already has 1 player which is enough to start

		component := LobbyBody(room, player, cfg, cardService)
		body := renderer.Render(component).GetHTML()

		if strings.Contains(body, "Start Game") || strings.Contains(body, `data-attr:disabled="$isStarting || !$canStartGame"`) {
			t.Fatalf("expected lobby body to omit start controls, got %q", body)
		}
		assertNoLobbyManagementDOM(t, body, room.Code)
	})

}
