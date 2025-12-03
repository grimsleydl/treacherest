package game

import (
	"fmt"
)

// TransferRole moves a role from one player to another
// The 'from' player will have their role set to nil
// If turnFaceDown is true, the transferred role will be face down (unless it's a Leader)
func (r *Room) TransferRole(from, to *Player, turnFaceDown bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if from == nil {
		return fmt.Errorf("source player cannot be nil")
	}
	if to == nil {
		return fmt.Errorf("target player cannot be nil")
	}
	if from.Role == nil {
		return fmt.Errorf("source player %s has no role to transfer", from.Name)
	}

	// Transfer the role
	transferredRole := from.Role
	from.Role = nil
	from.FaceUp = false
	from.RoleRevealed = false

	to.Role = transferredRole

	// Determine face up/down state
	// Leaders are always face up, other roles follow turnFaceDown parameter
	if transferredRole.GetRoleType() == RoleLeader {
		to.FaceUp = true
		to.RoleRevealed = true
	} else if turnFaceDown {
		to.FaceUp = false
		to.RoleRevealed = false
	} else {
		// Keep the current revealed state
		to.RoleRevealed = true
	}

	return nil
}

// SwapRoles exchanges roles between two players
// If turnFaceDown is true, both roles will be face down after the swap (unless a Leader)
func (r *Room) SwapRoles(player1, player2 *Player, turnFaceDown bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if player1 == nil || player2 == nil {
		return fmt.Errorf("both players must be provided")
	}

	// Store original roles
	role1 := player1.Role
	role2 := player2.Role

	// Swap the roles
	player1.Role = role2
	player2.Role = role1

	// Set face states for player1
	if role2 != nil {
		if role2.GetRoleType() == RoleLeader {
			player1.FaceUp = true
			player1.RoleRevealed = true
		} else if turnFaceDown {
			player1.FaceUp = false
			player1.RoleRevealed = false
		}
	} else {
		player1.FaceUp = false
		player1.RoleRevealed = false
	}

	// Set face states for player2
	if role1 != nil {
		if role1.GetRoleType() == RoleLeader {
			player2.FaceUp = true
			player2.RoleRevealed = true
		} else if turnFaceDown {
			player2.FaceUp = false
			player2.RoleRevealed = false
		}
	} else {
		player2.FaceUp = false
		player2.RoleRevealed = false
	}

	return nil
}

// StealRole removes a role from an eliminated player and gives it to the stealer
// The stealer's current role is removed from the game (not transferred)
// If turnFaceDown is true, the stolen role will be face down (unless it's a Leader)
func (r *Room) StealRole(stealer, victim *Player, turnFaceDown bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if stealer == nil {
		return fmt.Errorf("stealer cannot be nil")
	}
	if victim == nil {
		return fmt.Errorf("victim cannot be nil")
	}
	if !victim.IsEliminated {
		return fmt.Errorf("can only steal roles from eliminated players")
	}
	if victim.Role == nil {
		return fmt.Errorf("victim %s has no role to steal", victim.Name)
	}

	// Store the stolen role
	stolenRole := victim.Role

	// Remove roles from both players
	victim.Role = nil
	victim.FaceUp = false
	victim.RoleRevealed = false

	// The stealer's current role is removed from the game (Metamorph text: "remove The Metamorph from the game")
	// We don't transfer it anywhere, it's just gone
	stealer.Role = stolenRole

	// Determine face state for the stolen role
	if stolenRole.GetRoleType() == RoleLeader {
		stealer.FaceUp = true
		stealer.RoleRevealed = true
	} else if turnFaceDown {
		stealer.FaceUp = false
		stealer.RoleRevealed = false
	} else {
		stealer.FaceUp = true
		stealer.RoleRevealed = true
	}

	return nil
}
