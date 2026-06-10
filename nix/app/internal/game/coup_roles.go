package game

import (
	"fmt"
	"math/rand"
	"strings"
)

const (
	RoleKing             RoleType   = "King"
	RoleBlueKnight       RoleType   = "Blue Knight"
	RoleBlackKnight      RoleType   = "Black Knight"
	RoleRedKnight        RoleType   = "Red Knight"
	RoleGreenKnight      RoleType   = "Green Knight"
	RoleWasteland        RoleType   = "Wasteland Knight"
	CoupPresetFive       CoupPreset = "coup-5"
	CoupPresetSix        CoupPreset = "coup-6"
	CoupPresetSeven      CoupPreset = "coup-7"
	CoupPresetEight      CoupPreset = "coup-8"
	CoupPresetEightChaos CoupPreset = "coup-8-chaos"
	CoupPresetNine       CoupPreset = "coup-9"
)

// CoupPreset identifies a concrete Coup role distribution.
type CoupPreset string

var coupRoleCards = map[RoleType]*Card{
	RoleKing:        coupRoleCard(1001, string(RoleKing), "Starts revealed. Political center of the table.", "Win if alive when Black, Red, and Wasteland threats are eliminated."),
	RoleBlueKnight:  coupRoleCard(1002, string(RoleBlueKnight), "Protects the King. Can call Inquisition.", "Win with the King. Lose when the King loses."),
	RoleBlackKnight: coupRoleCard(1003, string(RoleBlackKnight), "Assassin hired to kill the King, then betray Red.", "Win if the King is dead, at least one Black Knight survives, and Red is dead."),
	RoleRedKnight:   coupRoleCard(1004, string(RoleRedKnight), "Usurper who hired Black to kill the King.", "Win if the King is dead, Red survives, and all Black Knights are dead."),
	RoleGreenKnight: coupRoleCard(1005, string(RoleGreenKnight), "Opportunist with conditional shared victories.", "Win with King or Red only when eligible under the selected Green rules."),
	RoleWasteland:   coupRoleCard(1006, string(RoleWasteland), "Chaos role for larger tables. Wants everyone else dead.", "Win only as the sole surviving player. Never share victory."),
}

var coupPresetRoles = map[CoupPreset][]RoleType{
	CoupPresetFive:       {RoleKing, RoleBlueKnight, RoleBlackKnight, RoleRedKnight, RoleGreenKnight},
	CoupPresetSix:        {RoleKing, RoleBlueKnight, RoleBlackKnight, RoleBlackKnight, RoleRedKnight, RoleGreenKnight},
	CoupPresetSeven:      {RoleKing, RoleBlueKnight, RoleBlueKnight, RoleBlackKnight, RoleBlackKnight, RoleRedKnight, RoleGreenKnight},
	CoupPresetEight:      {RoleKing, RoleBlueKnight, RoleBlueKnight, RoleBlackKnight, RoleBlackKnight, RoleBlackKnight, RoleRedKnight, RoleGreenKnight},
	CoupPresetEightChaos: {RoleKing, RoleBlueKnight, RoleBlueKnight, RoleBlackKnight, RoleBlackKnight, RoleRedKnight, RoleGreenKnight, RoleWasteland},
	CoupPresetNine:       {RoleKing, RoleBlueKnight, RoleBlueKnight, RoleBlackKnight, RoleBlackKnight, RoleBlackKnight, RoleRedKnight, RoleGreenKnight, RoleWasteland},
}

var coupPresetOrder = []CoupPreset{
	CoupPresetFive,
	CoupPresetSix,
	CoupPresetSeven,
	CoupPresetEight,
	CoupPresetEightChaos,
	CoupPresetNine,
}

var coupPresetLabels = map[CoupPreset]string{
	CoupPresetFive:       "5 players",
	CoupPresetSix:        "6 players",
	CoupPresetSeven:      "7 players",
	CoupPresetEight:      "8 players",
	CoupPresetEightChaos: "8 players, chaos",
	CoupPresetNine:       "9 players",
}

var coupRoleSummaryOrder = []RoleType{
	RoleKing,
	RoleBlueKnight,
	RoleBlackKnight,
	RoleRedKnight,
	RoleGreenKnight,
	RoleWasteland,
}

func coupRoleCard(id int, name, text, winCondition string) *Card {
	return &Card{
		ID:     id,
		Name:   name,
		Type:   "Coup Role",
		Rarity: "Coup",
		Types: CardTypes{
			Supertype: "Coup",
			Subtype:   name,
		},
		Text:    text,
		Rulings: []string{"Win Condition: " + winCondition},
	}
}

// NormalizeCoupPreset returns the default Coup preset when no preset is selected.
func NormalizeCoupPreset(preset CoupPreset) CoupPreset {
	if preset == "" {
		return CoupPresetFive
	}
	return preset
}

// CoupPresetPlayerCount returns the active player count required by a Coup preset.
func CoupPresetPlayerCount(preset CoupPreset) (int, bool) {
	roles, ok := coupPresetRoles[NormalizeCoupPreset(preset)]
	if !ok {
		return 0, false
	}
	return len(roles), true
}

// CoupPresetLabel returns a short user-facing label for a Coup preset.
func CoupPresetLabel(preset CoupPreset) string {
	label, ok := coupPresetLabels[NormalizeCoupPreset(preset)]
	if !ok {
		return "Unknown preset"
	}
	return label
}

// CoupPresetOptions returns supported Coup presets in setup display order.
func CoupPresetOptions() []CoupPreset {
	options := make([]CoupPreset, len(coupPresetOrder))
	copy(options, coupPresetOrder)
	return options
}

// CoupPresetSummary returns a readable role-count summary for a Coup preset.
func CoupPresetSummary(preset CoupPreset) string {
	roles, ok := coupPresetRoles[NormalizeCoupPreset(preset)]
	if !ok {
		return "Unknown preset"
	}

	counts := make(map[RoleType]int, len(coupRoleSummaryOrder))
	for _, role := range roles {
		counts[role]++
	}

	parts := make([]string, 0, len(coupRoleSummaryOrder))
	for _, role := range coupRoleSummaryOrder {
		count := counts[role]
		switch count {
		case 0:
			continue
		case 1:
			parts = append(parts, string(role))
		default:
			parts = append(parts, fmt.Sprintf("%d %s", count, pluralCoupRole(role)))
		}
	}

	return strings.Join(parts, ", ")
}

func pluralCoupRole(role RoleType) string {
	switch role {
	case RoleBlueKnight:
		return "Blue Knights"
	case RoleBlackKnight:
		return "Black Knights"
	default:
		return string(role) + "s"
	}
}

// AssignCoupRoles assigns a Coup role set for the selected preset.
func AssignCoupRoles(players []*Player, preset CoupPreset) error {
	activePlayers := make([]*Player, 0, len(players))
	for _, player := range players {
		if !player.IsHost {
			activePlayers = append(activePlayers, player)
		}
	}

	preset = NormalizeCoupPreset(preset)

	roleTypes, ok := coupPresetRoles[preset]
	if !ok {
		return fmt.Errorf("unknown Coup preset %q", preset)
	}

	if len(activePlayers) != len(roleTypes) {
		return fmt.Errorf("Coup preset %s requires exactly %d active players, got %d", preset, len(roleTypes), len(activePlayers))
	}

	shuffled := make([]*Player, len(activePlayers))
	copy(shuffled, activePlayers)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i, player := range shuffled {
		role := *coupRoleCards[roleTypes[i]]
		player.Role = &role
		if role.GetRoleType() == RoleKing {
			player.RoleRevealed = true
			player.FaceUp = true
		} else {
			player.RoleRevealed = false
			player.FaceUp = false
		}
	}

	return nil
}
