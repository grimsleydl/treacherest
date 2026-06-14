package pages

import (
	"treacherest/internal/config"
	"treacherest/internal/game"
)

type hostDashboardStartState struct {
	CanStart bool
	Message  string
}

func HostDashboardCanStart(room *game.Room, cfg *config.ServerConfig) bool {
	return hostDashboardStartStateFor(room, cfg).CanStart
}

func HostDashboardStartMessage(room *game.Room, cfg *config.ServerConfig) string {
	return hostDashboardStartStateFor(room, cfg).Message
}

func hostDashboardStartStateFor(room *game.Room, cfg *config.ServerConfig) hostDashboardStartState {
	if room == nil {
		return hostDashboardStartState{CanStart: false, Message: "Room is unavailable"}
	}

	if room.RulesMode == game.RulesModeCoup {
		if coupPresetMatchesActivePlayers(room) {
			return hostDashboardStartState{CanStart: true, Message: "Ready to start"}
		}
		return hostDashboardStartState{CanStart: false, Message: coupPresetValidationText(room)}
	}

	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	state := room.GetValidationState(game.NewRoleConfigService(cfg))
	if state.CanStart {
		if state.ValidationMessage != "" {
			return hostDashboardStartState{CanStart: true, Message: state.ValidationMessage}
		}
		return hostDashboardStartState{CanStart: true, Message: "Ready to start"}
	}
	if state.ValidationMessage != "" {
		return hostDashboardStartState{CanStart: false, Message: state.ValidationMessage}
	}
	return hostDashboardStartState{CanStart: false, Message: "Room is not ready to start"}
}
