package handlers

import (
	"net/http"
	"strconv"
	"strings"
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
	
	// Parse JSON body with signals
	var signals struct {
		RoleName string `json:"roleName"`
		Enabled  bool   `json:"enabled"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
		http.Error(w, "Invalid request body: " + err.Error(), http.StatusBadRequest)
		return
	}
	
	if signals.RoleName == "" {
		http.Error(w, "Role name required", http.StatusBadRequest)
		return
	}
	
	// Update role state based on the explicit enabled value
	room.RoleConfig.EnabledRoles[signals.RoleName] = signals.Enabled
	
	if signals.Enabled {
		// Set default count if enabling
		if _, exists := room.RoleConfig.RoleCounts[signals.RoleName]; !exists {
			if roleDef, exists := h.config.Roles.Available[signals.RoleName]; exists {
				room.RoleConfig.RoleCounts[signals.RoleName] = roleDef.MinCount
				if roleDef.MinCount == 0 {
					room.RoleConfig.RoleCounts[signals.RoleName] = 1
				}
			}
		}
	} else {
		// Remove count if disabling
		delete(room.RoleConfig.RoleCounts, signals.RoleName)
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
	
	// Re-render the entire role configuration component
	sse := datastar.NewSSE(w, r)
	roleService := game.NewRoleConfigService(h.config)
	tempSortedRoles := roleService.GetSortedRoles()
	
	// Convert to components.SortedRole type
	sortedRoles := make([]components.SortedRole, len(tempSortedRoles))
	for i, role := range tempSortedRoles {
		sortedRoles[i] = components.SortedRole{
			Name:       role.Name,
			Definition: role.Definition,
		}
	}
	
	component := components.RoleConfiguration(room, h.config, sortedRoles)
	html := renderToString(component)
	
	sse.MergeFragments(html,
		datastar.WithSelector("#role-config"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
	
	// Also send validation update
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
	
	// If disabling leaderless games and leader is not enabled, enable it
	if !body.Allowed && !room.RoleConfig.EnabledRoles["leader"] {
		room.RoleConfig.EnabledRoles["leader"] = true
		room.RoleConfig.RoleCounts["leader"] = 1
		room.RoleConfig.PresetName = "custom"
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
	roleService := game.NewRoleConfigService(h.config)
	tempSortedRoles := roleService.GetSortedRoles()
	
	// Convert to components.SortedRole type
	sortedRoles := make([]components.SortedRole, len(tempSortedRoles))
	for i, role := range tempSortedRoles {
		sortedRoles[i] = components.SortedRole{
			Name:       role.Name,
			Definition: role.Definition,
		}
	}
	
	component := components.RoleConfiguration(room, h.config, sortedRoles)
	html := renderToString(component)
	
	sse.MergeFragments(html,
		datastar.WithSelector("#role-config"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
	
	// Also send validation update
	h.sendRoleValidation(w, r, room)
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
	component := components.RoleValidation(errors, warnings)
	html := renderToString(component)
	
	sse.MergeFragments(html, 
		datastar.WithSelector("#role-validation"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}