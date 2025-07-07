package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test loading default config when file doesn't exist
	t.Run("LoadDefaultWhenMissing", func(t *testing.T) {
		config, err := LoadConfig("nonexistent.yaml")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if config == nil {
			t.Fatal("expected default config, got nil")
		}
		if config.Server.MaxPlayersPerRoom != 20 {
			t.Errorf("expected MaxPlayersPerRoom 20, got %d", config.Server.MaxPlayersPerRoom)
		}
	})

	// Test loading from YAML file
	t.Run("LoadFromYAML", func(t *testing.T) {
		// Create a temporary YAML file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-config.yaml")

		yamlContent := `
server:
  maxPlayersPerRoom: 10
  minPlayersPerRoom: 2
  roomCodeLength: 6
  roomTimeout: 12h

roles:
  available:
    leader:
      displayName: "Leader"
      category: "Leader"
      minCount: 1
      maxCount: 1
      alwaysRevealed: true
    guardian:
      displayName: "Guardian"
      category: "Guardian"
      minCount: 0
      maxCount: 5
  presets:
    test:
      name: "Test"
      description: "Test preset"
      distributions:
        2: {leader: 1, guardian: 1}
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// Verify loaded values
		if config.Server.MaxPlayersPerRoom != 10 {
			t.Errorf("expected MaxPlayersPerRoom 10, got %d", config.Server.MaxPlayersPerRoom)
		}
		if config.Server.RoomTimeout != 12*time.Hour {
			t.Errorf("expected RoomTimeout 12h, got %v", config.Server.RoomTimeout)
		}
		if len(config.Roles.Available) != 2 {
			t.Errorf("expected 2 available roles, got %d", len(config.Roles.Available))
		}
		if len(config.Roles.Presets) != 1 {
			t.Errorf("expected 1 preset, got %d", len(config.Roles.Presets))
		}
	})
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *ServerConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "ValidConfig",
			config: &ServerConfig{
				Server: ServerSettings{
					MaxPlayersPerRoom: 20,
					MinPlayersPerRoom: 1,
					RoomCodeLength:    5,
					RoomTimeout:       24 * time.Hour,
				},
				Roles: RolesConfig{
					Available: map[string]RoleDefinition{
						"leader": {
							DisplayName: "Leader",
							Category:    "Leader",
							MinCount:    1,
							MaxCount:    1,
						},
					},
					Presets: map[string]Preset{},
				},
			},
			wantError: false,
		},
		{
			name: "InvalidMaxPlayers",
			config: &ServerConfig{
				Server: ServerSettings{
					MaxPlayersPerRoom: 0,
					MinPlayersPerRoom: 1,
					RoomCodeLength:    5,
				},
				Roles: RolesConfig{
					Available: map[string]RoleDefinition{
						"leader": {Category: "Leader"},
					},
				},
			},
			wantError: true,
			errorMsg:  "maxPlayersPerRoom must be at least 1",
		},
		{
			name: "MinGreaterThanMax",
			config: &ServerConfig{
				Server: ServerSettings{
					MaxPlayersPerRoom: 5,
					MinPlayersPerRoom: 10,
					RoomCodeLength:    5,
				},
				Roles: RolesConfig{
					Available: map[string]RoleDefinition{
						"leader": {Category: "Leader"},
					},
				},
			},
			wantError: true,
			errorMsg:  "minPlayersPerRoom cannot be greater than maxPlayersPerRoom",
		},
		{
			name: "NoLeaderRole",
			config: &ServerConfig{
				Server: ServerSettings{
					MaxPlayersPerRoom: 20,
					MinPlayersPerRoom: 1,
					RoomCodeLength:    5,
				},
				Roles: RolesConfig{
					Available: map[string]RoleDefinition{
						"guardian": {Category: "Guardian"},
					},
				},
			},
			wantError: true,
			errorMsg:  "at least one Leader role must be defined",
		},
		{
			name: "InvalidRoleMinMax",
			config: &ServerConfig{
				Server: ServerSettings{
					MaxPlayersPerRoom: 20,
					MinPlayersPerRoom: 1,
					RoomCodeLength:    5,
				},
				Roles: RolesConfig{
					Available: map[string]RoleDefinition{
						"leader": {
							Category: "Leader",
							MinCount: 5,
							MaxCount: 1,
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "minCount cannot be greater than maxCount",
		},
		{
			name: "UnknownRoleInPreset",
			config: &ServerConfig{
				Server: ServerSettings{
					MaxPlayersPerRoom: 20,
					MinPlayersPerRoom: 1,
					RoomCodeLength:    5,
				},
				Roles: RolesConfig{
					Available: map[string]RoleDefinition{
						"leader": {Category: "Leader"},
					},
					Presets: map[string]Preset{
						"test": {
							Distributions: map[int]map[string]int{
								2: {"leader": 1, "unknown": 1},
							},
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "unknown role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestGetPreset(t *testing.T) {
	config := &ServerConfig{
		Roles: RolesConfig{
			Presets: map[string]Preset{
				"standard": {
					Name:        "Standard",
					Description: "Test preset",
				},
			},
		},
	}

	// Test existing preset
	preset, exists := config.GetPreset("standard")
	if !exists {
		t.Error("expected preset to exist")
	}
	if preset.Name != "Standard" {
		t.Errorf("expected preset name 'Standard', got '%s'", preset.Name)
	}

	// Test non-existing preset
	_, exists = config.GetPreset("nonexistent")
	if exists {
		t.Error("expected preset not to exist")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
