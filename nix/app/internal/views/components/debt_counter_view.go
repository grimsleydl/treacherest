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
			"Reminder: at The Debt Collector's upkeep, this card's controller loses %d life (1 per debt counter) and The Debt Collector gains that much — resolve at the table.",
			state.Count,
		)
	}

	// Coalesce consecutive placements by the same collector on this card into
	// one line; the full per-event history stays in state for the record.
	view.PlacementEvents = make([]string, 0, len(state.Placements))
	for i := 0; i < len(state.Placements); {
		event := state.Placements[i]
		j := i
		for j < len(state.Placements) && state.Placements[j].CollectorPlayerName == event.CollectorPlayerName {
			j++
		}
		run := state.Placements[i:j]
		last := run[len(run)-1]
		if len(run) == 1 {
			view.PlacementEvents = append(view.PlacementEvents, fmt.Sprintf(
				"%s placed a debt counter on %s's identity card (%d → %d). %s should draw a card at the table.",
				event.CollectorPlayerName, event.TargetPlayerName, event.Before, last.After, event.TargetPlayerName,
			))
		} else {
			view.PlacementEvents = append(view.PlacementEvents, fmt.Sprintf(
				"%s placed %d debt counters on %s's identity card (%d → %d). %s should draw a card for each at the table.",
				event.CollectorPlayerName, len(run), event.TargetPlayerName, event.Before, last.After, event.TargetPlayerName,
			))
		}
		i = j
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
