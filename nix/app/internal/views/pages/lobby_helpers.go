package pages

import (
	"fmt"
	"strings"
	"treacherest/internal/game"
)

// LobbySettingsSummary is server-side display text for read-only room settings.
func LobbySettingsSummary(room *game.Room) string {
	if room == nil {
		return "Treachery - 0 players"
	}

	if room.RulesMode == game.RulesModeCoup {
		return strings.Join([]string{
			"Coup",
			fmt.Sprintf("%d players", lobbySeatCount(room)),
			coupInquisitionSummary(room),
			coupGreenHuntSummary(room),
			coupInquisitionAmnestySummary(room),
			coupKingKnowledgeSummary(room),
		}, " - ")
	}

	return strings.Join([]string{
		"Treachery",
		fmt.Sprintf("%d players", lobbySeatCount(room)),
	}, " - ")
}

func LobbyWaitingStatus(room *game.Room) string {
	if room == nil {
		return "Waiting for Room Operator - 0 of 0 seats filled"
	}
	return fmt.Sprintf("Waiting for Room Operator - %d of %d seats filled", room.GetActivePlayerCount(), lobbySeatCount(room))
}

func lobbySeatCount(room *game.Room) int {
	if room == nil {
		return 0
	}
	if room.RulesMode == game.RulesModeCoup {
		total := 0
		for _, count := range game.CoupRoleCountsForRoom(room) {
			total += count
		}
		if total > 0 {
			return total
		}
		if count, ok := game.CoupPresetPlayerCount(room.CoupPreset); ok {
			return count
		}
	}
	if room.RoleConfig != nil && room.RoleConfig.MaxPlayers > 0 {
		return room.RoleConfig.MaxPlayers
	}
	if room.MaxPlayers > 0 {
		return room.MaxPlayers
	}
	return room.GetActivePlayerCount()
}

func coupInquisitionSummary(room *game.Room) string {
	if game.NormalizeCoupInquisitionResultPolicy(room.CoupInquisitionResultPolicy) == game.CoupInquisitionResultPrivate {
		return "Private inquisition"
	}
	return "Public inquisition"
}

func coupGreenHuntSummary(room *game.Room) string {
	if game.NormalizeCoupGreenHuntRequirement(room.CoupGreenHuntRequirement) == game.CoupGreenHuntAllBlues {
		return "All Blue hunt"
	}
	return "One Blue hunt"
}

func coupInquisitionAmnestySummary(room *game.Room) string {
	if game.NormalizeCoupInquisitionAmnesty(room.CoupInquisitionAmnesty) == game.CoupInquisitionAmnestyBroad {
		return "Broad Amnesty"
	}
	return "King victory only"
}

func coupKingKnowledgeSummary(room *game.Room) string {
	switch game.NormalizeCoupInformationPolicy(room.CoupInfoPolicy).KingToBlue {
	case game.CoupKingGetsBlueCandidates:
		return "Soft King knowledge"
	case game.CoupKingKnowsNoBlue:
		return "No King knowledge"
	default:
		return "Full King knowledge"
	}
}
