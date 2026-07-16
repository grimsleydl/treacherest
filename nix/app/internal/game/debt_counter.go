package game

import (
	"errors"
	"sort"
)

const TheDebtCollectorCardID = 53

var (
	ErrDebtCollectorUnavailable = errors.New("player is not The Debt Collector")
	ErrDebtCounterTarget        = errors.New("invalid debt counter target")
	ErrNoDebtCounters           = errors.New("no debt counters to assess")
)

// DebtCounterPlacement is the public record of a counter being placed on one
// identity card. It deliberately records player names but no target card
// contents, because the counter is visible even while that card is face down.
type DebtCounterPlacement struct {
	Number              int    `json:"number"`
	CollectorPlayerID   string `json:"collector_player_id"`
	CollectorPlayerName string `json:"collector_player_name"`
	TargetPlayerID      string `json:"target_player_id"`
	TargetPlayerName    string `json:"target_player_name"`
	Before              int    `json:"before"`
	After               int    `json:"after"`
}

// DebtCounterAssessment is one public life-loss advisory captured at a Debt
// Collector upkeep. Life totals remain table state and are never mutated here.
type DebtCounterAssessment struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Count      int    `json:"count"`
}

// DebtCounterUpkeepAdvisory records the public counter snapshot for one
// table-adjudicated Debt Collector upkeep.
type DebtCounterUpkeepAdvisory struct {
	Number              int                     `json:"number"`
	CollectorPlayerID   string                  `json:"collector_player_id"`
	CollectorPlayerName string                  `json:"collector_player_name"`
	Assessments         []DebtCounterAssessment `json:"assessments"`
	TotalLifeGain       int                     `json:"total_life_gain"`
}

// DebtCounterLossAdvisory records the Debt Collector's public draw reminder
// when a player with debt counters loses the game.
type DebtCounterLossAdvisory struct {
	Number              int    `json:"number"`
	CollectorPlayerID   string `json:"collector_player_id"`
	CollectorPlayerName string `json:"collector_player_name"`
	LostPlayerID        string `json:"lost_player_id"`
	LostPlayerName      string `json:"lost_player_name"`
	DrawCount           int    `json:"draw_count"`
}

// DebtCounterState is mutable public state attached to one dealt identity-card
// instance. Count and Placements describe counters physically on this card;
// the advisory logs are used by The Debt Collector card itself.
type DebtCounterState struct {
	Count            int                         `json:"count"`
	Placements       []DebtCounterPlacement      `json:"placements,omitempty"`
	UpkeepAdvisories []DebtCounterUpkeepAdvisory `json:"upkeep_advisories,omitempty"`
	LossAdvisories   []DebtCounterLossAdvisory   `json:"loss_advisories,omitempty"`
}

// PlaceDebtCounter records The Debt Collector's public end-step table action.
// The target draw is advisory-only; the app mutates only the visible counter.
func (r *Room) PlaceDebtCounter(collectorID, targetID string) (*DebtCounterPlacement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	collector := r.Players[collectorID]
	if !isDebtCollectorPlayer(collector) || collector.IsEliminated {
		return nil, ErrDebtCollectorUnavailable
	}
	target := r.Players[targetID]
	if target == nil || target.IsHost || target.IsEliminated || target.Role == nil || target.ID == collector.ID {
		return nil, ErrDebtCounterTarget
	}

	state := ensureDebtCounterState(target.Role)
	placement := DebtCounterPlacement{
		Number:              len(state.Placements) + 1,
		CollectorPlayerID:   collector.ID,
		CollectorPlayerName: collector.Name,
		TargetPlayerID:      target.ID,
		TargetPlayerName:    target.Name,
		Before:              state.Count,
		After:               state.Count + 1,
	}
	state.Count = placement.After
	state.Placements = append(state.Placements, placement)
	return &state.Placements[len(state.Placements)-1], nil
}

// RecordDebtCounterUpkeep snapshots all visible debt counters and records the
// resulting life-loss/life-gain reminder without changing table life totals.
func (r *Room) RecordDebtCounterUpkeep(collectorID string) (*DebtCounterUpkeepAdvisory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	collector := r.Players[collectorID]
	if !isDebtCollectorPlayer(collector) || collector.IsEliminated {
		return nil, ErrDebtCollectorUnavailable
	}

	assessments := make([]DebtCounterAssessment, 0)
	total := 0
	for _, player := range r.Players {
		if player == nil || player.IsHost || player.IsEliminated || player.Role == nil || player.Role.DebtCounters == nil {
			continue
		}
		count := player.Role.DebtCounters.Count
		if count == 0 {
			continue
		}
		assessments = append(assessments, DebtCounterAssessment{
			PlayerID:   player.ID,
			PlayerName: player.Name,
			Count:      count,
		})
		total += count
	}
	if total == 0 {
		return nil, ErrNoDebtCounters
	}
	sort.Slice(assessments, func(i, j int) bool {
		if assessments[i].PlayerName == assessments[j].PlayerName {
			return assessments[i].PlayerID < assessments[j].PlayerID
		}
		return assessments[i].PlayerName < assessments[j].PlayerName
	})

	state := ensureDebtCounterState(collector.Role)
	advisory := DebtCounterUpkeepAdvisory{
		Number:              len(state.UpkeepAdvisories) + 1,
		CollectorPlayerID:   collector.ID,
		CollectorPlayerName: collector.Name,
		Assessments:         assessments,
		TotalLifeGain:       total,
	}
	state.UpkeepAdvisories = append(state.UpkeepAdvisories, advisory)
	return &state.UpkeepAdvisories[len(state.UpkeepAdvisories)-1], nil
}

// RecordDebtCounterLossAdvisories records public draw reminders on every dealt
// Debt Collector card when a player with debt counters loses the game.
func (r *Room) RecordDebtCounterLossAdvisories(lostPlayerID string) []DebtCounterLossAdvisory {
	r.mu.Lock()
	defer r.mu.Unlock()

	lostPlayer := r.Players[lostPlayerID]
	if lostPlayer == nil || lostPlayer.Role == nil || lostPlayer.Role.DebtCounters == nil {
		return nil
	}
	drawCount := lostPlayer.Role.DebtCounters.Count
	if drawCount == 0 {
		return nil
	}

	advisories := make([]DebtCounterLossAdvisory, 0, 1)
	for _, collector := range r.Players {
		if !isDebtCollectorPlayer(collector) {
			continue
		}
		state := ensureDebtCounterState(collector.Role)
		advisory := DebtCounterLossAdvisory{
			Number:              len(state.LossAdvisories) + 1,
			CollectorPlayerID:   collector.ID,
			CollectorPlayerName: collector.Name,
			LostPlayerID:        lostPlayer.ID,
			LostPlayerName:      lostPlayer.Name,
			DrawCount:           drawCount,
		}
		state.LossAdvisories = append(state.LossAdvisories, advisory)
		advisories = append(advisories, advisory)
	}
	return advisories
}

func isDebtCollectorPlayer(player *Player) bool {
	return player != nil && player.Role != nil && player.Role.ID == TheDebtCollectorCardID
}

func ensureDebtCounterState(card *Card) *DebtCounterState {
	if card.DebtCounters == nil {
		card.DebtCounters = &DebtCounterState{}
	}
	return card.DebtCounters
}
