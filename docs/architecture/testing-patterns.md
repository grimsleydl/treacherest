# Treacherest Testing Patterns & Strategy

## Overview
This document provides concrete testing patterns and strategies to achieve >80% code coverage and ensure reliable, maintainable code. It covers unit, integration, and E2E testing approaches specific to the Treacherest architecture.

## Current Coverage Analysis

### Coverage Gaps
1. **cmd/server** (0% → Target: 80%)
   - Server initialization
   - Route configuration
   - Middleware setup

2. **internal/views** (0% → Target: 70%)
   - Template rendering
   - Component composition
   - HTML output validation

3. **internal/handlers** (66% → Target: 85%)
   - Missing integration tests
   - SSE edge cases
   - Error scenarios

## Testing Patterns by Package

### 1. Server Package Testing (`cmd/server`)

#### Pattern: Server Setup Testing
```go
// server_test.go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSetupServer(t *testing.T) {
    // Test server initialization
    handler := SetupServer()
    assert.NotNil(t, handler)
    
    // Test middleware is applied
    ts := httptest.NewServer(handler)
    defer ts.Close()
    
    // Verify static files are served
    resp, err := http.Get(ts.URL + "/static/test.css")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRoutes(t *testing.T) {
    tests := []struct {
        name       string
        method     string
        path       string
        wantStatus int
    }{
        {"home page", "GET", "/", http.StatusOK},
        {"create room", "POST", "/room/new", http.StatusBadRequest}, // No data
        {"join room", "GET", "/room/INVALID", http.StatusNotFound},
        {"sse lobby", "GET", "/sse/lobby/INVALID", http.StatusNotFound},
    }
    
    handler := SetupServer()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.path, nil)
            w := httptest.NewRecorder()
            
            handler.ServeHTTP(w, req)
            
            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}
```

#### Pattern: Main Function Testing
```go
// main_test.go
package main

import (
    "os"
    "testing"
    "time"
)

func TestMainFunction(t *testing.T) {
    // Test with custom port
    os.Setenv("PORT", "8899")
    
    // Run main in goroutine
    go main()
    
    // Give server time to start
    time.Sleep(100 * time.Millisecond)
    
    // Verify server is running
    resp, err := http.Get("http://localhost:8899/")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### 2. View Template Testing (`internal/views`)

#### Pattern: Template Rendering Tests
```go
// views_test.go
package views

import (
    "bytes"
    "context"
    "strings"
    "testing"
    "treacherest/internal/game"
    "treacherest/internal/views/pages"
    "github.com/stretchr/testify/assert"
)

func TestHomePageRendering(t *testing.T) {
    component := pages.Home()
    
    buf := &bytes.Buffer{}
    err := component.Render(context.Background(), buf)
    
    assert.NoError(t, err)
    html := buf.String()
    
    // Verify key elements
    assert.Contains(t, html, "Create Room")
    assert.Contains(t, html, "Join Room")
    assert.Contains(t, html, "data-on-submit")
}

func TestLobbyPageRendering(t *testing.T) {
    room := &game.Room{
        Code:       "ABC12",
        State:      game.StateLobby,
        Players:    make(map[string]*game.Player),
        MaxPlayers: 8,
    }
    
    player := &game.Player{
        ID:   "p1",
        Name: "Alice",
    }
    
    room.Players[player.ID] = player
    
    component := pages.LobbyPage(room, player)
    
    buf := &bytes.Buffer{}
    err := component.Render(context.Background(), buf)
    
    assert.NoError(t, err)
    html := buf.String()
    
    // Verify room code displayed
    assert.Contains(t, html, "ABC12")
    
    // Verify player listed
    assert.Contains(t, html, "Alice")
    
    // Verify SSE connection
    assert.Contains(t, html, "data-on-load")
    assert.Contains(t, html, "/sse/lobby/ABC12")
}

func TestGamePageConditionalRendering(t *testing.T) {
    tests := []struct {
        name          string
        setupRoom     func() *game.Room
        setupPlayer   func() *game.Player
        wantContains  []string
        wantNotContains []string
    }{
        {
            name: "lobby state shows start button for 4 players",
            setupRoom: func() *game.Room {
                room := &game.Room{
                    Code:    "TEST1",
                    State:   game.StateLobby,
                    Players: make(map[string]*game.Player),
                }
                // Add 4 players
                for i := 0; i < 4; i++ {
                    p := &game.Player{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Player%d", i)}
                    room.Players[p.ID] = p
                }
                return room
            },
            setupPlayer: func() *game.Player {
                return &game.Player{ID: "p0", Name: "Player0"}
            },
            wantContains: []string{"Start Game"},
            wantNotContains: []string{"Need at least 4 players"},
        },
        {
            name: "lobby state shows waiting message for < 4 players",
            setupRoom: func() *game.Room {
                room := &game.Room{
                    Code:    "TEST2",
                    State:   game.StateLobby,
                    Players: make(map[string]*game.Player),
                }
                // Add 2 players
                for i := 0; i < 2; i++ {
                    p := &game.Player{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Player%d", i)}
                    room.Players[p.ID] = p
                }
                return room
            },
            setupPlayer: func() *game.Player {
                return &game.Player{ID: "p0", Name: "Player0"}
            },
            wantContains: []string{"Need at least 4 players"},
            wantNotContains: []string{"Start Game"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            room := tt.setupRoom()
            player := tt.setupPlayer()
            
            component := pages.LobbyBody(room, player)
            
            buf := &bytes.Buffer{}
            err := component.Render(context.Background(), buf)
            assert.NoError(t, err)
            
            html := buf.String()
            
            for _, want := range tt.wantContains {
                assert.Contains(t, html, want)
            }
            
            for _, notWant := range tt.wantNotContains {
                assert.NotContains(t, html, notWant)
            }
        })
    }
}
```

#### Pattern: Component Testing
```go
func TestPlayerCardComponent(t *testing.T) {
    tests := []struct {
        name     string
        player   *game.Player
        wantHTML []string
    }{
        {
            name: "shows name without role",
            player: &game.Player{
                ID:   "p1",
                Name: "Alice",
            },
            wantHTML: []string{"Alice", "player-card"},
        },
        {
            name: "shows role when revealed",
            player: &game.Player{
                ID:   "p1",
                Name: "Bob",
                Role: &game.Role{
                    Type:     game.RoleLeader,
                    Name:     "Leader",
                    Revealed: true,
                },
            },
            wantHTML: []string{"Bob", "Leader", "role-revealed"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := components.PlayerCard(tt.player)
            html := renderToString(component)
            
            for _, want := range tt.wantHTML {
                assert.Contains(t, html, want)
            }
        })
    }
}
```

### 3. Handler Integration Testing

#### Pattern: Full Request/Response Testing
```go
// handlers_integration_test.go
package handlers

import (
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "testing"
    "treacherest/internal/store"
)

func TestCreateRoomIntegration(t *testing.T) {
    // Setup
    gameStore := store.NewMemoryStore()
    h := New(gameStore)
    
    // Create form data
    form := url.Values{}
    form.Add("playerName", "Alice")
    
    // Make request
    req := httptest.NewRequest("POST", "/room/new", strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    w := httptest.NewRecorder()
    
    // Execute
    h.CreateRoom(w, req)
    
    // Verify redirect
    assert.Equal(t, http.StatusSeeOther, w.Code)
    location := w.Header().Get("Location")
    assert.Regexp(t, `^/room/[A-Z0-9]{5}$`, location)
    
    // Verify room created
    roomCode := strings.TrimPrefix(location, "/room/")
    room, err := gameStore.GetRoom(roomCode)
    assert.NoError(t, err)
    assert.Equal(t, 1, len(room.Players))
    
    // Verify cookie set
    cookies := w.Result().Cookies()
    assert.True(t, hasCookie(cookies, "session"))
    assert.True(t, hasCookie(cookies, "player_"+roomCode))
}

func TestJoinRoomIntegration(t *testing.T) {
    // Setup
    gameStore := store.NewMemoryStore()
    h := New(gameStore)
    
    // Create room with initial player
    room, _ := gameStore.CreateRoom()
    creator := game.NewPlayer("p1", "Creator", "session1")
    room.AddPlayer(creator)
    gameStore.UpdateRoom(room)
    
    tests := []struct {
        name       string
        roomCode   string
        playerName string
        wantStatus int
        wantBody   string
    }{
        {
            name:       "successful join",
            roomCode:   room.Code,
            playerName: "Alice",
            wantStatus: http.StatusOK,
            wantBody:   "Game Lobby",
        },
        {
            name:       "room not found",
            roomCode:   "INVALID",
            playerName: "Alice",
            wantStatus: http.StatusNotFound,
            wantBody:   "Room not found",
        },
        {
            name:       "missing player name shows form",
            roomCode:   room.Code,
            playerName: "",
            wantStatus: http.StatusOK,
            wantBody:   "Enter your name",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Make request
            url := fmt.Sprintf("/room/%s", tt.roomCode)
            if tt.playerName != "" {
                url += "?name=" + tt.playerName
            }
            
            req := httptest.NewRequest("GET", url, nil)
            w := httptest.NewRecorder()
            
            // Set up Chi context
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("code", tt.roomCode)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            
            // Execute
            h.JoinRoom(w, req)
            
            // Verify
            assert.Equal(t, tt.wantStatus, w.Code)
            assert.Contains(t, w.Body.String(), tt.wantBody)
        })
    }
}
```

#### Pattern: SSE Testing
```go
func TestSSEStreamIntegration(t *testing.T) {
    // Setup
    gameStore := store.NewMemoryStore()
    h := New(gameStore)
    
    // Create room and player
    room, _ := gameStore.CreateRoom()
    player := game.NewPlayer("p1", "Alice", "session1")
    room.AddPlayer(player)
    gameStore.UpdateRoom(room)
    
    // Create SSE request
    req := httptest.NewRequest("GET", "/sse/lobby/"+room.Code, nil)
    req.Header.Set("Accept", "text/event-stream")
    
    // Add player cookie
    req.AddCookie(&http.Cookie{
        Name:  "player_" + room.Code,
        Value: player.ID,
    })
    
    // Set up Chi context
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("code", room.Code)
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
    
    // Create response recorder
    w := httptest.NewRecorder()
    
    // Start SSE in goroutine
    done := make(chan bool)
    go func() {
        h.StreamLobby(w, req)
        done <- true
    }()
    
    // Wait for initial render
    time.Sleep(50 * time.Millisecond)
    
    // Verify SSE headers
    assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
    assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
    
    // Verify initial fragment sent
    body := w.Body.String()
    assert.Contains(t, body, "event: datastar-merge-fragments")
    assert.Contains(t, body, "data: selector #lobby-container")
    assert.Contains(t, body, "Alice")
    
    // Simulate player join
    newPlayer := game.NewPlayer("p2", "Bob", "session2")
    room.AddPlayer(newPlayer)
    h.eventBus.Publish(Event{
        Type:     "player_joined",
        RoomCode: room.Code,
        Data:     room,
    })
    
    // Wait for update
    time.Sleep(50 * time.Millisecond)
    
    // Verify update contains new player
    body = w.Body.String()
    assert.Contains(t, body, "Bob")
}
```

### 4. Browser/E2E Testing Setup

#### Pattern: Playwright Test Configuration
```javascript
// e2e/playwright.config.js
module.exports = {
    testDir: './tests',
    timeout: 30000,
    use: {
        baseURL: process.env.BASE_URL || 'http://localhost:8080',
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
    },
    projects: [
        {
            name: 'Desktop Chrome',
            use: { ...devices['Desktop Chrome'] },
        },
        {
            name: 'Mobile Safari',
            use: { ...devices['iPhone 12'] },
        },
    ],
};
```

#### Pattern: E2E Test Structure
```javascript
// e2e/tests/game-flow.spec.js
const { test, expect } = require('@playwright/test');

test.describe('Complete Game Flow', () => {
    test('multiple players can play a full game', async ({ browser }) => {
        // Create multiple browser contexts for different players
        const player1Context = await browser.newContext();
        const player2Context = await browser.newContext();
        const player3Context = await browser.newContext();
        const player4Context = await browser.newContext();
        
        const player1 = await player1Context.newPage();
        const player2 = await player2Context.newPage();
        const player3 = await player3Context.newPage();
        const player4 = await player4Context.newPage();
        
        // Player 1 creates room
        await player1.goto('/');
        await player1.fill('input[name="playerName"]', 'Alice');
        await player1.click('button[type="submit"]');
        
        // Get room code
        await expect(player1).toHaveURL(/\/room\/[A-Z0-9]{5}/);
        const roomCode = player1.url().split('/').pop();
        
        // Other players join
        for (const [page, name] of [[player2, 'Bob'], [player3, 'Charlie'], [player4, 'David']]) {
            await page.goto(`/room/${roomCode}`);
            await page.fill('input[name="name"]', name);
            await page.click('button[type="submit"]');
        }
        
        // Verify all players see each other
        for (const page of [player1, player2, player3, player4]) {
            await expect(page.locator('.player-list')).toContainText('Alice');
            await expect(page.locator('.player-list')).toContainText('Bob');
            await expect(page.locator('.player-list')).toContainText('Charlie');
            await expect(page.locator('.player-list')).toContainText('David');
        }
        
        // Start game
        await player1.click('button:has-text("Start Game")');
        
        // Verify redirect to game page
        for (const page of [player1, player2, player3, player4]) {
            await expect(page).toHaveURL(`/game/${roomCode}`);
        }
        
        // Cleanup
        await player1Context.close();
        await player2Context.close();
        await player3Context.close();
        await player4Context.close();
    });
    
    test('handles player disconnection gracefully', async ({ page, context }) => {
        // Test SSE reconnection
        await page.goto('/');
        await page.fill('input[name="playerName"]', 'TestPlayer');
        await page.click('button[type="submit"]');
        
        // Simulate network interruption
        await context.setOffline(true);
        await page.waitForTimeout(1000);
        
        // Verify reconnection UI
        await expect(page.locator('.connection-status')).toContainText('Reconnecting');
        
        // Restore connection
        await context.setOffline(false);
        await page.waitForTimeout(1000);
        
        // Verify reconnected
        await expect(page.locator('.connection-status')).toContainText('Connected');
    });
});
```

### 5. Performance & Load Testing

#### Pattern: Benchmark Tests
```go
// benchmarks_test.go
package handlers

func BenchmarkCreateRoom(b *testing.B) {
    gameStore := store.NewMemoryStore()
    h := New(gameStore)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        form := url.Values{}
        form.Add("playerName", fmt.Sprintf("Player%d", i))
        
        req := httptest.NewRequest("POST", "/room/new", strings.NewReader(form.Encode()))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        w := httptest.NewRecorder()
        
        h.CreateRoom(w, req)
    }
}

func BenchmarkSSEBroadcast(b *testing.B) {
    gameStore := store.NewMemoryStore()
    h := New(gameStore)
    
    // Create room with players
    room, _ := gameStore.CreateRoom()
    for i := 0; i < 8; i++ {
        player := game.NewPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player%d", i), fmt.Sprintf("s%d", i))
        room.AddPlayer(player)
    }
    gameStore.UpdateRoom(room)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        h.eventBus.Publish(Event{
            Type:     "game_update",
            RoomCode: room.Code,
            Data:     room,
        })
    }
}
```

## Test Execution Strategy

### 1. Local Development
```bash
# Run all tests with coverage
test-coverage

# Run specific package tests
go test ./internal/handlers -v

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### 2. Continuous Integration
```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: cachix/install-nix-action@v22
      - run: nix develop --command test-coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./nix/app/build/coverage/coverage.out
```

### 3. Pre-commit Hooks
```yaml
# lefthook.yml
pre-commit:
  commands:
    tests:
      run: cd nix/app && go test ./...
    coverage:
      run: cd nix/app && go test -cover ./...
```

## Coverage Goals by Phase

### Phase 0 Completion Targets
1. **Week 1**: Server package tests (0% → 80%)
2. **Week 1**: View template tests (0% → 70%)
3. **Week 2**: Handler integration tests (66% → 85%)
4. **Week 2**: E2E test infrastructure setup
5. **Week 3**: Full test suite running, >80% total coverage

### Critical Path to Phase 1
1. All packages have baseline test coverage
2. Integration tests verify all endpoints
3. E2E tests cover happy paths
4. Performance benchmarks established
5. CI/CD pipeline running all tests

## Common Testing Pitfalls to Avoid

### 1. Test Isolation
```go
// BAD: Tests depend on execution order
var sharedRoom *game.Room

func TestCreateSharedRoom(t *testing.T) {
    sharedRoom = &game.Room{Code: "TEST1"}
}

func TestUseSharedRoom(t *testing.T) {
    // This fails if TestCreateSharedRoom didn't run first
    assert.Equal(t, "TEST1", sharedRoom.Code)
}

// GOOD: Each test is independent
func TestRoomOperations(t *testing.T) {
    room := &game.Room{Code: "TEST1"}
    assert.Equal(t, "TEST1", room.Code)
}
```

### 2. Timing Dependencies
```go
// BAD: Relies on sleep for async operations
func TestAsyncOperation(t *testing.T) {
    go doSomethingAsync()
    time.Sleep(100 * time.Millisecond) // Flaky!
    // assertions...
}

// GOOD: Use channels or synchronization
func TestAsyncOperation(t *testing.T) {
    done := make(chan bool)
    go func() {
        doSomethingAsync()
        done <- true
    }()
    
    select {
    case <-done:
        // assertions...
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for async operation")
    }
}
```

### 3. Resource Cleanup
```go
// GOOD: Always clean up resources
func TestWithTempFile(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test")
    assert.NoError(t, err)
    defer os.Remove(tmpFile.Name()) // Clean up!
    defer tmpFile.Close()
    
    // test logic...
}
```

## Summary

This testing strategy provides:
1. **Concrete patterns** for each package type
2. **Clear coverage targets** with timelines
3. **Integration approaches** for complex flows
4. **E2E testing setup** for user journeys
5. **Performance baselines** to prevent regression

Following these patterns will achieve >80% coverage while ensuring tests are maintainable, reliable, and provide real value in preventing bugs and enabling confident refactoring.