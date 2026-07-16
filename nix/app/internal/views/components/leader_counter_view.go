package components

import (
	"fmt"
	"strings"

	"treacherest/internal/game"
)

type leaderCounterPublicView struct {
	CardID             int
	Name               string
	CountText          string
	TurnStatus         string
	ActivationEvents   []string
	ActivationLabel    string
	ActivationDisabled bool
	ShowTurnReset      bool
}

func buildLeaderCounterPublicView(card *game.Card) (leaderCounterPublicView, bool) {
	if card == nil || card.LeaderCounterPool == nil {
		return leaderCounterPublicView{}, false
	}
	pool := card.LeaderCounterPool
	name := strings.ToLower(pool.Name)
	view := leaderCounterPublicView{
		CardID:    card.ID,
		Name:      name,
		CountText: fmt.Sprintf("%s counters: %d of %d", titleCounterName(name), pool.Remaining, pool.Initial),
	}

	if pool.OncePerTurn {
		if pool.UsedThisTurn {
			view.TurnStatus = "Once per turn · Used this turn"
			view.ShowTurnReset = true
		} else {
			view.TurnStatus = "Once per turn · Available this turn"
		}
	} else {
		view.TurnStatus = "No per-turn limit"
	}

	view.ActivationDisabled = pool.Remaining == 0 || (pool.OncePerTurn && pool.UsedThisTurn)
	switch {
	case pool.Remaining == 0:
		view.ActivationLabel = fmt.Sprintf("No %s counters remaining", name)
	case pool.OncePerTurn && pool.UsedThisTurn:
		view.ActivationLabel = fmt.Sprintf("%s activation used this turn", titleCounterName(name))
	default:
		view.ActivationLabel = fmt.Sprintf("Record %s activation", titleCounterName(name))
	}

	view.ActivationEvents = make([]string, 0, len(pool.Activations))
	for _, event := range pool.Activations {
		view.ActivationEvents = append(view.ActivationEvents, fmt.Sprintf(
			"Activation #%d: %s removed a %s counter (%d → %d).",
			event.Number,
			event.PlayerName,
			name,
			event.Before,
			event.After,
		))
	}
	return view, true
}

func titleCounterName(name string) string {
	if name == "" {
		return "Counter"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func showLeaderCounterOwnerControls(room *game.Room, player *game.Player) bool {
	return room != nil &&
		room.State == game.StatePlaying &&
		player != nil &&
		!player.IsEliminated &&
		player.FaceUp &&
		player.Role != nil &&
		player.Role.LeaderCounterPool != nil
}
