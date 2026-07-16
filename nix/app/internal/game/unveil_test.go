package game

import "testing"

func TestHasOtherRevealedIdentity(t *testing.T) {
	room := &Room{Players: make(map[string]*Player)}
	owner := &Player{ID: "owner", Role: &Card{ID: 17}, RoleRevealed: true}
	other := &Player{ID: "other", Role: &Card{ID: 2}}
	room.Players[owner.ID] = owner
	room.Players[other.ID] = other

	if HasOtherRevealedIdentity(room, owner) {
		t.Fatal("owner's own revealed state must not satisfy the Undercover condition")
	}

	other.RoleRevealed = true
	if !HasOtherRevealedIdentity(room, owner) {
		t.Fatal("another identity's public reveal should satisfy the observable Undercover condition")
	}
}
