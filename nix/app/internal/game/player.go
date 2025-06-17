package game

import (
	"time"
)

// Player represents a player in the game
type Player struct {
	ID           string
	Name         string
	Role         *Role
	RoleRevealed bool
	JoinedAt     time.Time
	SessionID    string // Used for reconnection
}

// NewPlayer creates a new player
func NewPlayer(id, name, sessionID string) *Player {
	return &Player{
		ID:        id,
		Name:      name,
		SessionID: sessionID,
		JoinedAt:  time.Now(),
	}
}
