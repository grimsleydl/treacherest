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
		// Load preset configuration using current player count from role config
		playerCount := room.RoleConfig.MaxPlayers
		if playerCount == 0 {
			// Fallback to default game size if not set
			playerCount = h.config.Server.DefaultGameSize
		}
		newConfig, err := h.roleConfigService.CreateFromPreset(presetName, playerCount)
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

	// Create player count display data
	playerCountDisplay := h.createPlayerCountDisplay(room)

	// Re-render just the role configuration component
	component := components.RoleConfigurationNew(room, h.config, h.cardService, playerCountDisplay)
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
		"canStartGame":       validationState.CanStart,
		"validationMessage":  validationState.ValidationMessage,
		"canAutoScale":       validationState.CanAutoScale,
		"autoScaleDetails":   validationState.AutoScaleDetails,
		"requiredRoles":      validationState.RequiredRoles,
		"configuredRoles":    validationState.ConfiguredRoles,
		"updatingLeaderless": false,                               // Reset loading state
		"allowLeaderless":    room.RoleConfig.AllowLeaderlessGame, // Sync checkbox state
	}

	log.Printf("  - Sending signals (with allowLeaderless): %+v", signals)
	sse.MarshalAndMergeSignals(signals)
}

func (h *Handler) createPlayerCountDisplay(room *game.Room) components.PlayerCountDisplay {
	// Use RoleConfig.MaxPlayers which represents the current game size
	currentPlayerCount := room.RoleConfig.MaxPlayers
	canDecrement := currentPlayerCount > h.config.Server.MinPlayersPerRoom && currentPlayerCount > len(room.Players)
	canIncrement := currentPlayerCount < h.config.Server.MaxPlayersPerRoom

	// Build tooltips
	incrementTooltip := "Increase player count"
	decrementTooltip := "Decrease player count"

	if !canIncrement {
		incrementTooltip = "Maximum player count reached"
	} else if room.RoleConfig.PresetName == "custom" {
		// Calculate which role would be added
		roleToAdd := h.calculateRoleAdjustment(room, true)
		if roleToAdd != "" {
			incrementTooltip = fmt.Sprintf("Will add 1 %s", roleToAdd)
		}
	}

	if !canDecrement {
		if currentPlayerCount <= h.config.Server.MinPlayersPerRoom {
			decrementTooltip = "Minimum player count reached"
		} else if currentPlayerCount <= len(room.Players) {
			decrementTooltip = fmt.Sprintf("Cannot reduce below %d connected players", len(room.Players))
		}
	} else if room.RoleConfig.PresetName == "custom" {
		// Calculate which role would be removed
		roleToRemove := h.calculateRoleAdjustment(room, false)
		if roleToRemove != "" {
			decrementTooltip = fmt.Sprintf("Will remove 1 %s", roleToRemove)
		}
	}

	return components.PlayerCountDisplay{
		IncrementTooltip: incrementTooltip,
		DecrementTooltip: decrementTooltip,
		CanIncrement:     canIncrement,
		CanDecrement:     canDecrement,
	}
}

func (h *Handler) calculateRoleAdjustment(room *game.Room, increment bool) string {
	config := room.RoleConfig

	// For preset mode, return empty (no specific role preview)
	if config.PresetName != "custom" {
		return ""
	}

	// Build map of roles that can be adjusted
	roleCounts := make(map[string]int)
	roleTypes := []string{}

	// Always include Guardian, Assassin, Traitor
	for _, role := range []string{"Guardian", "Assassin", "Traitor"} {
		if typeConfig, exists := config.RoleTypes[role]; exists {
			roleCounts[role] = typeConfig.Count
			roleTypes = append(roleTypes, role)
		}
	}

	// Include Leader only if count > 1
	if leaderConfig, exists := config.RoleTypes["Leader"]; exists && leaderConfig.Count > 1 {
		roleCounts["Leader"] = leaderConfig.Count
		roleTypes = append(roleTypes, "Leader")
	}

	// Calculate average (excluding roles not in calculation)
	total := 0
	for _, count := range roleCounts {
		total += count
	}

	if len(roleCounts) == 0 {
		return ""
	}

	avg := float64(total) / float64(len(roleCounts))

	if increment {
		// Find most underrepresented role
		maxDeviation := 0.0
		roleToAdd := ""

		for _, role := range roleTypes {
			deviation := avg - float64(roleCounts[role])
			if deviation > maxDeviation || (deviation == maxDeviation && role == "Guardian") {
				maxDeviation = deviation
				roleToAdd = role
			}
		}

		return roleToAdd
	} else {
		// Find most overrepresented role that can be safely reduced
		maxDeviation := 0.0
		roleToRemove := ""

		for _, role := range roleTypes {
			count := roleCounts[role]
			canReduce := true

			// Check minimum constraints
			switch role {
			case "Leader":
				canReduce = count > 1
			case "Assassin":
				canReduce = count > 1
			default:
				canReduce = count > 0
			}

			if canReduce {
				deviation := float64(count) - avg
				if deviation > maxDeviation {
					maxDeviation = deviation
					roleToRemove = role
				}
			}
		}

		return roleToRemove
	}
}

// IncrementPlayerCount increments the player count for a room
func (h *Handler) IncrementPlayerCount(w http.ResponseWriter, r *http.Request) {
	h.updatePlayerCount(w, r, "increment")
}

// DecrementPlayerCount decrements the player count for a room
func (h *Handler) DecrementPlayerCount(w http.ResponseWriter, r *http.Request) {
	h.updatePlayerCount(w, r, "decrement")
}

// updatePlayerCount handles the actual player count update logic
func (h *Handler) updatePlayerCount(w http.ResponseWriter, r *http.Request, action string) {
	roomCode := chi.URLParam(r, "code")

	// Get room
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.MergeFragments(`<div class="alert alert-error">Room not found</div>`,
			datastar.WithSelector("#role-validation"))
		return
	}

	// Verify player is room creator
	if !h.isRoomCreator(r, room) {
		sse := datastar.NewSSE(w, r)
		sse.MergeFragments(`<div class="alert alert-error">Unauthorized</div>`,
			datastar.WithSelector("#role-validation"))
		return
	}

	// Validate action
	currentPlayerCount := room.RoleConfig.MaxPlayers

	switch action {
	case "increment":
		if currentPlayerCount >= h.config.Server.MaxPlayersPerRoom {
			sse := datastar.NewSSE(w, r)
			sse.MergeFragments(`<div class="alert alert-error">Maximum player count reached</div>`,
				datastar.WithSelector("#role-validation"))
			return
		}
		room.RoleConfig.MaxPlayers++

	case "decrement":
		if currentPlayerCount <= h.config.Server.MinPlayersPerRoom {
			sse := datastar.NewSSE(w, r)
			sse.MergeFragments(`<div class="alert alert-error">Minimum player count reached</div>`,
				datastar.WithSelector("#role-validation"))
			return
		}

		// Check connected players constraint
		if currentPlayerCount <= len(room.Players) {
			sse := datastar.NewSSE(w, r)
			sse.MergeFragments(fmt.Sprintf(`<div class="alert alert-error">Cannot reduce below %d connected players</div>`, len(room.Players)),
				datastar.WithSelector("#role-validation"))
			return
		}

		room.RoleConfig.MaxPlayers--

	default:
		log.Printf("ERROR: Invalid action '%s' for player count update", action)
		return
	}

	// Now adjust role counts based on mode
	if room.RoleConfig.PresetName != "custom" {
		// Apply preset for new player count
		h.applyPresetForPlayerCount(room)
	} else {
		// Smart rebalancing for custom mode
		h.rebalanceCustomRoles(room, action == "increment")
	}

	// Update room
	h.store.UpdateRoom(room)

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "role_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	// Send updated UI
	h.sendUpdatedRoleConfigUI(w, r, room)
}

func (h *Handler) applyPresetForPlayerCount(room *game.Room) {
	presetName := room.RoleConfig.PresetName
	playerCount := room.RoleConfig.MaxPlayers

	// Get preset distribution
	preset, exists := h.config.Roles.Presets[presetName]
	if !exists {
		log.Printf("ERROR: Preset '%s' not found", presetName)
		return
	}

	distribution, exists := preset.Distributions[playerCount]
	if !exists {
		log.Printf("ERROR: No distribution for %d players in preset '%s'", playerCount, presetName)
		return
	}

	// Apply distribution
	if leaderConfig, exists := room.RoleConfig.RoleTypes["Leader"]; exists {
		leaderConfig.Count = distribution["leader"]
	}
	if guardianConfig, exists := room.RoleConfig.RoleTypes["Guardian"]; exists {
		guardianConfig.Count = distribution["guardian"]
	}
	if assassinConfig, exists := room.RoleConfig.RoleTypes["Assassin"]; exists {
		assassinConfig.Count = distribution["assassin"]
	}
	if traitorConfig, exists := room.RoleConfig.RoleTypes["Traitor"]; exists {
		traitorConfig.Count = distribution["traitor"]
	}
}

func (h *Handler) rebalanceCustomRoles(room *game.Room, increment bool) {
	// Use the same logic as calculateRoleAdjustment to determine which role to adjust
	roleToAdjust := h.calculateRoleAdjustment(room, increment)

	if roleToAdjust == "" {
		// Fallback: adjust the first available role
		if increment {
			roleToAdjust = "Guardian"
		} else {
			// Find a role that can be decremented
			for _, role := range []string{"Traitor", "Guardian", "Assassin", "Leader"} {
				if config, exists := room.RoleConfig.RoleTypes[role]; exists {
					switch role {
					case "Leader", "Assassin":
						if config.Count > 1 {
							roleToAdjust = role
							break
						}
					default:
						if config.Count > 0 {
							roleToAdjust = role
							break
						}
					}
				}
			}
		}
	}

	// Apply the adjustment
	if roleConfig, exists := room.RoleConfig.RoleTypes[roleToAdjust]; exists {
		if increment {
			roleConfig.Count++
		} else {
			roleConfig.Count--
		}
	}
}
