package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
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
		sse.MergeFragments(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndMergeSignals(map[string]interface{}{
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
		sse.MergeFragments(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndMergeSignals(map[string]interface{}{
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
		sse.MergeFragments(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndMergeSignals(map[string]interface{}{
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
		err := sse.MergeFragments(errorHTML, datastar.WithSelector("#error-container"))
		if err != nil {
			log.Printf("❌ Failed to send error fragment: %v", err)
		}

		// Also update button state and re-sync ALL validation signals
		err = sse.MarshalAndMergeSignals(map[string]interface{}{
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
		sse.MergeFragments(errorHTML, datastar.WithSelector("#error-container"))
		sse.MarshalAndMergeSignals(map[string]interface{}{
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
