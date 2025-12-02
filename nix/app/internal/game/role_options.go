package game

import (
	"fmt"
	"sync"
)

// RoleOptions stores card-specific configuration options
// Example: The Wearer of Masks might have options for "use_all_cards" or "max_reveal"
type RoleOptions struct {
	CardID  int
	Options map[string]interface{}
	mu      sync.RWMutex
}

// NewRoleOptions creates new role options for a specific card
func NewRoleOptions(cardID int) *RoleOptions {
	return &RoleOptions{
		CardID:  cardID,
		Options: make(map[string]interface{}),
	}
}

// SetOption sets an option value
func (ro *RoleOptions) SetOption(key string, value interface{}) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.Options[key] = value
}

// GetOption retrieves an option value
func (ro *RoleOptions) GetOption(key string) (interface{}, bool) {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	val, exists := ro.Options[key]
	return val, exists
}

// GetBoolOption retrieves a boolean option with type checking
func (ro *RoleOptions) GetBoolOption(key string) (bool, error) {
	val, exists := ro.GetOption(key)
	if !exists {
		return false, fmt.Errorf("option '%s' does not exist", key)
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("option '%s' is not a bool", key)
	}

	return boolVal, nil
}

// GetIntOption retrieves an integer option with type checking
func (ro *RoleOptions) GetIntOption(key string) (int, error) {
	val, exists := ro.GetOption(key)
	if !exists {
		return 0, fmt.Errorf("option '%s' does not exist", key)
	}

	intVal, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("option '%s' is not an int", key)
	}

	return intVal, nil
}

// GetStringOption retrieves a string option with type checking
func (ro *RoleOptions) GetStringOption(key string) (string, error) {
	val, exists := ro.GetOption(key)
	if !exists {
		return "", fmt.Errorf("option '%s' does not exist", key)
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("option '%s' is not a string", key)
	}

	return strVal, nil
}

// HasOption checks if an option exists
func (ro *RoleOptions) HasOption(key string) bool {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	_, exists := ro.Options[key]
	return exists
}

// DeleteOption removes an option
func (ro *RoleOptions) DeleteOption(key string) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	delete(ro.Options, key)
}

// ClearOptions removes all options
func (ro *RoleOptions) ClearOptions() {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.Options = make(map[string]interface{})
}

// RoleOptionsManager manages options for multiple cards
type RoleOptionsManager struct {
	options map[int]*RoleOptions // Card ID -> RoleOptions
	mu      sync.RWMutex
}

// NewRoleOptionsManager creates a new options manager
func NewRoleOptionsManager() *RoleOptionsManager {
	return &RoleOptionsManager{
		options: make(map[int]*RoleOptions),
	}
}

// GetOrCreateOptions retrieves or creates options for a card
func (rom *RoleOptionsManager) GetOrCreateOptions(cardID int) *RoleOptions {
	rom.mu.Lock()
	defer rom.mu.Unlock()

	if opts, exists := rom.options[cardID]; exists {
		return opts
	}

	opts := NewRoleOptions(cardID)
	rom.options[cardID] = opts
	return opts
}

// HasOptions checks if options exist for a card
func (rom *RoleOptionsManager) HasOptions(cardID int) bool {
	rom.mu.RLock()
	defer rom.mu.RUnlock()

	_, exists := rom.options[cardID]
	return exists
}

// DeleteOptions removes options for a card
func (rom *RoleOptionsManager) DeleteOptions(cardID int) {
	rom.mu.Lock()
	defer rom.mu.Unlock()

	delete(rom.options, cardID)
}
