package ability

import (
	"testing"
	"time"
)

// TestNewAbilityState tests ability state creation
func TestNewAbilityState(t *testing.T) {
	state := NewAbilityState()

	if state == nil {
		t.Fatal("Expected ability state to be created")
	}

	if state.PendingAbilities == nil {
		t.Error("Expected PendingAbilities to be initialized")
	}

	if state.ActiveEffects == nil {
		t.Error("Expected ActiveEffects to be initialized")
	}

	if state.TransformState != nil {
		t.Error("Expected TransformState to be nil initially")
	}
}

// TestAddPendingAbility tests adding pending abilities
func TestAddPendingAbility(t *testing.T) {
	state := NewAbilityState()

	ability := &PendingAbility{
		ID:          "ability-1",
		PlayerID:    "player-1",
		CardID:      31,
		AbilityType: "unveil",
		Data:        make(map[string]interface{}),
	}

	state.AddPendingAbility(ability)

	if len(state.PendingAbilities) != 1 {
		t.Errorf("Expected 1 pending ability, got %d", len(state.PendingAbilities))
	}

	if state.PendingAbilities[0].ID != "ability-1" {
		t.Errorf("Expected ability ID 'ability-1', got %s", state.PendingAbilities[0].ID)
	}
}

// TestGetPendingAbility tests retrieving pending abilities
func TestGetPendingAbility(t *testing.T) {
	state := NewAbilityState()

	ability := &PendingAbility{
		ID:       "ability-1",
		PlayerID: "player-1",
		CardID:   31,
	}

	state.AddPendingAbility(ability)

	retrieved := state.GetPendingAbility("ability-1")
	if retrieved == nil {
		t.Fatal("Expected to retrieve pending ability")
	}

	if retrieved.ID != "ability-1" {
		t.Errorf("Expected ID 'ability-1', got %s", retrieved.ID)
	}

	// Test non-existent ability
	notFound := state.GetPendingAbility("non-existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent ability")
	}
}

// TestResolvePendingAbility tests resolving (removing) pending abilities
func TestResolvePendingAbility(t *testing.T) {
	state := NewAbilityState()

	ability1 := &PendingAbility{ID: "ability-1", PlayerID: "player-1"}
	ability2 := &PendingAbility{ID: "ability-2", PlayerID: "player-1"}

	state.AddPendingAbility(ability1)
	state.AddPendingAbility(ability2)

	if len(state.PendingAbilities) != 2 {
		t.Errorf("Expected 2 pending abilities, got %d", len(state.PendingAbilities))
	}

	state.ResolvePendingAbility("ability-1")

	if len(state.PendingAbilities) != 1 {
		t.Errorf("Expected 1 pending ability after resolution, got %d", len(state.PendingAbilities))
	}

	if state.PendingAbilities[0].ID != "ability-2" {
		t.Error("Wrong ability was resolved")
	}
}

// TestAddActiveEffect tests adding active effects
func TestAddActiveEffect(t *testing.T) {
	state := NewAbilityState()

	effect := &ActiveEffect{
		ID:           "effect-1",
		SourceCardID: 31,
		EffectType:   EffectTransform,
		AppliedAt:    time.Now(),
	}

	state.AddActiveEffect(effect)

	if len(state.ActiveEffects) != 1 {
		t.Errorf("Expected 1 active effect, got %d", len(state.ActiveEffects))
	}
}

// TestRemoveActiveEffect tests removing active effects
func TestRemoveActiveEffect(t *testing.T) {
	state := NewAbilityState()

	effect1 := &ActiveEffect{ID: "effect-1", SourceCardID: 31}
	effect2 := &ActiveEffect{ID: "effect-2", SourceCardID: 25}

	state.AddActiveEffect(effect1)
	state.AddActiveEffect(effect2)

	state.RemoveActiveEffect("effect-1")

	if len(state.ActiveEffects) != 1 {
		t.Errorf("Expected 1 active effect, got %d", len(state.ActiveEffects))
	}

	if state.ActiveEffects[0].ID != "effect-2" {
		t.Error("Wrong effect was removed")
	}
}

// TestStartTransform tests starting a transformation
func TestStartTransform(t *testing.T) {
	state := NewAbilityState()

	originalCardID := 31    // The Wearer of Masks
	transformedCardID := 15 // The Bodyguard
	keepTypes := []string{"Traitor"}
	endCondition := "face_down"

	state.StartTransform(originalCardID, transformedCardID, keepTypes, endCondition)

	if state.TransformState == nil {
		t.Fatal("Expected transform state to be created")
	}

	if !state.TransformState.IsTransformed {
		t.Error("Expected IsTransformed to be true")
	}

	if state.TransformState.OriginalCardID != 31 {
		t.Errorf("Expected OriginalCardID 31, got %d", state.TransformState.OriginalCardID)
	}

	if state.TransformState.TransformedCardID != 15 {
		t.Errorf("Expected TransformedCardID 15, got %d", state.TransformState.TransformedCardID)
	}

	if len(state.TransformState.KeepTypes) != 1 || state.TransformState.KeepTypes[0] != "Traitor" {
		t.Errorf("Expected KeepTypes [Traitor], got %v", state.TransformState.KeepTypes)
	}

	if state.TransformState.EndCondition != "face_down" {
		t.Errorf("Expected EndCondition 'face_down', got %s", state.TransformState.EndCondition)
	}
}

// TestEndTransform tests ending a transformation
func TestEndTransform(t *testing.T) {
	state := NewAbilityState()

	originalCardID := 31    // The Wearer of Masks
	transformedCardID := 15 // The Bodyguard

	state.StartTransform(originalCardID, transformedCardID, []string{"Traitor"}, "face_down")

	if !state.TransformState.IsTransformed {
		t.Error("Expected IsTransformed to be true")
	}

	returnedID := state.EndTransform()

	if returnedID != 31 {
		t.Errorf("Expected original card ID 31, got %d", returnedID)
	}

	if state.TransformState != nil {
		t.Error("Expected TransformState to be nil after ending transform")
	}
}

// TestIsTransformed tests transformation status check
func TestIsTransformed(t *testing.T) {
	state := NewAbilityState()

	if state.IsTransformed() {
		t.Error("Expected IsTransformed to be false initially")
	}

	originalCardID := 31
	transformedCardID := 15

	state.StartTransform(originalCardID, transformedCardID, []string{"Traitor"}, "face_down")

	if !state.IsTransformed() {
		t.Error("Expected IsTransformed to be true after transform")
	}

	state.EndTransform()

	if state.IsTransformed() {
		t.Error("Expected IsTransformed to be false after ending transform")
	}
}

// TestCheckTransformEndCondition tests condition checking
func TestCheckTransformEndCondition(t *testing.T) {
	state := NewAbilityState()

	// No transform active
	if state.CheckTransformEndCondition("face_down") {
		t.Error("Expected false when no transform is active")
	}

	originalCardID := 31
	transformedCardID := 15

	state.StartTransform(originalCardID, transformedCardID, []string{"Traitor"}, "face_down")

	// Correct condition
	if !state.CheckTransformEndCondition("face_down") {
		t.Error("Expected true for matching condition")
	}

	// Wrong condition
	if state.CheckTransformEndCondition("lose_game") {
		t.Error("Expected false for non-matching condition")
	}
}

// TestHasPendingAbilities tests checking for pending abilities
func TestHasPendingAbilities(t *testing.T) {
	state := NewAbilityState()

	if state.HasPendingAbilities() {
		t.Error("Expected false when no pending abilities")
	}

	ability := &PendingAbility{ID: "ability-1", PlayerID: "player-1"}
	state.AddPendingAbility(ability)

	if !state.HasPendingAbilities() {
		t.Error("Expected true when pending abilities exist")
	}

	state.ResolvePendingAbility("ability-1")

	if state.HasPendingAbilities() {
		t.Error("Expected false after resolving all abilities")
	}
}

// TestGetTransformedCardID tests getting transformed card ID
func TestGetTransformedCardID(t *testing.T) {
	state := NewAbilityState()

	// No transformation - should return 0
	transformedID := state.GetTransformedCardID()
	if transformedID != 0 {
		t.Errorf("Expected 0 when not transformed, got %d", transformedID)
	}

	// With transformation - should return transformed card ID
	originalCardID := 31
	transformedCardID := 15
	state.StartTransform(originalCardID, transformedCardID, []string{"Traitor"}, "face_down")

	transformedID = state.GetTransformedCardID()
	if transformedID != 15 {
		t.Errorf("Expected transformed card ID 15, got %d", transformedID)
	}
}

// TestGetOriginalCardID tests getting original card ID
func TestGetOriginalCardID(t *testing.T) {
	state := NewAbilityState()

	// No transformation - should return 0
	originalID := state.GetOriginalCardID()
	if originalID != 0 {
		t.Errorf("Expected 0 when not transformed, got %d", originalID)
	}

	// With transformation - should return original card ID
	originalCardID := 31
	transformedCardID := 15
	state.StartTransform(originalCardID, transformedCardID, []string{"Traitor"}, "face_down")

	originalID = state.GetOriginalCardID()
	if originalID != 31 {
		t.Errorf("Expected original card ID 31, got %d", originalID)
	}
}

// TestDismissModal tests dismissing a modal for a pending ability
func TestDismissModal(t *testing.T) {
	state := NewAbilityState()

	ability := &PendingAbility{
		ID:       "ability-1",
		PlayerID: "player-1",
		CardID:   31,
	}

	state.AddPendingAbility(ability)

	// Initially not dismissed
	if state.IsModalDismissed("ability-1") {
		t.Error("Expected modal to not be dismissed initially")
	}

	// Dismiss modal
	success := state.DismissModal("ability-1")
	if !success {
		t.Error("Expected DismissModal to succeed")
	}

	// Check dismissed status
	if !state.IsModalDismissed("ability-1") {
		t.Error("Expected modal to be dismissed after DismissModal")
	}

	// Try to dismiss non-existent ability
	success = state.DismissModal("non-existent")
	if success {
		t.Error("Expected DismissModal to fail for non-existent ability")
	}
}

// TestRestoreModal tests restoring a dismissed modal
func TestRestoreModal(t *testing.T) {
	state := NewAbilityState()

	ability := &PendingAbility{
		ID:             "ability-1",
		PlayerID:       "player-1",
		CardID:         31,
		ModalDismissed: true, // Start dismissed
	}

	state.AddPendingAbility(ability)

	// Initially dismissed
	if !state.IsModalDismissed("ability-1") {
		t.Error("Expected modal to be dismissed initially")
	}

	// Restore modal
	success := state.RestoreModal("ability-1")
	if !success {
		t.Error("Expected RestoreModal to succeed")
	}

	// Check restored status
	if state.IsModalDismissed("ability-1") {
		t.Error("Expected modal to not be dismissed after RestoreModal")
	}

	// Try to restore non-existent ability
	success = state.RestoreModal("non-existent")
	if success {
		t.Error("Expected RestoreModal to fail for non-existent ability")
	}
}

// TestSetModalState tests setting modal state data
func TestSetModalState(t *testing.T) {
	state := NewAbilityState()

	ability := &PendingAbility{
		ID:       "ability-1",
		PlayerID: "player-1",
		CardID:   31,
	}

	state.AddPendingAbility(ability)

	// Set some modal state
	success := state.SetModalState("ability-1", "selected_card_id", 15)
	if !success {
		t.Error("Expected SetModalState to succeed")
	}

	success = state.SetModalState("ability-1", "filter", "Traitor")
	if !success {
		t.Error("Expected SetModalState to succeed")
	}

	// Retrieve state
	val, exists := state.GetModalState("ability-1", "selected_card_id")
	if !exists {
		t.Error("Expected modal state 'selected_card_id' to exist")
	}
	if cardID, ok := val.(int); !ok || cardID != 15 {
		t.Errorf("Expected selected_card_id to be 15, got %v", val)
	}

	val, exists = state.GetModalState("ability-1", "filter")
	if !exists {
		t.Error("Expected modal state 'filter' to exist")
	}
	if filter, ok := val.(string); !ok || filter != "Traitor" {
		t.Errorf("Expected filter to be 'Traitor', got %v", val)
	}

	// Try to set state for non-existent ability
	success = state.SetModalState("non-existent", "key", "value")
	if success {
		t.Error("Expected SetModalState to fail for non-existent ability")
	}
}

// TestGetModalState tests retrieving modal state data
func TestGetModalState(t *testing.T) {
	state := NewAbilityState()

	ability := &PendingAbility{
		ID:       "ability-1",
		PlayerID: "player-1",
		CardID:   31,
		ModalState: map[string]interface{}{
			"option1": "value1",
			"option2": 42,
		},
	}

	state.AddPendingAbility(ability)

	// Get existing state
	val, exists := state.GetModalState("ability-1", "option1")
	if !exists {
		t.Error("Expected modal state 'option1' to exist")
	}
	if str, ok := val.(string); !ok || str != "value1" {
		t.Errorf("Expected option1 to be 'value1', got %v", val)
	}

	val, exists = state.GetModalState("ability-1", "option2")
	if !exists {
		t.Error("Expected modal state 'option2' to exist")
	}
	if num, ok := val.(int); !ok || num != 42 {
		t.Errorf("Expected option2 to be 42, got %v", val)
	}

	// Get non-existent state
	val, exists = state.GetModalState("ability-1", "non-existent")
	if exists {
		t.Error("Expected non-existent state to not exist")
	}
	if val != nil {
		t.Errorf("Expected nil value for non-existent state, got %v", val)
	}

	// Get state for non-existent ability
	val, exists = state.GetModalState("non-existent", "option1")
	if exists {
		t.Error("Expected GetModalState to fail for non-existent ability")
	}
}
