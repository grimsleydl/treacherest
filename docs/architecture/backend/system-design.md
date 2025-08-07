# Treacherest Backend System Design

## Overview
Treacherest follows a server-side rendering architecture with real-time updates via Server-Sent Events (SSE). The backend is built in Go using the Chi router, with Templ for templating and Datastar for reactive UI updates.

## Core Architecture Principles

### 1. DOM as State Philosophy
Following the Datastar philosophy, the DOM itself is the state. We don't maintain separate client-side state - instead:
- Server renders complete UI fragments
- SSE sends updated DOM fragments when state changes
- Idiomorph handles efficient DOM diffing
- Full body morphs for simplicity over granular updates

### 2. Event-Driven Architecture
All real-time updates flow through a centralized event bus:
```go
Event {
    Type     string      // "player_joined", "game_started", etc.
    RoomCode string      // Room identifier
    Data     interface{} // Event-specific payload
}
```

### 3. Thread-Safe State Management
All game state is protected by appropriate synchronization:
- Read/Write mutexes for concurrent access
- Atomic operations where appropriate
- No shared mutable state between goroutines

## System Components

### 1. HTTP Server (`cmd/server/`)
- **Responsibilities**: HTTP routing, middleware, static file serving
- **Key Components**:
  - `main.go`: Entry point and server initialization
  - `server.go`: Router configuration and middleware setup
- **Patterns**:
  - Chi router for RESTful endpoints
  - Middleware for logging, recovery, and timeouts
  - Clean separation of concerns

### 2. Handlers (`internal/handlers/`)
- **Responsibilities**: HTTP request handling, SSE streaming, event publishing
- **Key Components**:
  - `handlers.go`: Base handler struct with dependencies
  - `pages.go`: Page rendering handlers
  - `actions.go`: Game action handlers
  - `sse.go`: SSE streaming handlers
- **Patterns**:
  - Dependency injection via handler struct
  - Session-based player identification
  - Event publishing for state changes

### 3. Game Logic (`internal/game/`)
- **Responsibilities**: Core game rules, state transitions, player management
- **Key Components**:
  - `room.go`: Game room state and operations
  - `player.go`: Player data and lifecycle
  - `roles.go`: Role definitions and assignment
  - `errors.go`: Domain-specific error types
- **Patterns**:
  - Immutable role definitions
  - Thread-safe room operations
  - Clear error semantics

### 4. Storage (`internal/store/`)
- **Responsibilities**: Game state persistence (currently in-memory)
- **Key Components**:
  - `memory.go`: In-memory storage implementation
- **Patterns**:
  - Repository pattern for data access
  - Thread-safe operations
  - Room code generation with collision detection

### 5. Views (`internal/views/`)
- **Responsibilities**: HTML generation via Templ templates
- **Key Components**:
  - `layouts/`: Base page layouts
  - `pages/`: Full page templates
  - `components/`: Reusable UI components
- **Patterns**:
  - Component-based architecture
  - Datastar attributes for reactivity
  - Separate body components for SSE updates

## Request Flow Architecture

### 1. Page Load Flow
```
Browser -> HTTP GET -> Handler -> Game State -> Template -> HTML Response
                                       |
                                       v
                                  Session Cookie
```

### 2. SSE Connection Flow
```
Browser -> SSE Connect -> Handler -> EventBus.Subscribe
                             |
                             v
                        Send Initial State
                             |
                             v
                        Wait for Events
                             |
                             v
                     Render & Send Fragments
```

### 3. Action Flow
```
Browser -> HTTP POST -> Handler -> Update Game State -> Publish Event
                                                             |
                                                             v
                                                      All SSE Subscribers
```

## State Management Patterns

### 1. Room State Lifecycle
```
StateLobby -> StateCountdown -> StatePlaying -> StateEnded
```

### 2. Player Session Management
- Session cookies for browser identification
- Room-specific player cookies for rejoining
- Player IDs separate from session IDs

### 3. Concurrent Access Patterns
```go
// Read operations
room.mu.RLock()
defer room.mu.RUnlock()
// ... read operations

// Write operations
room.mu.Lock()
defer room.mu.Unlock()
// ... write operations
```

## Event System Architecture

### 1. Event Types
- `player_joined`: New player added to room
- `player_left`: Player removed from room
- `game_started`: Game transitions from lobby
- `countdown_update`: Countdown timer tick
- `role_revealed`: Player role becomes visible
- `game_ended`: Game reaches end state

### 2. Event Bus Implementation
- Channel-based pub/sub system
- Per-room event channels
- Non-blocking publish with channel overflow protection
- Automatic cleanup on unsubscribe

### 3. SSE Fragment Patterns
```go
// Send complete UI updates
sse.MergeFragments(html,
    datastar.WithSelector("#container-id"),
    datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
```

## Security Considerations

### 1. Session Security
- HTTP-only cookies prevent XSS attacks
- SameSite=Lax for CSRF protection
- Cryptographically random session IDs

### 2. Input Validation
- Player name validation
- Room code format validation
- Rate limiting considerations

### 3. State Isolation
- Players only receive their own role information
- No global state leakage
- Proper authorization checks

## Performance Architecture

### 1. Scalability Patterns
- In-memory storage for low latency
- Event bus with buffered channels
- Efficient DOM diffing via Idiomorph
- Connection pooling for SSE

### 2. Resource Management
- Timeout middleware for long requests
- Channel buffer sizes to prevent blocking
- Goroutine lifecycle management
- Memory cleanup for disconnected players

### 3. Optimization Opportunities
- Room cleanup after inactivity
- Event deduplication
- Batch SSE updates
- Connection health monitoring

## Testing Architecture

### 1. Unit Testing Patterns
```go
// Table-driven tests for comprehensive coverage
func TestRoomOperations(t *testing.T) {
    tests := []struct {
        name string
        // ... test cases
    }{
        // ... test definitions
    }
}
```

### 2. Integration Testing Patterns
- HTTP test server setup
- Full request/response testing
- SSE event verification
- Concurrent operation testing

### 3. Mocking Strategies
- Interface-based mocking
- Test doubles for external dependencies
- Time-based testing with controlled clocks

## Error Handling Patterns

### 1. Domain Errors
```go
var (
    ErrRoomNotFound = errors.New("room not found")
    ErrRoomFull     = errors.New("room is full")
    ErrGameStarted  = errors.New("game already started")
)
```

### 2. HTTP Error Responses
- Appropriate status codes
- User-friendly error messages
- Consistent error format

### 3. Recovery Patterns
- Panic recovery middleware
- Graceful degradation
- Error logging and monitoring

## Future Architecture Considerations

### 1. Persistence Layer
- Database integration for room persistence
- Player statistics tracking
- Game history storage

### 2. Horizontal Scaling
- Redis for shared state
- WebSocket upgrade path
- Load balancer considerations

### 3. Monitoring & Observability
- Structured logging
- Metrics collection
- Distributed tracing

## Implementation Checklist

When implementing new features:
1. Define domain models in `internal/game/`
2. Add storage operations in `internal/store/`
3. Create handler methods in `internal/handlers/`
4. Define event types and publishing
5. Create Templ templates in `internal/views/`
6. Add routes in `cmd/server/server.go`
7. Write comprehensive tests at each layer
8. Verify SSE updates work correctly
9. Test concurrent operations
10. Document any new patterns

This architecture provides a solid foundation for building a reliable, scalable multiplayer game while maintaining simplicity and testability.