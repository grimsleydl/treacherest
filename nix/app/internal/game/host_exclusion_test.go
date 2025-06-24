package game

import (
	"testing"
)

func TestHostExclusion(t *testing.T) {
	t.Run("hosts excluded from role assignment", func(t *testing.T) {
		// Create players including a host
		players := []*Player{
			{ID: "p1", Name: "Player 1", IsHost: false},
			{ID: "p2", Name: "Player 2", IsHost: false},
			{ID: "host", Name: "Host Player", IsHost: true},
		}

		// Create card service
		cardService, err := NewCardService()
		if err != nil {
			t.Fatalf("Failed to create card service: %v", err)
		}

		// Assign roles
		AssignRoles(players, cardService)

		// Verify host has no role
		for _, p := range players {
			if p.IsHost {
				if p.Role != nil {
					t.Errorf("Host player should not have a role, but got: %v", p.Role)
				}
			} else {
				if p.Role == nil {
					t.Errorf("Non-host player %s should have a role", p.Name)
				}
			}
		}
	})

	t.Run("hosts excluded from CanStart count", func(t *testing.T) {
		room := &Room{
			Code:       "TEST",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 8,
		}

		// Add only a host
		host := &Player{ID: "host", Name: "Host", IsHost: true}
		room.AddPlayer(host)

		// Should not be able to start with only a host
		if room.CanStart() {
			t.Error("Room should not be able to start with only a host")
		}

		// Add one active player
		player := &Player{ID: "p1", Name: "Player 1", IsHost: false}
		room.AddPlayer(player)

		// Now should be able to start
		if !room.CanStart() {
			t.Error("Room should be able to start with one active player")
		}
	})

	t.Run("GetActivePlayers excludes hosts", func(t *testing.T) {
		room := &Room{
			Code:       "TEST",
			Players:    make(map[string]*Player),
			MaxPlayers: 8,
		}

		// Add mixed players
		room.AddPlayer(&Player{ID: "p1", Name: "Player 1", IsHost: false})
		room.AddPlayer(&Player{ID: "host", Name: "Host", IsHost: true})
		room.AddPlayer(&Player{ID: "p2", Name: "Player 2", IsHost: false})

		// Get active players
		activePlayers := room.GetActivePlayers()

		// Should have 2 active players
		if len(activePlayers) != 2 {
			t.Errorf("Expected 2 active players, got %d", len(activePlayers))
		}

		// Verify no hosts in active players
		for _, p := range activePlayers {
			if p.IsHost {
				t.Errorf("Host player %s should not be in active players", p.Name)
			}
		}
	})

	t.Run("GetActivePlayerCount excludes hosts", func(t *testing.T) {
		room := &Room{
			Code:       "TEST",  
			Players:    make(map[string]*Player),
			MaxPlayers: 8,
		}

		// Add mixed players
		room.AddPlayer(&Player{ID: "p1", Name: "Player 1", IsHost: false})
		room.AddPlayer(&Player{ID: "host1", Name: "Host 1", IsHost: true})
		room.AddPlayer(&Player{ID: "p2", Name: "Player 2", IsHost: false})
		room.AddPlayer(&Player{ID: "host2", Name: "Host 2", IsHost: true})

		// Should count only active players
		if count := room.GetActivePlayerCount(); count != 2 {
			t.Errorf("Expected 2 active players, got %d", count)
		}

		// Total players should be 4
		if len(room.Players) != 4 {
			t.Errorf("Expected 4 total players, got %d", len(room.Players))
		}
	})

	t.Run("role distribution with hosts", func(t *testing.T) {
		// Create 3 active players and 1 host
		players := []*Player{
			{ID: "p1", Name: "Player 1", IsHost: false},
			{ID: "p2", Name: "Player 2", IsHost: false},
			{ID: "p3", Name: "Player 3", IsHost: false},
			{ID: "host", Name: "Host", IsHost: true},
		}

		cardService, err := NewCardService()
		if err != nil {
			t.Fatalf("Failed to create card service: %v", err)
		}
		AssignRoles(players, cardService)

		// Count roles assigned
		leaderCount := 0
		guardianCount := 0
		traitorCount := 0

		for _, p := range players {
			if p.Role != nil {
				switch p.Role.GetRoleType() {
				case RoleLeader:
					leaderCount++
				case RoleGuardian:
					guardianCount++
				case RoleTraitor:
					traitorCount++
				}
			}
		}

		// Should have the distribution for 3 players
		if leaderCount != 1 {
			t.Errorf("Expected 1 leader, got %d", leaderCount)
		}
		if guardianCount != 1 {
			t.Errorf("Expected 1 guardian, got %d", guardianCount)
		}
		if traitorCount != 1 {
			t.Errorf("Expected 1 traitor, got %d", traitorCount)
		}
	})
}