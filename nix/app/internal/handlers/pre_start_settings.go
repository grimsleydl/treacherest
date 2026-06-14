package handlers

import (
	"net/http"
	"treacherest/internal/game"
)

const preStartSettingsLockedMessage = "Pre-start game settings are locked once the room leaves lobby."

func rejectPreStartSettingsMutationIfLocked(w http.ResponseWriter, room *game.Room) bool {
	if room.State == game.StateLobby {
		return false
	}

	http.Error(w, preStartSettingsLockedMessage, http.StatusConflict)
	return true
}
