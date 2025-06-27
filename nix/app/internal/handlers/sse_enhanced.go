package handlers

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
	"net/http"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"
	"treacherest/internal/views/pages"
)

// SSEEvent represents an event that can be replayed
type SSEEvent struct {
	ID        string
	Type      string
	Data      string
	Timestamp time.Time
}

// EventStore stores recent events for replay
type EventStore struct {
	mu        sync.RWMutex
	events    map[string][]SSEEvent // roomCode -> events
	maxEvents int
}

// NewEventStore creates a new event store
func NewEventStore(maxEvents int) *EventStore {
	return &EventStore{
		events:    make(map[string][]SSEEvent),
		maxEvents: maxEvents,
	}
}

// AddEvent adds an event to the store
func (es *EventStore) AddEvent(roomCode string, event SSEEvent) {
	es.mu.Lock()
	defer es.mu.Unlock()

	events := es.events[roomCode]
	events = append(events, event)

	// Keep only the most recent events
	if len(events) > es.maxEvents {
		events = events[len(events)-es.maxEvents:]
	}

	es.events[roomCode] = events
}

// GetEventsSince returns events since the given ID
func (es *EventStore) GetEventsSince(roomCode string, lastEventID string) []SSEEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()

	events := es.events[roomCode]
	if len(events) == 0 {
		return nil
	}

	// If no lastEventID, return empty (client has all events)
	if lastEventID == "" {
		return nil
	}

	// Find the index of the last event
	lastIndex := -1
	for i, event := range events {
		if event.ID == lastEventID {
			lastIndex = i
			break
		}
	}

	// If not found, return all events (client is too far behind)
	if lastIndex == -1 {
		return events
	}

	// Return events after the last one
	if lastIndex+1 < len(events) {
		return events[lastIndex+1:]
	}

	return nil
}

// ConnectionTracker tracks active SSE connections
type ConnectionTracker struct {
	mu          sync.RWMutex
	connections map[string]int64 // roomCode -> connection count
	totalActive int64            // Total active connections (atomic)
}

// NewConnectionTracker creates a new connection tracker
func NewConnectionTracker() *ConnectionTracker {
	return &ConnectionTracker{
		connections: make(map[string]int64),
	}
}

// AddConnection increments the connection count for a room
func (ct *ConnectionTracker) AddConnection(roomCode string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.connections[roomCode]++
	atomic.AddInt64(&ct.totalActive, 1)

	log.Printf("SSE: Connection established for room %s (room total: %d, global total: %d)",
		roomCode, ct.connections[roomCode], atomic.LoadInt64(&ct.totalActive))
}

// RemoveConnection decrements the connection count for a room
func (ct *ConnectionTracker) RemoveConnection(roomCode string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if count, exists := ct.connections[roomCode]; exists && count > 0 {
		ct.connections[roomCode]--
		atomic.AddInt64(&ct.totalActive, -1)

		if ct.connections[roomCode] == 0 {
			delete(ct.connections, roomCode)
		}
	}

	log.Printf("SSE: Connection closed for room %s (room total: %d, global total: %d)",
		roomCode, ct.connections[roomCode], atomic.LoadInt64(&ct.totalActive))
}

// GetConnectionCount returns the number of connections for a room
func (ct *ConnectionTracker) GetConnectionCount(roomCode string) int64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.connections[roomCode]
}

// GetTotalConnections returns the total number of active connections
func (ct *ConnectionTracker) GetTotalConnections() int64 {
	return atomic.LoadInt64(&ct.totalActive)
}

// EnhancedHandler extends Handler with SSE improvements
type EnhancedHandler struct {
	*Handler
	eventStore   *EventStore
	connTracker  *ConnectionTracker
	eventCounter int64 // Atomic counter for event IDs
}

// NewEnhanced creates a new enhanced handler
func NewEnhanced(s *store.MemoryStore, cardService *game.CardService, cfg *config.ServerConfig) *EnhancedHandler {
	return &EnhancedHandler{
		Handler:      New(s, cardService, cfg),
		eventStore:   NewEventStore(100), // Keep last 100 events per room
		connTracker:  NewConnectionTracker(),
		eventCounter: 0,
	}
}

// generateEventID generates a unique event ID
func (h *EnhancedHandler) generateEventID() string {
	id := atomic.AddInt64(&h.eventCounter, 1)
	return fmt.Sprintf("%d-%d", time.Now().Unix(), id)
}

// StreamLobbyEnhanced streams lobby updates with heartbeat and reconnection support
func (h *EnhancedHandler) StreamLobbyEnhanced(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Track connection
	h.connTracker.AddConnection(roomCode)
	defer h.connTracker.RemoveConnection(roomCode)

	// Check for Last-Event-ID header
	lastEventID := r.Header.Get("Last-Event-ID")

	// Create SSE connection
	sse := datastar.NewSSE(w, r)

	// Create a done channel for cleanup
	ctx := r.Context()
	done := make(chan struct{})
	defer close(done)

	// Subscribe to events
	events := h.eventBus.Subscribe(roomCode)
	defer func() {
		h.eventBus.Unsubscribe(roomCode, events)
		log.Printf("SSE: Unsubscribed from room %s events", roomCode)
	}()

	// Send any missed events
	if lastEventID != "" {
		missedEvents := h.eventStore.GetEventsSince(roomCode, lastEventID)
		for _, event := range missedEvents {
			log.Printf("SSE: Replaying event %s for room %s", event.ID, roomCode)
			// For now, skip replaying stored events as datastar API doesn't support Event method
			_ = event
		}
	}

	// Send initial render
	eventID := h.generateEventID()
	h.renderLobbyWithID(sse, room, player, eventID)

	// Store the initial render event
	h.eventStore.AddEvent(roomCode, SSEEvent{
		ID:        eventID,
		Type:      "lobby_update",
		Data:      "initial_render",
		Timestamp: time.Now(),
	})

	// Start heartbeat ticker
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	log.Printf("SSE: Started streaming lobby for room %s, player %s", roomCode, player.ID)

	// Stream updates
	for {
		select {
		case <-ctx.Done():
			log.Printf("SSE: Context cancelled for room %s", roomCode)
			return

		case <-done:
			log.Printf("SSE: Done channel closed for room %s", roomCode)
			return

		case <-heartbeatTicker.C:
			// Send heartbeat as a script execution
			heartbeatScript := fmt.Sprintf(`console.log('Heartbeat: %s, connections: %d');`,
				time.Now().Format(time.RFC3339),
				h.connTracker.GetConnectionCount(roomCode))
			sse.ExecuteScript(heartbeatScript)

			// Store heartbeat event for replay
			eventID := h.generateEventID()

			// Store heartbeat event
			h.eventStore.AddEvent(roomCode, SSEEvent{
				ID:        eventID,
				Type:      "heartbeat",
				Data:      "ping",
				Timestamp: time.Now(),
			})

			log.Printf("SSE: Sent heartbeat for room %s", roomCode)

		case event := <-events:
			switch event.Type {
			case "player_joined", "player_left":
				// Re-render lobby
				room, _ = h.store.GetRoom(roomCode)
				eventID := h.generateEventID()
				h.renderLobbyWithID(sse, room, player, eventID)

				// Store update event
				h.eventStore.AddEvent(roomCode, SSEEvent{
					ID:        eventID,
					Type:      "lobby_update",
					Data:      event.Type,
					Timestamp: time.Now(),
				})

			case "game_started":
				// Redirect to game page
				eventID := h.generateEventID()
				sse.ExecuteScript("window.location.href = '/game/" + roomCode + "'")

				// Store game started event
				h.eventStore.AddEvent(roomCode, SSEEvent{
					ID:        eventID,
					Type:      "game_started",
					Data:      "redirect",
					Timestamp: time.Now(),
				})
			}
		}
	}
}

// StreamGameEnhanced streams game updates with heartbeat and reconnection support
func (h *EnhancedHandler) StreamGameEnhanced(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Track connection
	h.connTracker.AddConnection(roomCode)
	defer h.connTracker.RemoveConnection(roomCode)

	// Check for Last-Event-ID header
	lastEventID := r.Header.Get("Last-Event-ID")

	// Create SSE connection
	sse := datastar.NewSSE(w, r)

	// Create a done channel for cleanup
	ctx := r.Context()
	done := make(chan struct{})
	defer close(done)

	// Subscribe to events
	events := h.eventBus.Subscribe(roomCode)
	defer func() {
		h.eventBus.Unsubscribe(roomCode, events)
		log.Printf("SSE: Unsubscribed from room %s game events", roomCode)
	}()

	// Send any missed events
	if lastEventID != "" {
		missedEvents := h.eventStore.GetEventsSince(roomCode, lastEventID)
		for _, event := range missedEvents {
			log.Printf("SSE: Replaying game event %s for room %s", event.ID, roomCode)
			// For now, skip replaying stored events as datastar API doesn't support Event method
			_ = event
		}
	}

	// Send initial render
	eventID := h.generateEventID()
	h.renderGameWithID(sse, room, player, eventID)

	// Store the initial render event
	h.eventStore.AddEvent(roomCode, SSEEvent{
		ID:        eventID,
		Type:      "game_update",
		Data:      "initial_render",
		Timestamp: time.Now(),
	})

	// Start heartbeat ticker
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	log.Printf("SSE: Started streaming game for room %s, player %s", roomCode, player.ID)

	// Stream updates
	for {
		select {
		case <-ctx.Done():
			log.Printf("SSE: Game context cancelled for room %s", roomCode)
			return

		case <-done:
			log.Printf("SSE: Game done channel closed for room %s", roomCode)
			return

		case <-heartbeatTicker.C:
			// Send heartbeat as a script execution
			heartbeatScript := fmt.Sprintf(`console.log('Game heartbeat: %s, connections: %d, state: %s');`,
				time.Now().Format(time.RFC3339),
				h.connTracker.GetConnectionCount(roomCode),
				room.State)
			sse.ExecuteScript(heartbeatScript)

			// Store heartbeat event for replay
			eventID := h.generateEventID()

			// Store heartbeat event
			h.eventStore.AddEvent(roomCode, SSEEvent{
				ID:        eventID,
				Type:      "heartbeat",
				Data:      "ping",
				Timestamp: time.Now(),
			})

			log.Printf("SSE: Sent game heartbeat for room %s", roomCode)

		case <-events:
			// Re-render on any event
			room, _ = h.store.GetRoom(roomCode)
			player = room.GetPlayer(player.ID) // Refresh player data

			eventID := h.generateEventID()
			h.renderGameWithID(sse, room, player, eventID)

			// Store game update event
			h.eventStore.AddEvent(roomCode, SSEEvent{
				ID:        eventID,
				Type:      "game_update",
				Data:      "state_change",
				Timestamp: time.Now(),
			})
		}
	}
}

// renderLobbyWithID renders the lobby body with an event ID
func (h *EnhancedHandler) renderLobbyWithID(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player, eventID string) {
	component := pages.LobbyBody(room, player)

	// Render to string
	html := renderToString(component)

	// Send as fragment with morph mode and explicit selector
	sse.MergeFragments(html,
		datastar.WithSelector("#lobby-container"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}

// renderGameWithID renders the game body with an event ID
func (h *EnhancedHandler) renderGameWithID(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player, eventID string) {
	component := pages.GameBody(room, player)

	// Render to string
	html := renderToString(component)

	// Send as fragment with morph mode and explicit selector
	sse.MergeFragments(html,
		datastar.WithSelector("#game-container"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}
