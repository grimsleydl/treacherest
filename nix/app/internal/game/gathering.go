package game

import "errors"

const TheGatheringCardID = 54

type GatheringMode string

const (
	GatheringModeWhite GatheringMode = "white"
	GatheringModeBlue  GatheringMode = "blue"
	GatheringModeBlack GatheringMode = "black"
	GatheringModeRed   GatheringMode = "red"
	GatheringModeGreen GatheringMode = "green"
)

var (
	ErrGatheringUnavailable = errors.New("player is not The Gathering")
	ErrGatheringModeInvalid = errors.New("invalid Gathering mode")
	ErrGatheringModeChosen  = errors.New("Gathering mode has already been chosen")
)

// GatheringModeDefinition is one mode printed on The Gathering. Effects are
// reference text only; the table remains responsible for resolving them.
type GatheringModeDefinition struct {
	Mode   GatheringMode
	Color  string
	Effect string
}

var gatheringModeDefinitions = [...]GatheringModeDefinition{
	{Mode: GatheringModeWhite, Color: "White", Effect: "Create four 1/1 white Soldier creature tokens."},
	{Mode: GatheringModeBlue, Color: "Blue", Effect: "Scry 4, then draw two cards."},
	{Mode: GatheringModeBlack, Color: "Black", Effect: "Return target creature card from your graveyard to the battlefield."},
	{Mode: GatheringModeRed, Color: "Red", Effect: "The Gathering deals 4 damage divided as you choose among any number of targets."},
	{Mode: GatheringModeGreen, Color: "Green", Effect: "Destroy target noncreature permanent."},
}

// GatheringModeDefinitions returns the modes in their printed order.
func GatheringModeDefinitions() []GatheringModeDefinition {
	definitions := make([]GatheringModeDefinition, len(gatheringModeDefinitions))
	copy(definitions, gatheringModeDefinitions[:])
	return definitions
}

// GatheringModeChoice is the public record created when the upkeep trigger is
// put on the stack. The battlefield effect itself remains table-adjudicated.
type GatheringModeChoice struct {
	Number     int           `json:"number"`
	Mode       GatheringMode `json:"mode"`
	PlayerID   string        `json:"player_id"`
	PlayerName string        `json:"player_name"`
}

// GatheringChecklist is mutable public state attached to one The Gathering
// identity-card instance. Its history is also the permanent chosen-mode list.
type GatheringChecklist struct {
	Choices []GatheringModeChoice `json:"choices,omitempty"`
}

func (c *GatheringChecklist) HasChosen(mode GatheringMode) bool {
	if c == nil {
		return false
	}
	for _, choice := range c.Choices {
		if choice.Mode == mode {
			return true
		}
	}
	return false
}

// ChooseGatheringMode permanently marks one mode on this identity card. Only
// the public choice is recorded; resolving the printed effect stays at table.
func (r *Room) ChooseGatheringMode(playerID string, mode GatheringMode) (*GatheringModeChoice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Players[playerID]
	if player == nil || player.Role == nil || player.Role.ID != TheGatheringCardID || player.Role.GatheringChecklist == nil {
		return nil, ErrGatheringUnavailable
	}
	if !isGatheringMode(mode) {
		return nil, ErrGatheringModeInvalid
	}
	checklist := player.Role.GatheringChecklist
	if checklist.HasChosen(mode) {
		return nil, ErrGatheringModeChosen
	}

	choice := GatheringModeChoice{
		Number:     len(checklist.Choices) + 1,
		Mode:       mode,
		PlayerID:   player.ID,
		PlayerName: player.Name,
	}
	checklist.Choices = append(checklist.Choices, choice)
	return &checklist.Choices[len(checklist.Choices)-1], nil
}

func isGatheringMode(mode GatheringMode) bool {
	for _, definition := range gatheringModeDefinitions {
		if definition.Mode == mode {
			return true
		}
	}
	return false
}
