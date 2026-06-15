package pages

import (
	"strings"
	"testing"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestHostDashboardPlaying_PublicStateBoard(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)

	room := &game.Room{
		Code:      "HOST1",
		State:     game.StatePlaying,
		RulesMode: game.RulesModeCoup,
		Players:   make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	hidden := &game.Player{
		ID:           "p1",
		Name:         "Blue Player",
		Role:         mockCoupCard(1002, "Blue Knight"),
		RoleRevealed: false,
		FaceUp:       false,
	}
	hidden.Role.Rulings = []string{"Private information: Red is Red Player"}
	revealed := &game.Player{
		ID:           "p2",
		Name:         "Red Player",
		Role:         mockCoupCard(1004, "Red Knight"),
		RoleRevealed: true,
		FaceUp:       true,
	}
	room.Players[host.ID] = host
	room.Players[hidden.ID] = hidden
	room.Players[revealed.ID] = revealed

	html := renderer.Render(HostDashboardPlaying(room, host)).GetHTML()
	for _, expected := range []string{
		`id="operator-live-dashboard"`,
		`id="operator-live-board"`,
		`id="operator-tile-p1"`,
		`id="operator-tile-p2"`,
		"Roles stay hidden from the Room Operator until they are public.",
		"Blue Player",
		"Red Player",
		"Presence: seat active",
		"Face Down",
		"Revealed: Red Knight",
		"Record Reveal",
		`@post(&#39;/room/HOST1/reveal/p1&#39;)`,
		"Record Elimination",
		`@post(&#39;/room/HOST1/player/p1/eliminate&#39;)`,
		"overflow-x-auto",
		"min-w-64",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected live operator dashboard to contain %q in %s", expected, html)
		}
	}
	for _, forbidden := range []string{
		"Blue Knight",
		"Private information:",
		"onclick=",
		"confirm(",
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("live operator dashboard rendered forbidden private/browser-confirm content %q in %s", forbidden, html)
		}
	}
}

func TestHostDashboardPlaying_PublicCoupFacts(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:                            "FACTS",
		State:                           game.StatePlaying,
		RulesMode:                       game.RulesModeCoup,
		CoupKingFallen:                  true,
		CoupGreenEligibleBeforeKingFall: true,
		CoupInquisition: &game.CoupInquisitionState{
			Succeeded: true,
			Last: &game.CoupInquisitionAttempt{
				InquisitorID: "blue",
				TargetID:     "red",
				Success:      true,
			},
		},
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true}
	blue := &game.Player{ID: "blue", Name: "Blue Player", Role: mockCoupCard(1002, "Blue Knight")}
	red := &game.Player{ID: "red", Name: "Red Player", Role: mockCoupCard(1004, "Red Knight")}
	room.Players[host.ID] = host
	room.Players[blue.ID] = blue
	room.Players[red.ID] = red

	renderer.Render(HostDashboardPlaying(room, host)).
		AssertContains("Public Coup facts").
		AssertContains("King fallen: yes").
		AssertContains("Green lock: eligible before King fall").
		AssertContains("Inquisition: succeeded").
		AssertNotContains("Blue Knight").
		AssertNotContains("Red Knight")
}

func TestHostDashboardPlaying_CoupAdvisoryWinControls(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, _ := makeCoupWinViewRoom()
	host := &game.Player{
		ID:        "host",
		Name:      "Host",
		IsHost:    true,
		SessionID: "session-host",
	}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID

	renderer.Render(HostDashboardPlaying(room, host)).
		AssertContains("Looks like Black might have just won???").
		AssertContains("King has fallen").
		AssertContains("Confirm Win").
		AssertContains("Reject Prompt").
		AssertContains(`@post(&#39;/room/COUPWIN/coup/win/confirm&#39;)`).
		AssertContains(`@post(&#39;/room/COUPWIN/coup/win/reject&#39;)`)
}

func TestHostDashboardLobby_DebugPanelGatedByConfig(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:        "host",
		Name:      "Host",
		IsHost:    true,
		SessionID: "session-host",
	}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID

	disabled := config.DefaultConfig()
	disabled.Server.DebugModeEnabled = false
	renderer.Render(HostDashboardLobby(room, host, disabled, nil)).
		AssertNotContains(`id="debug-control-surface"`).
		AssertNotContains(`id="debug-panel"`).
		AssertNotContains(`id="debug-panel-toggle"`).
		AssertNotContains(`id="debug-clear"`).
		AssertNotContains("Debug Insights").
		AssertNotContains("setupDebugPanel")

	enabled := config.DefaultConfig()
	enabled.Server.DebugModeEnabled = true
	renderer.Render(HostDashboardLobby(room, host, enabled, nil)).
		AssertContains(`id="debug-control-surface"`).
		AssertContains(`id="debug-panel"`).
		AssertContains(`id="debug-clear"`).
		AssertContains("Debug Insights").
		AssertContains("setupDebugPanel")
}

func TestHostDashboardLobby_DebugControlSurfaceShell(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:        "host",
		Name:      "Host",
		IsHost:    true,
		SessionID: "session-host",
	}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	html := renderer.Render(HostDashboardLobby(room, host, cfg, nil)).GetHTML()
	for _, expected := range []string{
		`id="debug-control-surface"`,
		"DEBUG BUILD",
		"Debug Control Surface",
		"right-4",
		"inset-y-4",
		"border-dashed",
		`data-signals="{_showDebugSpoilers: false}"`,
		`id="debug-show-hidden-roles"`,
		"Show hidden roles",
		`id="debug-view-as-player-container"`,
		`id="debug-start-override-controls"`,
		`id="debug-start-with-debug-players"`,
		`@post(&#39;/room/DEBUG/debug/start-with-debug-players&#39;)`,
		`id="debug-start-as-is"`,
		`@post(&#39;/room/DEBUG/debug/start-as-is&#39;)`,
		`id="debug-insights-container"`,
		`id="debug-persistence-controls"`,
		`id="debug-dump"`,
		`id="debug-clear"`,
		"_debugClearRoom",
		`@post(&#39;/room/DEBUG/debug/clear&#39;)`,
		`id="debug-restore"`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected debug surface shell detail %q in HTML: %s", expected, html)
		}
	}
	if strings.Contains(html, "confirm(\"Clear room") {
		t.Fatalf("debug clear should not use browser confirm: %s", html)
	}
	viewAsIndex := strings.Index(html, `id="debug-view-as-player-container"`)
	insightsIndex := strings.Index(html, `id="debug-insights-container"`)
	persistenceIndex := strings.Index(html, `id="debug-persistence-controls"`)
	if viewAsIndex < 0 || insightsIndex < 0 || persistenceIndex < 0 {
		t.Fatalf("expected debug sections in HTML: %s", html)
	}
	if !(viewAsIndex < insightsIndex && insightsIndex < persistenceIndex) {
		t.Fatalf("expected View As before insights and persistence last; indexes viewAs=%d insights=%d persistence=%d", viewAsIndex, insightsIndex, persistenceIndex)
	}
}

func TestHostDashboardLobby_DebugControlSurfaceCanBeMinimized(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:        "host",
		Name:      "Host",
		IsHost:    true,
		SessionID: "session-host",
	}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains(`id="debug-panel-toggle"`).
		AssertContains(`aria-controls="debug-panel"`).
		AssertContains(`aria-expanded="true"`).
		AssertContains("Minimize").
		AssertContains(`id="debug-panel-minimized"`).
		AssertContains("Debug Mode minimized").
		AssertContains("debug-panel-minimized").
		AssertContains(`surface.classList.toggle("debug-panel-minimized", minimized)`).
		AssertContains("treacherest_debug_panel_minimized_").
		AssertContains("setDebugPanelMinimized")
}

func TestHostDashboardLobby_CoupModeUsesCoupSetupControls(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "COUPH",
		State:      game.StateLobby,
		RulesMode:  game.RulesModeCoup,
		CoupPreset: game.CoupPresetFive,
		RoleConfig: &game.RoleConfiguration{
			MaxPlayers: 5,
			RoleTypes:  map[string]*game.RoleTypeConfig{},
		},
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	room.Players[host.ID] = host
	cfg := config.DefaultConfig()

	renderer.Render(HostDashboardLobby(room, host, cfg, &game.CardService{})).
		AssertContains("Player Count").
		AssertContains(`@post(&#39;/room/COUPH/config/coup-player-count/decrement&#39;)`).
		AssertContains(`@post(&#39;/room/COUPH/config/coup-player-count/increment&#39;)`).
		AssertContains("Coup Preset").
		AssertContains("Role Counts").
		AssertContains(`id="coup-role-counts-form"`).
		AssertContains(`name="king"`).
		AssertContains(`name="blueKnight"`).
		AssertContains(`name="blackKnight"`).
		AssertContains(`name="redKnight"`).
		AssertContains(`name="greenKnight"`).
		AssertContains(`name="wastelandKnight"`).
		AssertContains(`name="unsafeRoleCounts"`).
		AssertContains("Unsafe Role Count Override").
		AssertContains("game is probably broken").
		AssertContains("King-to-Blue Info").
		AssertContains("Red-to-Black Info").
		AssertContains("Royal Guard Blockers").
		AssertContains("Inquisition Result").
		AssertContains("Coup Rules Reference").
		AssertNotContains("Role Preset:").
		AssertNotContains("Allow Leaderless Games").
		AssertNotContains("Leaders").
		AssertNotContains("Guardians").
		AssertNotContains("Assassins").
		AssertNotContains("Traitors")
}

func TestHostDashboardLobby_InvalidCoupSetupDisablesStartWithValidationLine(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "BADST",
		State:      game.StateLobby,
		RulesMode:  game.RulesModeCoup,
		CoupPreset: game.CoupPresetFive,
		Players:    make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	player := &game.Player{ID: "p1", Name: "Player One", SessionID: "session-one"}
	room.Players[host.ID] = host
	room.Players[player.ID] = player
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()

	renderer.Render(HostDashboardLobby(room, host, cfg, &game.CardService{})).
		AssertContains(`id="operator-dashboard"`).
		AssertContains(`id="operator-start-controls"`).
		AssertContains(`id="operator-start-game"`).
		AssertContains(`disabled`).
		AssertContains(`aria-describedby="role-count-validation"`).
		AssertContains(`id="role-count-validation"`).
		AssertContains("Requires exactly 5 active players")
}

func TestHostDashboardLobby_RoleCountConfigurationRedesign(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "ROLEC",
		State:      game.StateLobby,
		RulesMode:  game.RulesModeCoup,
		CoupPreset: game.CoupPresetFive,
		Players:    make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()

	renderer.Render(HostDashboardLobby(room, host, cfg, &game.CardService{})).
		AssertContains("Role Count Configuration").
		AssertContains(`id="coup-role-counts-list"`).
		AssertContains(`card bg-base-100 border border-base-300 rounded-2xl overflow-hidden`).
		AssertContains(`id="role-row-king"`).
		AssertContains(`id="role-row-blueKnight"`).
		AssertContains(`id="role-row-blackKnight"`).
		AssertContains(`id="role-row-redKnight"`).
		AssertContains(`id="role-row-greenKnight"`).
		AssertContains(`id="role-row-wastelandKnight"`).
		AssertContains(`data-role-count-row="king"`).
		AssertContains(`data-role-count-row="blueKnight"`).
		AssertContains(`data-config-row="role-preset"`).
		AssertContains(`data-stepper-locked="king"`).
		AssertContains(`data-stepper-locked="redKnight"`).
		AssertContains(`badge badge-outline`).
		AssertContains("Locked").
		AssertContains(`data-stepper="blueKnight"`).
		AssertContains(`data-stepper="blackKnight"`).
		AssertContains(`data-stepper="greenKnight"`).
		AssertContains(`data-stepper="wastelandKnight"`).
		AssertContains(`@post(&#39;/room/ROLEC/config/coup-role-count/blueKnight/increment&#39;)`).
		AssertContains(`@post(&#39;/room/ROLEC/config/coup-role-count/blueKnight/decrement&#39;)`).
		AssertContains(`id="role-count-mode-label"`).
		AssertContains("Preset role counts").
		AssertContains(`data-config-row="unsafe-role-count-override"`).
		AssertContains(`class="toggle toggle-warning toggle-sm"`).
		AssertContains("Unsafe Role Count Override").
		AssertContains(`data-config-row="king-to-blue"`).
		AssertContains(`data-config-row="red-to-black"`).
		AssertContains("Rules Variants").
		AssertContains(`id="coup-rules-variants"`).
		AssertContains(`data-config-row="royal-guard"`).
		AssertContains(`data-config-row="inquisition-result"`).
		AssertContains(`select select-bordered select-sm w-auto`).
		AssertContains(`collapse collapse-arrow`).
		AssertContains(`id="role-accordion-blueKnight"`).
		AssertContains(`collapse-title font-bold flex items-center gap-4`).
		AssertContains(`btn btn-primary relative z-20`).
		AssertNotContains(`btn btn-square btn-sm btn-neutral`).
		AssertNotContains(`<summary class="collapse-title`).
		AssertNotContains(`aria-hidden="true">v</span>`).
		AssertNotContains(`onclick="event.stopPropagation(); const input`).
		AssertNotContains(`id="role-count-required"`).
		AssertNotContains(`id="role-count-flexible"`).
		AssertNotContains(`sm:grid-cols-2`)
}

func TestHostDashboardLobby_CoupUnsafeOverrideMakesRequiredRolesEditable(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:                      "UNSAF",
		State:                     game.StateLobby,
		RulesMode:                 game.RulesModeCoup,
		CoupPreset:                game.CoupPresetFive,
		CoupRoleCountsCustom:      true,
		CoupAllowUnsafeRoleCounts: true,
		CoupRoleCounts: game.CoupRoleCounts{
			game.RoleKing:        0,
			game.RoleBlueKnight:  1,
			game.RoleBlackKnight: 1,
			game.RoleRedKnight:   2,
			game.RoleGreenKnight: 1,
			game.RoleWasteland:   0,
		},
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()

	renderer.Render(HostDashboardLobby(room, host, cfg, &game.CardService{})).
		AssertContains("Unsafe Role Count Override").
		AssertContains("Unsafe custom counts").
		AssertContains(`data-stepper="king"`).
		AssertContains(`id="role-row-king"`).
		AssertContains(`@post(&#39;/room/UNSAF/config/coup-role-count/king/increment&#39;)`).
		AssertContains(`data-stepper="redKnight"`).
		AssertContains(`id="role-row-redKnight"`).
		AssertContains("Override").
		AssertNotContains(`data-stepper-locked="king"`).
		AssertNotContains(`data-stepper-locked="redKnight"`)
}

func TestHostDashboardLobby_TreacheryModeUsesUnifiedRoleCountConfiguration(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	cfg := config.DefaultConfig()
	roleConfig, err := game.NewRoleConfigService(cfg).CreateFromPreset("standard", 5)
	if err != nil {
		t.Fatalf("create role config: %v", err)
	}
	room := &game.Room{
		Code:       "TRCHY",
		State:      game.StateLobby,
		RulesMode:  game.RulesModeTreachery,
		RoleConfig: roleConfig,
		Players:    make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	room.Players[host.ID] = host

	renderer.Render(HostDashboardLobby(room, host, cfg, &game.CardService{})).
		AssertContains("Role Count Configuration").
		AssertContains(`data-config-row="player-count"`).
		AssertContains(`data-config-row="role-preset"`).
		AssertContains("Role Counts").
		AssertContains("Rules Variants").
		AssertContains(`id="treachery-rules-variants"`).
		AssertContains(`data-config-row="allow-leaderless"`).
		AssertContains(`data-config-row="hide-role-distribution"`).
		AssertContains(`data-config-row="fully-random-roles"`).
		AssertContains("Allow Leaderless Games").
		AssertContains("Hide Role Distribution").
		AssertContains("Fully Random Roles").
		AssertContains("Leaders").
		AssertContains("Guardians").
		AssertContains("Assassins").
		AssertContains("Traitors").
		AssertNotContains("Coup Preset").
		AssertNotContains("King-to-Blue Info").
		AssertNotContains("Advanced Options")
}

func TestHostDashboardLobby_DebugControlSurfaceRequiresRoomOperator(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	legacyHost := &game.Player{
		ID:        "host",
		Name:      "Legacy Host",
		IsHost:    true,
		SessionID: "session-host",
	}
	room.Players[legacyHost.ID] = legacyHost
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, legacyHost, cfg, nil)).
		AssertNotContains(`id="debug-control-surface"`).
		AssertNotContains("Debug Control Surface").
		AssertNotContains("Debug Insights").
		AssertNotContains(`id="debug-clear"`).
		AssertNotContains("setupDebugPanel")
}

func TestGamePage_DoesNotRenderDebugControlSurface(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StatePlaying,
		Players: make(map[string]*game.Player),
	}
	player := &game.Player{
		ID:   "player",
		Name: "Player",
	}
	room.Players[player.ID] = player

	renderer.Render(GamePage(room, player)).
		AssertNotContains(`id="debug-control-surface"`).
		AssertNotContains("Debug Control Surface").
		AssertNotContains("Debug Insights").
		AssertNotContains(`id="debug-clear"`).
		AssertNotContains("setupDebugPanel")
}

func TestHostDashboardLobby_DebugPlayersAreVisiblyMarked(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "ROOM1",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	debugPlayer := &game.Player{
		ID:      "debug-1",
		Name:    "Debug Player 1",
		IsDebug: true,
	}
	room.Players[host.ID] = host
	room.Players[debugPlayer.ID] = debugPlayer
	cfg := config.DefaultConfig()

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains("Debug Player 1").
		AssertContains(`<span class="badge badge-warning badge-sm">Debug</span>`)
}

func TestHostDashboardLobby_DebugInsightsShowRepresentativeCoupState(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	kingCard := mockCoupCard(1001, "King")
	kingCard.Rulings = []string{"Private information: Blue Knights: Blue Player"}
	room := &game.Room{
		Code:                            "DBG01",
		State:                           game.StatePlaying,
		RulesMode:                       game.RulesModeCoup,
		CoupPreset:                      game.CoupPresetFive,
		DebugStartMode:                  game.DebugStartModeAsIs,
		CoupKingFallen:                  true,
		CoupGreenEligibleBeforeKingFall: true,
		CoupInquisition: &game.CoupInquisitionState{
			Succeeded: true,
			Last: &game.CoupInquisitionAttempt{
				InquisitorID: "blue",
				TargetID:     "red",
				Success:      true,
			},
		},
		CoupWin: &game.CoupWinState{
			Confirmed: &game.CoupWinPrompt{
				Outcome: game.CoupWinOutcomeBlack,
				Facts:   []string{"King has fallen"},
			},
		},
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	king := &game.Player{ID: "king", Name: "King Player", Role: kingCard, RoleRevealed: true, FaceUp: true}
	blue := &game.Player{ID: "blue", Name: "Blue Player", Role: mockCoupCard(1002, "Blue Knight")}
	black := &game.Player{ID: "black", Name: "Black Player", Role: mockCoupCard(1003, "Black Knight"), IsEliminated: true, RoleRevealed: true}
	red := &game.Player{ID: "red", Name: "Red Player", Role: mockCoupCard(1004, "Red Knight"), RoleRevealed: true}
	debugPlayer := &game.Player{ID: "debug-1", Name: "Debug Player 1", Role: mockCoupCard(1005, "Green Knight"), IsDebug: true}
	for _, player := range []*game.Player{host, king, blue, black, red, debugPlayer} {
		room.Players[player.ID] = player
	}
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains("Debug Insights").
		AssertContains("Rules Mode: coup").
		AssertContains("Coup Preset: coup-5").
		AssertContains("Debug Start: as-is").
		AssertContains("King Player").
		AssertContains("Role:").
		AssertContains("Hidden role").
		AssertContains("Revealed: yes").
		AssertContains("Black Player").
		AssertContains("Eliminated: yes").
		AssertContains("Debug Player 1").
		AssertContains("Debug Player: yes").
		AssertContains("Private information hidden").
		AssertContains("King Fallen: yes").
		AssertContains("Green Red-Share Lock: eligible").
		AssertContains("Inquisition: succeeded").
		AssertContains("Advisory Win: black")
}

func TestHostDashboardLobby_DebugInsightsAreRedactedRoleColoredAndClickable(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:                "DBG10",
		State:               game.StatePlaying,
		RulesMode:           game.RulesModeCoup,
		CoupPreset:          game.CoupPresetNine,
		Players:             make(map[string]*game.Player),
		DebugViewedPlayerID: "debug-1",
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	players := []*game.Player{
		host,
		{ID: "king", Name: "King Player", Role: mockCoupCard(1001, "King")},
		{ID: "blue", Name: "Blue Player", Role: mockCoupCard(1002, "Blue Knight")},
		{ID: "black", Name: "Black Player", Role: mockCoupCard(1003, "Black Knight")},
		{ID: "red", Name: "Red Player", Role: mockCoupCard(1004, "Red Knight")},
		{ID: "green", Name: "Green Player", Role: mockCoupCard(1005, "Green Knight")},
		{ID: "wasteland", Name: "Wasteland Player", Role: mockCoupCard(1006, "Wasteland Knight")},
		{ID: "debug-1", Name: "Debug Player 1", Role: mockCoupCard(1005, "Green Knight"), IsDebug: true},
	}
	for _, player := range players {
		room.Players[player.ID] = player
	}
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	html := renderer.Render(HostDashboardLobby(room, host, cfg, nil)).GetHTML()
	for _, expected := range []string{
		`data-debug-spoiler="role"`,
		`data-debug-role-accent="gold"`,
		`data-debug-role-accent="blue"`,
		`data-debug-role-accent="black"`,
		`data-debug-role-accent="red"`,
		`data-debug-role-accent="green"`,
		`data-debug-role-accent="gray"`,
		`data-class:hidden="!$_showDebugSpoilers"`,
		`id="debug-insight-player-debug-1"`,
		`data-debug-viewed-player="true"`,
		"Debug Player 1",
		"Debug Player: yes",
		`fetch(&#39;/room/DBG10/debug/view-as/blue&#39;`,
		`fetch(&#39;/room/DBG10/debug/view-as/debug-1&#39;`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected debug insight detail %q in HTML: %s", expected, html)
		}
	}
	for _, forbidden := range []string{
		"Role: King",
		"Role: Blue Knight",
		"border-color:#",
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("expected hidden role/color detail %q to be suppressed by default: %s", forbidden, html)
		}
	}
}

func TestHostDashboardLobby_DebugInsightsDoNotLeakToNonOperator(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:      "DBG11",
		State:     game.StatePlaying,
		RulesMode: game.RulesModeCoup,
		Players:   make(map[string]*game.Player),
	}
	operator := &game.Player{ID: "operator", Name: "Operator", SessionID: "session-operator"}
	viewer := &game.Player{ID: "viewer", Name: "Viewer", SessionID: "session-viewer", Role: mockCoupCard(1004, "Red Knight")}
	room.Players[operator.ID] = operator
	room.Players[viewer.ID] = viewer
	room.OperatorSessionID = operator.SessionID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, viewer, cfg, nil)).
		AssertNotContains(`id="debug-control-surface"`).
		AssertNotContains(`data-debug-role-accent=`).
		AssertNotContains(`/debug/view-as/`)
}

func TestHostDashboardLobby_DebugInsightsShowGreenRedShareLockPendingBeforeKingFall(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:       "DBG02",
		State:      game.StatePlaying,
		RulesMode:  game.RulesModeCoup,
		CoupPreset: game.CoupPresetFive,
		Players:    make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	room.Players[host.ID] = host
	room.OperatorSessionID = host.SessionID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains("Green Red-Share Lock: pending").
		AssertNotContains("Green Eligible Before King Fall")
}

func TestHostDashboardLobby_ViewAsPlayerSelectorIncludesDebugPlayers(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "VIEW1",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true, SessionID: "session-host"}
	realPlayer := &game.Player{ID: "player-1", Name: "Real Player"}
	debugPlayer := &game.Player{ID: "debug-1", Name: "Debug Player 1", IsDebug: true}
	for _, player := range []*game.Player{host, realPlayer, debugPlayer} {
		room.Players[player.ID] = player
	}
	room.OperatorSessionID = host.SessionID
	room.DebugViewedPlayerID = debugPlayer.ID
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains(`id="debug-operator-view"`).
		AssertContains("Operator View").
		AssertContains(`fetch(&#39;/room/VIEW1/debug/operator-view&#39;`).
		AssertNotContains(`@get(&#39;/room/VIEW1/debug/operator-view&#39;)`).
		AssertContains(`id="debug-view-as-player-select"`).
		AssertContains(`value="player-1"`).
		AssertContains("Real Player").
		AssertContains(`value="debug-1"`).
		AssertContains(`selected`).
		AssertContains("Debug Player 1").
		AssertContains(`fetch(&#39;/room/VIEW1/debug/view-as/&#39; + encodeURIComponent(evt.target.value)`).
		AssertNotContains(`@get(&#39;/room/VIEW1/debug/view-as/&#39; + evt.target.value)`)
}
