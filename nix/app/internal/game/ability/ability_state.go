package ability

import (
	"time"
)

// AbilityState tracks all ability-related state for a player
type AbilityState struct {
	PendingAbilities []*PendingAbility
	ActiveEffects    []*ActiveEffect
	TransformState   *TransformState
	MetamorphState   *MetamorphState // Tracks The Metamorph's steal ability
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

	// Confirmation system - prevents peeking at ability choices before physically revealing
	RequiresConfirmation bool     // Whether this ability needs external confirmation
	ConfirmationRole     string   // Who must confirm: "leader", "all_players", "any_player"
	ConfirmedBy          []string // Player IDs who have confirmed
}

// IsConfirmed checks if the ability has received required confirmations
func (pa *PendingAbility) IsConfirmed() bool {
	if !pa.RequiresConfirmation {
		return true
	}
	return len(pa.ConfirmedBy) > 0
}

// AddConfirmation adds a player's confirmation
func (pa *PendingAbility) AddConfirmation(playerID string) {
	// Check if already confirmed by this player
	for _, id := range pa.ConfirmedBy {
		if id == playerID {
			return
		}
	}
	pa.ConfirmedBy = append(pa.ConfirmedBy, playerID)
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

// RoleChangeType distinguishes between different kinds of role changes
type RoleChangeType string

const (
	RoleChangeTransform RoleChangeType = "transform" // Wearer of Masks - reversible
	RoleChangeSteal     RoleChangeType = "steal"     // Metamorph - permanent
)

// TransformState tracks role changes (transformation or identity stealing)
// Stores card IDs to avoid circular imports with game package
type TransformState struct {
	IsTransformed     bool
	OriginalCardID    int      // ID of original card
	TransformedCardID int      // ID of new card
	KeepTypes         []string // Types to retain (empty for full identity change)
	EndCondition      string   // When to revert ("never" for permanent)

	// Extended fields for abstraction
	ChangeType        RoleChangeType // "transform" or "steal"
	IsPermanent       bool           // If true, cannot revert
	SourceCardRemoved bool           // If true, original card removed from game (Metamorph)
}

// MetamorphState tracks The Metamorph's temporary steal WINDOW (not the role change itself)
// The actual role change is tracked by TransformState with ChangeType: RoleChangeSteal
type MetamorphState struct {
	IsActive    bool      // Effect window is active (until end of turn)
	ActivatedAt time.Time // When window opened
	HasBeenUsed bool      // Once used to steal, window closes
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

// StartTransform begins a role transformation (Wearer of Masks - reversible)
func (as *AbilityState) StartTransform(originalCardID, transformedCardID int, keepTypes []string, endCondition string) {
	as.TransformState = &TransformState{
		IsTransformed:     true,
		OriginalCardID:    originalCardID,
		TransformedCardID: transformedCardID,
		KeepTypes:         keepTypes,
		EndCondition:      endCondition,
		ChangeType:        RoleChangeTransform,
		IsPermanent:       false,
		SourceCardRemoved: false,
	}
}

// StealIdentity begins a permanent identity steal (Metamorph - irreversible)
func (as *AbilityState) StealIdentity(originalCardID, stolenCardID int) {
	as.TransformState = &TransformState{
		IsTransformed:     true,
		OriginalCardID:    originalCardID,
		TransformedCardID: stolenCardID,
		KeepTypes:         []string{},    // No types kept - full identity change
		EndCondition:      "never",       // Permanent
		ChangeType:        RoleChangeSteal,
		IsPermanent:       true,
		SourceCardRemoved: true, // Metamorph removed from game
	}
}

// IsStolenIdentity checks if this is a permanent identity steal (Metamorph)
func (as *AbilityState) IsStolenIdentity() bool {
	return as.TransformState != nil && as.TransformState.ChangeType == RoleChangeSteal
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

// ActivateMetamorph activates The Metamorph's steal window
func (as *AbilityState) ActivateMetamorph() {
	as.MetamorphState = &MetamorphState{
		IsActive:    true,
		ActivatedAt: time.Now(),
		HasBeenUsed: false,
	}
}

// DeactivateMetamorph deactivates The Metamorph's steal window (end of turn)
func (as *AbilityState) DeactivateMetamorph() {
	if as.MetamorphState != nil {
		as.MetamorphState.IsActive = false
	}
}

// UseMetamorph marks The Metamorph ability window as used
// The actual identity theft is tracked via StealIdentity() in TransformState
func (as *AbilityState) UseMetamorph() {
	if as.MetamorphState != nil {
		as.MetamorphState.HasBeenUsed = true
		as.MetamorphState.IsActive = false
	}
}

// IsMetamorphActive checks if The Metamorph's steal ability is currently active
func (as *AbilityState) IsMetamorphActive() bool {
	return as.MetamorphState != nil && as.MetamorphState.IsActive && !as.MetamorphState.HasBeenUsed
}

// CanUseMetamorph checks if The Metamorph can be used to steal
func (as *AbilityState) CanUseMetamorph() bool {
	return as.MetamorphState != nil && as.MetamorphState.IsActive && !as.MetamorphState.HasBeenUsed
}
