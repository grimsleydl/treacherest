package game

import (
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

	mu sync.RWMutex
}

// AddPlayer adds a player to the room
func (r *Room) AddPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= r.MaxPlayers {
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

// CanStart checks if the game can start
func (r *Room) CanStart() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Players) >= 1 && r.State == StateLobby
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
