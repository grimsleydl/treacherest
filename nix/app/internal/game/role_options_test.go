package game

import (
	"testing"
)

// TestNewRoleOptions tests role options creation
func TestNewRoleOptions(t *testing.T) {
	opts := NewRoleOptions(31) // The Wearer of Masks

	if opts == nil {
		t.Fatal("Expected role options to be created")
	}

	if opts.CardID != 31 {
		t.Errorf("Expected CardID=31, got %d", opts.CardID)
	}

	if opts.Options == nil {
		t.Fatal("Expected Options map to be initialized")
	}

	if len(opts.Options) != 0 {
		t.Errorf("Expected empty options map, got %d entries", len(opts.Options))
	}
}

// TestSetOption tests setting options
func TestSetOption(t *testing.T) {
	opts := NewRoleOptions(31)

	opts.SetOption("use_all_cards", true)
	opts.SetOption("max_reveal", 5)
	opts.SetOption("custom_filter", "traitors_only")

	if len(opts.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(opts.Options))
	}

	// Check values
	if val, ok := opts.GetOption("use_all_cards"); !ok || val != true {
		t.Error("Expected use_all_cards=true")
	}

	if val, ok := opts.GetOption("max_reveal"); !ok || val != 5 {
		t.Error("Expected max_reveal=5")
	}
}

// TestGetOption tests retrieving options
func TestGetOption(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("test_key", "test_value")

	// Existing option
	val, exists := opts.GetOption("test_key")
	if !exists {
		t.Error("Expected option to exist")
	}

	if val != "test_value" {
		t.Errorf("Expected 'test_value', got %v", val)
	}

	// Non-existent option
	val, exists = opts.GetOption("non_existent")
	if exists {
		t.Error("Expected option not to exist")
	}

	if val != nil {
		t.Errorf("Expected nil value for non-existent option, got %v", val)
	}
}

// TestGetBoolOption tests type-safe bool retrieval
func TestGetBoolOption(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("enabled", true)
	opts.SetOption("disabled", false)
	opts.SetOption("not_a_bool", "string_value")

	// True value
	val, err := opts.GetBoolOption("enabled")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !val {
		t.Error("Expected true")
	}

	// False value
	val, err = opts.GetBoolOption("disabled")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val {
		t.Error("Expected false")
	}

	// Wrong type
	_, err = opts.GetBoolOption("not_a_bool")
	if err == nil {
		t.Error("Expected error for wrong type")
	}

	// Non-existent
	_, err = opts.GetBoolOption("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent option")
	}
}

// TestGetIntOption tests type-safe int retrieval
func TestGetIntOption(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("count", 5)
	opts.SetOption("zero", 0)
	opts.SetOption("not_an_int", "string")

	// Positive value
	val, err := opts.GetIntOption("count")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != 5 {
		t.Errorf("Expected 5, got %d", val)
	}

	// Zero value
	val, err = opts.GetIntOption("zero")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}

	// Wrong type
	_, err = opts.GetIntOption("not_an_int")
	if err == nil {
		t.Error("Expected error for wrong type")
	}

	// Non-existent
	_, err = opts.GetIntOption("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent option")
	}
}

// TestGetStringOption tests type-safe string retrieval
func TestGetStringOption(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("mode", "advanced")
	opts.SetOption("empty", "")
	opts.SetOption("not_a_string", 123)

	// String value
	val, err := opts.GetStringOption("mode")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "advanced" {
		t.Errorf("Expected 'advanced', got %s", val)
	}

	// Empty string
	val, err = opts.GetStringOption("empty")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "" {
		t.Errorf("Expected empty string, got %s", val)
	}

	// Wrong type
	_, err = opts.GetStringOption("not_a_string")
	if err == nil {
		t.Error("Expected error for wrong type")
	}

	// Non-existent
	_, err = opts.GetStringOption("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent option")
	}
}

// TestHasOption tests checking option existence
func TestHasOption(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("existing", "value")

	if !opts.HasOption("existing") {
		t.Error("Expected option to exist")
	}

	if opts.HasOption("non_existent") {
		t.Error("Expected option not to exist")
	}
}

// TestDeleteOption tests removing options
func TestDeleteOption(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("to_delete", "value")

	if !opts.HasOption("to_delete") {
		t.Error("Option should exist before deletion")
	}

	opts.DeleteOption("to_delete")

	if opts.HasOption("to_delete") {
		t.Error("Option should not exist after deletion")
	}

	// Delete non-existent (should not panic)
	opts.DeleteOption("non_existent")
}

// TestClearOptions tests clearing all options
func TestClearOptions(t *testing.T) {
	opts := NewRoleOptions(31)
	opts.SetOption("opt1", "val1")
	opts.SetOption("opt2", "val2")
	opts.SetOption("opt3", "val3")

	if len(opts.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(opts.Options))
	}

	opts.ClearOptions()

	if len(opts.Options) != 0 {
		t.Errorf("Expected 0 options after clear, got %d", len(opts.Options))
	}
}

// TestRoleOptionsManager tests the manager for multiple cards
func TestRoleOptionsManager(t *testing.T) {
	manager := NewRoleOptionsManager()

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	// Get options for card (should auto-create)
	opts := manager.GetOrCreateOptions(31)
	if opts == nil {
		t.Fatal("Expected options to be created")
	}

	if opts.CardID != 31 {
		t.Errorf("Expected CardID=31, got %d", opts.CardID)
	}

	// Set some options
	opts.SetOption("use_all_cards", true)

	// Retrieve again (should be same instance)
	opts2 := manager.GetOrCreateOptions(31)
	if opts2 != opts {
		t.Error("Expected same options instance")
	}

	val, _ := opts2.GetOption("use_all_cards")
	if val != true {
		t.Error("Expected option to persist")
	}

	// Different card
	opts3 := manager.GetOrCreateOptions(25)
	if opts3.CardID != 25 {
		t.Errorf("Expected CardID=25, got %d", opts3.CardID)
	}

	if len(opts3.Options) != 0 {
		t.Error("Expected new card options to be empty")
	}
}

// TestRoleOptionsManagerHas tests checking if options exist
func TestRoleOptionsManagerHas(t *testing.T) {
	manager := NewRoleOptionsManager()

	if manager.HasOptions(31) {
		t.Error("Should not have options initially")
	}

	manager.GetOrCreateOptions(31)

	if !manager.HasOptions(31) {
		t.Error("Should have options after creation")
	}
}

// TestRoleOptionsManagerDelete tests deleting card options
func TestRoleOptionsManagerDelete(t *testing.T) {
	manager := NewRoleOptionsManager()

	manager.GetOrCreateOptions(31)
	manager.GetOrCreateOptions(25)

	if !manager.HasOptions(31) {
		t.Error("Options should exist")
	}

	manager.DeleteOptions(31)

	if manager.HasOptions(31) {
		t.Error("Options should be deleted")
	}

	if !manager.HasOptions(25) {
		t.Error("Other options should still exist")
	}
}
