package game

import (
	"errors"
	"fmt"
)

const (
	HerSeedbornHighnessCardID = 55
	TheLichQueenCardID        = 58
	TheQueenOfLightCardID     = 60
	TheVoidTyrantCardID       = 62
)

var (
	ErrLeaderCounterUnavailable = errors.New("role has no leader counter pool")
	ErrLeaderCounterEmpty       = errors.New("leader counter pool is empty")
	ErrLeaderCounterUsed        = errors.New("leader counter already activated this turn")
)

// LeaderCounterDefinition describes the shared counter-pool mechanic for a
// supported Leader identity card.
type LeaderCounterDefinition struct {
	Name        string
	OncePerTurn bool
	HalfPlayers bool
}

// LeaderCounterActivation is a public record of one counter removal. The app
// records only this event and the count change; the table adjudicates effects.
type LeaderCounterActivation struct {
	Number     int    `json:"number"`
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Before     int    `json:"before"`
	After      int    `json:"after"`
}

// LeaderCounterPool is mutable state attached to one identity-card instance.
type LeaderCounterPool struct {
	Name         string                    `json:"name"`
	Initial      int                       `json:"initial"`
	Remaining    int                       `json:"remaining"`
	OncePerTurn  bool                      `json:"once_per_turn"`
	UsedThisTurn bool                      `json:"used_this_turn"`
	Activations  []LeaderCounterActivation `json:"activations,omitempty"`
}

func LeaderCounterDefinitionForCard(cardID int) (LeaderCounterDefinition, bool) {
	switch cardID {
	case HerSeedbornHighnessCardID:
		return LeaderCounterDefinition{Name: "seed", OncePerTurn: true}, true
	case TheLichQueenCardID:
		return LeaderCounterDefinition{Name: "grave", OncePerTurn: true}, true
	case TheQueenOfLightCardID:
		return LeaderCounterDefinition{Name: "purification", HalfPlayers: true}, true
	case TheVoidTyrantCardID:
		return LeaderCounterDefinition{Name: "void", HalfPlayers: true}, true
	default:
		return LeaderCounterDefinition{}, false
	}
}

// NewDealtRoleCard creates the identity-card instance owned by a dealt role.
// Counter Leaders are cloned and initialized here so mutable state never lands
// on CardService's shared catalog card.
func NewDealtRoleCard(card *Card, playerCount int) *Card {
	if card == nil {
		return nil
	}
	definition, ok := LeaderCounterDefinitionForCard(card.ID)
	if !ok {
		return card
	}

	dealt := *card
	initial := playerCount
	if definition.HalfPlayers {
		initial = playerCount / 2
	}
	dealt.LeaderCounterPool = &LeaderCounterPool{
		Name:        definition.Name,
		Initial:     initial,
		Remaining:   initial,
		OncePerTurn: definition.OncePerTurn,
	}
	return &dealt
}

// ActivateLeaderCounter removes one public counter from the identity card.
func (r *Room) ActivateLeaderCounter(playerID string) (*LeaderCounterActivation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Players[playerID]
	if player == nil || player.Role == nil || player.Role.LeaderCounterPool == nil {
		return nil, ErrLeaderCounterUnavailable
	}
	pool := player.Role.LeaderCounterPool
	if pool.Remaining == 0 {
		return nil, ErrLeaderCounterEmpty
	}
	if pool.OncePerTurn && pool.UsedThisTurn {
		return nil, ErrLeaderCounterUsed
	}

	activation := LeaderCounterActivation{
		Number:     len(pool.Activations) + 1,
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Before:     pool.Remaining,
		After:      pool.Remaining - 1,
	}
	pool.Remaining = activation.After
	pool.UsedThisTurn = pool.OncePerTurn
	pool.Activations = append(pool.Activations, activation)
	return &pool.Activations[len(pool.Activations)-1], nil
}

// ResetLeaderCounterTurn is the table-adjudicated turn boundary for Leaders
// with a once-per-turn limit. Treacherest has no turn engine, so the owner
// confirms this reset when the table advances to their next turn.
func (r *Room) ResetLeaderCounterTurn(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Players[playerID]
	if player == nil || player.Role == nil || player.Role.LeaderCounterPool == nil {
		return ErrLeaderCounterUnavailable
	}
	pool := player.Role.LeaderCounterPool
	if !pool.OncePerTurn {
		return fmt.Errorf("%w: counter has no per-turn limit", ErrLeaderCounterUnavailable)
	}
	pool.UsedThisTurn = false
	return nil
}
