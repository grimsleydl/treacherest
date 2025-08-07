package game

import (
	"testing"
	"treacherest/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeaderlessGameConfiguration(t *testing.T) {
	cfg := config.DefaultConfig()
	service := NewRoleConfigService(cfg)

	t.Run("allows leaderless configuration when enabled", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"guardian": true, "traitor": true},
			RoleCounts:          map[string]int{"guardian": 2, "traitor": 1},
			MinPlayers:          3,
			MaxPlayers:          3,
			AllowLeaderlessGame: true,
		}

		err := service.ValidateConfiguration(roleConfig)
		assert.NoError(t, err)
	})

	t.Run("rejects leaderless configuration when disabled", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"guardian": true, "traitor": true},
			RoleCounts:          map[string]int{"guardian": 2, "traitor": 1},
			MinPlayers:          3,
			MaxPlayers:          3,
			AllowLeaderlessGame: false,
		}

		err := service.ValidateConfiguration(roleConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have a leader role")
	})

	t.Run("allows leader with leaderless enabled", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"leader": true, "guardian": true, "traitor": true},
			RoleCounts:          map[string]int{"leader": 1, "guardian": 1, "traitor": 1},
			MinPlayers:          3,
			MaxPlayers:          3,
			AllowLeaderlessGame: true,
		}

		err := service.ValidateConfiguration(roleConfig)
		assert.NoError(t, err)
	})

	t.Run("rejects multiple leaders even with leaderless enabled", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"leader": true},
			RoleCounts:          map[string]int{"leader": 2},
			MinPlayers:          2,
			MaxPlayers:          2,
			AllowLeaderlessGame: true,
		}

		err := service.ValidateConfiguration(roleConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot have more than 1 leader")
	})
}

func TestLeaderlessRoleDistribution(t *testing.T) {
	cfg := config.DefaultConfig()
	service := NewRoleConfigService(cfg)

	t.Run("distributes roles without leader", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"guardian": true, "assassin": true, "traitor": true},
			RoleCounts:          map[string]int{"guardian": 2, "assassin": 1, "traitor": 1},
			MinPlayers:          4,
			MaxPlayers:          4,
			AllowLeaderlessGame: true,
		}

		distribution, err := service.GetDistributionForPlayerCount(roleConfig, 4)
		require.NoError(t, err)

		assert.Equal(t, 0, distribution[RoleLeader])
		assert.Equal(t, 2, distribution[RoleGuardian])
		assert.Equal(t, 1, distribution[RoleAssassin])
		assert.Equal(t, 1, distribution[RoleTraitor])
	})

	t.Run("does not auto-add leader when leaderless allowed", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"traitor": true},
			RoleCounts:          map[string]int{"traitor": 2},
			MinPlayers:          2,
			MaxPlayers:          2,
			AllowLeaderlessGame: true,
		}

		distribution, err := service.GetDistributionForPlayerCount(roleConfig, 2)
		require.NoError(t, err)

		assert.Equal(t, 0, distribution[RoleLeader])
		assert.Equal(t, 2, distribution[RoleTraitor])
	})

	t.Run("auto-adds leader when leaderless not allowed", func(t *testing.T) {
		roleConfig := &RoleConfiguration{
			PresetName:          "custom",
			EnabledRoles:        map[string]bool{"traitor": true},
			RoleCounts:          map[string]int{"traitor": 1},
			MinPlayers:          2,
			MaxPlayers:          2,
			AllowLeaderlessGame: false,
		}

		distribution, err := service.GetDistributionForPlayerCount(roleConfig, 2)
		require.NoError(t, err)

		assert.Equal(t, 1, distribution[RoleLeader])
		assert.Equal(t, 1, distribution[RoleTraitor])
	})
}

func TestLeaderDependentCards(t *testing.T) {
	tests := []struct {
		name           string
		cardName       string
		expectedResult bool
	}{
		{"The Golem is leader dependent", "The Golem", true},
		{"The Great Martyr is leader dependent", "The Great Martyr", true},
		{"The Oracle is leader dependent", "The Oracle", true},
		{"The Quellmaster is leader dependent", "The Quellmaster", true},
		{"The Metamorph is leader dependent", "The Metamorph", true},
		{"The Puppet Master is leader dependent", "The Puppet Master", true},
		{"Random card is not leader dependent", "The Bodyguard", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{Name: tt.cardName}
			assert.Equal(t, tt.expectedResult, card.IsLeaderDependent())
		})
	}
}

func TestLeaderlessWinConditions(t *testing.T) {
	tests := []struct {
		name              string
		subtype           string
		expectedCondition string
	}{
		{
			"Guardian win condition",
			"Guardian",
			"Win if majority of non-Traitor players survive",
		},
		{
			"Assassin win condition",
			"Assassin",
			"Win by eliminating your secret target",
		},
		{
			"Traitor win condition",
			"Traitor",
			"Be the last player standing",
		},
		{
			"Leader not applicable",
			"Leader",
			"Not applicable in leaderless games",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{Types: CardTypes{Subtype: tt.subtype}}
			assert.Equal(t, tt.expectedCondition, card.GetLeaderlessWinCondition())
		})
	}
}