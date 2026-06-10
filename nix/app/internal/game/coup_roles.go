package game

import (
	"fmt"
	"math/rand"
)

const (
	RoleKing          RoleType = "King"
	RoleBlueKnight    RoleType = "Blue Knight"
	RoleBlackKnight   RoleType = "Black Knight"
	RoleRedKnight     RoleType = "Red Knight"
	RoleGreenKnight   RoleType = "Green Knight"
	RoleWasteland     RoleType = "Wasteland Knight"
	coupFivePlayerSet          = 5
)

var coupFivePlayerRoles = []*Card{
	coupRoleCard(1001, string(RoleKing), "Starts revealed. Political center of the table.", "Win if alive when Black, Red, and Wasteland threats are eliminated."),
	coupRoleCard(1002, string(RoleBlueKnight), "Protects the King. Can call Inquisition.", "Win with the King. Lose when the King loses."),
	coupRoleCard(1003, string(RoleBlackKnight), "Assassin hired to kill the King, then betray Red.", "Win if the King is dead, at least one Black Knight survives, and Red is dead."),
	coupRoleCard(1004, string(RoleRedKnight), "Usurper who hired Black to kill the King.", "Win if the King is dead, Red survives, and all Black Knights are dead."),
	coupRoleCard(1005, string(RoleGreenKnight), "Opportunist with conditional shared victories.", "Win with King or Red only when eligible under the selected Green rules."),
}

func coupRoleCard(id int, name, text, winCondition string) *Card {
	return &Card{
		ID:     id,
		Name:   name,
		Type:   "Coup Role",
		Rarity: "Coup",
		Types: CardTypes{
			Supertype: "Coup",
			Subtype:   name,
		},
		Text:    text,
		Rulings: []string{"Win Condition: " + winCondition},
	}
}

// AssignCoupRoles assigns the initial five-player Coup role set.
func AssignCoupRoles(players []*Player) error {
	activePlayers := make([]*Player, 0, len(players))
	for _, player := range players {
		if !player.IsHost {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) != coupFivePlayerSet {
		return fmt.Errorf("Coup currently requires exactly 5 active players, got %d", len(activePlayers))
	}

	shuffled := make([]*Player, len(activePlayers))
	copy(shuffled, activePlayers)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i, player := range shuffled {
		role := *coupFivePlayerRoles[i]
		player.Role = &role
		if role.GetRoleType() == RoleKing {
			player.RoleRevealed = true
			player.FaceUp = true
		} else {
			player.RoleRevealed = false
			player.FaceUp = false
		}
	}

	return nil
}
