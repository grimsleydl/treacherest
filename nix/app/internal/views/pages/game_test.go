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
		AssertContains("King Player").
		AssertContains(`alt="King"`).
		AssertNotContains("Black Knight").
		AssertNotContains("Red Knight").
		AssertNotContains("Green Knight")
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
