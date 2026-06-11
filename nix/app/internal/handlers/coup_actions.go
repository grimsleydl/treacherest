package handlers

import (
	"log"
	"net/http"
	"strconv"
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

	effectivePlayer, ok := h.requireEffectivePlayer(w, r, room, roomCode)
	if !ok {
		return
	}

	player := room.GetPlayer(playerID)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}
	if effectivePlayer.ID != player.ID {
		http.Error(w, "You can only use Royal Guard for your own role", http.StatusForbidden)
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

// CallCoupInquisition starts the default public Inquisition flow.
func (h *Handler) CallCoupInquisition(w http.ResponseWriter, r *http.Request) {
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

	effectivePlayer, ok := h.requireEffectivePlayer(w, r, room, roomCode)
	if !ok {
		return
	}

	player := room.GetPlayer(playerID)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}
	if effectivePlayer.ID != player.ID {
		http.Error(w, "You can only call Inquisition for your own role", http.StatusForbidden)
		return
	}
	if player.Role == nil || player.Role.GetRoleType() != game.RoleBlueKnight {
		http.Error(w, "Only Blue Knights can call Inquisition", http.StatusForbidden)
		return
	}
	if player.IsEliminated {
		http.Error(w, "Eliminated players cannot call Inquisition", http.StatusBadRequest)
		return
	}

	state := game.EnsureCoupInquisitionState(room)
	if _, ok := state.Attempts[player.ID]; ok {
		http.Error(w, "This Blue Knight has already called Inquisition", http.StatusBadRequest)
		return
	}
	if state.Pending != nil {
		http.Error(w, "An Inquisition is already pending witness confirmation", http.StatusBadRequest)
		return
	}

	targetID := r.FormValue("targetID")
	target := room.GetPlayer(targetID)
	if target == nil || target.IsHost || target.IsEliminated {
		http.Error(w, "Invalid Inquisition target", http.StatusBadRequest)
		return
	}
	if target.Role != nil && target.Role.GetRoleType() == game.RoleKing {
		http.Error(w, "King is not a valid Inquisition target", http.StatusBadRequest)
		return
	}
	currentLife, err := strconv.Atoi(r.FormValue("currentLife"))
	if err != nil || currentLife < 0 {
		http.Error(w, "Invalid current life total", http.StatusBadRequest)
		return
	}

	player.RoleRevealed = true
	player.FaceUp = true
	attempt := game.CoupInquisitionAttempt{
		InquisitorID: player.ID,
		TargetID:     target.ID,
		CurrentLife:  currentLife,
		PenaltyLife:  game.CoupInquisitionPenalty(currentLife),
	}
	state.Pending = &attempt
	state.Attempts[player.ID] = attempt
	h.store.UpdateRoom(room)

	log.Printf("🔎 Blue Knight %s called Inquisition naming %s in room %s", player.Name, target.Name, roomCode)

	h.eventBus.Publish(Event{
		Type:     "coup_inquisition_called",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// ConfirmCoupInquisition resolves the pending Inquisition after a living non-Blue witness confirms notice.
func (h *Handler) ConfirmCoupInquisition(w http.ResponseWriter, r *http.Request) {
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

	state := game.EnsureCoupInquisitionState(room)
	if state.Pending == nil {
		http.Error(w, "No Inquisition pending confirmation", http.StatusBadRequest)
		return
	}

	witness, ok := h.requireEffectivePlayer(w, r, room, roomCode)
	if !ok {
		return
	}
	if witness.IsHost || witness.IsEliminated || witness.ID == state.Pending.InquisitorID ||
		(witness.Role != nil && witness.Role.GetRoleType() == game.RoleBlueKnight) {
		http.Error(w, "A living non-Blue witness must confirm Inquisition", http.StatusForbidden)
		return
	}

	target := room.GetPlayer(state.Pending.TargetID)
	if target == nil {
		http.Error(w, "Inquisition target not found", http.StatusBadRequest)
		return
	}

	result := *state.Pending
	result.ConfirmedBy = witness.ID
	result.Resolved = true
	result.Success = target.Role != nil && target.Role.GetRoleType() == game.RoleRedKnight
	if result.Success {
		if game.NormalizeCoupInquisitionResultPolicy(room.CoupInquisitionResultPolicy) == game.CoupInquisitionResultPublic {
			target.RoleRevealed = true
			target.FaceUp = true
		}
		state.Succeeded = true
	}
	state.Last = &result
	state.Attempts[result.InquisitorID] = result
	state.Pending = nil
	h.store.UpdateRoom(room)

	log.Printf("🔎 Inquisition in room %s confirmed by %s (success: %v)", roomCode, witness.Name, result.Success)

	h.eventBus.Publish(Event{
		Type:     "coup_inquisition_resolved",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}
