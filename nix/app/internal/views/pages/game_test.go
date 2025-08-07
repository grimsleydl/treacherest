package pages

import (
	"testing"
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
		Text: "Test Leader Card",
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
		Text: "Test Guardian Card",
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
		Text: "Test Assassin Card",
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
			AssertContains("Your Role: Villager").
			AssertHasElementWithID("game-container").
			AssertContains(`data-on-load="@get(&#39;/sse/game/GAME1&#39;)"`)
	})

	t.Run("shows player role", func(t *testing.T) {
		component := GamePage(room, player)

		renderer.Render(component).
			AssertContains("Your Role: Villager").
			AssertContains("A regular villager trying to survive")
	})

	t.Run("shows countdown state", func(t *testing.T) {
		room.State = game.StateCountdown
		room.CountdownRemaining = 3

		component := GamePage(room, player)

		renderer.Render(component).
			AssertContains("Revealing roles in 3...")
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
			AssertContains("Your Role: Assassin")
	})

	t.Run("shows win condition", func(t *testing.T) {
		component := GameBody(room, player)

		renderer.Render(component).
			AssertContains("Win Condition:").
			AssertContains("Win if the Leader is eliminated")
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
			AssertNotContains("Guardian") // Role should not be shown
	})

	t.Run("shows role class styling", func(t *testing.T) {
		component := GameBody(room, player)

		renderer.Render(component).
			AssertHasClass("role-card").
			AssertHasClass("assassin")
	})
}
