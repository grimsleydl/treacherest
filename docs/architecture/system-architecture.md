# Treacherest System Architecture

## 1. System Architecture Overview

### High-Level Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          Client Layer                             │
├─────────────────────────────────────────────────────────────────┤
│  Browser                                                          │
│  ├── Templ-rendered HTML (Initial Page Load)                     │
│  ├── Datastar JS (Reactivity & DOM Updates)                      │
│  └── SSE Connection (Real-time Updates)                          │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 │ HTTP/SSE
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Application Layer                         │
├─────────────────────────────────────────────────────────────────┤
│  Chi Router                                                       │
│  ├── Page Handlers (/, /room/{code}, /game/{code})               │
│  ├── Action Handlers (/room/new, /room/{code}/start, etc.)       │
│  └── SSE Handlers (/sse/lobby/{code}, /sse/game/{code})          │
├─────────────────────────────────────────────────────────────────┤
│  Handler Layer (internal/handlers)                               │
│  ├── Handler struct (Dependencies)                               │
│  ├── EventBus (Real-time event distribution)                     │
│  └── Session Management                                          │
├─────────────────────────────────────────────────────────────────┤
│  Business Logic Layer (internal/game)                            │
│  ├── Game State Machine                                          │
│  ├── Role System                                                 │
│  └── Game Rules Engine                                           │
├─────────────────────────────────────────────────────────────────┤
│  View Layer (internal/views)                                     │
│  ├── Templ Templates                                             │
│  ├── Layout Components                                           │
│  └── Page Components                                             │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Storage Layer                            │
├─────────────────────────────────────────────────────────────────┤
│  Memory Store (internal/store)                                   │
│  ├── Room Management                                             │
│  ├── Player Sessions                                             │
│  └── Game State Persistence                                      │
│                                                                   │
│  Future: Redis Store                                             │
│  └── Distributed State Management                                │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow Architecture

```
1. Initial Page Load:
   Browser → Chi Router → Handler → Store → Templ → HTML Response

2. User Actions:
   Browser → Datastar → HTTP POST → Handler → Store → EventBus → SSE

3. Real-time Updates:
   EventBus → SSE Handler → Datastar → DOM Update

4. Game State Changes:
   Game Logic → Store → EventBus → All Connected Clients
```

### Real-time Synchronization Approach

The system uses Server-Sent Events (SSE) for unidirectional real-time updates:

1. **Connection Management**: Each client maintains an SSE connection to their current room
2. **Event Distribution**: EventBus pattern for publishing events to all subscribers
3. **DOM Updates**: Datastar automatically updates DOM based on SSE fragments
4. **Reconnection**: Client-side automatic reconnection with exponential backoff

## 2. Component Design

### API Structure and Endpoints

```
# Page Routes (GET - Return full HTML pages)
GET  /                      # Home page
GET  /room/{code}          # Room lobby page
GET  /game/{code}          # Active game page

# Action Routes (POST - Perform actions, return fragments)
POST /room/new             # Create new room
POST /room/{code}/join     # Join existing room
POST /room/{code}/start    # Start game (host only)
POST /room/{code}/leave    # Leave room
POST /game/{code}/reveal   # Reveal role (game action)
POST /game/{code}/vote     # Vote action

# SSE Routes (GET - Real-time event streams)
GET  /sse/lobby/{code}     # Lobby updates stream
GET  /sse/game/{code}      # Game updates stream
```

### Game State Management

```go
// Centralized state machine for game progression
type GameStateMachine struct {
    transitions map[GameState][]GameState
    handlers    map[GameState]StateHandler
}

// State transitions
Lobby → Countdown → Playing → Ended
     ↖______________|

// Each state has:
- Entry actions
- Valid transitions
- Exit actions
- Allowed player actions
```

### Event System Architecture

```go
// Event types for real-time updates
type EventType string

const (
    // Lobby events
    EventPlayerJoined    EventType = "player_joined"
    EventPlayerLeft      EventType = "player_left"
    EventGameStarting    EventType = "game_starting"
    EventCountdownUpdate EventType = "countdown_update"
    
    // Game events
    EventRolesAssigned   EventType = "roles_assigned"
    EventLeaderRevealed  EventType = "leader_revealed"
    EventPlayerVoted     EventType = "player_voted"
    EventGameEnded       EventType = "game_ended"
)

// Event payload structure
type Event struct {
    Type     EventType
    RoomCode string
    Data     interface{} // Type-specific payload
    Target   string      // Datastar selector
}
```

### Session Management

```go
// Session layer for player identification
type SessionManager struct {
    // Cookie-based sessions (7-day expiry)
    // Maps session ID to player ID
    // Handles reconnection grace periods
}

// Player session lifecycle:
1. Create session on first visit
2. Associate with player when joining room
3. Maintain across page navigation
4. Handle reconnection within grace period
```

## 3. Data Models

### Core Game Entities

```go
// Room - Core game container
type Room struct {
    Code               string
    State              GameState
    Players            map[string]*Player
    MaxPlayers         int
    CreatedAt          time.Time
    StartedAt          time.Time
    CountdownRemaining int
    LeaderRevealed     bool
    GameConfig         GameConfig
    mu                 sync.RWMutex
}

// Player - Individual participant
type Player struct {
    ID          string
    SessionID   string
    Name        string
    Role        *Role
    IsHost      bool
    IsConnected bool
    LastSeen    time.Time
}

// Role - Player role definition
type Role struct {
    Type        RoleType
    Name        string
    Description string
    WinCondition string
    CanReveal   bool
    IsRevealed  bool
}

// GameConfig - Configurable game parameters
type GameConfig struct {
    CountdownDuration int
    RevealTimeout     int
    MaxPlayers        int
    MinPlayers        int
    RoleDistribution  RoleDistribution
}
```

### Role System

```go
type RoleType string

const (
    RoleLeader   RoleType = "leader"
    RoleGuardian RoleType = "guardian"
    RoleAssassin RoleType = "assassin"
    RoleTraitor  RoleType = "traitor"
    RoleFollower RoleType = "follower"
)

// Role distribution based on player count
type RoleDistribution struct {
    PlayerCount int
    Roles       map[RoleType]int
}

// Default distributions (from mtgtreachery.net)
var DefaultDistributions = []RoleDistribution{
    {4, map[RoleType]int{RoleLeader: 1, RoleGuardian: 1, RoleAssassin: 1, RoleFollower: 1}},
    {5, map[RoleType]int{RoleLeader: 1, RoleGuardian: 1, RoleAssassin: 1, RoleFollower: 2}},
    // ... etc
}
```

### SSE Event Payloads

```go
// Player update event
type PlayerUpdatePayload struct {
    Players []PlayerInfo
    Total   int
}

// Game state update
type GameStatePayload struct {
    State              GameState
    CountdownRemaining int
    LeaderRevealed     bool
}

// Role assignment (sent only to individual player)
type RoleAssignmentPayload struct {
    Role        *Role
    OtherRoles  []string // For assassins/traitors
}
```

## 4. Technology Decisions

### Core Stack Confirmation

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go | Performance, simplicity, great concurrency |
| Router | Chi | Lightweight, idiomatic, good middleware |
| Templates | Templ | Type-safe, Go-native, good performance |
| Reactivity | Datastar | SSE-native, minimal JS, server-driven |
| Storage | In-memory → Redis | Simple start, clear upgrade path |

### Testing Framework Selection

```go
// Unit Testing
- Standard library testing package
- testify/assert for assertions
- testify/mock for mocking

// Integration Testing
- httptest for handler testing
- Custom test harness for SSE testing

// E2E Testing
- Chromium via Playwright-go
- Test scenarios in headless mode
```

### Browser Testing Approach

```yaml
# E2E Test Structure
setup:
  - Start test server
  - Initialize Chromium driver
  
scenarios:
  - Create and join room
  - Start game with minimum players
  - Role reveal mechanics
  - Player disconnection/reconnection
  - Concurrent player actions

validation:
  - DOM state matches server state
  - SSE updates received correctly
  - No race conditions
```

### Scaling Considerations

1. **Phase 0-2**: In-memory storage sufficient
2. **Phase 3+**: Redis for distributed state
3. **Future**: Horizontal scaling via sticky sessions

## 5. Implementation Patterns

### Handler Pattern

```go
// Consistent handler structure
type Handler struct {
    store    Store
    eventBus *EventBus
    sessions SessionManager
}

// Handler methods follow pattern:
func (h *Handler) ActionName(w http.ResponseWriter, r *http.Request) {
    // 1. Get/validate session
    session := h.getOrCreateSession(w, r)
    
    // 2. Parse request
    roomCode := chi.URLParam(r, "code")
    
    // 3. Perform business logic
    room, err := h.store.GetRoom(roomCode)
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    // 4. Update state
    // 5. Publish events
    // 6. Return response (HTML fragment or redirect)
}
```

### Template Organization

```
internal/views/
├── layouts/
│   └── base.templ          # Main HTML structure
├── pages/
│   ├── home.templ          # Full pages
│   ├── lobby.templ
│   └── game.templ
├── fragments/
│   ├── player_list.templ   # SSE update fragments
│   ├── game_state.templ
│   └── role_card.templ
└── components/
    ├── button.templ        # Reusable UI components
    └── countdown.templ
```

### Datastar Integration Patterns

```html
<!-- Reactive container -->
<div data-signals='{"connected": false}'>
  <!-- SSE connection -->
  <div data-on-connected='connected = true'
       data-sse='/sse/lobby/{{.RoomCode}}'>
  </div>
  
  <!-- Conditional rendering -->
  <div data-show='connected'>
    Connected to room
  </div>
  
  <!-- Action binding -->
  <button data-on-click='@post(/room/{{.RoomCode}}/start)'>
    Start Game
  </button>
</div>
```

### Error Handling

```go
// Centralized error types
var (
    ErrRoomNotFound = errors.New("room not found")
    ErrRoomFull     = errors.New("room is full")
    ErrNotHost      = errors.New("only host can perform this action")
    ErrInvalidState = errors.New("invalid game state for action")
)

// Consistent error responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
    switch err {
    case ErrRoomNotFound:
        // Redirect to home with flash message
    case ErrRoomFull:
        // Show error fragment
    default:
        // Generic error page
    }
}
```

## 6. Testing Architecture

### Test Structure

```
tests/
├── unit/
│   ├── game/         # Business logic tests
│   ├── store/        # Storage tests
│   └── handlers/     # Handler tests
├── integration/
│   ├── api/          # HTTP endpoint tests
│   ├── sse/          # Real-time event tests
│   └── session/      # Session management tests
└── e2e/
    ├── scenarios/    # Full user flow tests
    └── fixtures/     # Test data
```

### Testing Patterns

```go
// Unit test pattern
func TestRoom_AddPlayer(t *testing.T) {
    // Arrange
    room := &Room{
        Code:       "ABCDE",
        MaxPlayers: 4,
        Players:    make(map[string]*Player),
    }
    
    // Act
    err := room.AddPlayer(newTestPlayer("p1"))
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, room.Players, 1)
}

// Integration test pattern
func TestHandler_CreateRoom(t *testing.T) {
    // Setup
    h := setupTestHandler()
    
    // Execute
    req := httptest.NewRequest("POST", "/room/new", nil)
    rec := httptest.NewRecorder()
    h.CreateRoom(rec, req)
    
    // Verify
    assert.Equal(t, http.StatusSeeOther, rec.Code)
    assert.Contains(t, rec.Header().Get("Location"), "/room/")
}

// E2E test pattern
func TestGameFlow_CompleteGame(t *testing.T) {
    // Start server
    // Launch browser
    // Create room
    // Join players
    // Start game
    // Verify roles assigned
    // Play through game
    // Verify end state
}
```

### Coverage Strategy

| Component | Target | Strategy |
|-----------|--------|----------|
| Game Logic | >95% | Comprehensive unit tests |
| Handlers | >85% | Integration + unit tests |
| Store | >90% | Unit tests with race detection |
| Templates | >70% | Render tests + E2E validation |
| **Total** | >80% | Mix of all test types |

## Architecture Decision Records

### ADR-001: Server-Side Rendering with SSE
- **Decision**: Use SSR + SSE instead of SPA
- **Rationale**: Simpler architecture, better SEO, reduced client complexity
- **Trade-offs**: Less client-side state, requires careful fragment design

### ADR-002: In-Memory Storage First
- **Decision**: Start with in-memory, plan for Redis
- **Rationale**: Faster development, sufficient for initial load
- **Trade-offs**: No persistence, single-server limitation

### ADR-003: Cookie-Based Sessions
- **Decision**: Use HTTP-only cookies for session management
- **Rationale**: Simple, secure, works with SSE
- **Trade-offs**: No JWT benefits, requires cookie support

### ADR-004: EventBus Pattern
- **Decision**: Use EventBus for real-time event distribution
- **Rationale**: Decouples publishers from subscribers, scales well
- **Trade-offs**: Additional abstraction layer

## Security Considerations

1. **Session Security**
   - HTTP-only cookies
   - Secure flag in production
   - CSRF protection via SameSite

2. **Input Validation**
   - Sanitize all user inputs
   - Validate room codes
   - Rate limit room creation

3. **Game Integrity**
   - Server authoritative state
   - Action validation
   - Role visibility controls

## Performance Optimization

1. **SSE Optimization**
   - Connection pooling
   - Efficient event serialization
   - Heartbeat for connection health

2. **Template Caching**
   - Pre-compile templates
   - Cache rendered fragments
   - Minimize re-renders

3. **State Management**
   - Minimize lock contention
   - Efficient player lookups
   - Lazy loading where possible

---

This architecture provides a solid foundation for implementing Treacherest with clear separation of concerns, testability, and a path for future scaling.