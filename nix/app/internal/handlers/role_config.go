package handlers

import (
	"net/http"
	"fmt"
	"time"
	"encoding/json"
	"treacherest/internal/game"
	"treacherest/internal/views/components"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
)

// UpdateRolePreset updates the role preset for a room
func (h *Handler) UpdateRolePreset(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get preset value from form (works with both urlencoded and multipart)
	presetName := r.FormValue("preset")
	if presetName == "" {
		http.Error(w, "Preset name required", http.StatusBadRequest)
		return
	}
	
	// Update role configuration
	if presetName == "custom" {
		// Keep current custom configuration
		room.RoleConfig.PresetName = "custom"
	} else {
		// Load preset configuration
		newConfig, err := h.roleConfigService.CreateFromPreset(presetName, room.MaxPlayers)
		if err != nil {
			http.Error(w, "Invalid preset", http.StatusBadRequest)
			return
		}
		room.RoleConfig = newConfig
	}
	
	h.store.UpdateRoom(room)
	
	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})
	
	// Send validation update
	h.sendRoleValidation(w, r, room)
}

// ToggleRole enables/disables a role
func (h *Handler) ToggleRole(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// This endpoint is deprecated - use UpdateRoleTypeCount and ToggleRoleCard instead
	http.Error(w, "This endpoint is deprecated", http.StatusBadRequest)
}

// UpdateRoleCount updates the count for a specific role
func (h *Handler) UpdateRoleCount(w http.ResponseWriter, r *http.Request) {
	// This endpoint is deprecated - use UpdateRoleTypeCount instead
	http.Error(w, "This endpoint is deprecated", http.StatusBadRequest)
}

// Helper functions

func (h *Handler) isRoomCreator(r *http.Request, room *game.Room) bool {
	playerCookie, err := r.Cookie("player_" + room.Code)
	if err != nil {
		return false
	}
	
	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		return false
	}
	
	// Check if this player is the host
	if player.IsHost {
		return true
	}
	
	// If no host, check if this is the first player (room creator)
	hasHost := false
	for _, p := range room.Players {
		if p.IsHost {
			hasHost = true
			break
		}
	}
	
	if !hasHost {
		// Find the first player
		var firstPlayer *game.Player
		var firstJoinTime time.Time
		
		for _, p := range room.Players {
			if !p.IsHost && (firstPlayer == nil || p.JoinedAt.Before(firstJoinTime)) {
				firstPlayer = p
				firstJoinTime = p.JoinedAt
			}
		}
		
		return firstPlayer != nil && player.ID == firstPlayer.ID
	}
	
	return false
}

func (h *Handler) updatePlayerLimits(room *game.Room) {
	// Deprecated - use updatePlayerLimitsNew
	h.updatePlayerLimitsNew(room)
}

// UpdateLeaderlessGame updates the leaderless game setting for a room
func (h *Handler) UpdateLeaderlessGame(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse JSON body
	var body struct {
		Allowed bool `json:"allowed"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the setting
	room.RoleConfig.AllowLeaderlessGame = body.Allowed
	
	// If disabling leaderless games and leader count is 0, set it to 1
	if !body.Allowed {
		if leaderConfig, exists := room.RoleConfig.RoleTypes["Leader"]; exists && leaderConfig.Count == 0 {
			leaderConfig.Count = 1
			room.RoleConfig.PresetName = "custom"
		}
	}
	
	h.store.UpdateRoom(room)
	
	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated", 
		RoomCode: room.Code,
		Data:     room,
	})
	
	// Re-render the entire role configuration component
	sse := datastar.NewSSE(w, r)
	component := components.RoleConfigurationNew(room, h.config, h.cardService)
	html := renderToString(component)
	
	sse.MergeFragments(html,
		datastar.WithSelector("#role-config"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
	
	// Also send validation update
	h.sendRoleValidation(w, r, room)
}

func (h *Handler) sendRoleValidation(w http.ResponseWriter, r *http.Request, room *game.Room) {
	// Deprecated - use sendRoleValidationNew
	h.sendRoleValidationNew(w, r, room)
}

// UpdateRoleTypeCount updates the count for a specific role type
func (h *Handler) UpdateRoleTypeCount(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse JSON body
	var body struct {
		RoleType string `json:"roleType"`
		Count    int    `json:"count"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate role type exists
	if _, exists := room.RoleConfig.RoleTypes[body.RoleType]; !exists {
		http.Error(w, "Invalid role type", http.StatusBadRequest)
		return
	}
	
	// Update count
	room.RoleConfig.RoleTypes[body.RoleType].Count = body.Count
	room.RoleConfig.PresetName = "custom"
	
	// Update min/max players based on role counts
	h.updatePlayerLimitsNew(room)
	
	h.store.UpdateRoom(room)
	
	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})
	
	// Send validation update
	h.sendRoleValidationNew(w, r, room)
}

// ToggleRoleCard enables/disables a specific role card
func (h *Handler) ToggleRoleCard(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse JSON body
	var body struct {
		RoleType string `json:"roleType"`
		CardName string `json:"cardName"`
		Enabled  bool   `json:"enabled"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate role type exists
	typeConfig, exists := room.RoleConfig.RoleTypes[body.RoleType]
	if !exists {
		http.Error(w, "Invalid role type", http.StatusBadRequest)
		return
	}
	
	// Check if EnabledCards is nil and initialize if needed
	if typeConfig.EnabledCards == nil {
		typeConfig.EnabledCards = make(map[string]bool)
	}
	
	// Update card state
	typeConfig.EnabledCards[body.CardName] = body.Enabled
	room.RoleConfig.PresetName = "custom"
	
	h.store.UpdateRoom(room)
	
	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})
	
	// Send validation update
	h.sendRoleValidationNew(w, r, room)
}

func (h *Handler) updatePlayerLimitsNew(room *game.Room) {
	// Calculate total roles needed
	totalRoles := 0
	
	for _, typeConfig := range room.RoleConfig.RoleTypes {
		totalRoles += typeConfig.Count
	}
	
	// Min players is the total roles needed (or server minimum)
	minPlayers := totalRoles
	if minPlayers < h.config.Server.MinPlayersPerRoom {
		minPlayers = h.config.Server.MinPlayersPerRoom
	}
	
	// Max players should be at least min players, up to server maximum
	maxPlayers := totalRoles
	if maxPlayers < minPlayers {
		maxPlayers = minPlayers
	}
	if maxPlayers > h.config.Server.MaxPlayersPerRoom {
		maxPlayers = h.config.Server.MaxPlayersPerRoom
	}
	
	// For presets, we might want to allow flexibility for more players
	// This allows the game to scale up with more of certain roles
	if room.RoleConfig.PresetName != "custom" && maxPlayers < h.config.Server.MaxPlayersPerRoom {
		// Allow scaling up to max server limit for presets
		maxPlayers = h.config.Server.MaxPlayersPerRoom
	}
	
	room.RoleConfig.MinPlayers = minPlayers
	room.RoleConfig.MaxPlayers = maxPlayers
}

func (h *Handler) sendRoleValidationNew(w http.ResponseWriter, r *http.Request, room *game.Room) {
	errors := []string{}
	warnings := []string{}
	
	// Validate role configuration
	totalRoles := 0
	hasLeader := false
	
	for roleType, typeConfig := range room.RoleConfig.RoleTypes {
		if typeConfig.Count == 0 {
			continue
		}
		
		// Count enabled cards
		enabledCount := 0
		for _, enabled := range typeConfig.EnabledCards {
			if enabled {
				enabledCount++
			}
		}
		
		// Check if we have enough enabled cards
		if typeConfig.Count > enabledCount {
			errors = append(errors, fmt.Sprintf("%s: need %d cards but only %d are enabled", roleType, typeConfig.Count, enabledCount))
		}
		
		totalRoles += typeConfig.Count
		
		if roleType == "Leader" && typeConfig.Count > 0 {
			hasLeader = true
		}
	}
	
	// Check for required leader role
	if !hasLeader {
		if room.RoleConfig.AllowLeaderlessGame {
			warnings = append(warnings, "⚠️ Leaderless game - All roles will be hidden. Guardians and Assassins must deduce their allies without a revealed Leader.")
		} else {
			errors = append(errors, "Leader role is required")
		}
	}
	
	// Check player count constraints
	activePlayerCount := room.GetActivePlayerCount()
	if totalRoles < activePlayerCount && activePlayerCount > 0 {
		errors = append(errors, fmt.Sprintf("Not enough roles (%d) for current players (%d)", totalRoles, activePlayerCount))
	}
	if totalRoles > h.config.Server.MaxPlayersPerRoom {
		errors = append(errors, fmt.Sprintf("Too many roles (%d), max is %d", totalRoles, h.config.Server.MaxPlayersPerRoom))
	}
	
	// Check if we have enough roles for min players
	if totalRoles < room.RoleConfig.MinPlayers {
		warnings = append(warnings, fmt.Sprintf("Total roles (%d) less than minimum players (%d)", totalRoles, room.RoleConfig.MinPlayers))
	}
	
	// Send validation component via SSE
	sse := datastar.NewSSE(w, r)
	
	// Build validation HTML directly
	var html string
	if len(errors) > 0 || len(warnings) > 0 {
		html = `<div id="role-validation" class="validation-messages">`
		for _, err := range errors {
			html += `<div class="validation-error">❌ ` + err + `</div>`
		}
		for _, warn := range warnings {
			html += `<div class="validation-warning">⚠️ ` + warn + `</div>`
		}
		html += `</div>`
	} else {
		html = `<div id="role-validation" class="validation-messages"></div>`
	}
	
	sse.MergeFragments(html, 
		datastar.WithSelector("#role-validation"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}