package game

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// GameState represents the current state of a game
type GameState string

const (
	StateLobby     GameState = "lobby"
	StateCountdown GameState = "countdown"
	StatePlaying   GameState = "playing"
	StateEnded     GameState = "ended"
)

// RoleTypeConfig represents configuration for a specific role type (Leader, Guardian, etc)
type RoleTypeConfig struct {
	Count        int             `json:"count"`        // Desired number of this type
	EnabledCards map[string]bool `json:"enabledCards"` // Which specific cards are available
}

// RoleConfiguration represents the role settings for a room
type RoleConfiguration struct {
	PresetName           string                     `json:"presetName"`           // e.g., "standard", "assassination", "custom"
	MinPlayers           int                        `json:"minPlayers"`           // Minimum players needed
	MaxPlayers           int                        `json:"maxPlayers"`           // Maximum players allowed
	AllowLeaderlessGame  bool                       `json:"allowLeaderlessGame"`  // Allow games without a leader role
	HideRoleDistribution bool                       `json:"hideRoleDistribution"` // Hide role count distribution from players
	FullyRandomRoles     bool                       `json:"fullyRandomRoles"`     // Completely randomize role distribution
	RoleTypes            map[string]*RoleTypeConfig `json:"roleTypes"`            // Role type configurations
}

// ValidationState represents the current validation status of a room
type ValidationState struct {
	Version           int64     `json:"version"`           // For detecting stale states
	Timestamp         time.Time `json:"timestamp"`         // When validated
	CanStart          bool      `json:"canStart"`          // Whether game can start
	ValidationMessage string    `json:"validationMessage"` // User-friendly message
	CanAutoScale      bool      `json:"canAutoScale"`      // If auto-scaling can help
	AutoScaleDetails  string    `json:"autoScaleDetails"`  // Details about auto-scaling
	RequiredRoles     int       `json:"requiredRoles"`     // Players needing roles
	ConfiguredRoles   int       `json:"configuredRoles"`   // Currently configured roles
}

// Room represents a game room
type Room struct {
	Code       string
	State      GameState
	Players    map[string]*Player
	MaxPlayers int
	CreatedAt  time.Time
	StartedAt  time.Time

	// Countdown state
	CountdownRemaining int

	// Game state
	LeaderRevealed bool

	// Role configuration
	RoleConfig *RoleConfiguration

	// Validation versioning to detect stale UI states
	ValidationVersion int64     `json:"-"`
	LastValidatedAt   time.Time `json:"-"`

	mu sync.RWMutex
}

// AddPlayer adds a player to the room
func (r *Room) AddPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate names (case-insensitive)
	playerNameLower := strings.ToLower(player.Name)
	for _, p := range r.Players {
		if strings.ToLower(p.Name) == playerNameLower {
			return ErrDuplicateName
		}
	}

	// Count non-host players only for capacity check
	activePlayerCount := 0
	for _, p := range r.Players {
		if !p.IsHost {
			activePlayerCount++
		}
	}

	// Check capacity against non-host players only
	// Allow hosts to join without counting against the player limit
	if !player.IsHost && activePlayerCount >= r.MaxPlayers {
		return ErrRoomFull
	}

	r.Players[player.ID] = player
	return nil
}

// RemovePlayer removes a player from the room
func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, playerID)
}

// GetPlayer retrieves a player by ID
func (r *Room) GetPlayer(playerID string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Players[playerID]
}

// GetPlayers returns a copy of all players
func (r *Room) GetPlayers() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]*Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, p)
	}
	return players
}

// GetActivePlayers returns a copy of all non-host players
func (r *Room) GetActivePlayers() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]*Player, 0, len(r.Players))
	for _, p := range r.Players {
		if !p.IsHost {
			players = append(players, p)
		}
	}
	return players
}

// GetActivePlayerCount returns the number of non-host players
func (r *Room) GetActivePlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, p := range r.Players {
		if !p.IsHost {
			count++
		}
	}
	return count
}

// CanStart checks if the game can start
func (r *Room) CanStart() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Must be in lobby state
	if r.State != StateLobby {
		return false
	}

	// Count only active (non-host) players
	activePlayerCount := 0
	for _, p := range r.Players {
		if !p.IsHost {
			activePlayerCount++
		}
	}

	// Need at least 1 player
	if activePlayerCount < 1 {
		return false
	}

	// If we have role configuration, validate it
	if r.RoleConfig != nil {
		// Count total configured roles
		totalRoles := 0
		hasLeader := false

		for roleType, typeConfig := range r.RoleConfig.RoleTypes {
			if typeConfig.Count > 0 {
				totalRoles += typeConfig.Count
				if roleType == "Leader" {
					hasLeader = true
				}
			}
		}

		// Check if we have enough roles for all players
		if totalRoles < activePlayerCount {
			return false
		}

		// Check if we need a leader
		if !hasLeader && !r.RoleConfig.AllowLeaderlessGame {
			return false
		}
	}

	return true
}

// GetStartValidationError returns a detailed error message if the game cannot start
func (r *Room) GetStartValidationError() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Must be in lobby state
	if r.State != StateLobby {
		return "Game is not in lobby state"
	}

	// Count only active (non-host) players
	activePlayerCount := 0
	for _, p := range r.Players {
		if !p.IsHost {
			activePlayerCount++
		}
	}

	// Need at least 1 player
	if activePlayerCount < 1 {
		return "Need at least 1 player to start"
	}

	// If we have role configuration, validate it
	if r.RoleConfig != nil {
		// Count total configured roles
		totalRoles := 0
		hasLeader := false

		for roleType, typeConfig := range r.RoleConfig.RoleTypes {
			if typeConfig.Count > 0 {
				totalRoles += typeConfig.Count
				if roleType == "Leader" {
					hasLeader = true
				}
			}
		}

		// Check if we have enough roles for all players
		if totalRoles < activePlayerCount {
			return fmt.Sprintf("Not enough roles configured (%d) for %d players", totalRoles, activePlayerCount)
		}

		// Check if we need a leader
		if !hasLeader && !r.RoleConfig.AllowLeaderlessGame {
			return "Leader role is required (or enable leaderless games)"
		}
	}

	return ""
}

// GetLeader returns the player with the Leader role
func (r *Room) GetLeader() *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.Players {
		if p.Role != nil && p.Role.GetRoleType() == RoleLeader {
			return p
		}
	}
	return nil
}

// GetHost returns the first player who joined (usually the host)
func (r *Room) GetHost() *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// In a real implementation, we'd track who created the room
	// For now, return the first player
	for _, p := range r.Players {
		return p
	}
	return nil
}

// ValidateRoleConfig validates if the current role configuration can work with the player count
func (r *Room) ValidateRoleConfig() error {
	if r.RoleConfig == nil {
		return fmt.Errorf("no role configuration set")
	}

	activeCount := r.GetActivePlayerCount()

	// Check player count bounds
	if activeCount < r.RoleConfig.MinPlayers {
		return fmt.Errorf("need at least %d players, have %d", r.RoleConfig.MinPlayers, activeCount)
	}
	if activeCount > r.RoleConfig.MaxPlayers {
		return fmt.Errorf("maximum %d players allowed, have %d", r.RoleConfig.MaxPlayers, activeCount)
	}

	// Count total roles and validate each type
	totalRoles := 0
	hasLeader := false

	for roleType, config := range r.RoleConfig.RoleTypes {
		// Count enabled cards for this type
		enabledCount := 0
		for _, enabled := range config.EnabledCards {
			if enabled {
				enabledCount++
			}
		}

		// Check if we have enough enabled cards for the desired count
		if config.Count > enabledCount {
			return fmt.Errorf("%s: need %d cards but only %d are enabled", roleType, config.Count, enabledCount)
		}

		totalRoles += config.Count

		if roleType == "Leader" && config.Count > 0 {
			hasLeader = true
		}
	}

	// Must have exactly one leader unless leaderless game is allowed
	if !hasLeader && !r.RoleConfig.AllowLeaderlessGame {
		return fmt.Errorf("must have exactly one leader role")
	}

	// For now, we allow flexible role counts
	// The assignment logic will handle distributing roles appropriately
	return nil
}

// GetValidationState returns comprehensive validation information
// THIS IS THE SINGLE SOURCE OF TRUTH for all validation
func (r *Room) GetValidationState(roleService *RoleConfigService) ValidationState {
	r.mu.Lock()
	r.ValidationVersion++
	r.LastValidatedAt = time.Now()
	version := r.ValidationVersion
	timestamp := r.LastValidatedAt
	r.mu.Unlock()

	state := ValidationState{
		Version:           version,
		Timestamp:         timestamp,
		CanStart:          true,
		ValidationMessage: "",
		CanAutoScale:      false,
	}

	// Check basic requirements
	activeCount := r.GetActivePlayerCount()
	if activeCount < 1 {
		state.CanStart = false
		state.ValidationMessage = "Need at least 1 player to start"
		return state
	}

	// Must be in lobby state
	if r.State != StateLobby {
		state.CanStart = false
		state.ValidationMessage = "Game is not in lobby state"
		return state
	}

	// Check role configuration
	if r.RoleConfig != nil {
		totalRoles := 0
		hasLeader := false

		for roleType, config := range r.RoleConfig.RoleTypes {
			if config.Count > 0 {
				totalRoles += config.Count
				if roleType == "Leader" {
					hasLeader = true
				}
			}
		}

		state.ConfiguredRoles = totalRoles
		state.RequiredRoles = activeCount

		// Check if we have enough roles
		if totalRoles < activeCount {
			// Check if auto-scaling can help
			if roleService != nil && r.RoleConfig.PresetName != "custom" {
				canScale, details := roleService.CanAutoScale(r.RoleConfig, activeCount)
				state.CanAutoScale = canScale
				state.AutoScaleDetails = details

				if canScale {
					state.CanStart = true
					state.ValidationMessage = fmt.Sprintf("Will auto-scale roles from %d to %d players", totalRoles, activeCount)
				} else {
					state.CanStart = false
					state.ValidationMessage = fmt.Sprintf("Not enough roles configured (%d) for %d players. %s", totalRoles, activeCount, details)
				}
			} else {
				state.CanStart = false
				state.ValidationMessage = fmt.Sprintf("Not enough roles configured (%d) for %d players", totalRoles, activeCount)
				if r.RoleConfig.PresetName == "custom" {
					state.AutoScaleDetails = "Custom configurations do not support auto-scaling"
				}
			}
		}

		// Check leader requirement
		if !hasLeader && !r.RoleConfig.AllowLeaderlessGame && state.CanStart {
			state.CanStart = false
			state.ValidationMessage = "Leader role is required (or enable leaderless games)"
		}
	}

	return state
}
