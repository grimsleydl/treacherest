package handlers

import (
	"net/http"
	"strconv"
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

// UpdateCoupInfoPolicy updates the selected Coup private information policy for a room.
func (h *Handler) UpdateCoupInfoPolicy(w http.ResponseWriter, r *http.Request) {
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

	policy := game.CoupInformationPolicy{
		KingToBlue:   game.CoupKingToBluePolicy(r.FormValue("kingToBlue")),
		RedToBlack:   game.CoupRedToBlackPolicy(r.FormValue("redToBlack")),
		BlackToRed:   game.CoupBlackToRedPolicy(r.FormValue("blackToRed")),
		BlackNetwork: game.CoupBlackNetworkPolicy(r.FormValue("blackNetwork")),
	}
	policy = game.NormalizeCoupInformationPolicy(policy)
	if !game.IsValidCoupInformationPolicy(policy) {
		http.Error(w, "Invalid Coup information policy", http.StatusBadRequest)
		return
	}

	room.CoupInfoPolicy = policy
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

// UpdateCoupRoyalGuardSettings updates Coup Royal Guard rule settings for a room.
func (h *Handler) UpdateCoupRoyalGuardSettings(w http.ResponseWriter, r *http.Request) {
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

	blockerLimit, err := strconv.Atoi(r.FormValue("blockerLimit"))
	if err != nil {
		http.Error(w, "Invalid Royal Guard blocker limit", http.StatusBadRequest)
		return
	}

	room.CoupRoyalGuardBlockerLimit = game.NormalizeCoupRoyalGuardBlockerLimit(blockerLimit)
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

// UpdateCoupInquisitionSettings updates Coup Inquisition rule settings for a room.
func (h *Handler) UpdateCoupInquisitionSettings(w http.ResponseWriter, r *http.Request) {
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

	policy := game.CoupInquisitionResultPolicy(r.FormValue("resultPolicy"))
	room.CoupInquisitionResultPolicy = game.NormalizeCoupInquisitionResultPolicy(policy)
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
