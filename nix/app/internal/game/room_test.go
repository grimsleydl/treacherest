package game

import (
	"testing"
)

func TestRoom_AddPlayer(t *testing.T) {
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
		player := NewPlayer(string(rune('a'+i)), "Player", "session")
		err := room.AddPlayer(player)
		if err != nil {
			t.Errorf("Failed to add player %d: %v", i, err)
		}
	}

	// Try to add one more
	player := NewPlayer("e", "Extra", "session")
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
		player := NewPlayer(string(rune('a'+i)), "Player", "session")
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
	player := NewPlayer("e", "Extra", "session")
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
