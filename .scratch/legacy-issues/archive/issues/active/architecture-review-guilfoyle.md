# Architecture Review - Critical Issues

**Date**: 2025-08-05
**Reviewer**: Guilfoyle (via Claude Code)
**Grade**: C+ (Shows promise, needs significant architectural improvements)

## Executive Summary

The Treacherest codebase has good initial structure but contains fundamental concurrency and architecture issues that will cause problems in production. This document outlines the critical issues discovered during architectural review.

## Critical Issues (Ordered by Severity)

### 1. Race Conditions - Thread Safety Issues

**Problem**: Game state mutations are not thread-safe. The `Players` map and other shared state can be corrupted by concurrent access in a real-time multiplayer environment.

**Evidence**:
- `internal/game/game.go`: Direct map access without synchronization
- No mutexes or channels protecting shared state
- Multiple goroutines can modify game state simultaneously

**Solution**:
- Implement proper synchronization with `sync.RWMutex`
- Consider using channels for state mutations instead of direct map access
- Add state validation before every mutation

### 2. Memory Leaks - SSE Connection Management

**Problem**: SSE connections aren't properly cleaned up, leading to resource exhaustion. Game rooms persist forever without cleanup logic.

**Current State**:
- Channels created for SSE connections but no guaranteed cleanup path
- No connection tracking or lifecycle management
- Game rooms never garbage collected after games end

**Proposed Solution**:
- Implement connection lifecycle management with proper cleanup
- Add heartbeat mechanism to detect disconnected clients
- Implement room timeout and cleanup after inactivity
- Track active connections per room

### 3. God Objects - Handler Responsibility Overload

**Problem**: Handlers (especially `game_handler.go`) are doing too much:
- HTTP request handling
- Game state manipulation
- SSE broadcasting
- Template rendering

**Solution**: Separate concerns with service layers:
```go
type GameService interface {
    CreateRoom(playerName string) (*Room, error)
    JoinRoom(roomCode, playerName string) (*Player, error)
    SubmitMove(roomCode, playerID string, move Move) error
}

type SSEBroadcaster interface {
    BroadcastToRoom(roomCode string, event SSEEvent)
}
```

### 4. Security Vulnerability - No Input Validation

**Problem**: All client input is trusted without validation. Room codes, player names, and move data are used directly.

**Risks**:
- SQL injection (if database is added)
- XSS attacks through player names
- Server crashes from malformed input

**Solution**:
- Validate all input at handler boundaries
- Sanitize player names and room codes
- Validate game moves against game rules

### 5. Poor Error Handling - Silent Failures

**Problem**: Errors are logged but not properly propagated, leading to poor user experience.

**Current Pattern**:
```go
if err != nil {
    log.Printf("error: %v", err)
    return
}
```

**Better Pattern**:
```go
if err != nil {
    return fmt.Errorf("failed to join room %s: %w", roomCode, err)
}
```

## Additional Issues

### Template Organization
- Templates are deeply nested without clear responsibility boundaries
- `game.templ` handles both layout and game state rendering
- Should be split into components: game_board, player_list, card_display

### Missing Features
- No comprehensive test coverage for concurrent scenarios
- No metrics or monitoring
- No graceful shutdown handling

## Immediate Action Items

1. **Fix race conditions** - Add proper synchronization to all game state mutations
2. **Implement connection cleanup** - Properly manage SSE connection lifecycles with heartbeat
3. **Add input validation** - Validate everything that comes from clients
4. **Separate concerns** - Break up god objects into focused services
5. **Improve error handling** - Stop swallowing errors silently

## Long-term Recommendations

1. Implement proper event sourcing for game state
2. Add comprehensive integration tests for multiplayer scenarios
3. Consider using a message queue for game events
4. Add observability (metrics, tracing, structured logging)
5. Implement proper graceful shutdown

## Positive Aspects

- Clean project structure with good separation of concerns
- Proper use of Templ + Datastar without overengineering
- Justfile for readable build automation
- Clear module boundaries

## Next Steps

Start with fixing the race conditions as they're the most dangerous issue that will corrupt game state in production. Then move to SSE connection management with proper cleanup and heartbeat implementation.