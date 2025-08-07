package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
	"log"
	"net/http"
	"strings"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/views/components"
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

	// Send updated UI using the helper
	h.sendUpdatedRoleConfigUI(w, r, room)
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
	log.Printf("🔍 UpdateLeaderlessGame called for room: %s", roomCode)

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("❌ Room not found: %s", roomCode)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		log.Printf("❌ Unauthorized access attempt for room: %s", roomCode)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse JSON body
	var body struct {
		AllowLeaderless bool `json:"allowLeaderless"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("❌ Invalid request body for room %s: %v", roomCode, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log state change
	previousState := room.RoleConfig.AllowLeaderlessGame
	leaderCount := 0
	if leaderConfig, exists := room.RoleConfig.RoleTypes["Leader"]; exists {
		leaderCount = leaderConfig.Count
	}
	
	log.Printf("📊 UpdateLeaderlessGame state change for room %s:", roomCode)
	log.Printf("  - Previous AllowLeaderlessGame: %v", previousState)
	log.Printf("  - New AllowLeaderlessGame: %v", body.AllowLeaderless)
	log.Printf("  - Current Leader count: %d", leaderCount)

	// Update the setting
	room.RoleConfig.AllowLeaderlessGame = body.AllowLeaderless

	// If disabling leaderless games and leader count is 0, set it to 1
	if !body.AllowLeaderless {
		if leaderConfig, exists := room.RoleConfig.RoleTypes["Leader"]; exists && leaderConfig.Count == 0 {
			log.Printf("  - Auto-adding 1 Leader because leaderless disabled and leader count was 0")
			leaderConfig.Count = 1
			room.RoleConfig.PresetName = "custom"
		}
	}

	h.store.UpdateRoom(room)
	log.Printf("✅ UpdateLeaderlessGame completed for room %s", roomCode)

	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	// Send updated UI using the helper
	h.sendUpdatedRoleConfigUI(w, r, room)
}

func (h *Handler) sendRoleValidation(w http.ResponseWriter, r *http.Request, room *game.Room) {
	// Deprecated - use sendRoleValidationNew
	h.sendRoleValidationNew(w, r, room)
}

// IncrementRoleTypeCount increments the count for a specific role type
func (h *Handler) IncrementRoleTypeCount(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEBUG: IncrementRoleTypeCount called")
	h.updateRoleTypeCount(w, r, "increment")
}

// DecrementRoleTypeCount decrements the count for a specific role type
func (h *Handler) DecrementRoleTypeCount(w http.ResponseWriter, r *http.Request) {
	h.updateRoleTypeCount(w, r, "decrement")
}

// updateRoleTypeCount handles the actual count update logic
func (h *Handler) updateRoleTypeCount(w http.ResponseWriter, r *http.Request, action string) {
	roomCode := chi.URLParam(r, "code")
	roleType := chi.URLParam(r, "roleType")

	// Get room
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		// Return error fragment
		sse := datastar.NewSSE(w, r)
		sse.MergeFragments(`<div class="alert alert-error">Room not found</div>`,
			datastar.WithSelector("#role-validation"))
		return
	}

	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		// Return error fragment
		sse := datastar.NewSSE(w, r)
		sse.MergeFragments(`<div class="alert alert-error">Unauthorized</div>`,
			datastar.WithSelector("#role-validation"))
		return
	}


	// Get the type config
	typeConfig, exists := room.RoleConfig.RoleTypes[roleType]
	if !exists {
		// Return error fragment
		sse := datastar.NewSSE(w, r)
		sse.MergeFragments(fmt.Sprintf(`<div class="alert alert-error">Invalid role type: %s</div>`, roleType),
			datastar.WithSelector("#role-validation"))
		return
	}

	// Update count based on action
	switch action {
	case "increment":
		typeConfig.Count++
		room.RoleConfig.PresetName = "custom" // Switch to custom when modified
	case "decrement":
		if typeConfig.Count > 0 {
			typeConfig.Count--
			room.RoleConfig.PresetName = "custom" // Switch to custom when modified
		}
	default:
		// This should never happen with our current implementation
		log.Printf("ERROR: Invalid action '%s'", action)
		return
	}

	// Update min/max players based on role counts
	h.updatePlayerLimitsNew(room)

	h.store.UpdateRoom(room)

	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	// CRITICAL: Return updated DOM fragments, not just signals
	h.sendUpdatedRoleConfigUI(w, r, room)
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

	// Parse JSON body with signal variables
	var body map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("ERROR: Failed to decode body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract values from the map
	cardId, _ := body["cardId"].(string)
	enabled, _ := body["cardChecked"].(bool)

	// Parse the card ID: "card-{roleType}-{cardAnchor}"
	parts := strings.Split(cardId, "-")
	if len(parts) < 3 || parts[0] != "card" {
		log.Printf("ERROR: Invalid card ID format: %s", cardId)
		http.Error(w, "Invalid card ID format", http.StatusBadRequest)
		return
	}

	roleType := parts[1]
	cardAnchor := strings.Join(parts[2:], "-") // Handle card names with hyphens

	// Find the card name from the anchor
	var cardName string
	cards := h.getCardsForRoleType(roleType)
	for _, card := range cards {
		if card.NameAnchor == cardAnchor {
			cardName = card.Name
			break
		}
	}

	if cardName == "" {
		log.Printf("ERROR: Card not found for anchor: '%s' in role type: '%s'", cardAnchor, roleType)
		http.Error(w, "Card not found", http.StatusBadRequest)
		return
	}

	// Validate role type exists
	typeConfig, exists := room.RoleConfig.RoleTypes[roleType]
	if !exists {
		log.Printf("ERROR: Invalid role type received: '%s'", roleType)
		http.Error(w, "Invalid role type", http.StatusBadRequest)
		return
	}

	// Check if EnabledCards is nil and initialize if needed
	if typeConfig.EnabledCards == nil {
		typeConfig.EnabledCards = make(map[string]bool)
	}

	// Update card state
	typeConfig.EnabledCards[cardName] = enabled
	room.RoleConfig.PresetName = "custom"

	h.store.UpdateRoom(room)

	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	// Send updated UI using the helper
	h.sendUpdatedRoleConfigUI(w, r, room)
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

// ToggleRoleCardFast handles card toggle with minimal response
func (h *Handler) ToggleRoleCardFast(w http.ResponseWriter, r *http.Request) {
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

	// Get data from headers
	roleType := r.Header.Get("X-Role-Type")
	cardAnchor := r.Header.Get("X-Card-Anchor")
	enabled := r.Header.Get("X-Enabled") == "true"

	// Find the card name from the anchor
	var cardName string
	cards := h.getCardsForRoleType(roleType)
	for _, card := range cards {
		if card.NameAnchor == cardAnchor {
			cardName = card.Name
			break
		}
	}

	if cardName == "" {
		http.Error(w, "Card not found", http.StatusBadRequest)
		return
	}

	// Validate role type exists
	typeConfig, exists := room.RoleConfig.RoleTypes[roleType]
	if !exists {
		http.Error(w, "Invalid role type", http.StatusBadRequest)
		return
	}

	// Check if EnabledCards is nil and initialize if needed
	if typeConfig.EnabledCards == nil {
		typeConfig.EnabledCards = make(map[string]bool)
	}

	// Update card state
	typeConfig.EnabledCards[cardName] = enabled
	room.RoleConfig.PresetName = "custom"

	h.store.UpdateRoom(room)

	// Don't publish events, just send minimal response
	sse := datastar.NewSSE(w, r)

	// Send empty response to acknowledge
	sse.ExecuteScript("// OK")
}

// ToggleRoleCardOptimistic handles card toggle with optimistic updates
func (h *Handler) ToggleRoleCardOptimistic(w http.ResponseWriter, r *http.Request) {
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

	// Send only validation update (checkbox already updated optimistically)
	h.sendRoleValidationNew(w, r, room)

	// If other players are watching, notify them
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})
}

func (h *Handler) getCardsForRoleType(roleType string) []*game.Card {
	switch roleType {
	case "Leader":
		return h.cardService.Leaders
	case "Guardian":
		return h.cardService.Guardians
	case "Assassin":
		return h.cardService.Assassins
	case "Traitor":
		return h.cardService.Traitors
	default:
		return nil
	}
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

func (h *Handler) sendUpdatedRoleConfigUI(w http.ResponseWriter, r *http.Request, room *game.Room) {
	log.Printf("📤 sendUpdatedRoleConfigUI called for room %s", room.Code)
	sse := datastar.NewSSE(w, r)
	
	// Log current state
	leaderCount := 0
	if leaderConfig, exists := room.RoleConfig.RoleTypes["Leader"]; exists {
		leaderCount = leaderConfig.Count
	}
	log.Printf("  - Current AllowLeaderlessGame: %v", room.RoleConfig.AllowLeaderlessGame)
	log.Printf("  - Current Leader count: %d", leaderCount)
	
	// Re-render just the role configuration component
	component := components.RoleConfigurationNew(room, h.config, h.cardService)
	html := renderToString(component)
	
	log.Printf("  - Sending role config update with selector #role-config")
	
	// Send the role config fragment
	sse.MergeFragments(html,
		datastar.WithSelector("#role-config"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
	
	// Also update validation state
	roleService := game.NewRoleConfigService(h.config)
	validationState := room.GetValidationState(roleService)
	
	log.Printf("  - Validation state: CanStart=%v, Message=%s", validationState.CanStart, validationState.ValidationMessage)
	
	signals := map[string]interface{}{
		"canStartGame": validationState.CanStart,
		"validationMessage": validationState.ValidationMessage,
		"canAutoScale": validationState.CanAutoScale,
		"autoScaleDetails": validationState.AutoScaleDetails,
		"requiredRoles": validationState.RequiredRoles,
		"configuredRoles": validationState.ConfiguredRoles,
		"updatingLeaderless": false, // Reset loading state
		"allowLeaderless": room.RoleConfig.AllowLeaderlessGame, // Sync checkbox state
	}
	
	log.Printf("  - Sending signals (with allowLeaderless): %+v", signals)
	sse.MarshalAndMergeSignals(signals)
}
