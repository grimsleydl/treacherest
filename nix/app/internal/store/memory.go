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
	mu          sync.RWMutex
	rooms       map[string]*game.Room
	config      *config.ServerConfig
	cardService *game.CardService
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore(cfg *config.ServerConfig) *MemoryStore {
	return &MemoryStore{
		rooms:  make(map[string]*game.Room),
		config: cfg,
	}
}

// SetCardService sets the card service for the store
func (s *MemoryStore) SetCardService(cardService *game.CardService) {
	s.cardService = cardService
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

	// Create default role configuration using standard preset with default game size
	roleService := game.NewRoleConfigService(s.config)
	roleService.SetCardService(s.cardService)
	roleConfig, _ := roleService.CreateFromPreset("standard", s.config.Server.DefaultGameSize)

	// Initialize card pool with all available cards
	var allCards []*game.Card
	if s.cardService != nil {
		allCards = s.cardService.GetAllCards()
	}

	room := &game.Room{
		Code:               code,
		State:              game.StateLobby,
		RulesMode:          game.RulesModeTreachery,
		Players:            make(map[string]*game.Player),
		CreatedAt:          time.Now(),
		MaxPlayers:         s.config.Server.MaxPlayersPerRoom,
		RoleConfig:         roleConfig,
		CardPool:           game.NewCardPool(allCards),
		RoleOptionsManager: game.NewRoleOptionsManager(),
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

	// For now, we don't need to fix anything with the new structure
	// The validation is handled by the room's ValidateRoleConfig method
	return false
}

// RegisterRestoredRoom registers a room that was restored from a backup
// This is used when recovering from a Cloud Run instance replacement
func (s *MemoryStore) RegisterRestoredRoom(room *game.Room) error {
	if room == nil {
		return fmt.Errorf("room is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if room already exists (another player might have restored it first)
	if _, exists := s.rooms[room.Code]; exists {
		return fmt.Errorf("room %s already exists", room.Code)
	}

	// Reinitialize CardPool with current cards from CardService
	// The backup may have outdated card data, so we need fresh references
	if s.cardService != nil {
		allCards := s.cardService.GetAllCards()
		room.CardPool = game.NewCardPool(allCards)

		// Re-mark assigned cards based on player roles
		for _, player := range room.Players {
			if player.Role != nil {
				// Find the matching card in the fresh card pool
				freshCard := room.CardPool.GetCardByID(player.Role.ID)
				if freshCard != nil {
					room.CardPool.MarkCardAssigned(freshCard.ID)
					// Update player's role reference to the fresh card
					player.Role = freshCard
				}
			}
		}
	}

	// Initialize RoleOptionsManager if nil
	if room.RoleOptionsManager == nil {
		room.RoleOptionsManager = game.NewRoleOptionsManager()
	}

	s.rooms[room.Code] = room
	return nil
}

// RoomExists checks if a room with the given code exists
func (s *MemoryStore) RoomExists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.rooms[code]
	return exists
}

// DeleteRoom removes a room from the store (used for debug/testing)
func (s *MemoryStore) DeleteRoom(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, code)
}
