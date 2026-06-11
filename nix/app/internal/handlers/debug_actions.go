package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"treacherest/internal/game"
)

func (h *Handler) DebugStartWithDebugPlayers(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	room, ok := h.requireDebugHostRoom(w, r, roomCode)
	if !ok {
		return
	}
	if room.State != game.StateLobby {
		http.Error(w, "Debug start override requires a room in lobby state", http.StatusConflict)
		return
	}

	targetCount, err := debugStartTargetPlayerCount(room)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := addDebugPlayers(room, targetCount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.store.UpdateRoom(room)

	h.StartGame(w, r)
}

func (h *Handler) requireDebugHostRoom(w http.ResponseWriter, r *http.Request, roomCode string) (*game.Room, bool) {
	if !h.config.Server.DebugModeEnabled {
		http.Error(w, "Debug endpoints only available when debugModeEnabled is true", http.StatusForbidden)
		return nil, false
	}

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return nil, false
	}

	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Host access required", http.StatusUnauthorized)
		return nil, false
	}
	player := room.GetPlayer(playerCookie.Value)
	if player == nil || !player.IsHost {
		http.Error(w, "Host access required", http.StatusForbidden)
		return nil, false
	}

	return room, true
}

func debugStartTargetPlayerCount(room *game.Room) (int, error) {
	if room.RulesMode == game.RulesModeCoup {
		targetCount, ok := game.CoupPresetPlayerCount(room.CoupPreset)
		if !ok {
			return 0, fmt.Errorf("unknown Coup preset %q", room.CoupPreset)
		}
		return targetCount, nil
	}
	if room.RoleConfig != nil && room.RoleConfig.MaxPlayers > 0 {
		return room.RoleConfig.MaxPlayers, nil
	}
	if room.MaxPlayers > 0 {
		return room.MaxPlayers, nil
	}
	return 0, fmt.Errorf("room has no target player count")
}

func addDebugPlayers(room *game.Room, targetCount int) error {
	for room.GetActivePlayerCount() < targetCount {
		index := nextDebugPlayerIndex(room)
		player := game.NewPlayer(
			fmt.Sprintf("debug-%s-%d", room.Code, index),
			fmt.Sprintf("Debug Player %d", index),
			fmt.Sprintf("debug-session-%s-%d", room.Code, index),
		)
		player.IsDebug = true
		if err := room.AddPlayer(player); err != nil {
			return err
		}
	}
	return nil
}

func nextDebugPlayerIndex(room *game.Room) int {
	next := 1
	for _, player := range room.GetPlayers() {
		var index int
		if _, err := fmt.Sscanf(player.Name, "Debug Player %d", &index); err == nil && index >= next {
			next = index + 1
		}
	}
	return next
}
