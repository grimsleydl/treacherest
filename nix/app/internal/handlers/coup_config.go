package handlers

import (
	"net/http"
	"strconv"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"

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
	if rejectPreStartSettingsMutationIfLocked(w, room) {
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
	counts, ok := game.CoupRoleCountsForPreset(preset)
	if !ok {
		http.Error(w, "Invalid Coup preset", http.StatusBadRequest)
		return
	}

	room.CoupPreset = preset
	room.CoupRoleCounts = counts
	room.CoupRoleCountsCustom = false
	room.CoupAllowUnsafeRoleCounts = false
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "coup_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	h.renderCoupConfigResponse(w, r, room)
}

// IncrementCoupPlayerCount selects the next supported default Coup preset.
func (h *Handler) IncrementCoupPlayerCount(w http.ResponseWriter, r *http.Request) {
	h.updateCoupPlayerCount(w, r, 1)
}

// DecrementCoupPlayerCount selects the previous supported default Coup preset.
func (h *Handler) DecrementCoupPlayerCount(w http.ResponseWriter, r *http.Request) {
	h.updateCoupPlayerCount(w, r, -1)
}

func (h *Handler) updateCoupPlayerCount(w http.ResponseWriter, r *http.Request, delta int) {
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
	if rejectPreStartSettingsMutationIfLocked(w, room) {
		return
	}

	currentCount, ok := game.CoupPresetPlayerCount(room.CoupPreset)
	if !ok {
		http.Error(w, "Current Coup preset is unsupported", http.StatusBadRequest)
		return
	}

	preset, ok := game.CoupDefaultPresetForPlayerCount(currentCount + delta)
	if !ok {
		h.renderCoupConfigResponse(w, r, room)
		return
	}
	counts, ok := game.CoupRoleCountsForPreset(preset)
	if !ok {
		http.Error(w, "Invalid Coup preset", http.StatusBadRequest)
		return
	}

	room.CoupPreset = preset
	if !room.CoupRoleCountsCustom {
		room.CoupRoleCounts = counts
		room.CoupAllowUnsafeRoleCounts = false
	}
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "coup_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	h.renderCoupConfigResponse(w, r, room)
}

// UpdateCoupRoleCounts updates the editable Coup role-count pool for a room.
func (h *Handler) UpdateCoupRoleCounts(w http.ResponseWriter, r *http.Request) {
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
	if rejectPreStartSettingsMutationIfLocked(w, room) {
		return
	}

	counts := make(game.CoupRoleCounts, len(game.CoupRoleCountOptions()))
	for _, role := range game.CoupRoleCountOptions() {
		rawCount := r.FormValue(game.CoupRoleCountFormName(role))
		if rawCount == "" {
			rawCount = "0"
		}
		count, err := strconv.Atoi(rawCount)
		if err != nil || count < 0 {
			http.Error(w, "Invalid Coup role count", http.StatusBadRequest)
			return
		}
		counts[role] = count
	}

	counts = game.NormalizeCoupRoleCounts(counts)
	unsafeRoleCounts := r.FormValue("unsafeRoleCounts") == "on"
	room.CoupRoleCounts = counts
	room.CoupAllowUnsafeRoleCounts = unsafeRoleCounts
	room.CoupRoleCountsCustom = unsafeRoleCounts || !coupRoleCountsMatchPreset(counts, room.CoupPreset)
	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "coup_config_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	h.renderCoupConfigResponse(w, r, room)
}

func coupRoleCountsMatchPreset(counts game.CoupRoleCounts, preset game.CoupPreset) bool {
	presetCounts, ok := game.CoupRoleCountsForPreset(preset)
	if !ok {
		return false
	}
	counts = game.NormalizeCoupRoleCounts(counts)
	presetCounts = game.NormalizeCoupRoleCounts(presetCounts)
	for _, role := range game.CoupRoleCountOptions() {
		if counts[role] != presetCounts[role] {
			return false
		}
	}
	return true
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
	if rejectPreStartSettingsMutationIfLocked(w, room) {
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

	h.renderCoupConfigResponse(w, r, room)
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
	if rejectPreStartSettingsMutationIfLocked(w, room) {
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

	h.renderCoupConfigResponse(w, r, room)
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
	if rejectPreStartSettingsMutationIfLocked(w, room) {
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

	h.renderCoupConfigResponse(w, r, room)
}

func (h *Handler) renderCoupConfigResponse(w http.ResponseWriter, r *http.Request, room *game.Room) {
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
	if room.IsOperatorSession(player.SessionID) || requestHasHostDashboardCookie(r, room.Code) {
		h.renderHostDashboardCoupConfigUpdate(sse, room)
		return
	}
	h.renderLobby(sse, room, player)
}

func (h *Handler) renderHostDashboardCoupConfigUpdate(sse *datastar.ServerSentEventGenerator, room *game.Room) {
	setupHTML := renderToString(pages.HostDashboardCoupSetup(room))
	sse.PatchElements(setupHTML, datastar.WithSelector("#host-dashboard-coup-setup"))

	startHTML := renderToString(pages.HostDashboardStartControls(room, h.config))
	sse.PatchElements(startHTML, datastar.WithSelector("#operator-start-controls"))
}

func requestHasHostDashboardCookie(r *http.Request, roomCode string) bool {
	hostCookie, err := r.Cookie("host_" + roomCode)
	return err == nil && hostCookie.Value == "true"
}
