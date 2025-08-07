package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
)

// MemoryStore holds all game state in memory
type MemoryStore struct {
	mu     sync.RWMutex
	rooms  map[string]*game.Room
	config *config.ServerConfig
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore(cfg *config.ServerConfig) *MemoryStore {
	return &MemoryStore{
		rooms:  make(map[string]*game.Room),
		config: cfg,
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

	// Create default role configuration using standard preset
	roleService := game.NewRoleConfigService(s.config)
	roleConfig, _ := roleService.CreateFromPreset("standard", s.config.Server.MaxPlayersPerRoom)

	room := &game.Room{
		Code:       code,
		State:      game.StateLobby,
		Players:    make(map[string]*game.Player),
		CreatedAt:  time.Now(),
		MaxPlayers: s.config.Server.MaxPlayersPerRoom,
		RoleConfig: roleConfig,
	}

	s.rooms[code] = room
	return room, nil
}

// GetRoom retrieves a room by code
func (s *MemoryStore) GetRoom(code string) (*game.Room, error) {
	s.mu.RLock()
	room, exists := s.rooms[code]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("room %s not found", code)
	}

	// Validate and fix role configuration if needed
	if s.validateAndFixRoleConfig(room) {
		// Save the fixed configuration
		s.UpdateRoom(room)
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

// validateAndFixRoleConfig checks and fixes invalid role configurations
// Returns true if any fixes were made
func (s *MemoryStore) validateAndFixRoleConfig(room *game.Room) bool {
	if room.RoleConfig == nil {
		return false
	}

	fixed := false

	// Check each enabled role has proper count
	for role, enabled := range room.RoleConfig.EnabledRoles {
		if !enabled {
			continue
		}

		count, hasCount := room.RoleConfig.RoleCounts[role]
		if roleDef, exists := s.config.Roles.Available[role]; exists {
			// Fix missing or invalid counts
			if !hasCount || count < roleDef.MinCount {
				if roleDef.MinCount > 0 {
					room.RoleConfig.RoleCounts[role] = roleDef.MinCount
				} else {
					room.RoleConfig.RoleCounts[role] = 1
				}
				fixed = true
				// Log the fix for debugging
				fmt.Printf("Fixed role %s count from %d to %d in room %s\n", 
					role, count, room.RoleConfig.RoleCounts[role], room.Code)
			}
		}
	}

	// Specifically check for leader role
	if room.RoleConfig.EnabledRoles["leader"] {
		if count, exists := room.RoleConfig.RoleCounts["leader"]; !exists || count < 1 {
			room.RoleConfig.RoleCounts["leader"] = 1
			fixed = true
			fmt.Printf("Fixed missing leader in room %s\n", room.Code)
		}
	}

	return fixed
}
