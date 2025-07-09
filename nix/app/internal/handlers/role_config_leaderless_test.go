package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateLeaderlessGame(t *testing.T) {
	// Setup
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	memStore.SetCardService(cardService)
	handler := New(memStore, cardService, cfg)

	// Create a test room with a player
	player := &game.Player{
		ID:     "player1",
		Name:   "Test Player",
		IsHost: true,
	}
	room := &game.Room{
		Code:    "TEST123",
		Players: map[string]*game.Player{"player1": player},
		State:   game.StateLobby,
		RoleConfig: &game.RoleConfiguration{
			MinPlayers:          3,
			MaxPlayers:          8,
			AllowLeaderlessGame: false,
			PresetName:          "custom",
			RoleTypes: map[string]*game.RoleTypeConfig{
				"Leader":   {Count: 0, EnabledCards: map[string]bool{}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
			},
		},
	}
	memStore.UpdateRoom(room)

	t.Run("Enable leaderless games", func(t *testing.T) {
		body := map[string]bool{"allowLeaderless": true}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/room/TEST123/config/leaderless", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "player_TEST123", Value: "player1"})

		// Add chi route params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateLeaderlessGame(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check room state
		updatedRoom, _ := memStore.GetRoom("TEST123")
		assert.True(t, updatedRoom.RoleConfig.AllowLeaderlessGame)
		// Leader count should remain 0 when enabling leaderless
		assert.Equal(t, 0, updatedRoom.RoleConfig.RoleTypes["Leader"].Count)
	})

	t.Run("Disable leaderless games with zero leaders", func(t *testing.T) {
		// Reset to leaderless enabled
		room.RoleConfig.AllowLeaderlessGame = true
		room.RoleConfig.RoleTypes["Leader"].Count = 0
		memStore.UpdateRoom(room)

		body := map[string]bool{"allowLeaderless": false}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/room/TEST123/config/leaderless", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "player_TEST123", Value: "player1"})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateLeaderlessGame(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check room state
		updatedRoom, _ := memStore.GetRoom("TEST123")
		assert.False(t, updatedRoom.RoleConfig.AllowLeaderlessGame)
		// Leader count should be set to 1 when disabling leaderless with 0 leaders
		assert.Equal(t, 1, updatedRoom.RoleConfig.RoleTypes["Leader"].Count)
		// Should switch to custom preset
		assert.Equal(t, "custom", updatedRoom.RoleConfig.PresetName)
	})

	t.Run("Disable leaderless games with existing leader", func(t *testing.T) {
		// Reset to leaderless enabled with 1 leader
		room.RoleConfig.AllowLeaderlessGame = true
		room.RoleConfig.RoleTypes["Leader"].Count = 1
		memStore.UpdateRoom(room)

		body := map[string]bool{"allowLeaderless": false}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/room/TEST123/config/leaderless", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "player_TEST123", Value: "player1"})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateLeaderlessGame(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check room state
		updatedRoom, _ := memStore.GetRoom("TEST123")
		assert.False(t, updatedRoom.RoleConfig.AllowLeaderlessGame)
		// Leader count should remain 1 (not changed)
		assert.Equal(t, 1, updatedRoom.RoleConfig.RoleTypes["Leader"].Count)
	})

	t.Run("Rapid toggle protection", func(t *testing.T) {
		// Test rapid toggling doesn't cause issues
		states := []bool{true, false, true, false, true}
		
		for _, state := range states {
			body := map[string]bool{"allowLeaderless": state}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/room/TEST123/config/leaderless", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "player_TEST123", Value: "player1"})

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", "TEST123")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.UpdateLeaderlessGame(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Final state should be true (last in states array)
		updatedRoom, _ := memStore.GetRoom("TEST123")
		assert.True(t, updatedRoom.RoleConfig.AllowLeaderlessGame)
	})

	t.Run("Unauthorized access", func(t *testing.T) {
		body := map[string]bool{"allowLeaderless": true}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/room/TEST123/config/leaderless", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		// No cookie or wrong player

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateLeaderlessGame(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/TEST123/config/leaderless", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "player_TEST123", Value: "player1"})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateLeaderlessGame(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Room not found", func(t *testing.T) {
		body := map[string]bool{"allowLeaderless": true}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/room/NOTFOUND/config/leaderless", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "player_NOTFOUND", Value: "player1"})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "NOTFOUND")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateLeaderlessGame(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUpdateLeaderlessGameConcurrency(t *testing.T) {
	// Test concurrent updates don't cause race conditions
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	memStore.SetCardService(cardService)
	handler := New(memStore, cardService, cfg)

	// Create a test room
	player := &game.Player{
		ID:     "player1",
		Name:   "Test Player",
		IsHost: true,
	}
	room := &game.Room{
		Code:    "CONCURRENT",
		Players: map[string]*game.Player{"player1": player},
		State:   game.StateLobby,
		RoleConfig: &game.RoleConfiguration{
			MinPlayers:          3,
			MaxPlayers:          8,
			AllowLeaderlessGame: false,
			PresetName:          "custom",
			RoleTypes: map[string]*game.RoleTypeConfig{
				"Leader": {Count: 0, EnabledCards: map[string]bool{}},
			},
		},
	}
	memStore.UpdateRoom(room)

	// Run multiple concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(iteration int) {
			body := map[string]bool{"allowed": iteration%2 == 0}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/room/CONCURRENT/config/leaderless", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "player_CONCURRENT", Value: "player1"})

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", "CONCURRENT")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.UpdateLeaderlessGame(w, req)
			
			require.Equal(t, http.StatusOK, w.Code)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
		time.Sleep(10 * time.Millisecond) // Small delay between requests
	}

	// Verify final state is consistent
	updatedRoom, _ := memStore.GetRoom("CONCURRENT")
	assert.NotNil(t, updatedRoom)
	// State should be consistent (either true or false, with appropriate leader count)
	if updatedRoom.RoleConfig.AllowLeaderlessGame {
		// If leaderless is allowed, leader count can be 0 or more
		assert.True(t, updatedRoom.RoleConfig.RoleTypes["Leader"].Count >= 0)
	} else {
		// If leaderless is not allowed, leader count must be at least 1
		assert.True(t, updatedRoom.RoleConfig.RoleTypes["Leader"].Count >= 1)
	}
}