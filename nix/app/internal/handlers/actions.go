package handlers

import (
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
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
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify player is in room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		log.Printf("❌ No player cookie for room: %s", roomCode)
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		log.Printf("❌ Player not found in room: %s, cookie: %s", roomCode, playerCookie.Value)
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ Player %s attempting to start game in room %s", player.Name, roomCode)

	// Check if game can start
	if !room.CanStart() {
		log.Printf("❌ Room cannot start: state=%s, players=%d", room.State, len(room.Players))
		http.Error(w, "Cannot start game", http.StatusBadRequest)
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
		http.Error(w, "Internal server error: CardService not available", http.StatusInternalServerError)
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
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
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
