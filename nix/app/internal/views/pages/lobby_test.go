package pages

import (
	"testing"
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
		ID:   "p1",
		Name: "Player One",
	}

	player2 := &game.Player{
		ID:   "p2",
		Name: "Player Two",
	}

	room.Players[player1.ID] = player1
	room.Players[player2.ID] = player2

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
			AssertContains("Game Lobby")
	})

	t.Run("shows SSE connection", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains(`data-on-load="@get(&#39;/sse/lobby/TEST1&#39;)"`)
	})

	t.Run("renders player list", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Players (2)").
			AssertContains("Player One").
			AssertContains("Player Two")
	})

	t.Run("shows need more players message", func(t *testing.T) {
		// Create empty room to test minimum message
		emptyRoom := &game.Room{
			Code:       "EMPTY",
			State:      game.StateLobby,
			Players:    make(map[string]*game.Player),
			MaxPlayers: 4,
		}

		component := LobbyPage(emptyRoom, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Need at least 1 player to start")
	})

	t.Run("shows start button when enough players", func(t *testing.T) {
		// Room already has 2 players (player1 and player2), which is >= 1
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Start Game").
			AssertContains(`@post(&#39;/room/TEST1/start&#39;)`)
	})

	t.Run("shows leave button", func(t *testing.T) {
		component := LobbyPage(room, player1, cfg, cardService)

		renderer.Render(component).
			AssertContains("Leave Room").
			AssertContains(`@post(&#39;/room/TEST1/leave&#39;)`)
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
			AssertContains("Players (1)")
	})

	t.Run("shows minimum players message", func(t *testing.T) {
		component := LobbyBody(room, player, cfg, cardService)

		renderer.Render(component).
			AssertContains("Start Game")
	})

	t.Run("enables start button when enough players", func(t *testing.T) {
		// Room already has 1 player which is enough to start

		component := LobbyBody(room, player, cfg, cardService)

		renderer.Render(component).
			AssertNotContains("Need at least").
			AssertContains("Start Game").
			AssertContains(`data-attr-disabled="$isStarting || !$canStartGame"`)
	})

}
