package components

import (
	"fmt"

	"treacherest/internal/game"
)

type gatheringModeView struct {
	Mode              string
	Color             string
	Effect            string
	ChosenText        string
	ChosenValue       string
	SelectableValue   string
	AriaDisabledValue string
	ItemClass         string
	ControlLabel      string
	ControlDisabled   bool
}

type gatheringChecklistView struct {
	CardID int
	Modes  []gatheringModeView
}

func buildGatheringChecklistView(card *game.Card) (gatheringChecklistView, bool) {
	if card == nil || card.ID != game.TheGatheringCardID || card.GatheringChecklist == nil {
		return gatheringChecklistView{}, false
	}

	choices := make(map[game.GatheringMode]game.GatheringModeChoice, len(card.GatheringChecklist.Choices))
	for _, choice := range card.GatheringChecklist.Choices {
		choices[choice.Mode] = choice
	}

	definitions := game.GatheringModeDefinitions()
	view := gatheringChecklistView{
		CardID: card.ID,
		Modes:  make([]gatheringModeView, 0, len(definitions)),
	}
	for _, definition := range definitions {
		choice, chosen := choices[definition.Mode]
		mode := gatheringModeView{
			Mode:              string(definition.Mode),
			Color:             definition.Color,
			Effect:            definition.Effect,
			ChosenValue:       fmt.Sprintf("%t", chosen),
			SelectableValue:   fmt.Sprintf("%t", !chosen),
			AriaDisabledValue: fmt.Sprintf("%t", chosen),
			ItemClass:         "rounded-box border border-base-300 bg-base-100 p-3",
			ChosenText:        "Available",
			ControlLabel:      "Choose " + definition.Color,
			ControlDisabled:   chosen,
		}
		if chosen {
			mode.ItemClass = "rounded-box border border-success/50 bg-success/10 p-3"
			mode.ChosenText = fmt.Sprintf("Chosen #%d by %s", choice.Number, choice.PlayerName)
			mode.ControlLabel = fmt.Sprintf("%s — already chosen", definition.Color)
		}
		view.Modes = append(view.Modes, mode)
	}
	return view, true
}

func buildGatheringOwnerView(room *game.Room, player *game.Player) (gatheringChecklistView, bool) {
	if room == nil || room.State != game.StatePlaying || player == nil || player.IsEliminated || !player.FaceUp ||
		player.Role == nil || player.Role.ID != game.TheGatheringCardID {
		return gatheringChecklistView{}, false
	}
	return buildGatheringChecklistView(player.Role)
}
