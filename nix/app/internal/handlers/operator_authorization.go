package handlers

import (
	"net/http"
	"treacherest/internal/game"
)

func (h *Handler) isRoomOperator(r *http.Request, room *game.Room) bool {
	if room == nil {
		return false
	}
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return false
	}
	return room.IsOperatorSession(sessionCookie.Value)
}

func (h *Handler) debugControlsEnabled(r *http.Request, room *game.Room) bool {
	return h.config.Server.DebugModeEnabled && h.isRoomOperator(r, room)
}

func (h *Handler) debugViewedPlayer(room *game.Room) *game.Player {
	if room == nil || room.DebugViewedPlayerID == "" {
		return nil
	}

	selected := room.GetPlayer(room.DebugViewedPlayerID)
	if selected == nil || selected.IsHost {
		room.DebugViewedPlayerID = ""
		h.store.UpdateRoom(room)
		return nil
	}

	return selected
}

func (h *Handler) requireEffectivePlayer(w http.ResponseWriter, r *http.Request, room *game.Room, roomCode string) (*game.Player, bool) {
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return nil, false
	}

	cookiePlayer := room.GetPlayer(playerCookie.Value)
	if cookiePlayer == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return nil, false
	}

	if h.debugControlsEnabled(r, room) {
		if viewedPlayer := h.debugViewedPlayer(room); viewedPlayer != nil {
			return viewedPlayer, true
		}
	}

	return cookiePlayer, true
}

func (h *Handler) effectivePlayerForRender(r *http.Request, room *game.Room, fallback *game.Player) *game.Player {
	if h.debugControlsEnabled(r, room) {
		if viewedPlayer := h.debugViewedPlayer(room); viewedPlayer != nil {
			return viewedPlayer
		}
	}
	if room == nil || fallback == nil {
		return nil
	}
	return room.GetPlayer(fallback.ID)
}
