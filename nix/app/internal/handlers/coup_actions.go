package handlers

import (
	"log"
	"net/http"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

// UseCoupRoyalGuard records that a Blue Knight is using Royal Guard.
// Royal Guard is a table action, so using it publicly reveals Blue.
func (h *Handler) UseCoupRoyalGuard(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	if room.RulesMode != game.RulesModeCoup {
		http.Error(w, "Room is not using Coup rules", http.StatusBadRequest)
		return
	}

	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}
	if playerCookie.Value != playerID {
		http.Error(w, "You can only use Royal Guard for your own role", http.StatusForbidden)
		return
	}

	player := room.GetPlayer(playerID)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}
	if player.Role == nil || player.Role.GetRoleType() != game.RoleBlueKnight {
		http.Error(w, "Only Blue Knights can use Royal Guard", http.StatusForbidden)
		return
	}
	if player.IsEliminated {
		http.Error(w, "Eliminated players cannot use Royal Guard", http.StatusBadRequest)
		return
	}

	player.RoleRevealed = true
	player.FaceUp = true
	h.store.UpdateRoom(room)

	log.Printf("🛡️ Blue Knight %s used Royal Guard in room %s", player.Name, roomCode)

	h.eventBus.Publish(Event{
		Type:     "role_revealed",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}
