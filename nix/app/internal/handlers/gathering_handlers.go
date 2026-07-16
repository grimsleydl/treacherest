package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"treacherest/internal/game"
)

// ChooseGatheringMode records the owner's public mode choice. Targets and the
// printed battlefield effect remain table-adjudicated.
func (h *Handler) ChooseGatheringMode(w http.ResponseWriter, r *http.Request) {
	room, owner, ok := h.requireGatheringOwner(w, r)
	if !ok {
		return
	}

	choice, err := room.ChooseGatheringMode(owner.ID, game.GatheringMode(chi.URLParam(r, "mode")))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, game.ErrGatheringModeChosen) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	h.store.UpdateRoom(room)
	h.eventBus.Publish(Event{
		Type:     "gathering_mode_chosen",
		RoomCode: room.Code,
		Data:     choice,
	})
	log.Printf("The Gathering mode %s chosen by %s in room %s", choice.Mode, owner.Name, room.Code)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requireGatheringOwner(w http.ResponseWriter, r *http.Request) (*game.Room, *game.Player, bool) {
	roomCode := chi.URLParam(r, "code")
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return nil, nil, false
	}
	if room.State != game.StatePlaying {
		http.Error(w, "The Gathering modes can only be chosen while the game is playing", http.StatusConflict)
		return nil, nil, false
	}

	effectivePlayer, ok := h.requireEffectivePlayer(w, r, room, roomCode)
	if !ok {
		return nil, nil, false
	}
	owner := room.GetPlayer(chi.URLParam(r, "playerID"))
	if owner == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return nil, nil, false
	}
	if effectivePlayer.ID != owner.ID {
		http.Error(w, "You can only choose modes for your own Gathering identity", http.StatusForbidden)
		return nil, nil, false
	}
	if owner.IsEliminated {
		http.Error(w, "Eliminated players cannot choose Gathering modes", http.StatusConflict)
		return nil, nil, false
	}
	if !owner.FaceUp {
		http.Error(w, "The Gathering identity must be face up", http.StatusConflict)
		return nil, nil, false
	}
	if owner.Role == nil || owner.Role.ID != game.TheGatheringCardID || owner.Role.GatheringChecklist == nil {
		http.Error(w, game.ErrGatheringUnavailable.Error(), http.StatusBadRequest)
		return nil, nil, false
	}
	return room, owner, true
}
