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
		AssertNotContains(`id="debug-panel"`).
		AssertNotContains(`id="debug-clear"`).
		AssertNotContains("setupDebugPanel")

	enabled := config.DefaultConfig()
	enabled.Server.DebugModeEnabled = true
	renderer.Render(HostDashboardLobby(room, host, enabled, nil)).
		AssertContains(`id="debug-panel"`).
		AssertContains(`id="debug-clear"`).
		AssertContains("setupDebugPanel")
}
