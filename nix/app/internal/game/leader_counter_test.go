package game

import (
	"errors"
	"fmt"
	"testing"
)

var leaderCounterTestCards = []struct {
	id          int
	name        string
	counterName string
	oncePerTurn bool
}{
	{HerSeedbornHighnessCardID, "Her Seedborn Highness", "seed", true},
	{TheLichQueenCardID, "The Lich Queen", "grave", true},
	{TheQueenOfLightCardID, "The Queen of Light", "purification", false},
	{TheVoidTyrantCardID, "The Void Tyrant", "void", false},
}

func TestLeaderCounterPoolsInitializeAtDealTime(t *testing.T) {
	for _, role := range leaderCounterTestCards {
		for _, playerCount := range []int{3, 6} {
			t.Run(fmt.Sprintf("%s/%d_players", role.name, playerCount), func(t *testing.T) {
				catalogCard := testLeaderCounterCard(role.id, role.name)
				service := &CardService{Leaders: []*Card{catalogCard}}
				players := make([]*Player, playerCount)
				for i := range players {
					players[i] = NewPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player %d", i), fmt.Sprintf("s%d", i))
				}
				config := &RoleConfiguration{
					PresetName: "custom",
					RoleTypes: map[string]*RoleTypeConfig{
						string(RoleLeader): {Count: 1, EnabledCards: map[string]bool{role.name: true}},
					},
				}

				AssignRolesWithConfig(players, service, config, nil)

				var dealt *Card
				for _, player := range players {
					if player.Role != nil {
						dealt = player.Role
						break
					}
				}
				if dealt == nil || dealt.LeaderCounterPool == nil {
					t.Fatal("dealt counter Leader did not receive an identity-card counter pool")
				}
				if dealt == catalogCard {
					t.Fatal("mutable counter state was attached to the shared catalog card")
				}
				want := playerCount
				if role.id == TheQueenOfLightCardID || role.id == TheVoidTyrantCardID {
					want = playerCount / 2
				}
				if got := dealt.LeaderCounterPool.Remaining; got != want {
					t.Fatalf("initial %s counters = %d, want %d", role.counterName, got, want)
				}
				if catalogCard.LeaderCounterPool != nil {
					t.Fatal("deal mutated the shared catalog card")
				}
			})
		}
	}
}

func TestCounterPoolLeadersAreSupported(t *testing.T) {
	for _, role := range leaderCounterTestCards {
		if !IsCardSupported(role.id) {
			t.Errorf("counter Leader %d (%s) remains excluded from deals", role.id, role.name)
		}
	}
	for _, cardID := range []int{53, 54} {
		if IsCardSupported(cardID) {
			t.Errorf("unimplemented card %d unexpectedly became dealable", cardID)
		}
	}
}

func TestLeaderCounterPoolsDecrementAndBlockAtZero(t *testing.T) {
	for _, role := range leaderCounterTestCards {
		t.Run(role.name, func(t *testing.T) {
			room, owner := testLeaderCounterRoom(role.id, role.name, 3)
			initial := owner.Role.LeaderCounterPool.Remaining

			for i := 0; i < initial; i++ {
				activation, err := room.ActivateLeaderCounter(owner.ID)
				if err != nil {
					t.Fatalf("activation %d failed: %v", i+1, err)
				}
				if activation.Before-1 != activation.After {
					t.Fatalf("activation did not decrement exactly once: %+v", activation)
				}
				if role.oncePerTurn && i < initial-1 {
					if err := room.ResetLeaderCounterTurn(owner.ID); err != nil {
						t.Fatalf("reset between table-adjudicated turns failed: %v", err)
					}
				}
			}

			if got := owner.Role.LeaderCounterPool.Remaining; got != 0 {
				t.Fatalf("remaining counters = %d, want 0", got)
			}
			if _, err := room.ActivateLeaderCounter(owner.ID); !errors.Is(err, ErrLeaderCounterEmpty) {
				t.Fatalf("activation at zero error = %v, want %v", err, ErrLeaderCounterEmpty)
			}
			if got := len(owner.Role.LeaderCounterPool.Activations); got != initial {
				t.Fatalf("public activation records = %d, want %d", got, initial)
			}
		})
	}
}

func TestLeaderCounterOncePerTurnState(t *testing.T) {
	for _, role := range leaderCounterTestCards[:2] {
		t.Run(role.name, func(t *testing.T) {
			room, owner := testLeaderCounterRoom(role.id, role.name, 6)
			if _, err := room.ActivateLeaderCounter(owner.ID); err != nil {
				t.Fatalf("first activation failed: %v", err)
			}
			if !owner.Role.LeaderCounterPool.UsedThisTurn {
				t.Fatal("once-per-turn state was not surfaced after activation")
			}
			if _, err := room.ActivateLeaderCounter(owner.ID); !errors.Is(err, ErrLeaderCounterUsed) {
				t.Fatalf("second same-turn activation error = %v, want %v", err, ErrLeaderCounterUsed)
			}
			if err := room.ResetLeaderCounterTurn(owner.ID); err != nil {
				t.Fatalf("owner-confirmed next-turn reset failed: %v", err)
			}
			if owner.Role.LeaderCounterPool.UsedThisTurn {
				t.Fatal("once-per-turn state remained used after reset")
			}
			if _, err := room.ActivateLeaderCounter(owner.ID); err != nil {
				t.Fatalf("activation after next-turn reset failed: %v", err)
			}
		})
	}
}

func TestUnlimitedLeaderCountersHaveNoPerTurnGate(t *testing.T) {
	for _, role := range leaderCounterTestCards[2:] {
		t.Run(role.name, func(t *testing.T) {
			room, owner := testLeaderCounterRoom(role.id, role.name, 6)
			for i := 0; i < 2; i++ {
				if _, err := room.ActivateLeaderCounter(owner.ID); err != nil {
					t.Fatalf("activation %d failed: %v", i+1, err)
				}
			}
			if owner.Role.LeaderCounterPool.UsedThisTurn {
				t.Fatal("unlimited counter pool gained once-per-turn state")
			}
		})
	}
}

func TestLeaderCounterStateMovesWithIdentityCard(t *testing.T) {
	assertMovedState := func(t *testing.T, destination *Player, card *Card, pool *LeaderCounterPool) {
		t.Helper()
		if destination.Role != card {
			t.Fatal("destination did not receive the same identity-card instance")
		}
		if destination.Role.LeaderCounterPool != pool || pool.Remaining != 5 || !pool.UsedThisTurn {
			t.Fatalf("card-attached state did not survive movement: %+v", destination.Role.LeaderCounterPool)
		}
	}
	newPlayers := func() (*Room, *Player, *Player, *Card, *LeaderCounterPool) {
		room, owner := testLeaderCounterRoom(HerSeedbornHighnessCardID, "Her Seedborn Highness", 6)
		target := NewPlayer("target", "Target", "target-session")
		target.Role = &Card{ID: 1, Name: "Other", Types: CardTypes{Subtype: "Guardian"}}
		room.Players[target.ID] = target
		_, _ = room.ActivateLeaderCounter(owner.ID)
		return room, owner, target, owner.Role, owner.Role.LeaderCounterPool
	}

	t.Run("face transitions", func(t *testing.T) {
		_, owner, _, card, pool := newPlayers()
		owner.FaceUp = false
		owner.FaceUp = true
		assertMovedState(t, owner, card, pool)
	})
	t.Run("TransferRole", func(t *testing.T) {
		room, owner, target, card, pool := newPlayers()
		if err := room.TransferRole(owner, target, true); err != nil {
			t.Fatal(err)
		}
		assertMovedState(t, target, card, pool)
	})
	t.Run("SwapRoles", func(t *testing.T) {
		room, owner, target, card, pool := newPlayers()
		if err := room.SwapRoles(owner, target, true); err != nil {
			t.Fatal(err)
		}
		assertMovedState(t, target, card, pool)
	})
	t.Run("StealRole", func(t *testing.T) {
		room, owner, target, card, pool := newPlayers()
		owner.IsEliminated = true
		if err := room.StealRole(target, owner, true); err != nil {
			t.Fatal(err)
		}
		assertMovedState(t, target, card, pool)
	})
	t.Run("Puppet Master redistribution", func(t *testing.T) {
		room, owner, target, card, pool := newPlayers()
		otherCard := target.Role
		if err := room.RedistributeRoles(map[string]*Card{
			owner.ID:  otherCard,
			target.ID: card,
		}); err != nil {
			t.Fatal(err)
		}
		assertMovedState(t, target, card, pool)
	})
}

func testLeaderCounterRoom(cardID int, name string, playerCount int) (*Room, *Player) {
	owner := NewPlayer("owner", "Alice", "owner-session")
	owner.Role = NewDealtRoleCard(testLeaderCounterCard(cardID, name), playerCount)
	owner.RoleRevealed = true
	owner.FaceUp = true
	room := &Room{
		Code:    "COUNT",
		State:   StatePlaying,
		Players: map[string]*Player{owner.ID: owner},
	}
	return room, owner
}

func testLeaderCounterCard(cardID int, name string) *Card {
	return &Card{ID: cardID, Name: name, Types: CardTypes{Subtype: "Leader"}}
}
