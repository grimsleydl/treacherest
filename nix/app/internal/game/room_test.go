package game

import (
	"testing"
)

func TestRoom_AddPlayer(t *testing.T) {
	t.Run("successfully adds new player", func(t *testing.T) {
		room := &Room{
			Code:       "TEST1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		player := NewPlayer("p1", "Alice", "session1")

		err := room.AddPlayer(player)
		if err != nil {
			t.Errorf("Failed to add player: %v", err)
		}

		if len(room.Players) != 1 {
			t.Errorf("Expected 1 player, got %d", len(room.Players))
		}

		if room.GetPlayer("p1") == nil {
			t.Error("Player not found after adding")
		}
	})

	t.Run("rejects duplicate names", func(t *testing.T) {
		room := &Room{
			Code:       "TEST1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add first player
		player1 := NewPlayer("p1", "Alice", "session1")
		err := room.AddPlayer(player1)
		if err != nil {
			t.Errorf("Failed to add first player: %v", err)
		}

		// Try to add second player with same name
		player2 := NewPlayer("p2", "Alice", "session2")
		err = room.AddPlayer(player2)
		if err != ErrDuplicateName {
			t.Errorf("Expected ErrDuplicateName, got %v", err)
		}

		// Verify only one player in room
		if len(room.Players) != 1 {
			t.Errorf("Expected 1 player, got %d", len(room.Players))
		}
	})

	t.Run("rejects duplicate names case-insensitive", func(t *testing.T) {
		room := &Room{
			Code:       "TEST1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add first player with lowercase
		player1 := NewPlayer("p1", "alice", "session1")
		err := room.AddPlayer(player1)
		if err != nil {
			t.Errorf("Failed to add first player: %v", err)
		}

		// Try to add second player with uppercase
		player2 := NewPlayer("p2", "ALICE", "session2")
		err = room.AddPlayer(player2)
		if err != ErrDuplicateName {
			t.Errorf("Expected ErrDuplicateName for case-insensitive match, got %v", err)
		}

		// Try mixed case
		player3 := NewPlayer("p3", "Alice", "session3")
		err = room.AddPlayer(player3)
		if err != ErrDuplicateName {
			t.Errorf("Expected ErrDuplicateName for mixed case, got %v", err)
		}

		// Verify still only one player
		if len(room.Players) != 1 {
			t.Errorf("Expected 1 player, got %d", len(room.Players))
		}
	})

	t.Run("allows different names", func(t *testing.T) {
		room := &Room{
			Code:       "TEST1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add multiple players with different names
		names := []string{"Alice", "Bob", "Charlie", "Dave"}
		for i, name := range names {
			player := NewPlayer(string(rune('a'+i)), name, "session"+string(rune('1'+i)))
			err := room.AddPlayer(player)
			if err != nil {
				t.Errorf("Failed to add player %s: %v", name, err)
			}
		}

		// Verify all players added
		if len(room.Players) != 4 {
			t.Errorf("Expected 4 players, got %d", len(room.Players))
		}
	})
}

func TestRoom_CanStart(t *testing.T) {
	room := &Room{
		Code:       "TEST1",
		State:      StateLobby,
		Players:    make(map[string]*Player),
		MaxPlayers: 4,
	}

	// Empty room should not be able to start
	if room.CanStart() {
		t.Error("Room should not be able to start with 0 players")
	}

	// Add 1 player - should be able to start
	player := NewPlayer("a", "Player", "session")
	room.AddPlayer(player)

	if !room.CanStart() {
		t.Error("Room should be able to start with 1 player")
	}
}

func TestRoom_MaxPlayers(t *testing.T) {
	room := &Room{
		Code:       "TEST1",
		State:      StateLobby,
		Players:    make(map[string]*Player),
		MaxPlayers: 4,
	}

	// Fill room
	for i := 0; i < 4; i++ {
		player := NewPlayer(string(rune('a'+i)), "Player"+string(rune('1'+i)), "session"+string(rune('1'+i)))
		err := room.AddPlayer(player)
		if err != nil {
			t.Errorf("Failed to add player %d: %v", i, err)
		}
	}

	// Try to add one more
	player := NewPlayer("e", "Extra", "session5")
	err := room.AddPlayer(player)
	if err != ErrRoomFull {
		t.Error("Expected ErrRoomFull when adding player to full room")
	}
}

func TestRoom_MaxPlayersWithHost(t *testing.T) {
	room := &Room{
		Code:       "TEST1",
		State:      StateLobby,
		Players:    make(map[string]*Player),
		MaxPlayers: 4,
	}

	// Add a host
	host := NewPlayer("host", "Host", "session-host")
	host.IsHost = true
	err := room.AddPlayer(host)
	if err != nil {
		t.Errorf("Failed to add host: %v", err)
	}

	// Should still be able to add 4 regular players
	for i := 0; i < 4; i++ {
		player := NewPlayer(string(rune('a'+i)), "Player"+string(rune('1'+i)), "session"+string(rune('1'+i)))
		err := room.AddPlayer(player)
		if err != nil {
			t.Errorf("Failed to add player %d with host present: %v", i, err)
		}
	}

	// Room should have 5 total (1 host + 4 players)
	if len(room.Players) != 5 {
		t.Errorf("Expected 5 total players (1 host + 4 players), got %d", len(room.Players))
	}

	// Active player count should be 4
	if room.GetActivePlayerCount() != 4 {
		t.Errorf("Expected 4 active players, got %d", room.GetActivePlayerCount())
	}

	// Try to add one more regular player - should fail
	player := NewPlayer("e", "Extra", "session5")
	err = room.AddPlayer(player)
	if err != ErrRoomFull {
		t.Error("Expected ErrRoomFull when adding 5th player to room with host")
	}

	// But should be able to add another host
	host2 := NewPlayer("host2", "Host2", "session-host2")
	host2.IsHost = true
	err = room.AddPlayer(host2)
	if err != nil {
		t.Errorf("Should be able to add another host: %v", err)
	}
}
