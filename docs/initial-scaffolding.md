# Initial Scaffolding - MTG Treacherest

## Overview

This document summarizes the initial architecture and implementation of MTG Treacherest, a real-time multiplayer game of deception and hidden roles.

## Technology Stack

- **Backend**: Go with Chi router
- **Templating**: Templ (type-safe Go templates)
- **Frontend Reactivity**: Datastar (SSE-based reactivity)
- **Real-time Updates**: Server-Sent Events (SSE)
- **Development Environment**: Nix with gomod2nix
- **Testing**: Go standard library

## Architecture Decisions

### 1. DOM as State Philosophy

Following the guidance from `datastar-notes.md`, we implemented a "DOM is the state" architecture:
- Full page morphing using idiomorph algorithm
- No complex state management or partial updates
- Server renders complete page bodies on any state change
- Datastar handles efficient DOM diffing client-side

### 2. Project Structure

```
nix/app/
├── cmd/server/         # Application entry point
├── internal/
│   ├── game/          # Core game logic (Room, Player, Roles)
│   ├── handlers/      # HTTP and SSE handlers
│   ├── store/         # In-memory game storage
│   └── views/         # Templ templates
├── static/            # Static assets
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
└── gomod2nix.toml     # Nix dependency management
```

### 3. Real-time Updates Pattern

```go
// Simple SSE pattern - re-render entire page on any event
func (h *Handler) StreamPage(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    events := h.eventBus.Subscribe(roomCode)
    
    // Initial render
    h.renderPage(sse, pageType, roomCode, playerID)
    
    // Re-render on ANY event
    for {
        select {
        case <-r.Context().Done():
            return
        case <-events:
            h.renderPage(sse, pageType, roomCode, playerID)
        }
    }
}
```

## Core Features Implemented

### 1. Room Management
- 5-character alphanumeric room codes (e.g., "4W47G")
- Create room with player name
- Join room by code or direct URL
- 4-8 players per room

### 2. Game Flow
- **Lobby State**: Players join, see real-time player list
- **Countdown State**: 5-second countdown after game starts
- **Playing State**: Roles revealed, game in progress

### 3. Role System
- **Leader**: Public role, wins by surviving
- **Guardian**: Protects Leader, wins with Leader
- **Assassin**: Eliminates Leader to win
- **Traitor**: Wins by being last standing

Role distribution based on player count (matching mtgtreachery.net).

### 4. Session Management
- HTTP-only cookies for player identification
- Per-room player cookies for reconnection
- No authentication required

### 5. Event System
- Central event bus for game state changes
- Events: player_joined, player_left, game_started, countdown_update, game_playing
- All events trigger full page re-render

## Development Workflow

### Nix Commands
All development commands are integrated into the nix shell:
- `dev` - Start with hot reload
- `run` - Run server
- `build` - Build binary
- `test` - Run tests
- `test-all` - Run with coverage
- `fmt` - Format code
- `update-deps` - Update Go dependencies

### Dependency Management
Using gomod2nix for reproducible builds:
1. Add dependencies with `go get`
2. Run `update-deps` to regenerate gomod2nix.toml
3. Nix handles the rest

## Testing
- Unit tests for game logic (roles, room management)
- All tests passing
- Coverage available via `test-all`

## Next Steps

### Phase 1 Priorities
1. Improve mobile UI/UX
2. Add MTG-themed styling
3. Implement role unveil abilities
4. Add game end detection
5. Player disconnect/reconnect handling

### Phase 2 Features
1. Life counter integration
2. Custom role configurations
3. Spectator mode
4. Game history/replay
5. Chat system

## Lessons Learned

1. **Simplicity Wins**: Full page morphing is simpler and more reliable than targeted updates
2. **Datastar Philosophy**: Let the framework handle DOM diffing, focus on server state
3. **Nix Integration**: gomod2nix provides excellent reproducibility for Go projects
4. **SSE Over WebSockets**: Simpler, proxy-friendly, sufficient for this use case

## Resources

- Original site: https://mtgtreachery.net/treacherer/
- Datastar docs: https://data-star.dev
- Project PRD: `/workspace/docs/PRD.org`