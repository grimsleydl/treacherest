package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"treacherest/internal/game"
	"treacherest/internal/store"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	store       *store.MemoryStore
	eventBus    *EventBus
	cardService *game.CardService
}

// New creates a new handler
func New(store *store.MemoryStore) *Handler {
	// Create CardService
	cardService, err := game.NewCardService()
	if err != nil {
		// Log error but continue with nil cardService for now
		// In production, this should probably return an error
		cardService = nil
	}

	return &Handler{
		store:       store,
		eventBus:    NewEventBus(),
		cardService: cardService,
	}
}

// Store returns the handler's store (for testing)
func (h *Handler) Store() *store.MemoryStore {
	return h.store
}

// Event represents a game event
type Event struct {
	Type     string
	RoomCode string
	Data     interface{}
}

// EventBus manages event subscriptions
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

// Subscribe subscribes to events for a room
func (eb *EventBus) Subscribe(roomCode string) chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 10)
	eb.subscribers[roomCode] = append(eb.subscribers[roomCode], ch)
	return ch
}

// Unsubscribe removes a subscription
func (eb *EventBus) Unsubscribe(roomCode string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs := eb.subscribers[roomCode]
	for i, sub := range subs {
		if sub == ch {
			eb.subscribers[roomCode] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, ch := range eb.subscribers[event.RoomCode] {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// getOrCreateSession gets or creates a session for the user
func getOrCreateSession(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err == nil {
		return cookie.Value
	}

	// Create new session
	b := make([]byte, 16)
	rand.Read(b)
	sessionID := hex.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})

	return sessionID
}

// generatePlayerID generates a unique player ID
func generatePlayerID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
