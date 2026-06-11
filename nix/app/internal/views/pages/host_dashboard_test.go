package pages

import (
	"testing"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/testhelpers"
)

func TestHostDashboardPlaying_CoupModeratorControls(t *testing.T) {
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
	room.Players[host.ID] = host
	room.Players[hidden.ID] = hidden

	renderer.Render(HostDashboardPlaying(room, host)).
		AssertContains("Blue Player").
		AssertContains("Record Reveal").
		AssertContains(`@post(&#39;/room/HOST1/reveal/p1&#39;)`).
		AssertContains("Record Elimination").
		AssertContains(`@post(&#39;/room/HOST1/player/p1/eliminate&#39;)`).
		AssertNotContains("Blue Knight")
}

func TestHostDashboardPlaying_CoupAdvisoryWinControls(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room, _ := makeCoupWinViewRoom()
	host := &game.Player{
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	room.Players[host.ID] = host

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
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	room.Players[host.ID] = host

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
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	room.Players[host.ID] = host
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains(`id="debug-control-surface"`).
		AssertContains("Debug Control Surface").
		AssertContains(`id="debug-persistence-controls"`).
		AssertContains(`id="debug-start-override-controls"`).
		AssertContains(`id="debug-start-with-debug-players"`).
		AssertContains(`@post(&#39;/room/DEBUG/debug/start-with-debug-players&#39;)`).
		AssertContains(`id="debug-start-as-is"`).
		AssertContains(`@post(&#39;/room/DEBUG/debug/start-as-is&#39;)`).
		AssertContains(`id="debug-insights-container"`).
		AssertContains(`id="debug-view-as-player-container"`).
		AssertContains(`id="debug-dump"`).
		AssertContains(`id="debug-clear"`).
		AssertContains(`id="debug-restore"`)
}

func TestHostDashboardLobby_DebugControlSurfaceCanBeMinimized(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:     "host",
		Name:   "Host",
		IsHost: true,
	}
	room.Players[host.ID] = host
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains(`id="debug-panel-toggle"`).
		AssertContains(`aria-controls="debug-panel"`).
		AssertContains(`aria-expanded="true"`).
		AssertContains("Minimize").
		AssertContains(`id="debug-panel-minimized"`).
		AssertContains("Debug Mode minimized").
		AssertContains("treacherest_debug_panel_minimized_").
		AssertContains("setDebugPanelMinimized")
}

func TestHostDashboardLobby_DebugControlSurfaceRequiresHostPlayer(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "DEBUG",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	nonHost := &game.Player{
		ID:   "player",
		Name: "Player",
	}
	room.Players[nonHost.ID] = nonHost
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, nonHost, cfg, nil)).
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
	host := &game.Player{ID: "host", Name: "Host", IsHost: true}
	king := &game.Player{ID: "king", Name: "King Player", Role: kingCard, RoleRevealed: true, FaceUp: true}
	blue := &game.Player{ID: "blue", Name: "Blue Player", Role: mockCoupCard(1002, "Blue Knight")}
	black := &game.Player{ID: "black", Name: "Black Player", Role: mockCoupCard(1003, "Black Knight"), IsEliminated: true, RoleRevealed: true}
	red := &game.Player{ID: "red", Name: "Red Player", Role: mockCoupCard(1004, "Red Knight"), RoleRevealed: true}
	debugPlayer := &game.Player{ID: "debug-1", Name: "Debug Player 1", Role: mockCoupCard(1005, "Green Knight"), IsDebug: true}
	for _, player := range []*game.Player{host, king, blue, black, red, debugPlayer} {
		room.Players[player.ID] = player
	}
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains("Debug Insights").
		AssertContains("Rules Mode: coup").
		AssertContains("Coup Preset: coup-5").
		AssertContains("Debug Start: as-is").
		AssertContains("King Player").
		AssertContains("Role: King").
		AssertContains("Revealed: yes").
		AssertContains("Black Player").
		AssertContains("Eliminated: yes").
		AssertContains("Debug Player 1").
		AssertContains("Debug Player: yes").
		AssertContains("Private information: Blue Knights: Blue Player").
		AssertContains("King Fallen: yes").
		AssertContains("Green Eligible Before King Fall: yes").
		AssertContains("Inquisition: succeeded").
		AssertContains("Advisory Win: black")
}

func TestHostDashboardLobby_ViewAsPlayerSelectorIncludesDebugPlayers(t *testing.T) {
	renderer := testhelpers.NewTemplateRenderer(t)
	room := &game.Room{
		Code:    "VIEW1",
		State:   game.StateLobby,
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{ID: "host", Name: "Host", IsHost: true}
	realPlayer := &game.Player{ID: "player-1", Name: "Real Player"}
	debugPlayer := &game.Player{ID: "debug-1", Name: "Debug Player 1", IsDebug: true}
	for _, player := range []*game.Player{host, realPlayer, debugPlayer} {
		room.Players[player.ID] = player
	}
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true

	renderer.Render(HostDashboardLobby(room, host, cfg, nil)).
		AssertContains(`id="debug-view-as-player-select"`).
		AssertContains(`value="player-1"`).
		AssertContains("Real Player").
		AssertContains(`value="debug-1"`).
		AssertContains("Debug Player 1").
		AssertContains(`@get(&#39;/room/VIEW1/debug/view-as/&#39; + evt.target.value)`)
}
