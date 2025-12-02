package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar-go/datastar"
	"html"
	"log"
	"net/http"
	"time"
	"treacherest/internal/game"
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

	// Authorization: only the player themselves can reveal their own role
	if me.ID != target.ID {
		log.Printf("❌ Player %s attempted to reveal %s's role (forbidden)", me.ID, target.ID)
		http.Error(w, "You can only reveal your own role", http.StatusForbidden)
		return
	}

	// Leaders cannot hide their role (they start face-up per game rules)
	if target.Role != nil && target.Role.GetRoleType() == game.RoleLeader && target.RoleRevealed {
		log.Printf("❌ Leader %s attempted to hide their role (not allowed)", target.Name)
		http.Error(w, "Leaders cannot hide their role", http.StatusForbidden)
		return
	}

	// Toggle the reveal state
	target.RoleRevealed = !target.RoleRevealed
	h.store.UpdateRoom(room)

	log.Printf("🎭 Player %s toggled role reveal to %v in room %s", target.Name, target.RoleRevealed, roomCode)

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

// SetRoleOption sets a role option for a specific card
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
