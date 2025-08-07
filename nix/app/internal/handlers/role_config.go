package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"fmt"
	"time"
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
		roleService := game.NewRoleConfigService(h.config)
		newConfig, err := roleService.CreateFromPreset(presetName, room.MaxPlayers)
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
	
	// Parse form data - handle both multipart and urlencoded
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.ParseMultipartForm(32 << 20) // 32MB max
	} else {
		r.ParseForm()
	}
	
	// Find which role was toggled
	var roleName string
	for key := range r.Form {
		if strings.HasPrefix(key, "role-") {
			roleName = strings.TrimPrefix(key, "role-")
			break
		}
	}
	
	if roleName == "" {
		http.Error(w, "Role name required", http.StatusBadRequest)
		return
	}
	
	// Toggle role
	if room.RoleConfig.EnabledRoles[roleName] {
		room.RoleConfig.EnabledRoles[roleName] = false
		delete(room.RoleConfig.RoleCounts, roleName)
	} else {
		room.RoleConfig.EnabledRoles[roleName] = true
		// Set default count
		if roleDef, exists := h.config.Roles.Available[roleName]; exists {
			room.RoleConfig.RoleCounts[roleName] = roleDef.MinCount
			if roleDef.MinCount == 0 {
				room.RoleConfig.RoleCounts[roleName] = 1
			}
		}
	}
	
	// Mark as custom since user modified it
	room.RoleConfig.PresetName = "custom"
	
	// Update min/max players based on enabled roles
	h.updatePlayerLimits(room)
	
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

// UpdateRoleCount updates the count for a specific role
func (h *Handler) UpdateRoleCount(w http.ResponseWriter, r *http.Request) {
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
	
	// Parse form data - handle both multipart and urlencoded
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.ParseMultipartForm(32 << 20) // 32MB max
	} else {
		r.ParseForm()
	}
	
	// Find role name and count
	var roleName string
	var count int
	for key, values := range r.Form {
		if strings.HasPrefix(key, "count-") && len(values) > 0 {
			roleName = strings.TrimPrefix(key, "count-")
			count, _ = strconv.Atoi(values[0])
			break
		}
	}
	
	if roleName == "" {
		http.Error(w, "Role name required", http.StatusBadRequest)
		return
	}
	
	// Validate count against role constraints
	if roleDef, exists := h.config.Roles.Available[roleName]; exists {
		if count < roleDef.MinCount {
			count = roleDef.MinCount
		}
		if count > roleDef.MaxCount {
			count = roleDef.MaxCount
		}
	}
	
	// Update count
	room.RoleConfig.RoleCounts[roleName] = count
	room.RoleConfig.PresetName = "custom"
	
	// Update min/max players based on role counts
	h.updatePlayerLimits(room)
	
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
	// Calculate total roles needed
	totalRoles := 0
	
	for roleName, enabled := range room.RoleConfig.EnabledRoles {
		if !enabled {
			continue
		}
		
		count := room.RoleConfig.RoleCounts[roleName]
		totalRoles += count
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

func (h *Handler) sendRoleValidation(w http.ResponseWriter, r *http.Request, room *game.Room) {
	errors := []string{}
	warnings := []string{}
	
	// Validate role configuration
	totalRoles := 0
	hasLeader := false
	for roleName, enabled := range room.RoleConfig.EnabledRoles {
		if !enabled {
			continue
		}
		count := room.RoleConfig.RoleCounts[roleName]
		totalRoles += count
		
		if roleName == "leader" && count > 0 {
			hasLeader = true
		}
	}
	
	// Check for required leader role
	if !hasLeader {
		errors = append(errors, "Leader role is required")
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
	component := components.RoleValidation(errors, warnings)
	html := renderToString(component)
	
	sse.MergeFragments(html, 
		datastar.WithSelector("#role-validation"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}