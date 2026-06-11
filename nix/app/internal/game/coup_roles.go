package game

import (
	"fmt"
	"math/rand"
	"sort"
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

// CoupKingToBluePolicy controls what the King learns about Blue Knights.
type CoupKingToBluePolicy string

const (
	CoupKingKnowsAllBlue       CoupKingToBluePolicy = "king-knows-all-blue"
	CoupKingGetsBlueCandidates CoupKingToBluePolicy = "king-gets-blue-candidates"
	CoupKingKnowsNoBlue        CoupKingToBluePolicy = "king-knows-no-blue"
)

// CoupRedToBlackPolicy controls what Red learns about Black Knights.
type CoupRedToBlackPolicy string

const (
	CoupRedKnowsAllBlack CoupRedToBlackPolicy = "red-knows-all-black"
	CoupRedKnowsOneBlack CoupRedToBlackPolicy = "red-knows-one-black"
	CoupRedKnowsNoBlack  CoupRedToBlackPolicy = "red-knows-no-black"
)

// CoupBlackToRedPolicy controls whether Black Knights learn Red.
type CoupBlackToRedPolicy string

const (
	CoupBlackToRedNone CoupBlackToRedPolicy = "black-knows-no-red"
	CoupBlackToRedOne  CoupBlackToRedPolicy = "one-black-knows-red"
	CoupBlackToRedAll  CoupBlackToRedPolicy = "all-black-know-red"
)

// CoupBlackNetworkPolicy controls whether Black Knights know each other.
type CoupBlackNetworkPolicy string

const (
	CoupBlackNetworkNone CoupBlackNetworkPolicy = "black-network-none"
	CoupBlackNetworkAll  CoupBlackNetworkPolicy = "black-network-all"
)

// CoupInformationPolicy configures private information dealt with Coup roles.
type CoupInformationPolicy struct {
	KingToBlue   CoupKingToBluePolicy
	RedToBlack   CoupRedToBlackPolicy
	BlackToRed   CoupBlackToRedPolicy
	BlackNetwork CoupBlackNetworkPolicy
}

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
		Text: text,
		Rulings: []string{
			"Role Goal: " + winCondition,
			"Win Condition: " + winCondition,
		},
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

// CoupRoyalGuardRuleText returns the player-facing Royal Guard rule for a blocker limit.
// A limit of 0 means any number of eligible blockers.
func CoupRoyalGuardRuleText(blockerLimit int) string {
	blockerLimit = NormalizeCoupRoyalGuardBlockerLimit(blockerLimit)
	blockerText := "any number of untapped creatures"
	switch {
	case blockerLimit == 1:
		blockerText = "one untapped creature"
	case blockerLimit > 1:
		blockerText = fmt.Sprintf("up to %d untapped creatures", blockerLimit)
	}

	return fmt.Sprintf("Once each combat, a revealed Blue Knight may have %s they control block creatures attacking the King player as though those creatures were attacking the Blue Knight. Normal blocking restrictions apply.", blockerText)
}

// NormalizeCoupRoyalGuardBlockerLimit treats negative limits as the default unlimited setting.
func NormalizeCoupRoyalGuardBlockerLimit(blockerLimit int) int {
	if blockerLimit < 0 {
		return 0
	}
	return blockerLimit
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

// NormalizeCoupInformationPolicy fills omitted policy fields with Coup defaults.
func NormalizeCoupInformationPolicy(policy CoupInformationPolicy) CoupInformationPolicy {
	if policy.KingToBlue == "" {
		policy.KingToBlue = CoupKingKnowsAllBlue
	}
	if policy.RedToBlack == "" {
		policy.RedToBlack = CoupRedKnowsAllBlack
	}
	if policy.BlackToRed == "" {
		policy.BlackToRed = CoupBlackToRedNone
	}
	if policy.BlackNetwork == "" {
		policy.BlackNetwork = CoupBlackNetworkNone
	}
	return policy
}

// IsValidCoupInformationPolicy reports whether all policy values are supported.
func IsValidCoupInformationPolicy(policy CoupInformationPolicy) bool {
	policy = NormalizeCoupInformationPolicy(policy)
	switch policy.KingToBlue {
	case CoupKingKnowsAllBlue, CoupKingGetsBlueCandidates, CoupKingKnowsNoBlue:
	default:
		return false
	}
	switch policy.RedToBlack {
	case CoupRedKnowsAllBlack, CoupRedKnowsOneBlack, CoupRedKnowsNoBlack:
	default:
		return false
	}
	switch policy.BlackToRed {
	case CoupBlackToRedNone, CoupBlackToRedOne, CoupBlackToRedAll:
	default:
		return false
	}
	switch policy.BlackNetwork {
	case CoupBlackNetworkNone, CoupBlackNetworkAll:
	default:
		return false
	}
	return true
}

// AssignCoupRoles assigns a Coup role set for the selected preset.
func AssignCoupRoles(players []*Player, preset CoupPreset) error {
	return AssignCoupRolesWithInformation(players, preset, CoupInformationPolicy{})
}

// AssignCoupRolesBestEffort assigns Coup roles to the current active players without
// requiring the selected preset's exact player count. It always includes King when
// at least one active player exists, then fills remaining players from the preset pool.
func AssignCoupRolesBestEffort(players []*Player, preset CoupPreset, policy CoupInformationPolicy) error {
	activePlayers := make([]*Player, 0, len(players))
	for _, player := range players {
		if !player.IsHost {
			activePlayers = append(activePlayers, player)
		}
	}
	if len(activePlayers) == 0 {
		return fmt.Errorf("Start As-Is requires at least one active player")
	}

	preset = NormalizeCoupPreset(preset)
	roleTypes, ok := coupPresetRoles[preset]
	if !ok {
		return fmt.Errorf("unknown Coup preset %q", preset)
	}
	if len(activePlayers) > len(roleTypes) {
		return fmt.Errorf("Coup preset %s has %d roles, got %d active players", preset, len(roleTypes), len(activePlayers))
	}

	shuffledPlayers := make([]*Player, len(activePlayers))
	copy(shuffledPlayers, activePlayers)
	rand.Shuffle(len(shuffledPlayers), func(i, j int) {
		shuffledPlayers[i], shuffledPlayers[j] = shuffledPlayers[j], shuffledPlayers[i]
	})

	remainingRoles := make([]RoleType, 0, len(roleTypes)-1)
	for _, roleType := range roleTypes {
		if roleType != RoleKing {
			remainingRoles = append(remainingRoles, roleType)
		}
	}
	rand.Shuffle(len(remainingRoles), func(i, j int) {
		remainingRoles[i], remainingRoles[j] = remainingRoles[j], remainingRoles[i]
	})

	selectedRoles := []RoleType{RoleKing}
	selectedRoles = append(selectedRoles, remainingRoles[:len(activePlayers)-1]...)
	for i, player := range shuffledPlayers {
		role := *coupRoleCards[selectedRoles[i]]
		player.Role = &role
		if role.GetRoleType() == RoleKing {
			player.RoleRevealed = true
			player.FaceUp = true
		} else {
			player.RoleRevealed = false
			player.FaceUp = false
		}
	}

	applyCoupInformation(activePlayers, NormalizeCoupInformationPolicy(policy))

	return nil
}

// AssignCoupRolesWithInformation assigns Coup roles and attaches recipient-scoped private information.
func AssignCoupRolesWithInformation(players []*Player, preset CoupPreset, policy CoupInformationPolicy) error {
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

	applyCoupInformation(activePlayers, NormalizeCoupInformationPolicy(policy))

	return nil
}

func applyCoupInformation(players []*Player, policy CoupInformationPolicy) {
	playersByRole := coupPlayersByRole(players)
	kings := playersByRole[RoleKing]
	blues := playersByRole[RoleBlueKnight]
	reds := playersByRole[RoleRedKnight]
	blacks := playersByRole[RoleBlackKnight]

	if len(kings) > 0 && len(blues) > 0 {
		switch policy.KingToBlue {
		case CoupKingKnowsAllBlue:
			appendCoupPrivateInfo(kings[0].Role, "Blue Knights: "+joinCoupPlayerNames(blues))
		case CoupKingGetsBlueCandidates:
			candidates := []*Player{blues[0]}
			if decoy := firstCoupPlayerWithoutRoles(players, RoleKing, RoleBlueKnight); decoy != nil {
				candidates = append(candidates, decoy)
			}
			appendCoupPrivateInfo(kings[0].Role, "Blue Knight candidates: "+joinCoupPlayerNames(candidates))
		}
	}
	if len(reds) > 0 && len(blacks) > 0 {
		switch policy.RedToBlack {
		case CoupRedKnowsAllBlack:
			appendCoupPrivateInfo(reds[0].Role, "Black Knights: "+joinCoupPlayerNames(blacks))
		case CoupRedKnowsOneBlack:
			appendCoupPrivateInfo(reds[0].Role, "Black Knights: "+joinCoupPlayerNames(blacks[:1]))
		}
	}
	if len(reds) > 0 && len(blacks) > 0 {
		switch policy.BlackToRed {
		case CoupBlackToRedAll:
			for _, black := range blacks {
				appendCoupPrivateInfo(black.Role, "Red Knight: "+reds[0].Name)
			}
		case CoupBlackToRedOne:
			appendCoupPrivateInfo(blacks[0].Role, "Red Knight: "+reds[0].Name)
		}
	}
	if len(blacks) > 0 && policy.BlackNetwork == CoupBlackNetworkAll {
		for _, black := range blacks {
			appendCoupPrivateInfo(black.Role, "Black Knights: "+joinCoupPlayerNames(blacks))
		}
	}
}

func coupPlayersByRole(players []*Player) map[RoleType][]*Player {
	playersByRole := make(map[RoleType][]*Player)
	for _, player := range players {
		if player.Role == nil {
			continue
		}
		roleType := player.Role.GetRoleType()
		playersByRole[roleType] = append(playersByRole[roleType], player)
	}
	for roleType := range playersByRole {
		sort.Slice(playersByRole[roleType], func(i, j int) bool {
			return playersByRole[roleType][i].Name < playersByRole[roleType][j].Name
		})
	}
	return playersByRole
}

func appendCoupPrivateInfo(card *Card, info string) {
	if card == nil || info == "" {
		return
	}
	card.Rulings = append(card.Rulings, "Private information: "+info)
}

func firstCoupPlayerWithoutRoles(players []*Player, excludedRoles ...RoleType) *Player {
	excluded := make(map[RoleType]bool, len(excludedRoles))
	for _, role := range excludedRoles {
		excluded[role] = true
	}

	candidates := make([]*Player, 0)
	for _, player := range players {
		if player.Role == nil || excluded[player.Role.GetRoleType()] {
			continue
		}
		candidates = append(candidates, player)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Name < candidates[j].Name
	})
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

func joinCoupPlayerNames(players []*Player) string {
	names := make([]string, 0, len(players))
	for _, player := range players {
		names = append(names, player.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
