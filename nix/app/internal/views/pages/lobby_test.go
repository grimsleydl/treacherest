package pages

import (
	"testing"
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
		MaxPlayers: 8,
	}
	
	player1 := &game.Player{
		ID:   "p1",
		Name: "Player One",
	}
	
	player2 := &game.Player{
		ID:   "p2",
		Name: "Player Two",
	}
	
	room.Players[player1.ID] = player1
	room.Players[player2.ID] = player2

	t.Run("renders lobby page structure", func(t *testing.T) {
		component := LobbyPage(room, player1)
		
		renderer.Render(component).
			AssertNotEmpty().
			AssertValid().
			AssertContains("TEST1").
			AssertHasElementWithID("lobby-container").
			AssertContains("Game Lobby")
	})

	t.Run("shows SSE connection", func(t *testing.T) {
		component := LobbyPage(room, player1)
		
		renderer.Render(component).
			AssertContains(`data-on-load="@get(&#39;/sse/lobby/TEST1&#39;)"`)
	})

	t.Run("renders player list", func(t *testing.T) {
		component := LobbyPage(room, player1)
		
		renderer.Render(component).
			AssertContains("Players (2/8)").
			AssertContains("Player One").
			AssertContains("Player Two")
	})

	t.Run("shows need more players message", func(t *testing.T) {
		component := LobbyPage(room, player1)
		
		renderer.Render(component).
			AssertContains("Need at least 4 players to start")
	})

	t.Run("shows start button when enough players", func(t *testing.T) {
		// Add more players to meet minimum
		room.Players["p3"] = &game.Player{ID: "p3", Name: "Player Three"}
		room.Players["p4"] = &game.Player{ID: "p4", Name: "Player Four"}
		
		component := LobbyPage(room, player1)
		
		renderer.Render(component).
			AssertContains("Start Game").
			AssertContains(`data-on-click="@post(&#39;/room/TEST1/start&#39;)"`)
	})

	t.Run("shows leave button", func(t *testing.T) {
		component := LobbyPage(room, player1)
		
		renderer.Render(component).
			AssertContains("Leave Room").
			AssertContains(`data-on-click="@post(&#39;/room/TEST1/leave&#39;)"`)
	})
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
		ID:   "p1",
		Name: "Test Player",
	}
	
	room.Players[player.ID] = player

	t.Run("renders lobby body fragment", func(t *testing.T) {
		component := LobbyBody(room, player)
		
		renderer.Render(component).
			AssertNotEmpty().
			AssertHasElementWithID("lobby-container").
			AssertContains("Players (1/4)")
	})

	t.Run("shows minimum players message", func(t *testing.T) {
		component := LobbyBody(room, player)
		
		renderer.Render(component).
			AssertContains("Need at least 4 players to start")
	})

	t.Run("enables start button when enough players", func(t *testing.T) {
		// Add more players
		for i := 2; i <= 4; i++ {
			p := &game.Player{
				ID:   string(rune('p' + i)),
				Name: "Player " + string(rune('0' + i)),
			}
			room.Players[p.ID] = p
		}
		
		component := LobbyBody(room, player)
		
		renderer.Render(component).
			AssertNotContains("Need at least").
			AssertContains("Start Game").
			AssertNotContains("disabled")
	})

}