package game

import (
	"testing"
	"treacherest/internal/config"
)

func TestRoleConfigService_CreateFromPreset(t *testing.T) {
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName: "Leader",
					Category:    "Leader",
					MinCount:    1,
					MaxCount:    1,
				},
				"guardian": {
					DisplayName: "Guardian",
					Category:    "Good",
					MinCount:    0,
					MaxCount:    10,
				},
				"assassin": {
					DisplayName: "Assassin",
					Category:    "Evil",
					MinCount:    0,
					MaxCount:    10,
				},
				"traitor": {
					DisplayName: "Traitor",
					Category:    "Evil",
					MinCount:    0,
					MaxCount:    10,
				},
			},
			Presets: map[string]config.Preset{
				"basic-3p": {
					Name:        "Basic 3 Player",
					Description: "Simple 3 player setup",
					Distributions: map[int]map[string]int{
						3: {
							"leader":   1,
							"guardian": 1,
							"traitor":  1,
						},
					},
				},
				"dynamic": {
					Name:        "Dynamic",
					Description: "Adapts to player count",
					Distributions: map[int]map[string]int{
						3: {
							"leader":   1,
							"guardian": 1,
							"traitor":  1,
						},
						4: {
							"leader":   1,
							"guardian": 2,
							"traitor":  1,
						},
						5: {
							"leader":   1,
							"guardian": 3,
							"assassin": 1,
							"traitor":  1,
						},
						6: {
							"leader":   1,
							"guardian": 3,
							"assassin": 1,
							"traitor":  1,
						},
					},
				},
			},
		},
	}

	service := NewRoleConfigService(cfg)

	tests := []struct {
		name        string
		presetName  string
		maxPlayers  int
		wantErr     bool
		checkConfig func(t *testing.T, config *RoleConfiguration)
	}{
		{
			name:       "valid basic preset",
			presetName: "basic-3p",
			maxPlayers: 10,
			wantErr:    false,
			checkConfig: func(t *testing.T, rc *RoleConfiguration) {
				if rc.PresetName != "basic-3p" {
					t.Errorf("expected preset name 'basic-3p', got %s", rc.PresetName)
				}
				if rc.MinPlayers != 3 || rc.MaxPlayers != 10 {
					t.Errorf("expected min/max players 3/10, got %d/%d", rc.MinPlayers, rc.MaxPlayers)
				}
			},
		},
		{
			name:       "valid dynamic preset",
			presetName: "dynamic",
			maxPlayers: 10,
			wantErr:    false,
			checkConfig: func(t *testing.T, rc *RoleConfiguration) {
				if rc.PresetName != "dynamic" {
					t.Errorf("expected preset name 'dynamic', got %s", rc.PresetName)
				}
				if rc.MinPlayers != 3 || rc.MaxPlayers != 10 {
					t.Errorf("expected min 3 max 10 players, got %d/%d", rc.MinPlayers, rc.MaxPlayers)
				}
			},
		},
		{
			name:       "invalid preset name",
			presetName: "nonexistent",
			maxPlayers: 10,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateFromPreset(tt.presetName, tt.maxPlayers)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFromPreset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkConfig != nil {
				tt.checkConfig(t, result)
			}
		})
	}
}

func TestRoleConfigService_GetDistributionForPlayerCount(t *testing.T) {
	cfg := &config.ServerConfig{
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader":   {DisplayName: "Leader", MinCount: 1, MaxCount: 1},
				"guardian": {DisplayName: "Guardian", MinCount: 0, MaxCount: 10},
				"traitor":  {DisplayName: "Traitor", MinCount: 0, MaxCount: 10},
			},
			Presets: map[string]config.Preset{
				"dynamic": {
					Name:        "Dynamic",
					Description: "Adapts to player count",
					Distributions: map[int]map[string]int{
						3: {
							"leader":   1,
							"guardian": 1,
							"traitor":  1,
						},
						4: {
							"leader":   1,
							"guardian": 2,
							"traitor":  1,
						},
						5: {
							"leader":   1,
							"guardian": 2,
							"assassin": 1,
							"traitor":  1,
						},
						6: {
							"leader":   1,
							"guardian": 3,
							"assassin": 1,
							"traitor":  1,
						},
					},
				},
			},
		},
	}

	service := NewRoleConfigService(cfg)
	
	// Create a dynamic preset config
	roleConfig, _ := service.CreateFromPreset("dynamic", 10)

	tests := []struct {
		name         string
		playerCount  int
		wantRoles    map[RoleType]int
		wantErr      bool
	}{
		{
			name:        "3 players matches 3-4 range",
			playerCount: 3,
			wantRoles: map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 1,
				RoleTraitor:  1,
			},
			wantErr: false,
		},
		{
			name:        "4 players matches 3-4 range",
			playerCount: 4,
			wantRoles: map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleTraitor:  1,
			},
			wantErr: false,
		},
		{
			name:        "5 players matches 5-6 range",
			playerCount: 5,
			wantRoles: map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleAssassin: 1,
				RoleTraitor:  1,
			},
			wantErr: false,
		},
		{
			name:        "player count outside all ranges uses closest",
			playerCount: 7,
			wantRoles: map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 4, // 2 + 2 extra to fill
				RoleAssassin: 1,
				RoleTraitor:  1,
			},
			wantErr: false,
		},
		{
			name:        "zero players",
			playerCount: 0,
			wantRoles:   nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetDistributionForPlayerCount(roleConfig, tt.playerCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDistributionForPlayerCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != len(tt.wantRoles) {
					t.Errorf("expected %d roles, got %d", len(tt.wantRoles), len(result))
				}
				for role, count := range tt.wantRoles {
					if result[role] != count {
						t.Errorf("expected %d %s, got %d", count, role, result[role])
					}
				}
			}
		})
	}
}

func TestRoleConfigService_ValidateConfiguration(t *testing.T) {
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName: "Leader",
					MinCount:    1,
					MaxCount:    1,
				},
				"guardian": {
					DisplayName: "Guardian",
					MinCount:    0,
					MaxCount:    10,
				},
			},
		},
	}

	service := NewRoleConfigService(cfg)

	tests := []struct {
		name       string
		roleConfig *RoleConfiguration
		wantValid  bool
	}{
		{
			name: "valid configuration",
			roleConfig: &RoleConfiguration{
				EnabledRoles: map[string]bool{
					"leader":   true,
					"guardian": true,
				},
				RoleCounts: map[string]int{
					"leader":   1,
					"guardian": 2,
				},
				MinPlayers: 3,
				MaxPlayers: 3,
			},
			wantValid:  true,
		},
		{
			name: "missing required role",
			roleConfig: &RoleConfiguration{
				EnabledRoles: map[string]bool{
					"guardian": true,
				},
				RoleCounts: map[string]int{
					"guardian": 2,
				},
				MinPlayers: 2,
				MaxPlayers: 2,
			},
			wantValid:  false, // leader is required
		},
		{
			name: "invalid role count",
			roleConfig: &RoleConfiguration{
				EnabledRoles: map[string]bool{
					"leader": true,
				},
				RoleCounts: map[string]int{
					"leader": 2, // max is 1
				},
				MinPlayers: 2,
				MaxPlayers: 2,
			},
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateConfiguration(tt.roleConfig)
			if (err != nil) != !tt.wantValid {
				t.Errorf("ValidateConfiguration() error = %v, wantValid %v", err, tt.wantValid)
			}
		})
	}
}

func TestRoleConfigService_CreateFromPreset_InitializesRoleCounts(t *testing.T) {
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName: "Leader",
					Category:    "Leader",
					MinCount:    1,
					MaxCount:    1,
				},
				"guardian": {
					DisplayName: "Guardian",
					Category:    "Good",
					MinCount:    0,
					MaxCount:    10,
				},
				"assassin": {
					DisplayName: "Assassin",
					Category:    "Evil",
					MinCount:    2,
					MaxCount:    5,
				},
			},
			Presets: map[string]config.Preset{
				"test-preset": {
					Name:        "Test Preset",
					Description: "Test preset for role count initialization",
					Distributions: map[int]map[string]int{
						5: {
							"leader":   1,
							"guardian": 3,
							"assassin": 1,
						},
					},
				},
			},
		},
	}

	service := NewRoleConfigService(cfg)

	// Test preset creation initializes counts based on MinCount
	result, err := service.CreateFromPreset("test-preset", 10)
	if err != nil {
		t.Fatalf("CreateFromPreset() error = %v", err)
	}

	// Check leader has MinCount of 1
	if count, exists := result.RoleCounts["leader"]; !exists || count != 1 {
		t.Errorf("expected leader count to be 1 (MinCount), got %d (exists: %v)", count, exists)
	}

	// Check guardian has default count of 1 (MinCount is 0)
	if count, exists := result.RoleCounts["guardian"]; !exists || count != 1 {
		t.Errorf("expected guardian count to be 1 (default), got %d (exists: %v)", count, exists)
	}

	// Check assassin has MinCount of 2
	if count, exists := result.RoleCounts["assassin"]; !exists || count != 2 {
		t.Errorf("expected assassin count to be 2 (MinCount), got %d (exists: %v)", count, exists)
	}
}

func TestRoleConfigService_CreateCustomConfiguration_RespectsMinCount(t *testing.T) {
	cfg := &config.ServerConfig{
		Server: config.ServerSettings{
			MaxPlayersPerRoom: 20,
			MinPlayersPerRoom: 1,
		},
		Roles: config.RolesConfig{
			Available: map[string]config.RoleDefinition{
				"leader": {
					DisplayName: "Leader",
					MinCount:    1,
					MaxCount:    1,
				},
				"guardian": {
					DisplayName: "Guardian",
					MinCount:    0,
					MaxCount:    10,
				},
			},
		},
	}

	service := NewRoleConfigService(cfg)

	// Test with invalid counts (below MinCount)
	enabledRoles := map[string]bool{
		"leader":   true,
		"guardian": true,
	}
	roleCounts := map[string]int{
		"leader":   0, // Invalid: below MinCount of 1
		"guardian": 0, // Will be set to default 1
	}

	result, err := service.CreateCustomConfiguration(enabledRoles, roleCounts)
	if err != nil {
		t.Fatalf("CreateCustomConfiguration() error = %v", err)
	}

	// Check leader was fixed to MinCount of 1
	if count := result.RoleCounts["leader"]; count != 1 {
		t.Errorf("expected leader count to be fixed to 1 (MinCount), got %d", count)
	}

	// Check guardian was set to default 1
	if count := result.RoleCounts["guardian"]; count != 1 {
		t.Errorf("expected guardian count to be 1 (default), got %d", count)
	}
}