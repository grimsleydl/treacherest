package game

import (
	"fmt"
	"sort"
	"strings"
)

// CoupWinOutcome identifies the faction or role that appears to have won.
type CoupWinOutcome string

const (
	CoupWinOutcomeKingSide  CoupWinOutcome = "king-side"
	CoupWinOutcomeBlack     CoupWinOutcome = "black"
	CoupWinOutcomeRed       CoupWinOutcome = "red"
	CoupWinOutcomeWasteland CoupWinOutcome = "wasteland"
)

// CoupWinPrompt is an advisory, table-confirmed win candidate.
type CoupWinPrompt struct {
	ID          string
	Outcome     CoupWinOutcome
	Title       string
	Summary     string
	Facts       []string
	GreenShares bool
}

// CoupWinState stores table decisions about advisory win prompts.
type CoupWinState struct {
	Confirmed          *CoupWinPrompt
	RejectedAdvisoryID string
}

type coupWinSnapshot struct {
	kingPresent bool
	kingAlive   bool

	blueTotal      int
	blueAlive      int
	blueEliminated int

	blackTotal int
	blackAlive int

	redTotal int
	redAlive int

	greenAlive int

	wastelandAlive int
	totalLiving    int
}

// EnsureCoupWinState returns a room's Coup win state, creating it if needed.
func EnsureCoupWinState(room *Room) *CoupWinState {
	if room.CoupWin == nil {
		room.CoupWin = &CoupWinState{}
	}
	return room.CoupWin
}

// CurrentCoupAdvisoryWin returns the current non-rejected advisory prompt.
func CurrentCoupAdvisoryWin(room *Room) *CoupWinPrompt {
	prompt := DetectCoupAdvisoryWin(room)
	if prompt == nil {
		return nil
	}
	if room.CoupWin != nil {
		if room.CoupWin.Confirmed != nil {
			return nil
		}
		if room.CoupWin.RejectedAdvisoryID == prompt.ID {
			return nil
		}
	}
	return prompt
}

// DetectCoupAdvisoryWin returns the tracked-state win candidate, if any.
func DetectCoupAdvisoryWin(room *Room) *CoupWinPrompt {
	if room == nil || room.RulesMode != RulesModeCoup || room.State != StatePlaying {
		return nil
	}

	snapshot := newCoupWinSnapshot(room)
	var prompt *CoupWinPrompt

	switch {
	case snapshot.wastelandAlive == 1 && snapshot.totalLiving == 1:
		prompt = &CoupWinPrompt{
			Outcome: CoupWinOutcomeWasteland,
			Title:   "Looks like Wasteland might have just won???",
			Summary: "Wasteland wins alone when every other player is eliminated.",
			Facts: []string{
				"Wasteland Knight is the sole surviving player.",
			},
		}
	case snapshot.kingAlive && snapshot.blackAlive == 0 && snapshot.redAlive == 0 && snapshot.wastelandAlive == 0:
		greenShares := coupGreenSharesKing(room, snapshot)
		prompt = &CoupWinPrompt{
			Outcome:     CoupWinOutcomeKingSide,
			Title:       "Looks like King-side might have just won???",
			Summary:     "King-side wins when the King is alive and anti-King threats are eliminated.",
			GreenShares: greenShares,
			Facts: []string{
				"King is alive.",
				"All Black Knights are eliminated.",
				"Red Knight is eliminated.",
				"No Wasteland Knight remains alive.",
			},
		}
		prompt.Facts = append(prompt.Facts, coupGreenKingFact(room, snapshot, greenShares))
	case snapshot.kingPresent && !snapshot.kingAlive && snapshot.blackAlive > 0 && snapshot.redAlive == 0:
		prompt = &CoupWinPrompt{
			Outcome: CoupWinOutcomeBlack,
			Title:   "Looks like Black might have just won???",
			Summary: "Black wins when the King is dead, at least one Black Knight survives, and Red is dead.",
			Facts: []string{
				"King has fallen.",
				"At least one Black Knight is alive.",
				"Red Knight is eliminated.",
				"Green Knight does not share Black victories.",
			},
		}
	case snapshot.kingPresent && !snapshot.kingAlive && snapshot.redAlive > 0 && snapshot.blackAlive == 0:
		greenShares := snapshot.greenAlive > 0 && room.CoupGreenEligibleBeforeKingFall
		prompt = &CoupWinPrompt{
			Outcome:     CoupWinOutcomeRed,
			Title:       "Looks like Red might have just won???",
			Summary:     "Red wins when the King is dead, Red survives, and all Black Knights are dead.",
			GreenShares: greenShares,
			Facts: []string{
				"King has fallen.",
				"Red Knight is alive.",
				"All Black Knights are eliminated.",
			},
		}
		prompt.Facts = append(prompt.Facts, coupGreenRedFact(room, snapshot, greenShares, room.CoupGreenEligibleBeforeKingFall))
	}

	if prompt == nil {
		return nil
	}
	prompt.ID = coupWinPromptID(room, prompt)
	return prompt
}

// ConfirmCoupWin records a table-confirmed Coup win prompt.
func ConfirmCoupWin(room *Room, prompt *CoupWinPrompt) {
	if room == nil || prompt == nil {
		return
	}
	state := EnsureCoupWinState(room)
	state.Confirmed = prompt
	state.RejectedAdvisoryID = ""
}

// RejectCoupWinPrompt records that the table rejected the current prompt.
func RejectCoupWinPrompt(room *Room, prompt *CoupWinPrompt) {
	if room == nil || prompt == nil {
		return
	}
	state := EnsureCoupWinState(room)
	state.RejectedAdvisoryID = prompt.ID
}

// RecordCoupKingFall locks Green Hunt satisfaction at the moment before King is eliminated.
func RecordCoupKingFall(room *Room) {
	if room == nil || room.CoupKingFallen {
		return
	}
	room.CoupKingFallen = true
	room.CoupGreenEligibleBeforeKingFall = coupGreenRedShareSatisfiedBeforeKingFall(room)
}

// CoupGreenHuntSatisfiedBeforeKingFall implements Green's Red-sharing Hunt lock.
func CoupGreenHuntSatisfiedBeforeKingFall(room *Room) bool {
	snapshot := newCoupWinSnapshot(room)
	return coupGreenHuntSatisfied(room, snapshot)
}

// CoupStrictGreenEligibleBeforeKingFall is kept for older callers; use
// CoupGreenHuntSatisfiedBeforeKingFall for new Coup code.
func CoupStrictGreenEligibleBeforeKingFall(room *Room) bool {
	return CoupGreenHuntSatisfiedBeforeKingFall(room)
}

func newCoupWinSnapshot(room *Room) coupWinSnapshot {
	var snapshot coupWinSnapshot
	if room == nil {
		return snapshot
	}

	for _, player := range room.GetActivePlayers() {
		if player.Role == nil {
			continue
		}
		roleType := player.Role.GetRoleType()
		if !player.IsEliminated {
			snapshot.totalLiving++
		}

		switch roleType {
		case RoleKing:
			snapshot.kingPresent = true
			if !player.IsEliminated {
				snapshot.kingAlive = true
			}
		case RoleBlueKnight:
			snapshot.blueTotal++
			if !player.IsEliminated {
				snapshot.blueAlive++
			} else {
				snapshot.blueEliminated++
			}
		case RoleBlackKnight:
			snapshot.blackTotal++
			if !player.IsEliminated {
				snapshot.blackAlive++
			}
		case RoleRedKnight:
			snapshot.redTotal++
			if !player.IsEliminated {
				snapshot.redAlive++
			}
		case RoleGreenKnight:
			if !player.IsEliminated {
				snapshot.greenAlive++
			}
		case RoleWasteland:
			if !player.IsEliminated {
				snapshot.wastelandAlive++
			}
		}
	}

	return snapshot
}

func coupGreenSharesKing(room *Room, snapshot coupWinSnapshot) bool {
	if snapshot.greenAlive == 0 {
		return false
	}
	return coupGreenHuntSatisfied(room, snapshot) || coupInquisitionSucceeded(room)
}

func coupGreenHuntSatisfied(room *Room, snapshot coupWinSnapshot) bool {
	if snapshot.blueTotal == 0 {
		return false
	}
	switch NormalizeCoupGreenHuntRequirement(room.CoupGreenHuntRequirement) {
	case CoupGreenHuntAllBlues:
		return snapshot.blueEliminated == snapshot.blueTotal
	default:
		return snapshot.blueEliminated > 0
	}
}

func coupGreenRedShareSatisfiedBeforeKingFall(room *Room) bool {
	if CoupGreenHuntSatisfiedBeforeKingFall(room) {
		return true
	}
	return NormalizeCoupInquisitionAmnesty(room.CoupInquisitionAmnesty) == CoupInquisitionAmnestyBroad && coupInquisitionSucceeded(room)
}

func coupGreenKingFact(room *Room, snapshot coupWinSnapshot, greenShares bool) string {
	if snapshot.greenAlive == 0 {
		return "No living Green Knight is available to share the King-side victory."
	}
	if greenShares {
		if coupInquisitionSucceeded(room) {
			return "Green Knight shares because Inquisition has succeeded."
		}
		return "Green Knight shares because Green Hunt is satisfied."
	}
	return "Green Knight is alive but does not share because Green Hunt is not satisfied and Inquisition has not succeeded."
}

func coupGreenRedFact(room *Room, snapshot coupWinSnapshot, greenShares bool, greenHuntBeforeKingFall bool) string {
	if snapshot.greenAlive == 0 {
		return "No living Green Knight is available to share the Red victory."
	}
	if greenShares {
		if NormalizeCoupInquisitionAmnesty(room.CoupInquisitionAmnesty) == CoupInquisitionAmnestyBroad && coupInquisitionSucceeded(room) && !CoupGreenHuntSatisfiedBeforeKingFall(room) {
			return "Green Knight shares because Broad Amnesty is on and Inquisition succeeded before the King fell."
		}
		return "Green Knight shares because Green Hunt was satisfied before the King fell."
	}
	if !greenHuntBeforeKingFall {
		if coupInquisitionSucceeded(room) {
			if NormalizeCoupInquisitionAmnesty(room.CoupInquisitionAmnesty) == CoupInquisitionAmnestyBroad {
				return "Green Knight does not share because Inquisition did not succeed before the King fell."
			}
			return "Green Knight does not share because Inquisition only helps Green share King victories under the selected settings."
		}
		return "Green Knight does not share because Green Hunt was not satisfied before the King fell."
	}
	return "Green Knight does not share this Red victory."
}

func coupInquisitionSucceeded(room *Room) bool {
	return room != nil && room.CoupInquisition != nil && room.CoupInquisition.Succeeded
}

func coupWinPromptID(room *Room, prompt *CoupWinPrompt) string {
	parts := []string{
		string(prompt.Outcome),
		fmt.Sprintf("green:%t", prompt.GreenShares),
		fmt.Sprintf("king-fallen:%t", room.CoupKingFallen),
		fmt.Sprintf("green-before-king:%t", room.CoupGreenEligibleBeforeKingFall),
		fmt.Sprintf("inq:%t", coupInquisitionSucceeded(room)),
	}

	players := room.GetActivePlayers()
	sort.Slice(players, func(i, j int) bool {
		return players[i].ID < players[j].ID
	})
	for _, player := range players {
		roleType := ""
		if player.Role != nil {
			roleType = string(player.Role.GetRoleType())
		}
		parts = append(parts, fmt.Sprintf("%s:%s:%t", player.ID, roleType, player.IsEliminated))
	}

	return strings.Join(parts, "|")
}
