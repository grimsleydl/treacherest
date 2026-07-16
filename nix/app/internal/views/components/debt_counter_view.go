package components

import (
	"fmt"
	"strings"

	"treacherest/internal/game"
)

type debtCounterPublicView struct {
	Count            int
	CountText        string
	CurrentAdvisory  string
	PlacementEvents  []string
	UpkeepAdvisories []string
	LossAdvisories   []string
}

type debtCounterTargetView struct {
	PlayerID   string
	PlayerName string
	Count      int
}

type debtCollectorOwnerView struct {
	Targets           []debtCounterTargetView
	TotalDebtCounters int
}

func buildDebtCounterPublicView(card *game.Card) (debtCounterPublicView, bool) {
	if card == nil || card.DebtCounters == nil {
		return debtCounterPublicView{}, false
	}
	state := card.DebtCounters
	if state.Count == 0 && len(state.Placements) == 0 && len(state.UpkeepAdvisories) == 0 && len(state.LossAdvisories) == 0 {
		return debtCounterPublicView{}, false
	}

	view := debtCounterPublicView{
		Count:     state.Count,
		CountText: fmt.Sprintf("Debt counters: %d", state.Count),
	}
	if state.Count > 0 {
		view.CurrentAdvisory = fmt.Sprintf(
			"Upkeep advisory: this card's controller should lose %d life during The Debt Collector's upkeep.",
			state.Count,
		)
	}

	view.PlacementEvents = make([]string, 0, len(state.Placements))
	for _, event := range state.Placements {
		view.PlacementEvents = append(view.PlacementEvents, fmt.Sprintf(
			"Placement #%d: %s placed a debt counter on %s's identity card (%d → %d). %s should draw a card at the table.",
			event.Number,
			event.CollectorPlayerName,
			event.TargetPlayerName,
			event.Before,
			event.After,
			event.TargetPlayerName,
		))
	}

	view.UpkeepAdvisories = make([]string, 0, len(state.UpkeepAdvisories))
	for _, event := range state.UpkeepAdvisories {
		assessments := make([]string, 0, len(event.Assessments))
		for _, assessment := range event.Assessments {
			assessments = append(assessments, fmt.Sprintf("%s loses %d life", assessment.PlayerName, assessment.Count))
		}
		view.UpkeepAdvisories = append(view.UpkeepAdvisories, fmt.Sprintf(
			"Upkeep advisory #%d: %s. %s should gain %d life total.",
			event.Number,
			strings.Join(assessments, "; "),
			event.CollectorPlayerName,
			event.TotalLifeGain,
		))
	}

	view.LossAdvisories = make([]string, 0, len(state.LossAdvisories))
	for _, event := range state.LossAdvisories {
		view.LossAdvisories = append(view.LossAdvisories, fmt.Sprintf(
			"Loss advisory #%d: %s lost with %d debt counters. %s should draw %d cards at the table.",
			event.Number,
			event.LostPlayerName,
			event.DrawCount,
			event.CollectorPlayerName,
			event.DrawCount,
		))
	}
	return view, true
}

func debtCounterCount(card *game.Card) (int, bool) {
	if card == nil || card.DebtCounters == nil || card.DebtCounters.Count == 0 {
		return 0, false
	}
	return card.DebtCounters.Count, true
}

func buildDebtCollectorOwnerView(room *game.Room, player *game.Player) (debtCollectorOwnerView, bool) {
	if room == nil || room.State != game.StatePlaying || player == nil || player.IsEliminated || !player.FaceUp ||
		player.Role == nil || player.Role.ID != game.TheDebtCollectorCardID {
		return debtCollectorOwnerView{}, false
	}

	view := debtCollectorOwnerView{}
	for _, target := range room.GetActivePlayers() {
		if target == nil || target.ID == player.ID || target.IsEliminated || target.Role == nil {
			continue
		}
		count := 0
		if target.Role.DebtCounters != nil {
			count = target.Role.DebtCounters.Count
		}
		view.Targets = append(view.Targets, debtCounterTargetView{
			PlayerID:   target.ID,
			PlayerName: target.Name,
			Count:      count,
		})
		view.TotalDebtCounters += count
	}
	return view, true
}
