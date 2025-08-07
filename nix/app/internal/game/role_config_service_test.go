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
			Presets: map[string]map[string]config.RoleDistribution{
				"basic-3p": {
					"3": {
						Roles: map[string]int{
							"leader":   1,
							"guardian": 1,
							"traitor":  1,
						},
					},
				},
				"dynamic": {
					"3-4": {
						Roles: map[string]int{
							"leader":   1,
							"guardian": 2,
							"traitor":  1,
						},
					},
					"5-6": {
						Roles: map[string]int{
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
				if rc.MinPlayers != 3 || rc.MaxPlayers != 3 {
					t.Errorf("expected min/max players 3, got %d/%d", rc.MinPlayers, rc.MaxPlayers)
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
				if rc.MinPlayers != 3 || rc.MaxPlayers != 6 {
					t.Errorf("expected min 3 max 6 players, got %d/%d", rc.MinPlayers, rc.MaxPlayers)
				}
			},
		},
		{
			name:       "invalid preset name",
			presetName: "nonexistent",
			maxPlayers: 10,
			wantErr:    true,
		},
		{
			name:       "empty preset name creates default",
			presetName: "",
			maxPlayers: 10,
			wantErr:    false,
			checkConfig: func(t *testing.T, rc *RoleConfiguration) {
				if rc.PresetName != "custom" {
					t.Errorf("expected preset name 'custom', got %s", rc.PresetName)
				}
				// Should have leader enabled by default
				if !rc.EnabledRoles["leader"] {
					t.Error("expected leader role to be enabled by default")
				}
			},
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
			Presets: map[string]map[string]config.RoleDistribution{
				"dynamic": {
					"3-4": {
						Roles: map[string]int{
							"leader":   1,
							"guardian": 2,
							"traitor":  1,
						},
					},
					"5-6": {
						Roles: map[string]int{
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
			name:        "player count outside all ranges",
			playerCount: 7,
			wantRoles:   nil,
			wantErr:     true,
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
		wantErrors int
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
			wantErrors: 0,
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
			wantValid:  false,
			wantErrors: 1, // leader is required
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
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errors := service.ValidateConfiguration(tt.roleConfig, 0)
			if valid != tt.wantValid {
				t.Errorf("ValidateConfiguration() valid = %v, want %v", valid, tt.wantValid)
			}
			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateConfiguration() returned %d errors, want %d", len(errors), tt.wantErrors)
			}
		})
	}
}