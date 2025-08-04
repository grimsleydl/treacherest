package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckboxHandlersReturnSSE(t *testing.T) {
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
			MinPlayers:           3,
			MaxPlayers:           8,
			AllowLeaderlessGame:  false,
			HideRoleDistribution: false,
			FullyRandomRoles:     false,
			PresetName:           "custom",
			RoleTypes: map[string]*game.RoleTypeConfig{
				"Leader":   {Count: 1, EnabledCards: map[string]bool{"The Commander": true}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"The Bodyguard": true}},
				"Assassin": {Count: 1, EnabledCards: map[string]bool{"The Assassin": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"The Cultist": true}},
			},
		},
	}
	memStore.UpdateRoom(room)

	tests := []struct {
		name     string
		endpoint string
		body     interface{}
		checkSSE func(t *testing.T, response string)
	}{
		{
			name:     "UpdateLeaderlessGame returns SSE with reset signal",
			endpoint: "/room/TEST123/config/leaderless",
			body:     map[string]interface{}{"allowLeaderless": true},
			checkSSE: func(t *testing.T, response string) {
				// Should contain SSE events
				assert.Contains(t, response, "event: datastar-merge-fragments")
				assert.Contains(t, response, "data: fragments")

				// Should reset the loading state
				assert.Contains(t, response, "event: datastar-merge-signals")
				assert.Contains(t, response, `"updatingLeaderless":false`)
			},
		},
		{
			name:     "UpdateHideDistribution returns SSE with reset signal",
			endpoint: "/room/TEST123/config/hide-distribution",
			body:     map[string]interface{}{"hideRoleDistribution": true},
			checkSSE: func(t *testing.T, response string) {
				// Should contain SSE events
				assert.Contains(t, response, "event: datastar-merge-fragments")
				assert.Contains(t, response, "data: fragments")

				// Should reset the loading state
				assert.Contains(t, response, "event: datastar-merge-signals")
				assert.Contains(t, response, `"updatingHideDistribution":false`)
			},
		},
		{
			name:     "UpdateFullyRandom returns SSE with reset signal",
			endpoint: "/room/TEST123/config/fully-random",
			body:     map[string]interface{}{"fullyRandomRoles": true},
			checkSSE: func(t *testing.T, response string) {
				// Should contain SSE events
				assert.Contains(t, response, "event: datastar-merge-fragments")
				assert.Contains(t, response, "data: fragments")

				// Should reset the loading state
				assert.Contains(t, response, "event: datastar-merge-signals")
				assert.Contains(t, response, `"updatingFullyRandom":false`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", tt.endpoint, bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "player_TEST123", Value: "player1"})

			// Add chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", "TEST123")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the appropriate handler
			if strings.Contains(tt.endpoint, "leaderless") {
				handler.UpdateLeaderlessGame(w, req)
			} else if strings.Contains(tt.endpoint, "hide-distribution") {
				handler.UpdateHideDistribution(w, req)
			} else if strings.Contains(tt.endpoint, "fully-random") {
				handler.UpdateFullyRandom(w, req)
			}

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")

			// Check SSE content
			responseBody := w.Body.String()
			require.NotEmpty(t, responseBody)
			tt.checkSSE(t, responseBody)
		})
	}
}
