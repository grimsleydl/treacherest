package game

import (
	"testing"
	"time"
)

func TestNewPlayer(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		playerName string
		sessionID  string
	}{
		{
			name:       "creates player with all fields",
			id:         "player-123",
			playerName: "Alice",
			sessionID:  "session-abc",
		},
		{
			name:       "creates player with empty name",
			id:         "player-456",
			playerName: "",
			sessionID:  "session-def",
		},
		{
			name:       "creates player with special characters in name",
			id:         "player-789",
			playerName: "Player@#$%",
			sessionID:  "session-ghi",
		},
		{
			name:       "creates player with unicode name",
			id:         "player-unicode",
			playerName: "プレイヤー",
			sessionID:  "session-unicode",
		},
		{
			name:       "creates player with very long name",
			id:         "player-long",
			playerName: "ThisIsAVeryLongPlayerNameThatMightExceedNormalLimits",
			sessionID:  "session-long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCreation := time.Now()
			player := NewPlayer(tt.id, tt.playerName, tt.sessionID)
			afterCreation := time.Now()

			// Verify basic fields
			if player.ID != tt.id {
				t.Errorf("ID = %v, want %v", player.ID, tt.id)
			}
			if player.Name != tt.playerName {
				t.Errorf("Name = %v, want %v", player.Name, tt.playerName)
			}
			if player.SessionID != tt.sessionID {
				t.Errorf("SessionID = %v, want %v", player.SessionID, tt.sessionID)
			}

			// Verify default values
			if player.Role != nil {
				t.Errorf("Role = %v, want nil", player.Role)
			}
			if player.RoleRevealed != false {
				t.Errorf("RoleRevealed = %v, want false", player.RoleRevealed)
			}

			// Verify JoinedAt is set to current time
			if player.JoinedAt.Before(beforeCreation) || player.JoinedAt.After(afterCreation) {
				t.Errorf("JoinedAt = %v, want between %v and %v", player.JoinedAt, beforeCreation, afterCreation)
			}
		})
	}
}

func TestPlayer_RoleAssignment(t *testing.T) {
	player := NewPlayer("player-1", "Bob", "session-1")

	// Test initial state
	if player.Role != nil {
		t.Errorf("Initial Role = %v, want nil", player.Role)
	}
	if player.RoleRevealed {
		t.Errorf("Initial RoleRevealed = %v, want false", player.RoleRevealed)
	}

	// Assign a role
	guardianRole := &Role{
		Type:         RoleGuardian,
		Name:         "Guardian",
		Description:  "Test Guardian",
		WinCondition: "Test Win Condition",
	}
	player.Role = guardianRole

	if player.Role != guardianRole {
		t.Errorf("Role after assignment = %v, want %v", player.Role, guardianRole)
	}

	// Test role reveal
	player.RoleRevealed = true
	if !player.RoleRevealed {
		t.Errorf("RoleRevealed after setting = %v, want true", player.RoleRevealed)
	}
}

func TestPlayer_MultiplePlayersUniqueness(t *testing.T) {
	// Create multiple players and ensure they have unique timestamps
	players := make([]*Player, 5)
	for i := 0; i < 5; i++ {
		players[i] = NewPlayer(
			string(rune('a'+i)),
			"Player"+string(rune('A'+i)),
			"session-"+string(rune('1'+i)),
		)
		// Small delay to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// Verify all players have unique IDs and timestamps
	seen := make(map[string]bool)
	seenTimes := make(map[time.Time]bool)

	for i, p := range players {
		if seen[p.ID] {
			t.Errorf("Duplicate ID found: %v", p.ID)
		}
		seen[p.ID] = true

		// While timestamps might theoretically be the same, they should generally be different
		if seenTimes[p.JoinedAt] {
			t.Logf("Warning: Duplicate timestamp found at index %d: %v", i, p.JoinedAt)
		}
		seenTimes[p.JoinedAt] = true
	}
}

func TestPlayer_ZeroValues(t *testing.T) {
	// Test with empty/zero values
	player := NewPlayer("", "", "")

	if player.ID != "" {
		t.Errorf("ID with empty string = %v, want empty", player.ID)
	}
	if player.Name != "" {
		t.Errorf("Name with empty string = %v, want empty", player.Name)
	}
	if player.SessionID != "" {
		t.Errorf("SessionID with empty string = %v, want empty", player.SessionID)
	}

	// Even with zero values, JoinedAt should be set
	if player.JoinedAt.IsZero() {
		t.Errorf("JoinedAt should not be zero even with empty inputs")
	}
}

// Benchmark for performance testing
func BenchmarkNewPlayer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPlayer("player-bench", "BenchPlayer", "session-bench")
	}
}
