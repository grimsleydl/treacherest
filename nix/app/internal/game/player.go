package game

import (
	"time"
)

// Player represents a player in the game
type Player struct {
	ID           string
	Name         string
	Role         *Card
	RoleRevealed bool
	JoinedAt     time.Time
	SessionID    string // Used for reconnection
	IsHost       bool   // Indicates if the player is the host who created the room but doesn't participate
}

// NewPlayer creates a new player
func NewPlayer(id, name, sessionID string) *Player {
	return &Player{
		ID:        id,
		Name:      name,
		SessionID: sessionID,
		JoinedAt:  time.Now(),
		IsHost:    false, // Default to false, must be explicitly set for hosts
	}
}
