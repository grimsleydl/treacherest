package game

import (
	"log"
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

	// Check for hide role distribution mode
	if roleConfig != nil && roleConfig.HideRoleDistribution {
		handleHiddenDistribution(shuffled, cardService, roleConfig, roleService)
		return
	}

	// Check for fully random roles mode
	if roleConfig != nil && roleConfig.FullyRandomRoles {
		handleFullyRandomDistribution(shuffled, cardService, roleConfig)
		return
	}

	// Get role distribution for the actual player count
	var roleDistribution map[RoleType]int
	if roleConfig != nil && roleConfig.PresetName == "custom" {
		// For custom games, use the exact counts from the configuration
		roleDistribution = make(map[RoleType]int)
		for roleTypeName, typeConfig := range roleConfig.RoleTypes {
			if typeConfig.Count > 0 {
				roleDistribution[RoleType(roleTypeName)] = typeConfig.Count
			}
		}
	} else if roleService != nil && roleConfig != nil {
		// For preset-based games, get the distribution from the service (which can auto-scale)
		dist, err := roleService.GetDistributionForPlayerCount(roleConfig, count)
		if err == nil {
			roleDistribution = dist
		}
	}

	// Fallback if no distribution could be determined
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

	// Create ordered list of role types to ensure consistent assignment order
	// Leaders should always be assigned first when not allowing leaderless games
	roleOrder := []RoleType{RoleLeader, RoleGuardian, RoleAssassin, RoleTraitor}
	
	// If allowing leaderless games, process in any order
	// Otherwise, ensure leaders are assigned first
	for _, roleType := range roleOrder {
		neededCount, exists := roleDistribution[roleType]
		if !exists || neededCount == 0 {
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
		
		// If no available cards for this role type, skip
		if len(availableCards) == 0 {
			continue
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
			if playerIndex >= len(shuffled) {
				// We've assigned roles to all players, stop processing
				goto done
			}
			if cardsAssigned >= neededCount {
				// We've assigned enough of this role type
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
done:
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
	// Use ordered iteration to ensure leaders are assigned first
	roleOrder := []RoleType{RoleLeader, RoleGuardian, RoleAssassin, RoleTraitor}
	playerIndex := 0
	
	for _, roleType := range roleOrder {
		count, exists := roleDistribution[roleType]
		if !exists || count == 0 {
			continue
		}
		
		// Get random cards for this role type
		cards := cardService.GetRandomCards(roleType, count)

		// Assign cards to players
		for _, card := range cards {
			if playerIndex >= len(shuffled) {
				goto doneLegacy
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
doneLegacy:
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

// handleHiddenDistribution randomly selects a preset and applies its distribution
func handleHiddenDistribution(shuffled []*Player, cardService *CardService, roleConfig *RoleConfiguration, roleService *RoleConfigService) {
	// Get available presets
	presets := []string{"standard", "assassination", "guardian"}
	
	// Randomly select a preset
	selectedPreset := presets[rand.Intn(len(presets))]
	log.Printf("🎲 Hidden distribution mode: randomly selected preset '%s' for %d players", selectedPreset, len(shuffled))
	
	// Create a temporary role config with the selected preset
	tempConfig, err := roleService.CreateFromPreset(selectedPreset, len(shuffled))
	if err != nil {
		log.Printf("❌ Failed to create config from preset %s: %v", selectedPreset, err)
		// Fallback to basic distribution
		fallbackDistribution := make(map[RoleType]int)
		if len(shuffled) > 0 {
			fallbackDistribution[RoleLeader] = 1
			if len(shuffled) > 1 {
				fallbackDistribution[RoleGuardian] = len(shuffled) - 1
			}
		}
		assignRolesFromDistribution(shuffled, cardService, fallbackDistribution, roleConfig)
		return
	}
	
	// Build role distribution from the preset
	roleDistribution := make(map[RoleType]int)
	for roleTypeName, typeConfig := range tempConfig.RoleTypes {
		if typeConfig.Count > 0 {
			roleDistribution[RoleType(roleTypeName)] = typeConfig.Count
		}
	}
	
	// Apply the distribution
	assignRolesFromDistribution(shuffled, cardService, roleDistribution, roleConfig)
}

// handleFullyRandomDistribution assigns completely random roles
func handleFullyRandomDistribution(shuffled []*Player, cardService *CardService, roleConfig *RoleConfiguration) {
	count := len(shuffled)
	log.Printf("🎲 Fully random distribution mode for %d players", count)
	
	// Ensure at least 1 leader unless leaderless is allowed
	minLeaders := 0
	if !roleConfig.AllowLeaderlessGame {
		minLeaders = 1
	}
	
	// Build a pool of all available role types
	rolePool := []RoleType{}
	
	// Add required leaders
	for i := 0; i < minLeaders; i++ {
		rolePool = append(rolePool, RoleLeader)
	}
	
	// Calculate remaining slots
	remainingSlots := count - minLeaders
	
	// Define role types and their weights for random selection
	// Weights ensure reasonable distribution
	roleWeights := map[RoleType]int{
		RoleLeader:   1, // Can have more leaders
		RoleGuardian: 3, // More common
		RoleAssassin: 2, // Medium frequency
		RoleTraitor:  1, // Less common
	}
	
	// Build weighted pool
	weightedPool := []RoleType{}
	for role, weight := range roleWeights {
		for i := 0; i < weight; i++ {
			weightedPool = append(weightedPool, role)
		}
	}
	
	// Fill remaining slots randomly
	for i := 0; i < remainingSlots; i++ {
		randomRole := weightedPool[rand.Intn(len(weightedPool))]
		rolePool = append(rolePool, randomRole)
	}
	
	// Shuffle the role pool
	rand.Shuffle(len(rolePool), func(i, j int) {
		rolePool[i], rolePool[j] = rolePool[j], rolePool[i]
	})
	
	// Count distribution for logging
	distribution := make(map[RoleType]int)
	for _, role := range rolePool {
		distribution[role]++
	}
	
	log.Printf("🎲 Generated distribution: Leaders=%d, Guardians=%d, Assassins=%d, Traitors=%d",
		distribution[RoleLeader], distribution[RoleGuardian], distribution[RoleAssassin], distribution[RoleTraitor])
	
	// Apply the distribution
	assignRolesFromDistribution(shuffled, cardService, distribution, roleConfig)
}

// assignRolesFromDistribution is a helper that assigns roles based on a distribution map
func assignRolesFromDistribution(shuffled []*Player, cardService *CardService, roleDistribution map[RoleType]int, roleConfig *RoleConfiguration) {
	// Map role types to card categories
	categoryToCards := map[RoleType][]*Card{
		RoleLeader:   cardService.Leaders,
		RoleGuardian: cardService.Guardians,
		RoleAssassin: cardService.Assassins,
		RoleTraitor:  cardService.Traitors,
	}
	
	// Create ordered list of role types
	roleOrder := []RoleType{RoleLeader, RoleGuardian, RoleAssassin, RoleTraitor}
	
	playerIndex := 0
	for _, roleType := range roleOrder {
		neededCount, exists := roleDistribution[roleType]
		if !exists || neededCount == 0 {
			continue
		}
		
		// Get enabled cards for this role type
		var enabledCardNames map[string]bool
		if roleConfig != nil && roleConfig.RoleTypes != nil {
			if typeConfig, exists := roleConfig.RoleTypes[string(roleType)]; exists {
				enabledCardNames = typeConfig.EnabledCards
			}
		}
		
		// Filter cards to only include enabled ones
		availableCards := make([]*Card, 0)
		for _, card := range categoryToCards[roleType] {
			if enabledCardNames == nil || enabledCardNames[card.Name] {
				availableCards = append(availableCards, card)
			}
		}
		
		// If no available cards for this role type, use all cards
		if len(availableCards) == 0 {
			availableCards = categoryToCards[roleType]
		}
		
		// Shuffle available cards
		shuffledCards := make([]*Card, len(availableCards))
		copy(shuffledCards, availableCards)
		rand.Shuffle(len(shuffledCards), func(i, j int) {
			shuffledCards[i], shuffledCards[j] = shuffledCards[j], shuffledCards[i]
		})
		
		// Assign cards to players
		for i := 0; i < neededCount && playerIndex < len(shuffled); i++ {
			// Use modulo to reuse cards if needed
			card := shuffledCards[i%len(shuffledCards)]
			shuffled[playerIndex].Role = card
			log.Printf("Assigned %s to player %s", card.Name, shuffled[playerIndex].Name)
			playerIndex++
		}
	}
}
