package store

import (
	"sync"
	"testing"
	"time"
	"treacherest/internal/game"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	if store == nil {
		t.Fatal("NewMemoryStore returned nil")
	}

	if store.rooms == nil {
		t.Fatal("rooms map not initialized")
	}

	if len(store.rooms) != 0 {
		t.Errorf("expected empty rooms map, got %d rooms", len(store.rooms))
	}
}

func TestCreateRoom(t *testing.T) {
	store := NewMemoryStore()

	t.Run("creates room with unique code", func(t *testing.T) {
		room, err := store.CreateRoom()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if room == nil {
			t.Fatal("CreateRoom returned nil room")
		}

		if room.Code == "" {
			t.Error("room code is empty")
		}

		if len(room.Code) != 5 {
			t.Errorf("expected room code length 5, got %d", len(room.Code))
		}

		// Verify room code contains only alphanumeric characters
		for _, char := range room.Code {
			if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				t.Errorf("room code contains invalid character: %c", char)
			}
		}
	})

	t.Run("creates room with correct initial state", func(t *testing.T) {
		room, err := store.CreateRoom()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if room.State != game.StateLobby {
			t.Errorf("expected state %s, got %s", game.StateLobby, room.State)
		}

		if room.Players == nil {
			t.Error("players map not initialized")
		}

		if len(room.Players) != 0 {
			t.Errorf("expected empty players map, got %d players", len(room.Players))
		}

		if room.MaxPlayers != 8 {
			t.Errorf("expected max players 8, got %d", room.MaxPlayers)
		}

		if room.CreatedAt.IsZero() {
			t.Error("CreatedAt not set")
		}
	})

	t.Run("creates multiple rooms with unique codes", func(t *testing.T) {
		codes := make(map[string]bool)

		for i := 0; i < 100; i++ {
			room, err := store.CreateRoom()
			if err != nil {
				t.Fatalf("unexpected error on iteration %d: %v", i, err)
			}

			if codes[room.Code] {
				t.Errorf("duplicate room code generated: %s", room.Code)
			}
			codes[room.Code] = true
		}
	})

	t.Run("stores room in internal map", func(t *testing.T) {
		room, err := store.CreateRoom()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Access internal map to verify storage
		store.mu.RLock()
		storedRoom, exists := store.rooms[room.Code]
		store.mu.RUnlock()

		if !exists {
			t.Error("room not stored in internal map")
		}

		if storedRoom != room {
			t.Error("stored room is not the same instance")
		}
	})
}

func TestGetRoom(t *testing.T) {
	store := NewMemoryStore()

	t.Run("returns error for non-existent room", func(t *testing.T) {
		_, err := store.GetRoom("ABCDE")
		if err == nil {
			t.Error("expected error for non-existent room")
		}

		expectedErr := "room ABCDE not found"
		if err.Error() != expectedErr {
			t.Errorf("expected error %q, got %q", expectedErr, err.Error())
		}
	})

	t.Run("returns existing room", func(t *testing.T) {
		// Create a room first
		createdRoom, err := store.CreateRoom()
		if err != nil {
			t.Fatalf("failed to create room: %v", err)
		}

		// Get the room
		retrievedRoom, err := store.GetRoom(createdRoom.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrievedRoom != createdRoom {
			t.Error("retrieved room is not the same instance")
		}
	})

	t.Run("returns correct room among multiple rooms", func(t *testing.T) {
		// Create multiple rooms
		rooms := make([]*game.Room, 10)
		for i := range rooms {
			room, err := store.CreateRoom()
			if err != nil {
				t.Fatalf("failed to create room %d: %v", i, err)
			}
			rooms[i] = room
		}

		// Retrieve each room and verify
		for i, expectedRoom := range rooms {
			retrievedRoom, err := store.GetRoom(expectedRoom.Code)
			if err != nil {
				t.Errorf("failed to get room %d: %v", i, err)
				continue
			}

			if retrievedRoom != expectedRoom {
				t.Errorf("room %d: retrieved room is not the same instance", i)
			}
		}
	})
}

func TestUpdateRoom(t *testing.T) {
	store := NewMemoryStore()

	t.Run("updates existing room", func(t *testing.T) {
		// Create a room
		room, err := store.CreateRoom()
		if err != nil {
			t.Fatalf("failed to create room: %v", err)
		}

		// Modify the room
		room.State = game.StateCountdown
		room.Players["player1"] = &game.Player{
			ID:   "player1",
			Name: "Test Player",
		}

		// Update the room
		err = store.UpdateRoom(room)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify the update
		retrievedRoom, err := store.GetRoom(room.Code)
		if err != nil {
			t.Fatalf("failed to get room: %v", err)
		}

		if retrievedRoom.State != game.StateCountdown {
			t.Errorf("expected state %s, got %s", game.StateCountdown, retrievedRoom.State)
		}

		if len(retrievedRoom.Players) != 1 {
			t.Errorf("expected 1 player, got %d", len(retrievedRoom.Players))
		}
	})

	t.Run("can update non-existent room", func(t *testing.T) {
		// This is current behavior - UpdateRoom doesn't check if room exists
		room := &game.Room{
			Code:       "NEWRM",
			State:      game.StatePlaying,
			Players:    make(map[string]*game.Player),
			CreatedAt:  time.Now(),
			MaxPlayers: 6,
		}

		err := store.UpdateRoom(room)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify room was added
		retrievedRoom, err := store.GetRoom("NEWRM")
		if err != nil {
			t.Fatalf("failed to get room: %v", err)
		}

		if retrievedRoom.State != game.StatePlaying {
			t.Errorf("expected state %s, got %s", game.StatePlaying, retrievedRoom.State)
		}
	})
}

func TestGenerateRoomCode(t *testing.T) {
	t.Run("generates 5 character code", func(t *testing.T) {
		code := generateRoomCode()
		if len(code) != 5 {
			t.Errorf("expected code length 5, got %d", len(code))
		}
	})

	t.Run("generates alphanumeric code", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			code := generateRoomCode()
			for _, char := range code {
				if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
					t.Errorf("invalid character in code: %c", char)
				}
			}
		}
	})

	t.Run("generates different codes", func(t *testing.T) {
		codes := make(map[string]int)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			code := generateRoomCode()
			codes[code]++
		}

		// With 36^5 possible codes, duplicates are extremely unlikely in 1000 iterations
		duplicates := 0
		for code, count := range codes {
			if count > 1 {
				duplicates++
				t.Logf("Code %s appeared %d times", code, count)
			}
		}

		// Allow for a small number of duplicates due to randomness
		if duplicates > 5 {
			t.Errorf("too many duplicate codes: %d duplicates in %d iterations", duplicates, iterations)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	t.Run("concurrent room creation", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 100

		rooms := make([]*game.Room, numGoroutines)
		errors := make([]error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				room, err := store.CreateRoom()
				rooms[idx] = room
				errors[idx] = err
			}(i)
		}

		wg.Wait()

		// Check for errors
		for i, err := range errors {
			if err != nil {
				t.Errorf("goroutine %d got error: %v", i, err)
			}
		}

		// Verify all rooms were created with unique codes
		codes := make(map[string]bool)
		for i, room := range rooms {
			if room == nil {
				t.Errorf("goroutine %d created nil room", i)
				continue
			}
			if codes[room.Code] {
				t.Errorf("duplicate room code: %s", room.Code)
			}
			codes[room.Code] = true
		}
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		// Create some initial rooms
		var rooms []*game.Room
		for i := 0; i < 10; i++ {
			room, err := store.CreateRoom()
			if err != nil {
				t.Fatalf("failed to create initial room: %v", err)
			}
			rooms = append(rooms, room)
		}

		var wg sync.WaitGroup

		// Readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					room := rooms[j%len(rooms)]
					_, err := store.GetRoom(room.Code)
					if err != nil {
						t.Errorf("failed to get room: %v", err)
					}
				}
			}()
		}

		// Writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					room := rooms[j%len(rooms)]
					room.State = game.StateCountdown
					err := store.UpdateRoom(room)
					if err != nil {
						t.Errorf("failed to update room: %v", err)
					}
				}
			}()
		}

		// Creators
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					_, err := store.CreateRoom()
					if err != nil {
						t.Errorf("failed to create room: %v", err)
					}
				}
			}()
		}

		wg.Wait()
	})
}

func TestMemoryStoreEdgeCases(t *testing.T) {
	t.Run("handles empty room code in GetRoom", func(t *testing.T) {
		store := NewMemoryStore()
		_, err := store.GetRoom("")
		if err == nil {
			t.Error("expected error for empty room code")
		}
	})

	t.Run("handles nil room in UpdateRoom", func(t *testing.T) {
		store := NewMemoryStore()
		// This will panic in current implementation
		// Documenting expected behavior
		defer func() {
			if r := recover(); r != nil {
				t.Logf("UpdateRoom with nil room panics as expected: %v", r)
			}
		}()

		err := store.UpdateRoom(nil)
		if err == nil {
			t.Error("expected error or panic for nil room")
		}
	})
}
