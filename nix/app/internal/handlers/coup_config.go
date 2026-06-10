package handlers

import (
	"net/http"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar-go/datastar"
)

// UpdateCoupPreset updates the selected Coup role preset for a room.
func (h *Handler) UpdateCoupPreset(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if room.RulesMode != game.RulesModeCoup {
		http.Error(w, "Room is not using Coup rules", http.StatusBadRequest)
		return
	}

	if !h.isRoomCreator(r, room) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	preset := game.CoupPreset(r.FormValue("preset"))
	if preset == "" {
		http.Error(w, "Preset required", http.StatusBadRequest)
		return
	}
	if _, ok := game.CoupPresetPlayerCount(preset); !ok {
		http.Error(w, "Invalid Coup preset", http.StatusBadRequest)
		return
	}

	room.CoupPreset = preset
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "coup_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	playerCookie, err := r.Cookie("player_" + room.Code)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	sse := datastar.NewSSE(w, r)
	h.renderLobby(sse, room, player)
}
