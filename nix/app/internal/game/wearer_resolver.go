package game

import (
	"fmt"
	"math/rand"
	"time"
	"treacherest/internal/game/ability"
)

// WearerOfMasksResolver implements custom logic for The Wearer of Masks (Card ID 31)
// Ability: "As The Wearer of Masks is unveiled, reveal up to X non-Leader identity cards
// from outside the game, where X is the amount paid. Choose one. The Wearer of Masks
// becomes a copy of that card, except it's still a Traitor in addition to its other types.
// When The Wearer of Masks is turned face down, it loses this ability."
type WearerOfMasksResolver struct{}

// OnTrigger is called when The Wearer of Masks is unveiled
func (r *WearerOfMasksResolver) OnTrigger(ctx *ability.AbilityContext) (*ability.PendingAbility, error) {
	if ctx.GameState == nil {
		return nil, fmt.Errorf("game state not available")
	}

	// Get X value from temp data (paid cost)
	xValue := 0
	if x, ok := ctx.TempData["X"]; ok {
		if xInt, ok := x.(int); ok {
			xValue = xInt
		}
	}

	// Check role options for custom configuration
	useAllCards := false
	maxReveal := xValue
	if val, ok := ctx.GameState.GetRoleOption(ctx.CardID, "use_all_cards"); ok {
		if boolVal, ok := val.(bool); ok {
			useAllCards = boolVal
		}
	}
	if val, ok := ctx.GameState.GetRoleOption(ctx.CardID, "max_reveal"); ok {
		if intVal, ok := val.(int); ok && intVal > 0 {
			if intVal < maxReveal {
				maxReveal = intVal
			}
		}
	}

	// Get available non-Leader cards from outside the game
	filters := []ability.Filter{
		{
			Type:   ability.FilterRole,
			Value:  "Leader",
			Negate: true, // non-Leader
		},
	}

	var availableCards []ability.CardLike
	if useAllCards {
		// Use all cards regardless of assignment status
		filters = append(filters, ability.Filter{
			Type:  ability.FilterRole,
			Value: "all",
		})
	}
	availableCards = ctx.GameState.GetAvailableCards(filters)

	// Limit to max reveal count
	if len(availableCards) > maxReveal {
		// Randomly select maxReveal cards
		rand.Shuffle(len(availableCards), func(i, j int) {
			availableCards[i], availableCards[j] = availableCards[j], availableCards[i]
		})
		availableCards = availableCards[:maxReveal]
	}

	// If no cards available, ability fizzles
	if len(availableCards) == 0 {
		return nil, fmt.Errorf("no non-Leader cards available to reveal")
	}

	// Create pending ability with revealed cards
	cardIDs := make([]int, len(availableCards))
	cardData := make([]map[string]interface{}, len(availableCards))
	for i, card := range availableCards {
		cardIDs[i] = card.GetID()
		cardData[i] = map[string]interface{}{
			"id":   card.GetID(),
			"text": card.GetText(),
		}
	}

	return &ability.PendingAbility{
		ID:          fmt.Sprintf("wearer-%s-%d", ctx.PlayerID, time.Now().UnixNano()),
		PlayerID:    ctx.PlayerID,
		CardID:      ctx.CardID,
		AbilityType: "wearer_transform",
		Data: map[string]interface{}{
			"revealed_cards":   cardData,
			"revealed_ids":     cardIDs,
			"requires_choice":  true,
			"original_card_id": ctx.CardID,
		},
		ModalDismissed: false,
		ModalState:     make(map[string]interface{}),
		CreatedAt:      time.Now(),
	}, nil
}

// OnChoice is called when player chooses a card to copy
func (r *WearerOfMasksResolver) OnChoice(ctx *ability.AbilityContext, choice interface{}) error {
	// Extract chosen card ID
	var chosenCardID int
	switch v := choice.(type) {
	case int:
		chosenCardID = v
	case float64:
		chosenCardID = int(v)
	case map[string]interface{}:
		if id, ok := v["card_id"]; ok {
			if idInt, ok := id.(int); ok {
				chosenCardID = idInt
			} else if idFloat, ok := id.(float64); ok {
				chosenCardID = int(idFloat)
			}
		}
	default:
		return fmt.Errorf("invalid choice format: expected card ID")
	}

	if chosenCardID == 0 {
		return fmt.Errorf("no card chosen")
	}

	// Store transformation in temp data for handler to apply
	ctx.TempData["transform_to_card_id"] = chosenCardID
	ctx.TempData["keep_types"] = []string{"Traitor"}
	ctx.TempData["end_condition"] = "face_down"

	return nil
}

// CanActivate checks if the ability can be activated
func (r *WearerOfMasksResolver) CanActivate(ctx *ability.AbilityContext) bool {
	// This is a triggered ability (as unveiled), not activated
	// It triggers automatically when unveiled with X paid
	return true
}

// ApplyEffect applies the transformation effect
func (r *WearerOfMasksResolver) ApplyEffect(ctx *ability.AbilityContext) error {
	// The actual transformation is handled by the game logic
	// This is called after OnChoice to apply the effect
	return nil
}

// RemoveEffect removes the transformation (when turned face down)
func (r *WearerOfMasksResolver) RemoveEffect(ctx *ability.AbilityContext) error {
	// The game logic handles reverting transformation when face down
	return nil
}

// NewWearerOfMasksResolver creates a new Wearer of Masks resolver
func NewWearerOfMasksResolver() *WearerOfMasksResolver {
	return &WearerOfMasksResolver{}
}
