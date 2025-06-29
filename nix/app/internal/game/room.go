package game

import (
	"fmt"
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

// RoleConfiguration represents the role settings for a room
type RoleConfiguration struct {
	PresetName          string           `json:"presetName"`          // e.g., "standard", "assassination", "custom"
	EnabledRoles        map[string]bool  `json:"enabledRoles"`        // Which roles from master pool are enabled
	RoleCounts          map[string]int   `json:"roleCounts"`          // Exact counts per role
	MinPlayers          int              `json:"minPlayers"`          // Minimum players needed
	MaxPlayers          int              `json:"maxPlayers"`          // Maximum players allowed
	AllowLeaderlessGame bool             `json:"allowLeaderlessGame"` // Allow games without a leader role
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

	mu sync.RWMutex
}

// AddPlayer adds a player to the room
func (r *Room) AddPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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

	// Count only active (non-host) players
	activePlayerCount := 0
	for _, p := range r.Players {
		if !p.IsHost {
			activePlayerCount++
		}
	}

	return activePlayerCount >= 1 && r.State == StateLobby
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

	// Count total roles
	totalRoles := 0
	hasLeader := false
	for role, count := range r.RoleConfig.RoleCounts {
		if r.RoleConfig.EnabledRoles[role] {
			totalRoles += count
			if role == "leader" {
				hasLeader = true
			}
		}
	}

	// Must have exactly one leader
	if !hasLeader {
		return fmt.Errorf("must have exactly one leader role")
	}

	// For now, we allow flexible role counts
	// The assignment logic will handle distributing roles appropriately
	return nil
}
