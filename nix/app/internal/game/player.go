package game

import (
	"time"
	"treacherest/internal/game/ability"
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

	// Ability system
	AbilityState *ability.AbilityState // Tracks pending abilities, transformations, active effects
	FaceUp       bool                  // Explicit face up/down state for role card

	// Elimination
	IsEliminated bool      // Player has been eliminated from the game
	EliminatedAt time.Time // When elimination occurred
}

// NewPlayer creates a new player
func NewPlayer(id, name, sessionID string) *Player {
	return &Player{
		ID:           id,
		Name:         name,
		SessionID:    sessionID,
		JoinedAt:     time.Now(),
		IsHost:       false, // Default to false, must be explicitly set for hosts
		AbilityState: ability.NewAbilityState(),
		FaceUp:       true, // Default to face up (will be managed by game logic)
	}
}

// MarkEliminated marks the player as eliminated from the game
func (p *Player) MarkEliminated() {
	p.IsEliminated = true
	p.EliminatedAt = time.Now()
	p.RoleRevealed = true
	p.FaceUp = true
}

// IsActiveInGame returns true if the player is actively participating (not eliminated and not host)
func (p *Player) IsActiveInGame() bool {
	return !p.IsEliminated && !p.IsHost
}
