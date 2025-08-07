package views

import (
	"bytes"
	"context"
	"testing"
	"treacherest/internal/game"
	"treacherest/internal/views/layouts"
	"treacherest/internal/views/pages"
)

// Test that templates can render without panicking
func TestTemplateRendering(t *testing.T) {
	// Create test data
	room := &game.Room{
		Code:       "TEST1",
		State:      game.StateLobby,
		Players:    make(map[string]*game.Player),
		MaxPlayers: 8,
	}

	player := &game.Player{
		ID:   "p1",
		Name: "Test Player",
		Role: &game.Card{
			ID:   1,
			Name: "Villager",
			Text: "A regular villager",
			Types: game.CardTypes{
				Supertype: "Creature",
				Subtype:   "Guardian",
			},
		},
	}

	room.Players[player.ID] = player

	ctx := context.Background()

	t.Run("Base layout renders", func(t *testing.T) {
		buf := &bytes.Buffer{}
		component := layouts.Base("Test Title")

		// This will panic if there's an issue
		err := component.Render(ctx, buf)
		if err != nil {
			t.Errorf("Base template failed to render: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Base template rendered empty content")
		}
	})

	t.Run("Home page renders", func(t *testing.T) {
		buf := &bytes.Buffer{}
		component := pages.Home()

		err := component.Render(ctx, buf)
		if err != nil {
			t.Errorf("Home template failed to render: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Home template rendered empty content")
		}
	})

	t.Run("LobbyPage renders", func(t *testing.T) {
		buf := &bytes.Buffer{}
		component := pages.LobbyPage(room, player)

		err := component.Render(ctx, buf)
		if err != nil {
			t.Errorf("LobbyPage template failed to render: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("LobbyPage template rendered empty content")
		}
	})

	t.Run("LobbyBody renders", func(t *testing.T) {
		buf := &bytes.Buffer{}
		component := pages.LobbyBody(room, player)

		err := component.Render(ctx, buf)
		if err != nil {
			t.Errorf("LobbyBody template failed to render: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("LobbyBody template rendered empty content")
		}
	})

	t.Run("GamePage renders", func(t *testing.T) {
		buf := &bytes.Buffer{}
		component := pages.GamePage(room, player)

		err := component.Render(ctx, buf)
		if err != nil {
			t.Errorf("GamePage template failed to render: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("GamePage template rendered empty content")
		}
	})

	t.Run("GameBody renders", func(t *testing.T) {
		buf := &bytes.Buffer{}
		component := pages.GameBody(room, player)

		err := component.Render(ctx, buf)
		if err != nil {
			t.Errorf("GameBody template failed to render: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("GameBody template rendered empty content")
		}
	})
}
