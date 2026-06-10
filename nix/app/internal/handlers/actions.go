package handlers

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar-go/datastar"
	"html"
	"log"
	"net/http"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/game/ability"
	"treacherest/internal/views/components"
)

// StartGame starts a game
func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	log.Printf("🚀 StartGame called for room: %s", roomCode)

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("❌ Room not found: %s", roomCode)
		// Return SSE error instead of HTTP error
		sse := datastar.NewSSE(w, r)
		errorHTML := `<div id="start-game-error" class="alert alert-error mt-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<span>Room not found</span>
		</div>`
		sse.PatchElements(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"isStarting": false,
		})
		return
	}

	// Verify player is in room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		log.Printf("❌ No player cookie for room: %s", roomCode)
		sse := datastar.NewSSE(w, r)
		errorHTML := `<div id="start-game-error" class="alert alert-error mt-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<span>You are not in this room</span>
		</div>`
		sse.PatchElements(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"isStarting": false,
		})
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		log.Printf("❌ Player not found in room: %s, cookie: %s", roomCode, playerCookie.Value)
		sse := datastar.NewSSE(w, r)
		errorHTML := `<div id="start-game-error" class="alert alert-error mt-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<span>You are not in this room</span>
		</div>`
		sse.PatchElements(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"isStarting": false,
		})
		return
	}

	log.Printf("✅ Player %s attempting to start game in room %s", player.Name, roomCode)
	log.Printf("🔍 Room %s has %d total players, %d active players", roomCode, len(room.Players), room.GetActivePlayerCount())
	for _, p := range room.Players {
		log.Printf("  - Player: %s (Host: %v)", p.Name, p.IsHost)
	}

	if room.RulesMode == game.RulesModeCoup {
		h.startCoupGame(w, r, room)
		return
	}

	// CRITICAL: Use the same validation function as SSE updates
	roleService := game.NewRoleConfigService(h.config)
	validationState := room.GetValidationState(roleService)

	log.Printf("🔍 Validation state: CanStart=%v, RequiredRoles=%d, ConfiguredRoles=%d, Message=%s",
		validationState.CanStart, validationState.RequiredRoles, validationState.ConfiguredRoles, validationState.ValidationMessage)

	if !validationState.CanStart {
		log.Printf("❌ Room cannot start: %s", validationState.ValidationMessage)

		// Always return HTTP 200 with error fragment
		sse := datastar.NewSSE(w, r)

		// Send error as HTML fragment
		errorHTML := fmt.Sprintf(`<div id="start-game-error" class="alert alert-error mt-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<span>%s</span>
		</div>`, html.EscapeString(validationState.ValidationMessage))

		// Send fragment targeting error container
		err := sse.PatchElements(errorHTML, datastar.WithSelector("#error-container"))
		if err != nil {
			log.Printf("❌ Failed to send error fragment: %v", err)
		}

		// Also update button state and re-sync ALL validation signals
		err = sse.MarshalAndPatchSignals(map[string]interface{}{
			"isStarting": false,
			"startError": validationState.ValidationMessage,
			// IMPORTANT: Re-sync all validation signals to ensure consistency
			"canStartGame":      validationState.CanStart,
			"validationMessage": validationState.ValidationMessage,
			"canAutoScale":      validationState.CanAutoScale,
			"autoScaleDetails":  validationState.AutoScaleDetails,
		})
		if err != nil {
			log.Printf("❌ Failed to update signals: %v", err)
		}

		// Flush to ensure immediate delivery
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		return
	}

	// Assign roles using room configuration
	players := room.GetPlayers()
	log.Printf("🎲 Assigning roles to %d players", len(players))
	if h.cardService != nil {
		if room.RoleConfig != nil {
			log.Printf("🎲 Using role configuration: %+v", room.RoleConfig)
			roleService := game.NewRoleConfigService(h.config)
			game.AssignRolesWithConfig(players, h.cardService, room.RoleConfig, roleService)
		} else {
			// Fallback to legacy assignment
			log.Printf("🎲 Using legacy role assignment")
			game.AssignRoles(players, h.cardService)
		}
		// Log assigned roles
		for _, p := range players {
			if p.Role != nil {
				log.Printf("🎲 Player %s assigned role: %s", p.Name, p.Role.Name)
			} else {
				log.Printf("❌ Player %s has no role assigned!", p.Name)
			}
		}
	} else {
		log.Printf("❌ CardService is nil, cannot assign roles")
		sse := datastar.NewSSE(w, r)
		errorHTML := `<div id="start-game-error" class="alert alert-error mt-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<span>Internal server error: Cannot assign roles</span>
		</div>`
		sse.PatchElements(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"isStarting": false,
		})
		return
	}

	// Update game state
	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	// Start countdown immediately
	go h.runCountdown(room)

	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "game_started",
		RoomCode: room.Code,
		Data:     room,
	})

	log.Printf("✅ Game started successfully for room %s", roomCode)

	// Use datastar to redirect directly in the POST response
	sse := datastar.NewSSE(w, r)
	sse.ExecuteScript("window.location.href = '/game/" + roomCode + "'")
	// Flush the response to ensure redirect is sent immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *Handler) startCoupGame(w http.ResponseWriter, r *http.Request, room *game.Room) {
	if err := game.AssignCoupRolesWithInformation(room.GetPlayers(), room.CoupPreset, room.CoupInfoPolicy); err != nil {
		log.Printf("❌ Room cannot start: %s", err.Error())

		sse := datastar.NewSSE(w, r)
		errorHTML := fmt.Sprintf(`<div id="start-game-error" class="alert alert-error mt-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 0118 0z" />
			</svg>
			<span>%s</span>
		</div>`, html.EscapeString(err.Error()))
		sse.PatchElements(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"isStarting":        false,
			"startError":        err.Error(),
			"canStartGame":      false,
			"validationMessage": err.Error(),
		})
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		return
	}

	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	go h.runCountdown(room)

	h.eventBus.Publish(Event{
		Type:     "game_started",
		RoomCode: room.Code,
		Data:     room,
	})

	log.Printf("✅ Coup game started successfully for room %s", room.Code)

	sse := datastar.NewSSE(w, r)
	sse.ExecuteScript("window.location.href = '/game/" + room.Code + "'")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// LeaveRoom removes a player from a room
func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		// Use SSE for consistency
		sse := datastar.NewSSE(w, r)
		sse.ExecuteScript("window.location.href = '/'")
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		// Use SSE for consistency
		sse := datastar.NewSSE(w, r)
		sse.ExecuteScript("window.location.href = '/'")
		return
	}

	// Remove player
	room.RemovePlayer(playerCookie.Value)
	h.store.UpdateRoom(room)

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "player_" + room.Code,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Notify other players
	h.eventBus.Publish(Event{
		Type:     "player_left",
		RoomCode: room.Code,
		Data:     room,
	})

	// Use datastar to redirect since this is called via @post
	sse := datastar.NewSSE(w, r)
	sse.ExecuteScript("window.location.href = '/'")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// ToggleReveal toggles the public reveal state of a player's role
func (h *Handler) ToggleReveal(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("❌ Room not found: %s", roomCode)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		log.Printf("❌ No player cookie for room: %s", roomCode)
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	me := room.GetPlayer(playerCookie.Value)
	if me == nil {
		log.Printf("❌ Player not found in room: %s", roomCode)
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Get the target player
	target := room.GetPlayer(playerID)
	if target == nil {
		log.Printf("❌ Target player not found: %s", playerID)
		http.Error(w, "Target player not found", http.StatusBadRequest)
		return
	}

	// Authorization: players reveal themselves; hosts may record public table reveals.
	if me.ID != target.ID && !me.IsHost {
		log.Printf("❌ Player %s attempted to reveal %s's role (forbidden)", me.ID, target.ID)
		http.Error(w, "You can only reveal your own role", http.StatusForbidden)
		return
	}

	if room.RulesMode == game.RulesModeCoup {
		target.RoleRevealed = true
		target.FaceUp = true
	} else {
		// Leaders cannot hide their role (they start face-up per game rules)
		if target.Role != nil && target.Role.GetRoleType() == game.RoleLeader && target.RoleRevealed {
			log.Printf("❌ Leader %s attempted to hide their role (not allowed)", target.Name)
			http.Error(w, "Leaders cannot hide their role", http.StatusForbidden)
			return
		}

		// Toggle the reveal state
		target.RoleRevealed = !target.RoleRevealed
		// Revealing also turns the card face up (you can't reveal a face-down card)
		// Hiding does NOT turn face down - use the separate "Turn Face Down" action for that
		if target.RoleRevealed {
			target.FaceUp = true
		}
	}
	h.store.UpdateRoom(room)

	log.Printf("🎭 Player %s toggled role reveal to %v (FaceUp: %v) in room %s", target.Name, target.RoleRevealed, target.FaceUp, roomCode)

	// Publish event to update all connected clients
	h.eventBus.Publish(Event{
		Type:     "role_revealed",
		RoomCode: room.Code,
		Data:     room,
	})

	// Return success - SSE will handle the UI update
	w.WriteHeader(http.StatusOK)
}

// ToggleFaceState toggles the face up/down state of a player's role
// This is separate from reveal - face state affects ability conditions (e.g., Wearer of Masks)
func (h *Handler) ToggleFaceState(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("❌ Room not found: %s", roomCode)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		log.Printf("❌ No player cookie for room: %s", roomCode)
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	me := room.GetPlayer(playerCookie.Value)
	if me == nil {
		log.Printf("❌ Player not found in room: %s", roomCode)
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Get the target player
	target := room.GetPlayer(playerID)
	if target == nil {
		log.Printf("❌ Target player not found: %s", playerID)
		http.Error(w, "Target player not found", http.StatusBadRequest)
		return
	}

	// Authorization: only the player themselves can toggle their face state
	if me.ID != target.ID {
		log.Printf("❌ Player %s attempted to toggle %s's face state (forbidden)", me.ID, target.ID)
		http.Error(w, "You can only toggle your own face state", http.StatusForbidden)
		return
	}

	// Toggle the face state
	target.FaceUp = !target.FaceUp
	log.Printf("🃏 Player %s toggled face state to %v in room %s", target.Name, target.FaceUp, roomCode)

	// Check for transformation end conditions
	if target.AbilityState != nil && target.AbilityState.TransformState != nil {
		if target.AbilityState.CheckTransformEndCondition("face_down") && !target.FaceUp {
			// End transformation
			log.Printf("🔄 Ending transformation for %s (turned face down)", target.Name)
			originalCardID := target.AbilityState.EndTransform()

			// Restore original role
			if room.CardPool != nil {
				if originalCard := room.CardPool.GetCardByID(originalCardID); originalCard != nil {
					target.Role = originalCard
					log.Printf("✅ Restored original role %s for %s", originalCard.GetText(), target.Name)
				}
			}
		}
	}

	h.store.UpdateRoom(room)

	// Publish event to update all connected clients
	h.eventBus.Publish(Event{
		Type:     "face_state_changed",
		RoomCode: room.Code,
		Data:     room,
	})

	// Return success - SSE will handle the UI update
	w.WriteHeader(http.StatusOK)
}

// GetRoleOptions retrieves role options for a specific card
func (h *Handler) GetRoleOptions(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	cardID := r.URL.Query().Get("card_id")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Parse card ID
	var cardIDInt int
	if _, err := fmt.Sscanf(cardID, "%d", &cardIDInt); err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Get options
	if room.RoleOptionsManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}

	if !room.RoleOptionsManager.HasOptions(cardIDInt) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}

	opts := room.RoleOptionsManager.GetOrCreateOptions(cardIDInt)

	// Return options as JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"card_id": %d, "options": %v}`, cardIDInt, opts.Options)))
}

// DismissModal dismisses a modal for a pending ability
func (h *Handler) DismissModal(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Verify the ability belongs to this player
	ability := player.AbilityState.GetPendingAbility(abilityID)
	if ability == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	if ability.PlayerID != player.ID {
		http.Error(w, "Ability does not belong to this player", http.StatusForbidden)
		return
	}

	// Dismiss the modal
	success := player.AbilityState.DismissModal(abilityID)
	if !success {
		http.Error(w, "Failed to dismiss modal", http.StatusInternalServerError)
		return
	}

	h.store.UpdateRoom(room)

	log.Printf("🙈 Player %s dismissed modal for ability %s in room %s", player.Name, abilityID, roomCode)

	// Publish event to update UI
	h.eventBus.Publish(Event{
		Type:     "modal_dismissed",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// RestoreModal restores a dismissed modal for a pending ability
func (h *Handler) RestoreModal(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Verify the ability belongs to this player
	ability := player.AbilityState.GetPendingAbility(abilityID)
	if ability == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	if ability.PlayerID != player.ID {
		http.Error(w, "Ability does not belong to this player", http.StatusForbidden)
		return
	}

	// Restore the modal
	success := player.AbilityState.RestoreModal(abilityID)
	if !success {
		http.Error(w, "Failed to restore modal", http.StatusInternalServerError)
		return
	}

	h.store.UpdateRoom(room)

	log.Printf("👁️ Player %s restored modal for ability %s in room %s", player.Name, abilityID, roomCode)

	// Publish event to update UI
	h.eventBus.Publish(Event{
		Type:     "modal_restored",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SetRoleOption(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room and is host
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Only host can modify role options
	if !player.IsHost {
		http.Error(w, "Only host can modify role options", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		CardID int                    `json:"card_id"`
		Key    string                 `json:"key"`
		Value  interface{}            `json:"value"`
	}

	if err := r.ParseForm(); err == nil && r.Method == "POST" {
		// Handle form data
		cardID := r.FormValue("card_id")
		key := r.FormValue("key")
		value := r.FormValue("value")

		var cardIDInt int
		if _, err := fmt.Sscanf(cardID, "%d", &cardIDInt); err != nil {
			http.Error(w, "Invalid card ID", http.StatusBadRequest)
			return
		}

		req.CardID = cardIDInt
		req.Key = key

		// Try to parse value as different types
		if value == "true" {
			req.Value = true
		} else if value == "false" {
			req.Value = false
		} else if intVal, err := fmt.Sscanf(value, "%d", new(int)); err == nil && intVal == 1 {
			var parsed int
			fmt.Sscanf(value, "%d", &parsed)
			req.Value = parsed
		} else {
			req.Value = value
		}
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.CardID == 0 || req.Key == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set option
	if room.RoleOptionsManager == nil {
		room.RoleOptionsManager = game.NewRoleOptionsManager()
	}

	opts := room.RoleOptionsManager.GetOrCreateOptions(req.CardID)
	opts.SetOption(req.Key, req.Value)

	h.store.UpdateRoom(room)

	log.Printf("🔧 Role option set: card=%d, key=%s, value=%v", req.CardID, req.Key, req.Value)

	// Publish event to update all connected clients
	h.eventBus.Publish(Event{
		Type:     "role_options_changed",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// runCountdown runs the countdown timer
func (h *Handler) runCountdown(room *game.Room) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := 5; i > 0; i-- {
		room.CountdownRemaining = i
		h.store.UpdateRoom(room)
		log.Printf("⏰ Publishing countdown_update for room %s: %d", room.Code, i)

		h.eventBus.Publish(Event{
			Type:     "countdown_update",
			RoomCode: room.Code,
			Data:     room,
		})

		<-ticker.C
	}

	// Transition to playing state
	room.State = game.StatePlaying
	room.CountdownRemaining = 0
	room.LeaderRevealed = true
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "game_playing",
		RoomCode: room.Code,
		Data:     room,
	})
}

// UnveilPlayer handles the universal unveil action for any card
// For cards without special requirements, this simply sets them face up
// For cards with requirements, it redirects to the appropriate flow
func (h *Handler) UnveilPlayer(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("❌ Room not found: %s", roomCode)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		log.Printf("❌ No player cookie for room: %s", roomCode)
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	me := room.GetPlayer(playerCookie.Value)
	if me == nil {
		log.Printf("❌ Player not found in room: %s", roomCode)
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Get the target player
	target := room.GetPlayer(playerID)
	if target == nil {
		log.Printf("❌ Target player not found: %s", playerID)
		http.Error(w, "Target player not found", http.StatusBadRequest)
		return
	}

	// Authorization: only the player themselves can unveil their own card
	if me.ID != target.ID {
		log.Printf("❌ Player %s attempted to unveil %s's card (forbidden)", me.ID, target.ID)
		http.Error(w, "You can only unveil your own card", http.StatusForbidden)
		return
	}

	// Already face up - nothing to do
	if target.FaceUp {
		log.Printf("ℹ️ Player %s's card is already face up", target.Name)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if this card has special unveil requirements
	if target.Role != nil {
		req := ability.GetUnveilRequirements(target.Role.GetID())

		// If requires input, redirect to modal (this shouldn't normally happen
		// since the UI should show the modal button instead)
		if req.InputType != ability.NoInput {
			log.Printf("⚠️ Card %s requires input but simple unveil called - redirect needed", target.Role.Name)
			http.Error(w, "This card requires input before unveiling", http.StatusBadRequest)
			return
		}
	}

	// Simple unveil: set face up and mark as revealed
	target.FaceUp = true
	target.RoleRevealed = true
	h.store.UpdateRoom(room)

	log.Printf("🎭 Player %s unveiled their card (simple unveil) in room %s", target.Name, roomCode)

	// Publish event to update all connected clients
	h.eventBus.Publish(Event{
		Type:     "role_revealed",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// GetUnveilModal returns the X input modal for cards that need it
func (h *Handler) GetUnveilModal(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("❌ Room not found: %s", roomCode)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		log.Printf("❌ No player cookie for room: %s", roomCode)
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	me := room.GetPlayer(playerCookie.Value)
	if me == nil {
		log.Printf("❌ Player not found in room: %s", roomCode)
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Get the target player
	target := room.GetPlayer(playerID)
	if target == nil {
		log.Printf("❌ Target player not found: %s", playerID)
		http.Error(w, "Target player not found", http.StatusBadRequest)
		return
	}

	// Authorization: only the player themselves can see their unveil modal
	if me.ID != target.ID {
		log.Printf("❌ Player %s attempted to get %s's unveil modal (forbidden)", me.ID, target.ID)
		http.Error(w, "You can only unveil your own card", http.StatusForbidden)
		return
	}

	// Verify target has a role
	if target.Role == nil {
		http.Error(w, "Player has no role assigned", http.StatusBadRequest)
		return
	}

	// Get unveil requirements for this card
	req := ability.GetUnveilRequirements(target.Role.GetID())

	// Calculate max available cards based on CardPool and options
	maxAvailableCards := 0
	if room.CardPool != nil {
		// Check if Leaders should be included
		var useAllCards bool = false
		if room.RoleOptionsManager != nil && room.RoleOptionsManager.HasOptions(31) {
			opts := room.RoleOptionsManager.GetOrCreateOptions(31)
			if val, err := opts.GetBoolOption("use_all_cards"); err == nil {
				useAllCards = val
			}
		}

		// Get available cards based on options
		var filterTypes []string
		if !useAllCards {
			filterTypes = []string{"Guardian", "Assassin", "Traitor"}
		}
		availableCards := room.CardPool.GetCardsByTypes(filterTypes...)
		maxAvailableCards = len(availableCards)
	}

	// Ensure at least 1 as default if no cards available
	if maxAvailableCards == 0 {
		maxAvailableCards = 10 // Fallback
	}

	// Render the modal
	var buf bytes.Buffer
	if err := components.XInputModal(room, target, req, maxAvailableCards).Render(r.Context(), &buf); err != nil {
		log.Printf("❌ Failed to render X input modal: %v", err)
		http.Error(w, "Failed to render modal", http.StatusInternalServerError)
		return
	}

	// Send as SSE fragment using inner mode to preserve #modal-container element
	sse := datastar.NewSSE(w, r)
	sse.PatchElements(buf.String(),
		datastar.WithSelector("#modal-container"),
		datastar.WithModeInner())
}

// RestoreRoom restores a room from a client-provided encrypted backup
// This is called when a player attempts to access a room that doesn't exist
// (e.g., after Cloud Run instance replacement)
func (h *Handler) RestoreRoom(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		RoomCode string `json:"roomCode"`
		Backup   string `json:"backup"`
		PlayerID string `json:"playerID"`
	}

	if err := r.ParseForm(); err == nil && r.Method == "POST" {
		req.RoomCode = r.FormValue("roomCode")
		req.Backup = r.FormValue("backup")
		req.PlayerID = r.FormValue("playerID")
	}

	if req.RoomCode == "" || req.Backup == "" {
		log.Printf("❌ RestoreRoom: missing roomCode or backup")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	log.Printf("🔄 RestoreRoom attempt for room %s by player %s", req.RoomCode, req.PlayerID)

	// Check if room already exists (another player might have restored it first)
	if h.store.RoomExists(req.RoomCode) {
		log.Printf("✅ Room %s already exists (restored by another player)", req.RoomCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "exists"}`))
		return
	}

	// Check if backup service is available
	if h.backupService == nil {
		log.Printf("❌ RestoreRoom: backup service not available")
		http.Error(w, "Backup service not available", http.StatusServiceUnavailable)
		return
	}

	// Attempt restore from backup
	room, err := h.backupService.RestoreBackup(req.Backup, req.RoomCode)
	if err != nil {
		log.Printf("❌ RestoreRoom: backup restore failed for %s: %v", req.RoomCode, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Re-register the restored room
	if err := h.store.RegisterRestoredRoom(room); err != nil {
		log.Printf("❌ RestoreRoom: failed to register restored room %s: %v", req.RoomCode, err)
		http.Error(w, "Failed to restore room", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Room %s restored from backup by player %s", req.RoomCode, req.PlayerID)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "restored"}`))
}

// DebugClearRoom deletes a room to simulate Cloud Run instance restart
// This is only available when debugModeEnabled is true
func (h *Handler) DebugClearRoom(w http.ResponseWriter, r *http.Request) {
	// Only allow in debug mode
	if !h.config.Server.DebugModeEnabled {
		http.Error(w, "Debug endpoints only available when debugModeEnabled is true", http.StatusForbidden)
		return
	}

	roomCode := chi.URLParam(r, "code")
	if roomCode == "" {
		http.Error(w, "Room code required", http.StatusBadRequest)
		return
	}

	// Just verify the room exists
	if !h.store.RoomExists(roomCode) {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Delete the room
	h.store.DeleteRoom(roomCode)

	log.Printf("🗑️ DEBUG: Room %s cleared (simulating instance restart)", roomCode)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "cleared"}`))
}
