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
	DefaultGameSize   int           `yaml:"defaultGameSize"`
	RoomCodeLength    int           `yaml:"roomCodeLength"`
	RoomTimeout       time.Duration `yaml:"roomTimeout"`
	
	// Server settings
	Port              string        `yaml:"port" envconfig:"PORT" required:"true"`
	Host              string        `yaml:"host" envconfig:"HOST" required:"true"`
	ReadTimeout       time.Duration `yaml:"readTimeout" envconfig:"READ_TIMEOUT" default:"15s"`
	WriteTimeout      time.Duration `yaml:"writeTimeout" envconfig:"WRITE_TIMEOUT" default:"15s"`
	IdleTimeout       time.Duration `yaml:"idleTimeout" envconfig:"IDLE_TIMEOUT" default:"0s"` // 0 for SSE support
	ShutdownTimeout   time.Duration `yaml:"shutdownTimeout" envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	
	// Rate limiting (using golang.org/x/time/rate)
	RateLimit         float64       `yaml:"rateLimit" envconfig:"RATE_LIMIT" default:"10"`        // requests per second
	RateLimitBurst    int           `yaml:"rateLimitBurst" envconfig:"RATE_LIMIT_BURST" default:"20"` // burst size
	
	// Request limits
	MaxRequestSize    int64         `yaml:"maxRequestSize" envconfig:"MAX_REQUEST_SIZE" default:"1048576"` // 1MB
	MaxSSEConnections int           `yaml:"maxSSEConnections" envconfig:"MAX_SSE_CONNECTIONS" default:"1000"`
	
	// Monitoring
	EnableMetrics     bool          `yaml:"enableMetrics" envconfig:"ENABLE_METRICS" default:"false"`
	MetricsPort       string        `yaml:"metricsPort" envconfig:"METRICS_PORT"` // No default - must be set if metrics enabled
	LogLevel          string        `yaml:"logLevel" envconfig:"LOG_LEVEL" default:"info"`
	LogFormat         string        `yaml:"logFormat" envconfig:"LOG_FORMAT" default:"text"`
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
	Name          string                 `yaml:"name"`
	Description   string                 `yaml:"description"`
	Distributions map[int]map[string]int `yaml:"distributions"`
}

// LoadConfig loads the server configuration from a YAML file and environment variables
func LoadConfig(path string) (*ServerConfig, error) {
	// Start with default config
	config := DefaultConfig()
	
	// If path provided, load from YAML file
	if path == "" {
		path = "config/server.yaml"
	}
	
	// Try to read the file
	data, err := os.ReadFile(path)
	if err == nil {
		// Parse YAML if file exists
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else if !os.IsNotExist(err) {
		// Return error if it's not just a missing file
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Override with environment variables
	loadFromEnv(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *ServerConfig) {
	// Check each environment variable and override if set
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		cfg.Server.Host = host
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Server.LogLevel = logLevel
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		cfg.Server.LogFormat = logFormat
	}
	
	// Parse numeric values
	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		if val, err := fmt.Sscanf(rateLimit, "%f", &cfg.Server.RateLimit); err == nil && val == 1 {
			// Successfully parsed
		}
	}
	if burst := os.Getenv("RATE_LIMIT_BURST"); burst != "" {
		if val, err := fmt.Sscanf(burst, "%d", &cfg.Server.RateLimitBurst); err == nil && val == 1 {
			// Successfully parsed
		}
	}
	if maxReq := os.Getenv("MAX_REQUEST_SIZE"); maxReq != "" {
		if val, err := fmt.Sscanf(maxReq, "%d", &cfg.Server.MaxRequestSize); err == nil && val == 1 {
			// Successfully parsed
		}
	}
	if maxSSE := os.Getenv("MAX_SSE_CONNECTIONS"); maxSSE != "" {
		if val, err := fmt.Sscanf(maxSSE, "%d", &cfg.Server.MaxSSEConnections); err == nil && val == 1 {
			// Successfully parsed
		}
	}
	
	// Parse boolean values
	if metrics := os.Getenv("ENABLE_METRICS"); metrics == "true" {
		cfg.Server.EnableMetrics = true
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		Server: ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
			DefaultGameSize:   5,
			RoomCodeLength:    5,
			RoomTimeout:       24 * time.Hour,
			
			// Server defaults
			Port:              "", // Must be set via env
			Host:              "", // Must be set via env
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       0, // 0 for SSE support
			ShutdownTimeout:   30 * time.Second,
			
			// Rate limiting defaults
			RateLimit:         10, // 10 requests per second
			RateLimitBurst:    20,
			
			// Request limits
			MaxRequestSize:    1048576, // 1MB
			MaxSSEConnections: 1000,
			
			// Monitoring defaults
			EnableMetrics:     false,
			MetricsPort:       "", // Must be set if metrics enabled
			LogLevel:          "info",
			LogFormat:         "text",
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
	// Required fields
	if c.Server.Port == "" {
		return fmt.Errorf("PORT environment variable must be set")
	}
	if c.Server.Host == "" {
		return fmt.Errorf("HOST environment variable must be set")
	}
	
	// If metrics are enabled, port must be set
	if c.Server.EnableMetrics && c.Server.MetricsPort == "" {
		return fmt.Errorf("METRICS_PORT must be set when ENABLE_METRICS is true")
	}
	
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

	// Validate and fix DefaultGameSize
	if c.Server.DefaultGameSize == 0 {
		c.Server.DefaultGameSize = 5 // Default to 5 if not set
	}
	if c.Server.DefaultGameSize < c.Server.MinPlayersPerRoom {
		c.Server.DefaultGameSize = c.Server.MinPlayersPerRoom
	}
	if c.Server.DefaultGameSize > c.Server.MaxPlayersPerRoom {
		c.Server.DefaultGameSize = c.Server.MaxPlayersPerRoom
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
