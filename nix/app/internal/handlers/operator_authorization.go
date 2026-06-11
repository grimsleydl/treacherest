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
