package ability

import (
	"time"
)

// AbilityState tracks all ability-related state for a player
type AbilityState struct {
	PendingAbilities []*PendingAbility
	ActiveEffects    []*ActiveEffect
	TransformState   *TransformState
}

// PendingAbility represents an ability awaiting player choice or resolution
type PendingAbility struct {
	ID             string
	PlayerID       string
	CardID         int
	AbilityType    string
	Data           map[string]interface{} // Flexible data for ability-specific info
	ModalDismissed bool
	ModalState     map[string]interface{}
	CreatedAt      time.Time
}

// ActiveEffect represents an ongoing effect on a player
type ActiveEffect struct {
	ID           string
	SourceCardID int
	EffectType   EffectType
	Params       EffectParams
	Duration     *Duration
	AppliedAt    time.Time
}

// TransformState tracks role transformation (e.g., The Wearer of Masks)
// Stores card IDs to avoid circular imports with game package
type TransformState struct {
	IsTransformed     bool
	OriginalCardID    int      // ID of original card
	TransformedCardID int      // ID of transformed card
	KeepTypes         []string // Types to retain (e.g., ["Traitor"])
	EndCondition      string   // When to revert (e.g., "face_down")
}

// NewAbilityState creates a new ability state
func NewAbilityState() *AbilityState {
	return &AbilityState{
		PendingAbilities: make([]*PendingAbility, 0),
		ActiveEffects:    make([]*ActiveEffect, 0),
		TransformState:   nil,
	}
}

// AddPendingAbility adds a new pending ability
func (as *AbilityState) AddPendingAbility(ability *PendingAbility) {
	if ability.CreatedAt.IsZero() {
		ability.CreatedAt = time.Now()
	}
	as.PendingAbilities = append(as.PendingAbilities, ability)
}

// GetPendingAbility retrieves a pending ability by ID
func (as *AbilityState) GetPendingAbility(abilityID string) *PendingAbility {
	for _, ability := range as.PendingAbilities {
		if ability.ID == abilityID {
			return ability
		}
	}
	return nil
}

// ResolvePendingAbility removes a pending ability (after resolution)
func (as *AbilityState) ResolvePendingAbility(abilityID string) {
	for i, ability := range as.PendingAbilities {
		if ability.ID == abilityID {
			// Remove by slicing
			as.PendingAbilities = append(as.PendingAbilities[:i], as.PendingAbilities[i+1:]...)
			return
		}
	}
}

// HasPendingAbilities checks if there are any pending abilities
func (as *AbilityState) HasPendingAbilities() bool {
	return len(as.PendingAbilities) > 0
}

// AddActiveEffect adds a new active effect
func (as *AbilityState) AddActiveEffect(effect *ActiveEffect) {
	if effect.AppliedAt.IsZero() {
		effect.AppliedAt = time.Now()
	}
	as.ActiveEffects = append(as.ActiveEffects, effect)
}

// RemoveActiveEffect removes an active effect by ID
func (as *AbilityState) RemoveActiveEffect(effectID string) {
	for i, effect := range as.ActiveEffects {
		if effect.ID == effectID {
			as.ActiveEffects = append(as.ActiveEffects[:i], as.ActiveEffects[i+1:]...)
			return
		}
	}
}

// StartTransform begins a role transformation
func (as *AbilityState) StartTransform(originalCardID, transformedCardID int, keepTypes []string, endCondition string) {
	as.TransformState = &TransformState{
		IsTransformed:     true,
		OriginalCardID:    originalCardID,
		TransformedCardID: transformedCardID,
		KeepTypes:         keepTypes,
		EndCondition:      endCondition,
	}
}

// EndTransform ends a role transformation and returns the original card ID
func (as *AbilityState) EndTransform() int {
	if as.TransformState == nil {
		return 0
	}

	originalCardID := as.TransformState.OriginalCardID
	as.TransformState = nil
	return originalCardID
}

// IsTransformed checks if the player is currently transformed
func (as *AbilityState) IsTransformed() bool {
	return as.TransformState != nil && as.TransformState.IsTransformed
}

// CheckTransformEndCondition checks if the transform should end
func (as *AbilityState) CheckTransformEndCondition(condition string) bool {
	if !as.IsTransformed() {
		return false
	}
	return as.TransformState.EndCondition == condition
}

// GetTransformedCardID returns the transformed card ID if transformed, otherwise 0
func (as *AbilityState) GetTransformedCardID() int {
	if as.IsTransformed() {
		return as.TransformState.TransformedCardID
	}
	return 0
}

// GetOriginalCardID returns the original card ID if transformed, otherwise 0
func (as *AbilityState) GetOriginalCardID() int {
	if as.IsTransformed() {
		return as.TransformState.OriginalCardID
	}
	return 0
}

// DismissModal sets the modal dismissed flag for a pending ability
func (as *AbilityState) DismissModal(abilityID string) bool {
	ability := as.GetPendingAbility(abilityID)
	if ability == nil {
		return false
	}
	ability.ModalDismissed = true
	return true
}

// RestoreModal clears the modal dismissed flag for a pending ability
func (as *AbilityState) RestoreModal(abilityID string) bool {
	ability := as.GetPendingAbility(abilityID)
	if ability == nil {
		return false
	}
	ability.ModalDismissed = false
	return true
}

// IsModalDismissed checks if a modal is dismissed for a pending ability
func (as *AbilityState) IsModalDismissed(abilityID string) bool {
	ability := as.GetPendingAbility(abilityID)
	if ability == nil {
		return false
	}
	return ability.ModalDismissed
}

// SetModalState stores state data for a modal
func (as *AbilityState) SetModalState(abilityID string, key string, value interface{}) bool {
	ability := as.GetPendingAbility(abilityID)
	if ability == nil {
		return false
	}
	if ability.ModalState == nil {
		ability.ModalState = make(map[string]interface{})
	}
	ability.ModalState[key] = value
	return true
}

// GetModalState retrieves state data for a modal
func (as *AbilityState) GetModalState(abilityID string, key string) (interface{}, bool) {
	ability := as.GetPendingAbility(abilityID)
	if ability == nil || ability.ModalState == nil {
		return nil, false
	}
	value, exists := ability.ModalState[key]
	return value, exists
}
