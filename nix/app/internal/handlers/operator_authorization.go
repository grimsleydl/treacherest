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
