package game

import (
	"testing"
	"treacherest"
	"treacherest/internal/config"
)

func TestAssignRoles(t *testing.T) {
	tests := []struct {
		name        string
		playerCount int
		wantLeaders int
		wantRoles   int
	}{
		{"1 player", 1, 1, 1},
		{"2 players", 2, 1, 2},
		{"3 players", 3, 1, 3},
		{"4 players", 4, 1, 4},
		{"5 players", 5, 1, 5},
		{"6 players", 6, 1, 6},
		{"7 players", 7, 1, 7},
		{"8 players", 8, 1, 8},
	}

	// Create CardService for testing
	cardService, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create players
			players := make([]*Player, tt.playerCount)
			for i := 0; i < tt.playerCount; i++ {
				players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
			}

			// Assign roles
			AssignRoles(players, cardService)

			// Count roles
			leaderCount := 0
			rolesAssigned := 0

			for _, p := range players {
				if p.Role != nil {
					rolesAssigned++
					if p.Role.GetRoleType() == RoleLeader {
						leaderCount++
						// Leader should be revealed
						if !p.RoleRevealed {
							t.Error("Leader should be revealed")
						}
					}
				}
			}

			if leaderCount != tt.wantLeaders {
				t.Errorf("Expected %d leader, got %d", tt.wantLeaders, leaderCount)
			}

			if rolesAssigned != tt.wantRoles {
				t.Errorf("Expected %d roles assigned, got %d", tt.wantRoles, rolesAssigned)
			}
		})
	}
}

func TestGetRoleDistribution(t *testing.T) {
	tests := []struct {
		playerCount int
		wantRoles   map[RoleType]int
	}{
		{
			1,
			map[RoleType]int{
				RoleLeader: 1,
			},
		},
		{
			2,
			map[RoleType]int{
				RoleLeader:  1,
				RoleTraitor: 1,
			},
		},
		{
			3,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 1,
				RoleTraitor:  1,
			},
		},
		{
			4,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleTraitor:  1,
			},
		},
		{
			5,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleAssassin: 1,
				RoleTraitor:  1,
			},
		},
		{
			6,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 2,
				RoleAssassin: 2,
				RoleTraitor:  1,
			},
		},
		{
			7,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 3,
				RoleAssassin: 2,
				RoleTraitor:  1,
			},
		},
		{
			8,
			map[RoleType]int{
				RoleLeader:   1,
				RoleGuardian: 3,
				RoleAssassin: 2,
				RoleTraitor:  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.playerCount))+" players", func(t *testing.T) {
			roles := getRoleDistribution(tt.playerCount)

			// Check counts match expected
			for roleType, wantCount := range tt.wantRoles {
				if roles[roleType] != wantCount {
					t.Errorf("Role %s: want %d, got %d", roleType, wantCount, roles[roleType])
				}
			}

			// Verify total count
			totalCount := 0
			for _, count := range roles {
				totalCount += count
			}
			if totalCount != tt.playerCount {
				t.Errorf("Total role count %d doesn't match player count %d", totalCount, tt.playerCount)
			}
		})
	}
}

func TestAssignRoles_NoDuplicateCards(t *testing.T) {
	// Create CardService for testing
	cardService, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test with maximum players to increase chance of duplicates
	players := make([]*Player, 8)
	for i := 0; i < 8; i++ {
		players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
	}

	// Run multiple times to test randomness
	for run := 0; run < 10; run++ {
		// Reset players' roles
		for _, p := range players {
			p.Role = nil
			p.RoleRevealed = false
		}

		// Assign roles
		AssignRoles(players, cardService)

		// Track used cards to check for duplicates
		usedCards := make(map[*Card]bool)
		for _, p := range players {
			if p.Role == nil {
				t.Errorf("Player has nil role in run %d", run)
				continue
			}

			if usedCards[p.Role] {
				t.Errorf("Duplicate card assigned in run %d: %s", run, p.Role.Name)
			}
			usedCards[p.Role] = true
		}
	}
}

func TestAssignRoles_CorrectRoleTypes(t *testing.T) {
	// Create CardService for testing
	cardService, err := NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatalf("Failed to create CardService: %v", err)
	}

	// Test each player count configuration
	for playerCount := 1; playerCount <= 8; playerCount++ {
		t.Run(string(rune('0'+playerCount))+" players", func(t *testing.T) {
			players := make([]*Player, playerCount)
			for i := 0; i < playerCount; i++ {
				players[i] = NewPlayer(string(rune('a'+i)), "Player", "session")
			}

			// Assign roles
			AssignRoles(players, cardService)

			// Get expected distribution
			expectedDist := getRoleDistribution(playerCount)

			// Count actual role types
			actualDist := make(map[RoleType]int)
			for _, p := range players {
				if p.Role != nil {
					roleType := p.Role.GetRoleType()
					actualDist[roleType]++
				}
			}

			// Compare distributions
			for roleType, expectedCount := range expectedDist {
				if actualDist[roleType] != expectedCount {
					t.Errorf("Role %s: expected %d, got %d", roleType, expectedCount, actualDist[roleType])
				}
			}
		})
	}
}

func TestHiddenDistribution(t *testing.T) {
	// Create test card service
	cardService := &CardService{
		Leaders: []*Card{
			{Name: "Leader1", NameAnchor: "leader1", Types: CardTypes{Subtype: "Leader"}},
			{Name: "Leader2", NameAnchor: "leader2", Types: CardTypes{Subtype: "Leader"}},
		},
		Guardians: []*Card{
			{Name: "Guardian1", NameAnchor: "guardian1", Types: CardTypes{Subtype: "Guardian"}},
			{Name: "Guardian2", NameAnchor: "guardian2", Types: CardTypes{Subtype: "Guardian"}},
			{Name: "Guardian3", NameAnchor: "guardian3", Types: CardTypes{Subtype: "Guardian"}},
		},
		Assassins: []*Card{
			{Name: "Assassin1", NameAnchor: "assassin1", Types: CardTypes{Subtype: "Assassin"}},
			{Name: "Assassin2", NameAnchor: "assassin2", Types: CardTypes{Subtype: "Assassin"}},
		},
		Traitors: []*Card{
			{Name: "Traitor1", NameAnchor: "traitor1", Types: CardTypes{Subtype: "Traitor"}},
			{Name: "Traitor2", NameAnchor: "traitor2", Types: CardTypes{Subtype: "Traitor"}},
		},
	}

	// Create test config
	cfg := &config.ServerConfig{
		Roles: config.RolesConfig{
			Presets: map[string]config.Preset{
				"standard": {
					Name:        "Standard",
					Description: "Balanced gameplay",
					Distributions: map[int]map[string]int{
						1: {"leader": 1},
						2: {"leader": 1, "guardian": 1},
						3: {"leader": 1, "guardian": 1, "traitor": 1},
						4: {"leader": 1, "guardian": 2, "traitor": 1},
						5: {"leader": 1, "guardian": 2, "assassin": 1, "traitor": 1},
					},
				},
				"assassination": {
					Name:        "Assassination",
					Description: "Heavy on assassins",
					Distributions: map[int]map[string]int{
						3: {"leader": 1, "assassin": 2},
						4: {"leader": 1, "guardian": 1, "assassin": 2},
						5: {"leader": 1, "guardian": 1, "assassin": 3},
					},
				},
			},
		},
	}

	roleService := NewRoleConfigService(cfg)

	t.Run("hidden distribution assigns roles from random preset", func(t *testing.T) {
		players := []*Player{
			{ID: "1", Name: "Player1"},
			{ID: "2", Name: "Player2"},
			{ID: "3", Name: "Player3"},
			{ID: "4", Name: "Player4"},
		}

		roleConfig := &RoleConfiguration{
			HideRoleDistribution: true,
			AllowLeaderlessGame:  false,
			MinPlayers:           1,
			MaxPlayers:           8,
			RoleTypes: map[string]*RoleTypeConfig{
				"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader1": true, "Leader2": true}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian1": true, "Guardian2": true, "Guardian3": true}},
				"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin1": true, "Assassin2": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"Traitor1": true, "Traitor2": true}},
			},
		}

		AssignRolesWithConfig(players, cardService, roleConfig, roleService)

		// Check all players have roles
		for _, p := range players {
			if p.Role == nil {
				t.Errorf("Player %s has no role assigned", p.Name)
			}
		}

		// Check we have at least one leader
		hasLeader := false
		for _, p := range players {
			if p.Role != nil && p.Role.GetRoleType() == RoleLeader {
				hasLeader = true
				break
			}
		}
		if !hasLeader {
			t.Error("No leader assigned when leaderless games are not allowed")
		}
	})

	t.Run("fully random distribution assigns random roles", func(t *testing.T) {
		players := []*Player{
			{ID: "1", Name: "Player1"},
			{ID: "2", Name: "Player2"},
			{ID: "3", Name: "Player3"},
			{ID: "4", Name: "Player4"},
			{ID: "5", Name: "Player5"},
		}

		roleConfig := &RoleConfiguration{
			FullyRandomRoles:    true,
			AllowLeaderlessGame: false,
			MinPlayers:          1,
			MaxPlayers:          8,
			RoleTypes: map[string]*RoleTypeConfig{
				"Leader":   {Count: 1, EnabledCards: map[string]bool{"Leader1": true, "Leader2": true}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian1": true, "Guardian2": true, "Guardian3": true}},
				"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin1": true, "Assassin2": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"Traitor1": true, "Traitor2": true}},
			},
		}

		AssignRolesWithConfig(players, cardService, roleConfig, roleService)

		// Check all players have roles
		for _, p := range players {
			if p.Role == nil {
				t.Errorf("Player %s has no role assigned", p.Name)
			}
		}

		// Check we have at least one leader
		hasLeader := false
		for _, p := range players {
			if p.Role != nil && p.Role.GetRoleType() == RoleLeader {
				hasLeader = true
				break
			}
		}
		if !hasLeader {
			t.Error("No leader assigned when leaderless games are not allowed")
		}

		// Check no duplicate cards
		usedCards := make(map[string]bool)
		for _, p := range players {
			if p.Role != nil {
				if usedCards[p.Role.Name] {
					t.Errorf("Duplicate card assigned: %s", p.Role.Name)
				}
				usedCards[p.Role.Name] = true
			}
		}
	})

	t.Run("fully random with leaderless allowed", func(t *testing.T) {
		players := []*Player{
			{ID: "1", Name: "Player1"},
			{ID: "2", Name: "Player2"},
			{ID: "3", Name: "Player3"},
		}

		roleConfig := &RoleConfiguration{
			FullyRandomRoles:    true,
			AllowLeaderlessGame: true,
			MinPlayers:          1,
			MaxPlayers:          8,
			RoleTypes: map[string]*RoleTypeConfig{
				"Leader":   {Count: 0, EnabledCards: map[string]bool{"Leader1": true, "Leader2": true}},
				"Guardian": {Count: 2, EnabledCards: map[string]bool{"Guardian1": true, "Guardian2": true, "Guardian3": true}},
				"Assassin": {Count: 1, EnabledCards: map[string]bool{"Assassin1": true, "Assassin2": true}},
				"Traitor":  {Count: 1, EnabledCards: map[string]bool{"Traitor1": true, "Traitor2": true}},
			},
		}

		AssignRolesWithConfig(players, cardService, roleConfig, roleService)

		// Check all players have roles
		for _, p := range players {
			if p.Role == nil {
				t.Errorf("Player %s has no role assigned", p.Name)
			}
		}

		// With leaderless allowed, we might not have a leader
		// Just check that roles were assigned
		roleCount := 0
		for _, p := range players {
			if p.Role != nil {
				roleCount++
			}
		}
		if roleCount != len(players) {
			t.Errorf("Expected %d roles assigned, got %d", len(players), roleCount)
		}
	})
}
