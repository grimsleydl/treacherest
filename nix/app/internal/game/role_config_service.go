package game

import (
	"fmt"
	"treacherest/internal/config"
)

// RoleConfigService handles role configuration logic
type RoleConfigService struct {
	config *config.ServerConfig
}

// NewRoleConfigService creates a new role configuration service
func NewRoleConfigService(cfg *config.ServerConfig) *RoleConfigService {
	return &RoleConfigService{
		config: cfg,
	}
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

	// Create role configuration
	roleConfig := &RoleConfiguration{
		PresetName:   presetName,
		EnabledRoles: make(map[string]bool),
		RoleCounts:   make(map[string]int),
		MinPlayers:   minPlayers,
		MaxPlayers:   maxPlayers,
	}

	// Enable all roles that appear in any distribution
	for _, dist := range preset.Distributions {
		for role := range dist {
			roleConfig.EnabledRoles[role] = true
		}
	}

	return roleConfig, nil
}

// CreateCustomConfiguration creates a custom role configuration
func (s *RoleConfigService) CreateCustomConfiguration(enabledRoles map[string]bool, roleCounts map[string]int) (*RoleConfiguration, error) {
	// Validate that all roles exist
	for role := range enabledRoles {
		if _, exists := s.config.GetRoleDefinition(role); !exists {
			return nil, fmt.Errorf("unknown role: %s", role)
		}
	}

	// Calculate min/max players based on role counts
	minPlayers := 0
	maxPlayers := 0
	
	for role, count := range roleCounts {
		if enabledRoles[role] && count > 0 {
			roleDef, _ := s.config.GetRoleDefinition(role)
			minPlayers += roleDef.MinCount
			maxPlayers += count
		}
	}

	// Ensure we have at least the server minimum
	if minPlayers < s.config.Server.MinPlayersPerRoom {
		minPlayers = s.config.Server.MinPlayersPerRoom
	}

	// Cap at server maximum
	if maxPlayers > s.config.Server.MaxPlayersPerRoom {
		maxPlayers = s.config.Server.MaxPlayersPerRoom
	}

	return &RoleConfiguration{
		PresetName:   "custom",
		EnabledRoles: enabledRoles,
		RoleCounts:   roleCounts,
		MinPlayers:   minPlayers,
		MaxPlayers:   maxPlayers,
	}, nil
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
				result[RoleType(role)] = count
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
				result[RoleType(role)] = count
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

	// For custom configurations, use the specified counts
	result := make(map[RoleType]int)
	totalRoles := 0

	for role, count := range config.RoleCounts {
		if config.EnabledRoles[role] && count > 0 {
			result[RoleType(role)] = count
			totalRoles += count
		}
	}

	// Ensure we have at least one leader
	if result[RoleLeader] == 0 {
		result[RoleLeader] = 1
		totalRoles++
	}

	// Adjust for player count mismatch
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
	totalMinRoles := 0
	totalMaxRoles := 0

	for role, enabled := range config.EnabledRoles {
		if !enabled {
			continue
		}

		count, hasCount := config.RoleCounts[role]
		if !hasCount || count == 0 {
			continue
		}

		roleDef, exists := s.config.GetRoleDefinition(role)
		if !exists {
			return fmt.Errorf("unknown role: %s", role)
		}

		if role == "leader" {
			hasLeader = true
			if count != 1 {
				return fmt.Errorf("must have exactly 1 leader, got %d", count)
			}
		}

		// Validate count bounds
		if count < roleDef.MinCount {
			return fmt.Errorf("role %s: count %d is less than minimum %d", role, count, roleDef.MinCount)
		}
		if count > roleDef.MaxCount {
			return fmt.Errorf("role %s: count %d exceeds maximum %d", role, count, roleDef.MaxCount)
		}

		totalMinRoles += roleDef.MinCount
		totalMaxRoles += count
	}

	if !hasLeader {
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