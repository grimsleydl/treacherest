package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"treacherest/internal/game"
)

// PlaceDebtCounter records The Debt Collector's public end-step table action.
func (h *Handler) PlaceDebtCounter(w http.ResponseWriter, r *http.Request) {
	room, collector, ok := h.requireDebtCollectorOwner(w, r)
	if !ok {
		return
	}

	targetID := chi.URLParam(r, "targetID")
	placement, err := room.PlaceDebtCounter(collector.ID, targetID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, game.ErrDebtCounterTarget) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	h.store.UpdateRoom(room)
	h.eventBus.Publish(Event{
		Type:     "debt_counter_placed",
		RoomCode: room.Code,
		Data:     placement,
	})
	log.Printf("Debt counter placed by %s on %s's identity card in room %s: %d -> %d", collector.Name, placement.TargetPlayerName, room.Code, placement.Before, placement.After)
	w.WriteHeader(http.StatusNoContent)
}

// RecordDebtCounterUpkeep records a public advisory snapshot. Life totals and
// the resulting life gain remain table state.
func (h *Handler) RecordDebtCounterUpkeep(w http.ResponseWriter, r *http.Request) {
	room, collector, ok := h.requireDebtCollectorOwner(w, r)
	if !ok {
		return
	}

	advisory, err := room.RecordDebtCounterUpkeep(collector.ID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, game.ErrNoDebtCounters) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	h.store.UpdateRoom(room)
	h.eventBus.Publish(Event{
		Type:     "debt_counter_upkeep_recorded",
		RoomCode: room.Code,
		Data:     advisory,
	})
	log.Printf("Debt counter upkeep advisory recorded by %s in room %s for %d total counters", collector.Name, room.Code, advisory.TotalLifeGain)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requireDebtCollectorOwner(w http.ResponseWriter, r *http.Request) (*game.Room, *game.Player, bool) {
	roomCode := chi.URLParam(r, "code")
	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return nil, nil, false
	}
	if room.State != game.StatePlaying {
		http.Error(w, "Debt counters can only be used while the game is playing", http.StatusConflict)
		return nil, nil, false
	}

	effectivePlayer, ok := h.requireEffectivePlayer(w, r, room, roomCode)
	if !ok {
		return nil, nil, false
	}
	collectorID := chi.URLParam(r, "playerID")
	collector := room.GetPlayer(collectorID)
	if collector == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return nil, nil, false
	}
	if effectivePlayer.ID != collector.ID {
		http.Error(w, "You can only use your own Debt Collector identity", http.StatusForbidden)
		return nil, nil, false
	}
	if collector.IsEliminated {
		http.Error(w, "Eliminated players cannot place debt counters", http.StatusConflict)
		return nil, nil, false
	}
	if !collector.FaceUp {
		http.Error(w, "The Debt Collector identity must be face up", http.StatusConflict)
		return nil, nil, false
	}
	if collector.Role == nil || collector.Role.ID != game.TheDebtCollectorCardID {
		http.Error(w, game.ErrDebtCollectorUnavailable.Error(), http.StatusBadRequest)
		return nil, nil, false
	}
	return room, collector, true
}
