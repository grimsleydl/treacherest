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
	// Use legacy role distribution
	AssignRolesLegacy(players, cardService)
}

// AssignRolesWithConfig assigns roles to players using the room's role configuration
func AssignRolesWithConfig(players []*Player, cardService *CardService, roleConfig *RoleConfiguration, roleService *RoleConfigService) {
	// Filter out hosts from role assignment
	activePlayers := make([]*Player, 0, len(players))
	for _, p := range players {
		if !p.IsHost {
			activePlayers = append(activePlayers, p)
		}
	}

	count := len(activePlayers)
	if count == 0 {
		return // No active players to assign roles to
	}

	// Shuffle players first
	shuffled := make([]*Player, count)
	copy(shuffled, activePlayers)
	rand.Shuffle(count, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Get role distribution for the actual player count
	var roleDistribution map[RoleType]int
	if roleService != nil {
		dist, err := roleService.GetDistributionForPlayerCount(roleConfig, count)
		if err == nil {
			roleDistribution = dist
		}
	}
	
	// If no distribution yet (roleService is nil or error), use counts from config directly
	if roleDistribution == nil && roleConfig != nil {
		roleDistribution = make(map[RoleType]int)
		for roleTypeName, typeConfig := range roleConfig.RoleTypes {
			if typeConfig.Count > 0 {
				roleDistribution[RoleType(roleTypeName)] = typeConfig.Count
			}
		}
	}

	// Assign cards based on role distribution
	playerIndex := 0
	usedCards := make(map[*Card]bool)
	
	// Map for getting cards by type
	categoryToCards := map[RoleType][]*Card{
		RoleLeader:   cardService.Leaders,
		RoleGuardian: cardService.Guardians,
		RoleAssassin: cardService.Assassins,
		RoleTraitor:  cardService.Traitors,
	}

	for roleType, neededCount := range roleDistribution {
		if neededCount == 0 {
			continue
		}

		// Get the category name for this role type
		categoryName := string(roleType)
		
		// Get enabled cards for this type from config
		var enabledCardNames map[string]bool
		if typeConfig, exists := roleConfig.RoleTypes[categoryName]; exists {
			enabledCardNames = typeConfig.EnabledCards
		}

		// Filter cards to only include enabled ones
		availableCards := make([]*Card, 0)
		for _, card := range categoryToCards[roleType] {
			if enabledCardNames == nil || enabledCardNames[card.Name] {
				availableCards = append(availableCards, card)
			}
		}

		// Shuffle available cards
		shuffledCards := make([]*Card, len(availableCards))
		copy(shuffledCards, availableCards)
		rand.Shuffle(len(shuffledCards), func(i, j int) {
			shuffledCards[i], shuffledCards[j] = shuffledCards[j], shuffledCards[i]
		})

		// Assign cards to players
		cardsAssigned := 0
		for _, card := range shuffledCards {
			if playerIndex >= len(shuffled) || cardsAssigned >= neededCount {
				break
			}

			// Skip if card already used
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
			cardsAssigned++
		}
	}
}

// AssignRolesLegacy uses the old hardcoded role distribution
func AssignRolesLegacy(players []*Player, cardService *CardService) {
	// Filter out hosts from role assignment
	activePlayers := make([]*Player, 0, len(players))
	for _, p := range players {
		if !p.IsHost {
			activePlayers = append(activePlayers, p)
		}
	}

	count := len(activePlayers)
	if count == 0 {
		return // No active players to assign roles to
	}

	// Shuffle players first
	shuffled := make([]*Player, count)
	copy(shuffled, activePlayers)
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
