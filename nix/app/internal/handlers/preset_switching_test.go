package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

func TestPresetSwitchingLeaderCount(t *testing.T) {
	// Create test config with multiple presets
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName: "Leader",
					Category:    "Leader",
					MinCount:    1,
					MaxCount:    1,
				},
				"guardian": {
					DisplayName: "Guardian",
					Category:    "Guardian",
					MinCount:    0,
					MaxCount:    10,
				},
				"assassin": {
					DisplayName: "Assassin",
					Category:    "Assassin",
					MinCount:    0,
					MaxCount:    10,
				},
				"traitor": {
					DisplayName: "Traitor",
					Category:    "Traitor",
					MinCount:    0,
					MaxCount:    10,
				},
			},
			Presets: map[string]config.Preset{
				"standard": {
					Name: "Standard",
					Distributions: map[int]map[string]int{
						5: {"leader": 1, "guardian": 2, "assassin": 1, "traitor": 1},
					},
				},
				"simple": {
					Name: "Simple",
					Distributions: map[int]map[string]int{
						3: {"leader": 1, "guardian": 1, "assassin": 1},
					},
				},
			},
		},
	}

	// Create handler
	s := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	s.SetCardService(cardService)
	h := New(s, cardService, cfg)

	// Create a room with standard preset
	room := &game.Room{
		Code:       "TEST1",
		MaxPlayers: 8,
		Players:    make(map[string]*game.Player),
		State:      game.StateLobby,
	}
	
	// Initialize with standard preset
	roleService := game.NewRoleConfigService(cfg)
	roleService.SetCardService(cardService)
	room.RoleConfig, _ = roleService.CreateFromPreset("standard", room.MaxPlayers)
	
	player := &game.Player{
		ID:       "player1",
		Name:     "Test Player",
		IsHost:   true,
		JoinedAt: time.Now(),
	}
	room.Players[player.ID] = player
	s.UpdateRoom(room)

	// Verify initial leader count
	t.Run("initial leader count", func(t *testing.T) {
		if room.RoleConfig.RoleTypes["Leader"] == nil || room.RoleConfig.RoleTypes["Leader"].Count != 1 {
			t.Errorf("Initial leader count should be 1, got %v", room.RoleConfig.RoleTypes["Leader"])
		}
	})

	// Switch to simple preset
	t.Run("switch to simple preset", func(t *testing.T) {
		form := url.Values{}
		form.Add("preset", "simple")
		req := httptest.NewRequest("POST", "/room/TEST1/config/preset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "player_TEST1", Value: player.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.UpdateRolePreset(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check leader count is still 1
		updatedRoom, _ := s.GetRoom("TEST1")
		if updatedRoom.RoleConfig.RoleTypes["Leader"] == nil || updatedRoom.RoleConfig.RoleTypes["Leader"].Count != 1 {
			t.Errorf("Leader count should remain 1 after preset switch, got %v", updatedRoom.RoleConfig.RoleTypes["Leader"])
		}
	})

	// Toggle a role to switch to custom mode
	t.Run("switch to custom mode", func(t *testing.T) {
		form := url.Values{}
		form.Add("role-traitor", "on")
		req := httptest.NewRequest("POST", "/room/TEST1/config/toggle", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "player_TEST1", Value: player.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.ToggleRole(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check leader count is still 1
		updatedRoom, _ := s.GetRoom("TEST1")
		if updatedRoom.RoleConfig.RoleTypes["Leader"] == nil || updatedRoom.RoleConfig.RoleTypes["Leader"].Count != 1 {
			t.Errorf("Leader count should remain 1 in custom mode, got %v", updatedRoom.RoleConfig.RoleTypes["Leader"])
		}
		
		// Check preset is now custom
		if updatedRoom.RoleConfig.PresetName != "custom" {
			t.Errorf("Preset should be 'custom' after toggling role, got %s", updatedRoom.RoleConfig.PresetName)
		}
		
		// Check traitor was enabled with proper count
		if updatedRoom.RoleConfig.RoleTypes["Traitor"] == nil || updatedRoom.RoleConfig.RoleTypes["Traitor"].Count != 1 {
			t.Errorf("Traitor count should be 1, got %v", updatedRoom.RoleConfig.RoleTypes["Traitor"])
		}
	})

	// Switch back to standard preset
	t.Run("switch back to standard preset", func(t *testing.T) {
		form := url.Values{}
		form.Add("preset", "standard")
		req := httptest.NewRequest("POST", "/room/TEST1/config/preset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "player_TEST1", Value: player.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "TEST1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.UpdateRolePreset(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check leader count is still 1
		updatedRoom, _ := s.GetRoom("TEST1")
		if updatedRoom.RoleConfig.RoleTypes["Leader"] == nil || updatedRoom.RoleConfig.RoleTypes["Leader"].Count != 1 {
			t.Errorf("Leader count should remain 1 after switching back to preset, got %v", updatedRoom.RoleConfig.RoleTypes["Leader"])
		}
		
		// Check all role types have proper counts
		for typeName, typeConfig := range updatedRoom.RoleConfig.RoleTypes {
			if typeConfig.Count > 0 {
				// At least one card should be enabled for non-zero count
				hasEnabled := false
				for _, enabled := range typeConfig.EnabledCards {
					if enabled {
						hasEnabled = true
						break
					}
				}
				if !hasEnabled {
					t.Errorf("Role type %s has count %d but no enabled cards", typeName, typeConfig.Count)
				}
			}
		}
	})
}

func TestCustomModeRoleCountInit(t *testing.T) {
	// Create test config
	cfg := config.DefaultConfig()

	// Create handler
	s := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	s.SetCardService(cardService)
	h := New(s, cardService, cfg)

	// Create a room starting in custom mode
	room := &game.Room{
		Code:       "CUSTOM1",
		MaxPlayers: 8,
		Players:    make(map[string]*game.Player),
		State:      game.StateLobby,
		RoleConfig: &game.RoleConfiguration{
			PresetName: "custom",
			MinPlayers: 1,
			MaxPlayers: 8,
			RoleTypes:  make(map[string]*game.RoleTypeConfig),
		},
	}
	
	player := &game.Player{
		ID:       "player1",
		Name:     "Test Player",
		IsHost:   true,
		JoinedAt: time.Now(),
	}
	room.Players[player.ID] = player
	s.UpdateRoom(room)

	// Enable leader role
	t.Run("enable leader role", func(t *testing.T) {
		form := url.Values{}
		form.Add("role-leader", "on")
		req := httptest.NewRequest("POST", "/room/CUSTOM1/config/toggle", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "player_CUSTOM1", Value: player.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "CUSTOM1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.ToggleRole(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check leader was enabled with count 1
		updatedRoom, _ := s.GetRoom("CUSTOM1")
		if updatedRoom.RoleConfig.RoleTypes["Leader"] == nil || updatedRoom.RoleConfig.RoleTypes["Leader"].Count != 1 {
			t.Errorf("Leader count should be 1 (MinCount), got %v", updatedRoom.RoleConfig.RoleTypes["Leader"])
		}
	})

	// Enable guardian role (MinCount = 0)
	t.Run("enable guardian role", func(t *testing.T) {
		form := url.Values{}
		form.Add("role-guardian", "on")
		req := httptest.NewRequest("POST", "/room/CUSTOM1/config/toggle", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "player_CUSTOM1", Value: player.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "CUSTOM1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.ToggleRole(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check guardian was enabled with count 1 (default for MinCount = 0)
		updatedRoom, _ := s.GetRoom("CUSTOM1")
		if updatedRoom.RoleConfig.RoleTypes["Guardian"] == nil || updatedRoom.RoleConfig.RoleTypes["Guardian"].Count != 1 {
			t.Errorf("Guardian count should be 1 (default), got %v", updatedRoom.RoleConfig.RoleTypes["Guardian"])
		}
	})
}