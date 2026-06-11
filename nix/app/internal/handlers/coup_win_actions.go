package handlers

import (
	"net/http"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

// ConfirmCoupWinPrompt records table confirmation of the current advisory win prompt.
func (h *Handler) ConfirmCoupWinPrompt(w http.ResponseWriter, r *http.Request) {
	room, _, ok := h.coupWinDecisionContext(w, r)
	if !ok {
		return
	}

	prompt := game.CurrentCoupAdvisoryWin(room)
	if prompt == nil {
		http.Error(w, "No advisory Coup win prompt is available", http.StatusBadRequest)
		return
	}

	game.ConfirmCoupWin(room, prompt)
	room.State = game.StateEnded
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "game_ended",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// RejectCoupWinPrompt records table rejection of the current advisory win prompt.
func (h *Handler) RejectCoupWinPrompt(w http.ResponseWriter, r *http.Request) {
	room, _, ok := h.coupWinDecisionContext(w, r)
	if !ok {
		return
	}

	prompt := game.CurrentCoupAdvisoryWin(room)
	if prompt == nil {
		http.Error(w, "No advisory Coup win prompt is available", http.StatusBadRequest)
		return
	}

	game.RejectCoupWinPrompt(room, prompt)
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "coup_win_prompt_rejected",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) coupWinDecisionContext(w http.ResponseWriter, r *http.Request) (*game.Room, *game.Player, bool) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return nil, nil, false
	}
	if room.RulesMode != game.RulesModeCoup {
		http.Error(w, "Room is not using Coup rules", http.StatusBadRequest)
		return nil, nil, false
	}
	if room.State != game.StatePlaying {
		http.Error(w, "Coup win prompts can only be decided while the game is playing", http.StatusBadRequest)
		return nil, nil, false
	}

	playerCookie, err := r.Cookie("player_" + room.Code)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return nil, nil, false
	}
	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return nil, nil, false
	}
	if !player.IsHost && !player.IsActiveInGame() {
		http.Error(w, "Only active players or host/spectators can decide advisory win prompts", http.StatusForbidden)
		return nil, nil, false
	}

	return room, player, true
}
