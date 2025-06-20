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
