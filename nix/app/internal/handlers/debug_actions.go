package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar-go/datastar"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"
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
	room.DebugStartMode = game.DebugStartModeWithDebugPlayers
	h.store.UpdateRoom(room)

	h.StartGame(w, r)
}

func (h *Handler) DebugStartAsIs(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	room, ok := h.requireDebugHostRoom(w, r, roomCode)
	if !ok {
		return
	}
	if room.State != game.StateLobby {
		http.Error(w, "Debug start override requires a room in lobby state", http.StatusConflict)
		return
	}
	if room.GetActivePlayerCount() == 0 {
		http.Error(w, "Start As-Is requires at least one active player", http.StatusBadRequest)
		return
	}

	if room.RulesMode == game.RulesModeCoup {
		if err := game.AssignCoupRolesBestEffort(room.GetPlayers(), room.CoupPreset, room.CoupInfoPolicy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if h.cardService == nil {
			http.Error(w, "Internal server error: Cannot assign roles", http.StatusInternalServerError)
			return
		}
		roleService := game.NewRoleConfigService(h.config)
		game.AssignRolesWithConfig(room.GetPlayers(), h.cardService, room.RoleConfig, roleService)
	}

	room.DebugStartMode = game.DebugStartModeAsIs
	h.finishDebugStartedRoom(w, r, room)
}

func (h *Handler) DebugViewAsPlayer(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	room, ok := h.requireDebugHostRoom(w, r, roomCode)
	if !ok {
		return
	}

	playerID := chi.URLParam(r, "playerID")
	selected := room.GetPlayer(playerID)
	if selected == nil || selected.IsHost {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	pages.DebugViewAsPlayerPerspective(room, selected, h.config, h.cardService).Render(r.Context(), w)
}

func (h *Handler) finishDebugStartedRoom(w http.ResponseWriter, r *http.Request, room *game.Room) {
	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	go h.runCountdown(room)

	h.eventBus.Publish(Event{
		Type:     "game_started",
		RoomCode: room.Code,
		Data:     room,
	})

	sse := datastar.NewSSE(w, r)
	sse.ExecuteScript("window.location.href = '/game/" + room.Code + "'")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
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
