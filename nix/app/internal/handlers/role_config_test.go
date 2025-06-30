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


func TestUpdateRolePreset(t *testing.T) {
	// Create test config
	cfg := config.DefaultConfig()
	cfg.Roles.Presets["test-preset"] = config.Preset{
		Name: "Test Preset",
		Distributions: map[int]map[string]int{
			3: {
				"leader":   1,
				"guardian": 1,
				"traitor":  1,
			},
		},
	}

	// Create handler
	s := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	s.SetCardService(cardService)
	h := New(s, cardService, cfg)

	// Create a test room with a player
	room := &game.Room{
		Code:       "TEST1",
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

	tests := []struct {
		name       string
		preset     string
		playerID   string
		wantStatus int
	}{
		{
			name:       "valid preset update",
			preset:     "test-preset",
			playerID:   "player1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "custom preset",
			preset:     "custom",
			playerID:   "player1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized player",
			preset:     "test-preset",
			playerID:   "unauthorized",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty preset",
			preset:     "",
			playerID:   "player1",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			form := url.Values{}
			form.Add("preset", tt.preset)
			req := httptest.NewRequest("POST", "/room/TEST1/config/preset", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			
			// Set cookie
			req.AddCookie(&http.Cookie{
				Name:  "player_TEST1",
				Value: tt.playerID,
			})

			// Add route params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", "TEST1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			h.UpdateRolePreset(rr, req)

			// Check status
			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			// If successful, verify room was updated
			if tt.wantStatus == http.StatusOK && tt.preset != "custom" {
				updatedRoom, _ := s.GetRoom("TEST1")
				if updatedRoom.RoleConfig.PresetName != tt.preset {
					t.Errorf("preset not updated: got %v want %v", updatedRoom.RoleConfig.PresetName, tt.preset)
				}
			}
		})
	}
}

// Commented out - tests for deprecated ToggleRole handler
/* func TestToggleRole(t *testing.T) {
	// Create test config
	cfg := config.DefaultConfig()

	// Create handler
	s := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	s.SetCardService(cardService)
	h := New(s, cardService, cfg)

	// Create a test room
	room := &game.Room{
		Code:       "TEST2",
		MaxPlayers: 8,
		Players:    make(map[string]*game.Player),
		State:      game.StateLobby,
		RoleConfig: &game.RoleConfiguration{
			PresetName: "custom",
			MinPlayers: 1,
			MaxPlayers: 8,
			RoleTypes: map[string]*game.RoleTypeConfig{
				"Leader": {Count: 1, EnabledCards: map[string]bool{"The Usurper": true}},
			},
		},
	}
	
	player := &game.Player{
		ID:       "player1",
		Name:     "Test Player",
		IsHost:   false,
		JoinedAt: time.Now(),
	}
	room.Players[player.ID] = player
	s.UpdateRoom(room)

	tests := []struct {
		name         string
		roleName     string
		expectEnable bool
	}{
		{
			name:         "enable guardian role",
			roleName:     "guardian",
			expectEnable: true,
		},
		{
			name:         "disable leader role",
			roleName:     "leader",
			expectEnable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get initial state
			initialEnabled := room.RoleConfig.EnabledRoles[tt.roleName]

			// Create request
			form := url.Values{}
			form.Add("role-"+tt.roleName, "on")
			req := httptest.NewRequest("POST", "/room/TEST2/config/toggle", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			
			// Set cookie
			req.AddCookie(&http.Cookie{
				Name:  "player_TEST2",
				Value: "player1",
			})

			// Add route params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", "TEST2")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			h.ToggleRole(rr, req)

			// Check status
			if rr.Code != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}

			// Verify role was toggled
			updatedRoom, _ := s.GetRoom("TEST2")
			if updatedRoom.RoleConfig.EnabledRoles[tt.roleName] == initialEnabled {
				t.Errorf("role %s was not toggled", tt.roleName)
			}

			// Verify preset was set to custom
			if updatedRoom.RoleConfig.PresetName != "custom" {
				t.Error("preset should be set to custom after toggle")
			}
		})
	}
} */

// Commented out - tests for deprecated UpdateRoleCount handler
/* func TestUpdateRoleCount(t *testing.T) {
	// Create test config
	cfg := config.DefaultConfig()

	// Create handler
	s := store.NewMemoryStore(cfg)
	cardService := createMockCardService()
	s.SetCardService(cardService)
	h := New(s, cardService, cfg)

	// Create a test room
	room := &game.Room{
		Code:       "TEST3",
		MaxPlayers: 8,
		Players:    make(map[string]*game.Player),
		State:      game.StateLobby,
		RoleConfig: &game.RoleConfiguration{
			PresetName: "custom",
			EnabledRoles: map[string]bool{
				"guardian": true,
			},
			RoleCounts: map[string]int{
				"guardian": 2,
			},
			MinPlayers: 2,
			MaxPlayers: 8,
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

	tests := []struct {
		name      string
		roleName  string
		count     string
		wantCount int
	}{
		{
			name:      "update guardian count",
			roleName:  "guardian",
			count:     "4",
			wantCount: 4,
		},
		{
			name:      "invalid count defaults to min",
			roleName:  "guardian",
			count:     "-1",
			wantCount: 0, // min count for guardian
		},
		{
			name:      "exceed max count",
			roleName:  "guardian",
			count:     "20",
			wantCount: 10, // max count for guardian
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			form := url.Values{}
			form.Add("count-"+tt.roleName, tt.count)
			req := httptest.NewRequest("POST", "/room/TEST3/config/count", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			
			// Set cookie
			req.AddCookie(&http.Cookie{
				Name:  "player_TEST3",
				Value: "player1",
			})

			// Add route params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", "TEST3")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			h.UpdateRoleCount(rr, req)

			// Check status
			if rr.Code != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}

			// Verify count was updated
			updatedRoom, _ := s.GetRoom("TEST3")
			if updatedRoom.RoleConfig.RoleCounts[tt.roleName] != tt.wantCount {
				t.Errorf("role count not updated correctly: got %v want %v", 
					updatedRoom.RoleConfig.RoleCounts[tt.roleName], tt.wantCount)
			}
		})
	}
} */

func TestIsRoomCreator(t *testing.T) {
	// Create handler
	cfg := config.DefaultConfig()
	s := store.NewMemoryStore(cfg)
	h := New(s, nil, cfg)

	// Test with host mode
	hostRoom := &game.Room{
		Code:    "HOST1",
		Players: make(map[string]*game.Player),
	}
	host := &game.Player{
		ID:       "host1",
		IsHost:   true,
		JoinedAt: time.Now(),
	}
	player1 := &game.Player{
		ID:       "player1",
		IsHost:   false,
		JoinedAt: time.Now().Add(1 * time.Second),
	}
	hostRoom.Players[host.ID] = host
	hostRoom.Players[player1.ID] = player1

	// Test with non-host mode (first player)
	nonHostRoom := &game.Room{
		Code:    "NOHOST1",
		Players: make(map[string]*game.Player),
	}
	firstPlayer := &game.Player{
		ID:       "first1",
		IsHost:   false,
		JoinedAt: time.Now(),
	}
	secondPlayer := &game.Player{
		ID:       "second1",
		IsHost:   false,
		JoinedAt: time.Now().Add(1 * time.Second),
	}
	nonHostRoom.Players[firstPlayer.ID] = firstPlayer
	nonHostRoom.Players[secondPlayer.ID] = secondPlayer

	tests := []struct {
		name     string
		room     *game.Room
		playerID string
		want     bool
	}{
		{
			name:     "host is room creator",
			room:     hostRoom,
			playerID: "host1",
			want:     true,
		},
		{
			name:     "non-host is not room creator when host exists",
			room:     hostRoom,
			playerID: "player1",
			want:     false,
		},
		{
			name:     "first player is room creator when no host",
			room:     nonHostRoom,
			playerID: "first1",
			want:     true,
		},
		{
			name:     "second player is not room creator when no host",
			room:     nonHostRoom,
			playerID: "second1",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  "player_" + tt.room.Code,
				Value: tt.playerID,
			})

			result := h.isRoomCreator(req, tt.room)
			if result != tt.want {
				t.Errorf("isRoomCreator() = %v, want %v", result, tt.want)
			}
		})
	}
}