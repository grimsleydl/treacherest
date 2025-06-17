package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
	"treacherest/internal/game"
)

// StartGame starts a game
func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify player is in room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	// Check if game can start
	if !room.CanStart() {
		http.Error(w, "Cannot start game", http.StatusBadRequest)
		return
	}

	// Assign roles
	players := room.GetPlayers()
	game.AssignRoles(players)

	// Update game state
	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	// Start countdown
	go h.runCountdown(room)

	// Notify all players
	h.eventBus.Publish(Event{
		Type:     "game_started",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
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

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
