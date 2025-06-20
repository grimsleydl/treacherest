package game

import (
	"sync"
	"testing"
	"time"
)

func TestRoom_StateTransitions(t *testing.T) {
	t.Run("lobby to countdown transition", func(t *testing.T) {
		room := &Room{
			Code:       "TEST1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add 1 player (minimum to start)
		player := NewPlayer("a", "Player", "session")
		room.AddPlayer(player)

		// Verify we can start
		if !room.CanStart() {
			t.Error("Should be able to start with 1 player")
		}

		// Transition to countdown
		room.mu.Lock()
		room.State = StateCountdown
		room.CountdownRemaining = 5
		room.mu.Unlock()

		// Verify we can't start when not in lobby
		if room.CanStart() {
			t.Error("Should not be able to start when in countdown state")
		}
	})

	t.Run("countdown to playing transition", func(t *testing.T) {
		room := &Room{
			Code:               "TEST2",
			State:              StateCountdown,
			Players:            make(map[string]*Player),
			MaxPlayers:         8,
			CountdownRemaining: 1,
		}

		// Add players with roles
		players := []*Player{
			{ID: "p1", Name: "Alice", Role: LeaderRole},
			{ID: "p2", Name: "Bob", Role: GuardianRole},
			{ID: "p3", Name: "Charlie", Role: GuardianRole},
			{ID: "p4", Name: "Dave", Role: TraitorRole},
		}

		for _, p := range players {
			room.Players[p.ID] = p
		}

		// Transition to playing
		room.mu.Lock()
		room.State = StatePlaying
		room.CountdownRemaining = 0
		room.StartedAt = time.Now()
		room.mu.Unlock()

		// Check leader
		leader := room.GetLeader()
		if leader == nil {
			t.Error("Should have a leader in playing state")
		}
		if leader.ID != "p1" {
			t.Errorf("Expected leader ID p1, got %s", leader.ID)
		}
	})

	t.Run("playing to ended transition", func(t *testing.T) {
		room := &Room{
			Code:       "TEST3",
			State:      StatePlaying,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
			StartedAt:  time.Now(),
		}

		// Transition to ended
		room.mu.Lock()
		room.State = StateEnded
		room.mu.Unlock()

		// Verify state
		room.mu.RLock()
		if room.State != StateEnded {
			t.Errorf("Expected state %v, got %v", StateEnded, room.State)
		}
		room.mu.RUnlock()
	})
}

func TestRoom_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent player additions", func(t *testing.T) {
		room := &Room{
			Code:       "CONC1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 20,
		}

		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		// Try to add 10 players concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				player := NewPlayer(string(rune('a'+id)), "Player", "session")
				err := room.AddPlayer(player)
				if err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Errorf("Concurrent add failed: %v", err)
		}

		// Verify all players were added
		if len(room.Players) != 10 {
			t.Errorf("Expected 10 players, got %d", len(room.Players))
		}
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		room := &Room{
			Code:       "CONC2",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add some initial players
		for i := 0; i < 4; i++ {
			player := NewPlayer(string(rune('a'+i)), "Player", "session")
			room.AddPlayer(player)
		}

		var wg sync.WaitGroup
		done := make(chan bool)

		// Reader goroutine - continuously read players
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					players := room.GetPlayers()
					_ = len(players) // Use the value
					time.Sleep(time.Microsecond)
				}
			}
		}()

		// Writer goroutine - add and remove players
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if i%2 == 0 {
					player := NewPlayer("temp"+string(rune(i)), "Temp", "session")
					room.AddPlayer(player)
				} else {
					room.RemovePlayer("temp" + string(rune(i-1)))
				}
				time.Sleep(time.Microsecond)
			}
		}()

		// State reader goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = room.CanStart()
					time.Sleep(time.Microsecond)
				}
			}
		}()

		// Let it run for a bit
		time.Sleep(10 * time.Millisecond)
		close(done)
		wg.Wait()
	})
}

func TestRoom_RemovePlayer(t *testing.T) {
	room := &Room{
		Code:       "REM1",
		State:      StateLobby,
		Players:    make(map[string]*Player),
		MaxPlayers: 8,
	}

	// Add players
	players := []*Player{
		NewPlayer("p1", "Alice", "session1"),
		NewPlayer("p2", "Bob", "session2"),
		NewPlayer("p3", "Charlie", "session3"),
	}

	for _, p := range players {
		room.AddPlayer(p)
	}

	// Remove a player
	room.RemovePlayer("p2")

	// Verify removal
	if room.GetPlayer("p2") != nil {
		t.Error("Player p2 should have been removed")
	}

	if len(room.Players) != 2 {
		t.Errorf("Expected 2 players after removal, got %d", len(room.Players))
	}

	// Remove non-existent player (should not panic)
	room.RemovePlayer("nonexistent")
}

func TestRoom_GetPlayers(t *testing.T) {
	room := &Room{
		Code:       "GET1",
		State:      StateLobby,
		Players:    make(map[string]*Player),
		MaxPlayers: 8,
	}

	// Test empty room
	players := room.GetPlayers()
	if len(players) != 0 {
		t.Errorf("Expected 0 players in empty room, got %d", len(players))
	}

	// Add players
	expectedPlayers := []*Player{
		NewPlayer("p1", "Alice", "session1"),
		NewPlayer("p2", "Bob", "session2"),
		NewPlayer("p3", "Charlie", "session3"),
	}

	for _, p := range expectedPlayers {
		room.AddPlayer(p)
	}

	// Get players
	players = room.GetPlayers()
	if len(players) != 3 {
		t.Errorf("Expected 3 players, got %d", len(players))
	}

	// GetPlayers returns a new slice but with pointers to the actual players
	// Find the player with ID "p1" in the returned slice
	var p1FromSlice *Player
	for _, p := range players {
		if p.ID == "p1" {
			p1FromSlice = p
			break
		}
	}

	if p1FromSlice == nil {
		t.Fatal("Could not find player p1 in returned slice")
	}

	// Modifying the player through the slice should affect the original
	p1FromSlice.Name = "Modified"
	originalPlayer := room.GetPlayer("p1")
	if originalPlayer.Name != "Modified" {
		t.Error("GetPlayers returns player pointers for real-time updates")
	}

	// Reset the name for cleanliness
	p1FromSlice.Name = "Alice"
}

func TestRoom_GetLeader(t *testing.T) {
	t.Run("no leader", func(t *testing.T) {
		room := &Room{
			Code:       "LEAD1",
			State:      StatePlaying,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add players without roles
		for i := 0; i < 4; i++ {
			player := NewPlayer(string(rune('a'+i)), "Player", "session")
			room.AddPlayer(player)
		}

		leader := room.GetLeader()
		if leader != nil {
			t.Error("Should return nil when no leader exists")
		}
	})

	t.Run("with leader", func(t *testing.T) {
		room := &Room{
			Code:       "LEAD2",
			State:      StatePlaying,
			Players:    make(map[string]*Player),
			MaxPlayers: 4,
		}

		// Add players with roles
		players := []*Player{
			{ID: "p1", Name: "Alice", Role: GuardianRole},
			{ID: "p2", Name: "Bob", Role: LeaderRole},
			{ID: "p3", Name: "Charlie", Role: TraitorRole},
		}

		for _, p := range players {
			room.Players[p.ID] = p
		}

		leader := room.GetLeader()
		if leader == nil {
			t.Fatal("Should have found a leader")
		}
		if leader.ID != "p2" {
			t.Errorf("Expected leader ID p2, got %s", leader.ID)
		}
	})
}

func TestRoom_EdgeCases(t *testing.T) {
	t.Run("zero max players", func(t *testing.T) {
		room := &Room{
			Code:       "EDGE1",
			State:      StateLobby,
			Players:    make(map[string]*Player),
			MaxPlayers: 0,
		}

		player := NewPlayer("p1", "Alice", "session")
		err := room.AddPlayer(player)
		if err != ErrRoomFull {
			t.Error("Should return ErrRoomFull when MaxPlayers is 0")
		}
	})

	t.Run("can't start with wrong state", func(t *testing.T) {
		states := []GameState{StateCountdown, StatePlaying, StateEnded}

		for _, state := range states {
			room := &Room{
				Code:       "EDGE2",
				State:      state,
				Players:    make(map[string]*Player),
				MaxPlayers: 4,
			}

			// Add 1 player (enough to start)
			player := NewPlayer("a", "Player", "session")
			room.AddPlayer(player)

			if room.CanStart() {
				t.Errorf("Should not be able to start in %v state", state)
			}
		}
	})
}

func TestRoom_Timestamps(t *testing.T) {
	now := time.Now()
	room := &Room{
		Code:       "TIME1",
		State:      StateLobby,
		Players:    make(map[string]*Player),
		MaxPlayers: 8,
		CreatedAt:  now,
	}

	if !room.CreatedAt.Equal(now) {
		t.Error("CreatedAt should be set correctly")
	}

	// Simulate game start
	room.mu.Lock()
	room.State = StatePlaying
	room.StartedAt = time.Now()
	room.mu.Unlock()

	if room.StartedAt.Before(room.CreatedAt) {
		t.Error("StartedAt should be after CreatedAt")
	}
}
