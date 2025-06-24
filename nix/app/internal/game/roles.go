package game

import (
	"math/rand"
)

// RoleType represents the type of role
type RoleType string

const (
	RoleLeader   RoleType = "Leader"
	RoleGuardian RoleType = "Guardian"
	RoleAssassin RoleType = "Assassin"
	RoleTraitor  RoleType = "Traitor"
)

// AssignRoles assigns roles to players based on player count using cards from CardService
func AssignRoles(players []*Player, cardService *CardService) {
	count := len(players)

	// Shuffle players first
	shuffled := make([]*Player, count)
	copy(shuffled, players)
	rand.Shuffle(count, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Get role distribution based on player count
	roleDistribution := getRoleDistribution(count)

	// Track used cards to prevent duplicates
	usedCards := make(map[*Card]bool)

	// Assign cards based on role distribution
	playerIndex := 0
	for roleType, count := range roleDistribution {
		// Get random cards for this role type
		cards := cardService.GetRandomCards(roleType, count)

		// Assign cards to players
		for _, card := range cards {
			if playerIndex >= len(shuffled) {
				break
			}

			// Skip if card already used (shouldn't happen with GetRandomCards, but be safe)
			if usedCards[card] {
				continue
			}

			shuffled[playerIndex].Role = card
			usedCards[card] = true

			// Leader is always revealed
			if card.GetRoleType() == RoleLeader {
				shuffled[playerIndex].RoleRevealed = true
			}

			playerIndex++
		}
	}
}

// getRoleDistribution returns the role distribution based on player count
func getRoleDistribution(playerCount int) map[RoleType]int {
	switch playerCount {
	case 1:
		return map[RoleType]int{
			RoleLeader: 1,
		}
	case 2:
		return map[RoleType]int{
			RoleLeader:  1,
			RoleTraitor: 1,
		}
	case 3:
		return map[RoleType]int{
			RoleLeader:   1,
			RoleGuardian: 1,
			RoleTraitor:  1,
		}
	case 4:
		return map[RoleType]int{
			RoleLeader:   1,
			RoleGuardian: 2,
			RoleTraitor:  1,
		}
	case 5:
		return map[RoleType]int{
			RoleLeader:   1,
			RoleGuardian: 2,
			RoleAssassin: 1,
			RoleTraitor:  1,
		}
	case 6:
		return map[RoleType]int{
			RoleLeader:   1,
			RoleGuardian: 2,
			RoleAssassin: 2,
			RoleTraitor:  1,
		}
	case 7:
		return map[RoleType]int{
			RoleLeader:   1,
			RoleGuardian: 3,
			RoleAssassin: 2,
			RoleTraitor:  1,
		}
	case 8:
		return map[RoleType]int{
			RoleLeader:   1,
			RoleGuardian: 3,
			RoleAssassin: 2,
			RoleTraitor:  2,
		}
	default:
		// Fallback for edge cases - just assign Leader and Guardians
		distribution := map[RoleType]int{
			RoleLeader: 1,
		}
		if playerCount > 1 {
			distribution[RoleGuardian] = playerCount - 1
		}
		return distribution
	}
}
