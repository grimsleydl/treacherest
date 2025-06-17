package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
	"treacherest/internal/game"
)

// MemoryStore holds all game state in memory
type MemoryStore struct {
	mu    sync.RWMutex
	rooms map[string]*game.Room
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		rooms: make(map[string]*game.Room),
	}
}

// CreateRoom creates a new game room
func (s *MemoryStore) CreateRoom() (*game.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate unique room code
	var code string
	for i := 0; i < 10; i++ { // Try up to 10 times
		code = generateRoomCode()
		if _, exists := s.rooms[code]; !exists {
			break
		}
	}

	room := &game.Room{
		Code:       code,
		State:      game.StateLobby,
		Players:    make(map[string]*game.Player),
		CreatedAt:  time.Now(),
		MaxPlayers: 8,
	}

	s.rooms[code] = room
	return room, nil
}

// GetRoom retrieves a room by code
func (s *MemoryStore) GetRoom(code string) (*game.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, exists := s.rooms[code]
	if !exists {
		return nil, fmt.Errorf("room %s not found", code)
	}

	return room, nil
}

// UpdateRoom updates a room
func (s *MemoryStore) UpdateRoom(room *game.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rooms[room.Code] = room
	return nil
}

// generateRoomCode generates a 5-character alphanumeric code
func generateRoomCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 5)
	rand.Read(b)

	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}

	return string(b)
}
