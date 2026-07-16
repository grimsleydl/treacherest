package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"treacherest/internal/game"
)

// ActivateLeaderCounter records the owner's public identity-card activation.
// Battlefield targets, costs, and effects remain table-adjudicated.
func (h *Handler) ActivateLeaderCounter(w http.ResponseWriter, r *http.Request) {
	room, owner, ok := h.requireLeaderCounterOwner(w, r)
	if !ok {
		return
	}
	if !owner.FaceUp {
		http.Error(w, "Leader identity must be face up", http.StatusConflict)
		return
	}

	activation, err := room.ActivateLeaderCounter(owner.ID)
	if err != nil {
		switch {
		case errors.Is(err, game.ErrLeaderCounterEmpty), errors.Is(err, game.ErrLeaderCounterUsed):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	h.store.UpdateRoom(room)
	h.eventBus.Publish(Event{
		Type:     "leader_counter_activated",
		RoomCode: room.Code,
		Data:     activation,
	})
	log.Printf("Leader counter activated by %s in room %s: %d -> %d", owner.Name, room.Code, activation.Before, activation.After)
	w.WriteHeader(http.StatusNoContent)
}

// ResetLeaderCounterTurn lets the owner confirm the table has advanced to a
// new turn. The app deliberately does not invent a turn engine.
func (h *Handler) ResetLeaderCounterTurn(w http.ResponseWriter, r *http.Request) {
	room, owner, ok := h.requireLeaderCounterOwner(w, r)
	if !ok {
		return
	}
	if err := room.ResetLeaderCounterTurn(owner.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.store.UpdateRoom(room)
	h.eventBus.Publish(Event{
		Type:     "leader_counter_turn_reset",
		RoomCode: room.Code,
		Data:     owner.Role,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requireLeaderCounterOwner(w http.ResponseWriter, r *http.Request) (*game.Room, *game.Player, bool) {
	roomCode := chi.URLParam(r, "code")
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return nil, nil, false
	}
	if room.State != game.StatePlaying {
		http.Error(w, "Leader counters can only be used while the game is playing", http.StatusConflict)
		return nil, nil, false
	}

	effectivePlayer, ok := h.requireEffectivePlayer(w, r, room, roomCode)
	if !ok {
		return nil, nil, false
	}
	playerID := chi.URLParam(r, "playerID")
	owner := room.GetPlayer(playerID)
	if owner == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return nil, nil, false
	}
	if effectivePlayer.ID != owner.ID {
		http.Error(w, "You can only activate your own Leader identity", http.StatusForbidden)
		return nil, nil, false
	}
	if owner.IsEliminated {
		http.Error(w, "Eliminated players cannot activate Leader counters", http.StatusConflict)
		return nil, nil, false
	}
	if owner.Role == nil || owner.Role.LeaderCounterPool == nil {
		http.Error(w, game.ErrLeaderCounterUnavailable.Error(), http.StatusBadRequest)
		return nil, nil, false
	}
	return room, owner, true
}
