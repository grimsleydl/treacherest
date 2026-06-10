package pages

import (
	"testing"
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
