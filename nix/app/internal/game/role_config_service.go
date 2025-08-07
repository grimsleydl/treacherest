package game

import (
	"fmt"
	"sort"
	"treacherest/internal/config"
)

// RoleConfigService handles role configuration logic
type RoleConfigService struct {
	config      *config.ServerConfig
	cardService *CardService
}

// NewRoleConfigService creates a new role configuration service
func NewRoleConfigService(cfg *config.ServerConfig) *RoleConfigService {
	return &RoleConfigService{
		config: cfg,
	}
}

// SetCardService sets the card service for the role config service
func (s *RoleConfigService) SetCardService(cs *CardService) {
	s.cardService = cs
}

// CreateFromPreset creates a RoleConfiguration from a preset name
func (s *RoleConfigService) CreateFromPreset(presetName string, maxPlayers int) (*RoleConfiguration, error) {
	preset, exists := s.config.GetPreset(presetName)
	if !exists {
		return nil, fmt.Errorf("preset '%s' not found", presetName)
	}

	// Find the appropriate distribution for all player counts
	minPlayers := maxPlayers
	for playerCount := range preset.Distributions {
		if playerCount < minPlayers {
			minPlayers = playerCount
		}
	}

	// Create role configuration with new structure
	roleConfig := &RoleConfiguration{
		PresetName: presetName,
		MinPlayers: minPlayers,
		MaxPlayers: maxPlayers,
		RoleTypes:  make(map[string]*RoleTypeConfig),
	}

	// Initialize all role types with all cards enabled
	for _, roleDef := range s.config.Roles.Available {
		if roleConfig.RoleTypes[roleDef.Category] == nil {
			roleConfig.RoleTypes[roleDef.Category] = &RoleTypeConfig{
				Count:        0,
				EnabledCards: make(map[string]bool),
			}
		}
	}

	// Enable all cards for each type
	if s.cardService != nil {
		for _, card := range s.cardService.Leaders {
			if roleConfig.RoleTypes["Leader"] != nil {
				roleConfig.RoleTypes["Leader"].EnabledCards[card.Name] = true
			}
		}
		for _, card := range s.cardService.Guardians {
			if roleConfig.RoleTypes["Guardian"] != nil {
				roleConfig.RoleTypes["Guardian"].EnabledCards[card.Name] = true
			}
		}
		for _, card := range s.cardService.Assassins {
			if roleConfig.RoleTypes["Assassin"] != nil {
				roleConfig.RoleTypes["Assassin"].EnabledCards[card.Name] = true
			}
		}
		for _, card := range s.cardService.Traitors {
			if roleConfig.RoleTypes["Traitor"] != nil {
				roleConfig.RoleTypes["Traitor"].EnabledCards[card.Name] = true
			}
		}
	}

	// Set counts based on the preset's closest distribution
	if dist, exists := preset.Distributions[maxPlayers]; exists {
		for role, count := range dist {
			if roleDef, ok := s.config.Roles.Available[role]; ok {
				if roleConfig.RoleTypes[roleDef.Category] != nil {
					roleConfig.RoleTypes[roleDef.Category].Count = count
				}
			}
		}
	}

	return roleConfig, nil
}

// CreateDefaultConfiguration creates a new role configuration with all cards enabled
func (s *RoleConfigService) CreateDefaultConfiguration() *RoleConfiguration {
	roleConfig := &RoleConfiguration{
		PresetName: "custom",
		MinPlayers: s.config.Server.MinPlayersPerRoom,
		MaxPlayers: s.config.Server.MaxPlayersPerRoom,
		RoleTypes:  make(map[string]*RoleTypeConfig),
	}

	// Initialize all role types with all cards enabled
	for _, roleDef := range s.config.Roles.Available {
		if roleConfig.RoleTypes[roleDef.Category] == nil {
			roleConfig.RoleTypes[roleDef.Category] = &RoleTypeConfig{
				Count:        0,
				EnabledCards: make(map[string]bool),
			}
		}
	}

	// Enable all cards for each type
	if s.cardService != nil {
		for _, card := range s.cardService.Leaders {
			if roleConfig.RoleTypes["Leader"] != nil {
				roleConfig.RoleTypes["Leader"].EnabledCards[card.Name] = true
			}
		}
		for _, card := range s.cardService.Guardians {
			if roleConfig.RoleTypes["Guardian"] != nil {
				roleConfig.RoleTypes["Guardian"].EnabledCards[card.Name] = true
			}
		}
		for _, card := range s.cardService.Assassins {
			if roleConfig.RoleTypes["Assassin"] != nil {
				roleConfig.RoleTypes["Assassin"].EnabledCards[card.Name] = true
			}
		}
		for _, card := range s.cardService.Traitors {
			if roleConfig.RoleTypes["Traitor"] != nil {
				roleConfig.RoleTypes["Traitor"].EnabledCards[card.Name] = true
			}
		}
	}

	// Set default leader count
	if roleConfig.RoleTypes["Leader"] != nil {
		roleConfig.RoleTypes["Leader"].Count = 1
	}

	return roleConfig
}

// GetDistributionForPlayerCount returns the role distribution for a specific player count
func (s *RoleConfigService) GetDistributionForPlayerCount(config *RoleConfiguration, playerCount int) (map[RoleType]int, error) {
	// If using a preset, get the exact distribution
	if config.PresetName != "custom" {
		preset, exists := s.config.GetPreset(config.PresetName)
		if !exists {
			return nil, fmt.Errorf("preset '%s' not found", config.PresetName)
		}

		// Look for exact player count match
		if dist, ok := preset.Distributions[playerCount]; ok {
			// Convert string map to RoleType map
			result := make(map[RoleType]int)
			for role, count := range dist {
				// Map lowercase role names to RoleType constants
				switch role {
				case "leader":
					result[RoleLeader] = count
				case "guardian":
					result[RoleGuardian] = count
				case "assassin":
					result[RoleAssassin] = count
				case "traitor":
					result[RoleTraitor] = count
				}
			}
			return result, nil
		}

		// If no exact match, find closest distribution
		closestCount := 0
		closestDiff := playerCount
		for count := range preset.Distributions {
			diff := abs(playerCount - count)
			if diff < closestDiff {
				closestDiff = diff
				closestCount = count
			}
		}

		if closestCount > 0 {
			dist := preset.Distributions[closestCount]
			result := make(map[RoleType]int)

			// Start with the base distribution
			totalRoles := 0
			for role, count := range dist {
				// Map lowercase role names to RoleType constants
				switch role {
				case "leader":
					result[RoleLeader] = count
				case "guardian":
					result[RoleGuardian] = count
				case "assassin":
					result[RoleAssassin] = count
				case "traitor":
					result[RoleTraitor] = count
				}
				totalRoles += count
			}

			// Adjust for player count difference
			if totalRoles < playerCount {
				// Add more guardians to fill the gap
				result[RoleGuardian] += playerCount - totalRoles
			} else if totalRoles > playerCount {
				// Remove roles starting with non-essential ones
				// This is a simplified approach - in production, you'd want more sophisticated logic
				for totalRoles > playerCount && result[RoleGuardian] > 1 {
					result[RoleGuardian]--
					totalRoles--
				}
			}

			return result, nil
		}
	}

	// For custom configurations, use the counts from RoleTypes
	result := make(map[RoleType]int)
	totalRoles := 0

	// Map category names to RoleType constants
	categoryToRoleType := map[string]RoleType{
		"Leader":   RoleLeader,
		"Guardian": RoleGuardian,
		"Assassin": RoleAssassin,
		"Traitor":  RoleTraitor,
	}

	for category, typeConfig := range config.RoleTypes {
		if roleType, ok := categoryToRoleType[category]; ok && typeConfig.Count > 0 {
			result[roleType] = typeConfig.Count
			totalRoles += typeConfig.Count
		}
	}

	// For custom configurations, return exact counts without adjustment
	// This respects the user's configuration exactly as specified
	if config.PresetName == "custom" {
		// Only validate that we don't exceed player count
		if totalRoles > playerCount {
			return nil, fmt.Errorf("too many roles (%d) for player count (%d)", totalRoles, playerCount)
		}
		// Return exact configured counts, even if less than player count
		return result, nil
	}

	// For presets, we still adjust to match player count
	// Ensure we have at least one leader (unless leaderless games are allowed)
	if result[RoleLeader] == 0 && !config.AllowLeaderlessGame {
		result[RoleLeader] = 1
		totalRoles++
	}

	// Adjust for player count mismatch (only for presets)
	if totalRoles < playerCount {
		// Add more guardians to fill
		result[RoleGuardian] += playerCount - totalRoles
	} else if totalRoles > playerCount {
		// This shouldn't happen with proper validation, but handle it gracefully
		// Reduce counts proportionally
		return nil, fmt.Errorf("too many roles (%d) for player count (%d)", totalRoles, playerCount)
	}

	return result, nil
}

// ValidateConfiguration validates a role configuration
func (s *RoleConfigService) ValidateConfiguration(config *RoleConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Check that we have at least one leader
	hasLeader := false
	totalRoles := 0

	for category, typeConfig := range config.RoleTypes {
		if typeConfig.Count == 0 {
			continue
		}

		// Count enabled cards
		enabledCount := 0
		for _, enabled := range typeConfig.EnabledCards {
			if enabled {
				enabledCount++
			}
		}

		// Validate we have enough cards
		if typeConfig.Count > enabledCount {
			return fmt.Errorf("%s: need %d cards but only %d are enabled", category, typeConfig.Count, enabledCount)
		}

		totalRoles += typeConfig.Count

		if category == "Leader" {
			hasLeader = typeConfig.Count > 0
			if typeConfig.Count > 1 {
				return fmt.Errorf("cannot have more than 1 leader, got %d", typeConfig.Count)
			}
		}
	}

	if !hasLeader && !config.AllowLeaderlessGame {
		return fmt.Errorf("must have a leader role")
	}

	// Validate player bounds
	if config.MinPlayers < s.config.Server.MinPlayersPerRoom {
		return fmt.Errorf("minimum players %d is less than server minimum %d",
			config.MinPlayers, s.config.Server.MinPlayersPerRoom)
	}
	if config.MaxPlayers > s.config.Server.MaxPlayersPerRoom {
		return fmt.Errorf("maximum players %d exceeds server maximum %d",
			config.MaxPlayers, s.config.Server.MaxPlayersPerRoom)
	}

	return nil
}

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// GetSortedRoles returns role definitions in a consistent sorted order
// Order: Leader roles first (always revealed), then by category, then alphabetical
func (s *RoleConfigService) GetSortedRoles() []struct {
	Name       string
	Definition config.RoleDefinition
} {
	var roles []struct {
		Name       string
		Definition config.RoleDefinition
	}

	// Collect all roles
	for name, def := range s.config.Roles.Available {
		roles = append(roles, struct {
			Name       string
			Definition config.RoleDefinition
		}{Name: name, Definition: def})
	}

	// Sort roles: Leader first, then by category, then by display name
	sort.Slice(roles, func(i, j int) bool {
		// Leader always comes first
		if roles[i].Definition.AlwaysRevealed && !roles[j].Definition.AlwaysRevealed {
			return true
		}
		if !roles[i].Definition.AlwaysRevealed && roles[j].Definition.AlwaysRevealed {
			return false
		}

		// Then sort by category
		if roles[i].Definition.Category != roles[j].Definition.Category {
			// Define category order
			categoryOrder := map[string]int{
				"Leader":   1,
				"Good":     2,
				"Guardian": 3,
				"Evil":     4,
				"Traitor":  5,
				"Assassin": 6,
			}

			orderI, hasI := categoryOrder[roles[i].Definition.Category]
			orderJ, hasJ := categoryOrder[roles[j].Definition.Category]

			if hasI && hasJ {
				return orderI < orderJ
			}
			if hasI {
				return true
			}
			if hasJ {
				return false
			}

			// Fallback to alphabetical by category
			return roles[i].Definition.Category < roles[j].Definition.Category
		}

		// Finally sort by display name
		return roles[i].Definition.DisplayName < roles[j].Definition.DisplayName
	})

	return roles
}
