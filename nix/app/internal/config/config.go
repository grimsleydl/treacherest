package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ServerConfig represents the server configuration
type ServerConfig struct {
	Server ServerSettings `yaml:"server"`
	Roles  RolesConfig    `yaml:"roles"`
}

// ServerSettings contains server-wide settings
type ServerSettings struct {
	MaxPlayersPerRoom int           `yaml:"maxPlayersPerRoom"`
	MinPlayersPerRoom int           `yaml:"minPlayersPerRoom"`
	RoomCodeLength    int           `yaml:"roomCodeLength"`
	RoomTimeout       time.Duration `yaml:"roomTimeout"`
}

// RolesConfig contains role definitions and presets
type RolesConfig struct {
	Available map[string]RoleDefinition `yaml:"available"`
	Presets   map[string]Preset         `yaml:"presets"`
}

// RoleDefinition defines a single role type
type RoleDefinition struct {
	DisplayName    string `yaml:"displayName"`
	Category       string `yaml:"category"`
	MinCount       int    `yaml:"minCount"`
	MaxCount       int    `yaml:"maxCount"`
	AlwaysRevealed bool   `yaml:"alwaysRevealed"`
}

// Preset defines a named role distribution preset
type Preset struct {
	Name          string                       `yaml:"name"`
	Description   string                       `yaml:"description"`
	Distributions map[int]map[string]int       `yaml:"distributions"`
}

// LoadConfig loads the server configuration from a YAML file
func LoadConfig(path string) (*ServerConfig, error) {
	// If no path provided, use default
	if path == "" {
		path = "config/server.yaml"
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		Server: ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
			RoomCodeLength:    5,
			RoomTimeout:       24 * time.Hour,
		},
		Roles: RolesConfig{
			Available: map[string]RoleDefinition{
				"leader": {
					DisplayName:    "Leader",
					Category:       "Leader",
					MinCount:       1,
					MaxCount:       1,
					AlwaysRevealed: true,
				},
				"guardian": {
					DisplayName: "Guardian",
					Category:    "Guardian",
					MinCount:    0,
					MaxCount:    10,
				},
				"assassin": {
					DisplayName: "Assassin",
					Category:    "Assassin",
					MinCount:    0,
					MaxCount:    10,
				},
				"traitor": {
					DisplayName: "Traitor",
					Category:    "Traitor",
					MinCount:    0,
					MaxCount:    10,
				},
			},
			Presets: map[string]Preset{
				"standard": {
					Name:        "Standard",
					Description: "Balanced gameplay",
					Distributions: map[int]map[string]int{
						1: {"leader": 1},
						2: {"leader": 1, "traitor": 1},
						3: {"leader": 1, "guardian": 1, "traitor": 1},
						4: {"leader": 1, "guardian": 2, "traitor": 1},
						5: {"leader": 1, "guardian": 2, "assassin": 1, "traitor": 1},
						6: {"leader": 1, "guardian": 2, "assassin": 2, "traitor": 1},
						7: {"leader": 1, "guardian": 3, "assassin": 2, "traitor": 1},
						8: {"leader": 1, "guardian": 3, "assassin": 2, "traitor": 2},
					},
				},
			},
		},
	}
}

// Validate checks if the configuration is valid
func (c *ServerConfig) Validate() error {
	if c.Server.MaxPlayersPerRoom < 1 {
		return fmt.Errorf("maxPlayersPerRoom must be at least 1")
	}
	if c.Server.MinPlayersPerRoom < 1 {
		return fmt.Errorf("minPlayersPerRoom must be at least 1")
	}
	if c.Server.MinPlayersPerRoom > c.Server.MaxPlayersPerRoom {
		return fmt.Errorf("minPlayersPerRoom cannot be greater than maxPlayersPerRoom")
	}
	if c.Server.RoomCodeLength < 3 {
		return fmt.Errorf("roomCodeLength must be at least 3")
	}

	// Validate roles
	hasLeader := false
	for name, role := range c.Roles.Available {
		if role.MinCount > role.MaxCount {
			return fmt.Errorf("role %s: minCount cannot be greater than maxCount", name)
		}
		if role.Category == "Leader" {
			hasLeader = true
		}
	}
	if !hasLeader {
		return fmt.Errorf("at least one Leader role must be defined")
	}

	// Validate presets
	for presetName, preset := range c.Roles.Presets {
		for playerCount, distribution := range preset.Distributions {
			if playerCount < 1 || playerCount > c.Server.MaxPlayersPerRoom {
				return fmt.Errorf("preset %s: invalid player count %d", presetName, playerCount)
			}
			
			// Check that all roles in distribution exist
			for roleName := range distribution {
				if _, exists := c.Roles.Available[roleName]; !exists {
					return fmt.Errorf("preset %s: unknown role %s", presetName, roleName)
				}
			}
		}
	}

	return nil
}

// GetPreset returns a preset by name
func (c *ServerConfig) GetPreset(name string) (*Preset, bool) {
	preset, exists := c.Roles.Presets[name]
	if !exists {
		return nil, false
	}
	return &preset, true
}

// GetRoleDefinition returns a role definition by name
func (c *ServerConfig) GetRoleDefinition(name string) (*RoleDefinition, bool) {
	role, exists := c.Roles.Available[name]
	if !exists {
		return nil, false
	}
	return &role, true
}