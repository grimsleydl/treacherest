# Test Coverage Implementation Guide

## Executive Summary
This guide provides step-by-step instructions to achieve >80% test coverage for Treacherest, focusing on the immediate gaps in `cmd/server` (0%) and `internal/views` (0%) packages.

## Current State Analysis

### Coverage Breakdown
```
Package                 Current  Target  Gap
cmd/server              0%       80%     -80%
internal/views          0%       70%     -70%
internal/handlers       66%      85%     -19%
internal/game           97.8%    98%     -0.2%
internal/store          100%     100%    0%
----------------------------------------
TOTAL                   28.8%    80%     -51.2%
```

## Week 1: Server Package Tests (0% → 80%)

### Day 1-2: Server Setup Tests

#### Step 1: Create `cmd/server/server_test.go`
```go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSetupServer(t *testing.T) {
    // Test server creation
    handler := SetupServer()
    require.NotNil(t, handler, "SetupServer should return a handler")
    
    // Create test server
    ts := httptest.NewServer(handler)
    defer ts.Close()
    
    // Test middleware chain
    t.Run("middleware applied", func(t *testing.T) {
        // Logger middleware test - check response headers
        resp, err := http.Get(ts.URL + "/")
        require.NoError(t, err)
        defer resp.Body.Close()
        
        // Verify request ID header (from logger middleware)
        assert.NotEmpty(t, resp.Header.Get("X-Request-Id"))
    })
    
    // Test static file serving
    t.Run("static files served", func(t *testing.T) {
        // Create a test static file
        testFile := "static/test.txt"
        err := os.MkdirAll("static", 0755)
        require.NoError(t, err)
        defer os.RemoveAll("static")
        
        err = os.WriteFile(testFile, []byte("test content"), 0644)
        require.NoError(t, err)
        
        resp, err := http.Get(ts.URL + "/static/test.txt")
        require.NoError(t, err)
        defer resp.Body.Close()
        
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        body, _ := io.ReadAll(resp.Body)
        assert.Equal(t, "test content", string(body))
    })
}

func TestRouteConfiguration(t *testing.T) {
    handler := SetupServer()
    ts := httptest.NewServer(handler)
    defer ts.Close()
    
    tests := []struct {
        name       string
        method     string
        path       string
        body       io.Reader
        wantStatus int
    }{
        {
            name:       "GET / returns home page",
            method:     "GET",
            path:       "/",
            body:       nil,
            wantStatus: http.StatusOK,
        },
        {
            name:       "POST /room/new without data",
            method:     "POST",
            path:       "/room/new",
            body:       nil,
            wantStatus: http.StatusBadRequest,
        },
        {
            name:       "GET /room/INVALID",
            method:     "GET",
            path:       "/room/INVALID",
            body:       nil,
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "GET /game/INVALID",
            method:     "GET",
            path:       "/game/INVALID",
            body:       nil,
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "POST /room/INVALID/start",
            method:     "POST",
            path:       "/room/INVALID/start",
            body:       nil,
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "POST /room/INVALID/leave",
            method:     "POST",
            path:       "/room/INVALID/leave",
            body:       nil,
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "GET /sse/lobby/INVALID",
            method:     "GET",
            path:       "/sse/lobby/INVALID",
            body:       nil,
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "GET /sse/game/INVALID",
            method:     "GET",
            path:       "/sse/game/INVALID",
            body:       nil,
            wantStatus: http.StatusNotFound,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, err := http.NewRequest(tt.method, ts.URL+tt.path, tt.body)
            require.NoError(t, err)
            
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, tt.wantStatus, resp.StatusCode)
        })
    }
}

func TestMiddlewareChain(t *testing.T) {
    handler := SetupServer()
    
    t.Run("timeout middleware", func(t *testing.T) {
        // Create a slow handler to test timeout
        slowPath := "/test-slow"
        
        // Wrap our handler to add the test route
        testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.URL.Path == slowPath {
                time.Sleep(2 * time.Second)
                w.WriteHeader(http.StatusOK)
            } else {
                handler.ServeHTTP(w, r)
            }
        })
        
        req := httptest.NewRequest("GET", slowPath, nil)
        w := httptest.NewRecorder()
        
        // Add timeout context
        ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
        defer cancel()
        req = req.WithContext(ctx)
        
        testHandler.ServeHTTP(w, req)
        
        // Should timeout
        assert.Equal(t, http.StatusServiceUnavailable, w.Code)
    })
    
    t.Run("recoverer middleware", func(t *testing.T) {
        // Test panic recovery
        panicPath := "/test-panic"
        
        panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.URL.Path == panicPath {
                panic("test panic")
            }
            handler.ServeHTTP(w, r)
        })
        
        // Wrap with recoverer
        safeHandler := middleware.Recoverer(panicHandler)
        
        req := httptest.NewRequest("GET", panicPath, nil)
        w := httptest.NewRecorder()
        
        // Should not panic
        assert.NotPanics(t, func() {
            safeHandler.ServeHTTP(w, req)
        })
        
        assert.Equal(t, http.StatusInternalServerError, w.Code)
    })
}
```

#### Step 2: Create `cmd/server/main_test.go`
```go
package main

import (
    "context"
    "net/http"
    "os"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMainFunction(t *testing.T) {
    // Save original args
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    
    // Test with custom port
    os.Setenv("PORT", "18899")
    defer os.Unsetenv("PORT")
    
    // Run main in goroutine
    mainDone := make(chan bool)
    go func() {
        main()
        mainDone <- true
    }()
    
    // Wait for server to start
    time.Sleep(200 * time.Millisecond)
    
    // Test server is running
    resp, err := http.Get("http://localhost:18899/")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Shutdown server (simulate interrupt)
    // In real implementation, you'd expose a shutdown channel
    select {
    case <-mainDone:
        // Server stopped
    case <-time.After(1 * time.Second):
        t.Error("Server didn't stop in time")
    }
}

func TestPortConfiguration(t *testing.T) {
    tests := []struct {
        name     string
        portEnv  string
        wantPort string
    }{
        {
            name:     "default port",
            portEnv:  "",
            wantPort: "8080",
        },
        {
            name:     "custom port from env",
            portEnv:  "3000",
            wantPort: "3000",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.portEnv != "" {
                os.Setenv("PORT", tt.portEnv)
                defer os.Unsetenv("PORT")
            }
            
            // In real implementation, extract getPort() function
            port := getPort()
            assert.Equal(t, tt.wantPort, port)
        })
    }
}

// Helper function to extract from main
func getPort() string {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    return port
}
```

### Day 3-4: View Template Tests

#### Step 1: Create `internal/views/views_test.go`
```go
package views

import (
    "bytes"
    "context"
    "strings"
    "testing"
    "treacherest/internal/game"
    "treacherest/internal/views/pages"
    "treacherest/internal/views/layouts"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Helper to render component to string
func renderToString(t *testing.T, component templ.Component) string {
    buf := &bytes.Buffer{}
    err := component.Render(context.Background(), buf)
    require.NoError(t, err)
    return buf.String()
}

func TestBaseLayout(t *testing.T) {
    // Test base layout renders correctly
    content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, err := w.Write([]byte("<div>Test Content</div>"))
        return err
    })
    
    component := layouts.Base("Test Title", content)
    html := renderToString(t, component)
    
    // Verify HTML structure
    assert.Contains(t, html, "<!DOCTYPE html>")
    assert.Contains(t, html, "<title>Test Title</title>")
    assert.Contains(t, html, "<div>Test Content</div>")
    assert.Contains(t, html, "data-star")  // Datastar script
}

func TestHomePage(t *testing.T) {
    component := pages.Home()
    html := renderToString(t, component)
    
    // Verify form elements
    assert.Contains(t, html, `name="playerName"`)
    assert.Contains(t, html, `type="submit"`)
    assert.Contains(t, html, `data-on-submit="@post('/room/new')"`)
    
    // Verify required attributes
    assert.Contains(t, html, `required`)
    
    // Verify page structure
    assert.Contains(t, html, "Create Room")
    assert.Contains(t, html, "Join Room")
}

func TestLobbyPage(t *testing.T) {
    tests := []struct {
        name         string
        setupRoom    func() *game.Room
        player       *game.Player
        wantContains []string
        wantMissing  []string
    }{
        {
            name: "shows start button when enough players",
            setupRoom: func() *game.Room {
                room := &game.Room{
                    Code:       "TEST1",
                    State:      game.StateLobby,
                    Players:    make(map[string]*game.Player),
                    MaxPlayers: 8,
                }
                for i := 0; i < 4; i++ {
                    p := &game.Player{
                        ID:   fmt.Sprintf("p%d", i),
                        Name: fmt.Sprintf("Player%d", i),
                    }
                    room.Players[p.ID] = p
                }
                return room
            },
            player: &game.Player{ID: "p0", Name: "Player0"},
            wantContains: []string{
                "TEST1",
                "Start Game",
                "Player0",
                "Player1",
                "4/8",
                `data-on-click="@post('/room/TEST1/start')"`,
            },
            wantMissing: []string{
                "Need at least 4 players",
            },
        },
        {
            name: "shows waiting message when not enough players",
            setupRoom: func() *game.Room {
                room := &game.Room{
                    Code:       "TEST2",
                    State:      game.StateLobby,
                    Players:    make(map[string]*game.Player),
                    MaxPlayers: 8,
                }
                p := &game.Player{ID: "p1", Name: "Alice"}
                room.Players[p.ID] = p
                return room
            },
            player: &game.Player{ID: "p1", Name: "Alice"},
            wantContains: []string{
                "TEST2",
                "Need at least 4 players",
                "Alice",
                "1/8",
            },
            wantMissing: []string{
                "Start Game",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            room := tt.setupRoom()
            component := pages.LobbyPage(room, tt.player)
            html := renderToString(t, component)
            
            for _, want := range tt.wantContains {
                assert.Contains(t, html, want)
            }
            
            for _, missing := range tt.wantMissing {
                assert.NotContains(t, html, missing)
            }
            
            // Always verify SSE setup
            assert.Contains(t, html, fmt.Sprintf(`data-on-load="@get('/sse/lobby/%s')"`, room.Code))
        })
    }
}

func TestGamePage(t *testing.T) {
    room := &game.Room{
        Code:           "GAME1",
        State:          game.StatePlaying,
        Players:        make(map[string]*game.Player),
        LeaderRevealed: false,
    }
    
    leader := &game.Player{
        ID:   "p1",
        Name: "Leader",
        Role: &game.Role{
            Type:     game.RoleLeader,
            Name:     "Leader",
            Revealed: false,
        },
    }
    
    follower := &game.Player{
        ID:   "p2",
        Name: "Follower",
        Role: &game.Role{
            Type:     game.RoleFollower,
            Name:     "Follower",
            Revealed: false,
        },
    }
    
    room.Players[leader.ID] = leader
    room.Players[follower.ID] = follower
    
    t.Run("leader view before reveal", func(t *testing.T) {
        component := pages.GamePage(room, leader)
        html := renderToString(t, component)
        
        // Leader sees their role
        assert.Contains(t, html, "You are the Leader")
        
        // SSE connection
        assert.Contains(t, html, `data-on-load="@get('/sse/game/GAME1')"`)
    })
    
    t.Run("follower view before reveal", func(t *testing.T) {
        component := pages.GamePage(room, follower)
        html := renderToString(t, component)
        
        // Follower sees their role
        assert.Contains(t, html, "You are a Follower")
        
        // Doesn't see who leader is
        assert.NotContains(t, html, "Leader is")
    })
    
    t.Run("after leader revealed", func(t *testing.T) {
        room.LeaderRevealed = true
        leader.Role.Revealed = true
        
        component := pages.GamePage(room, follower)
        html := renderToString(t, component)
        
        // Now follower sees leader
        assert.Contains(t, html, "Leader is: Leader")
    })
}

func TestComponentHelpers(t *testing.T) {
    t.Run("player card styling", func(t *testing.T) {
        player := &game.Player{ID: "p1", Name: "Test"}
        currentPlayer := &game.Player{ID: "p1", Name: "Test"}
        
        component := components.PlayerCard(player, currentPlayer)
        html := renderToString(t, component)
        
        // Should have current player styling
        assert.Contains(t, html, "current-player")
    })
    
    t.Run("room code display", func(t *testing.T) {
        component := components.RoomCode("ABC12")
        html := renderToString(t, component)
        
        assert.Contains(t, html, "ABC12")
        assert.Contains(t, html, "room-code")
    })
}

func TestDatastarAttributes(t *testing.T) {
    // Test that Datastar attributes are properly rendered
    tests := []struct {
        name      string
        component templ.Component
        wantAttrs []string
    }{
        {
            name:      "SSE connection attribute",
            component: pages.LobbyBody(mockRoom(), mockPlayer()),
            wantAttrs: []string{
                "data-on-load",
                "@get('/sse/lobby/",
            },
        },
        {
            name:      "click handler attribute",
            component: components.ActionButton("Test", "/test/action"),
            wantAttrs: []string{
                "data-on-click",
                "@post('/test/action')",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            html := renderToString(t, tt.component)
            
            for _, attr := range tt.wantAttrs {
                assert.Contains(t, html, attr)
            }
        })
    }
}

// Test helpers
func mockRoom() *game.Room {
    return &game.Room{
        Code:       "MOCK1",
        State:      game.StateLobby,
        Players:    make(map[string]*game.Player),
        MaxPlayers: 8,
    }
}

func mockPlayer() *game.Player {
    return &game.Player{
        ID:   "mock-player",
        Name: "Mock Player",
    }
}
```

### Day 5: Integration & Coverage Verification

#### Step 1: Run Coverage Analysis
```bash
# Run tests with coverage
cd nix/app
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Check specific package coverage
go test -cover ./cmd/server
go test -cover ./internal/views
```

#### Step 2: Fill Coverage Gaps
```go
// Add tests for any uncovered lines identified in coverage report
// Common gaps include:
// - Error paths
// - Edge cases
// - Panic recovery
// - Configuration variations
```

## Week 2: Handler Integration Tests (66% → 85%)

### Day 6-7: Complete Handler Coverage

#### Missing Handler Tests
```go
// handlers_complete_test.go
package handlers

func TestStartGame(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    
    // Create room with enough players
    room, _ := store.CreateRoom()
    for i := 0; i < 4; i++ {
        p := game.NewPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player%d", i), fmt.Sprintf("s%d", i))
        room.AddPlayer(p)
    }
    store.UpdateRoom(room)
    
    // Set player cookie
    req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
    req.AddCookie(&http.Cookie{
        Name:  "player_" + room.Code,
        Value: "p0",
    })
    
    w := httptest.NewRecorder()
    
    // Add Chi context
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("code", room.Code)
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
    
    // Execute
    h.StartGame(w, req)
    
    // Verify
    assert.Equal(t, http.StatusOK, w.Code)
    
    // Check room state updated
    updatedRoom, _ := store.GetRoom(room.Code)
    assert.Equal(t, game.StateCountdown, updatedRoom.State)
    
    // Verify roles assigned
    for _, p := range updatedRoom.Players {
        assert.NotNil(t, p.Role)
    }
}

func TestLeaveRoom(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    
    // Create room with player
    room, _ := store.CreateRoom()
    player := game.NewPlayer("p1", "Alice", "s1")
    room.AddPlayer(player)
    store.UpdateRoom(room)
    
    tests := []struct {
        name         string
        setupReq     func() *http.Request
        wantStatus   int
        wantRedirect string
        checkRoom    func(*testing.T, *game.Room)
    }{
        {
            name: "successful leave",
            setupReq: func() *http.Request {
                req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
                req.AddCookie(&http.Cookie{
                    Name:  "player_" + room.Code,
                    Value: player.ID,
                })
                return req
            },
            wantStatus:   http.StatusSeeOther,
            wantRedirect: "/",
            checkRoom: func(t *testing.T, r *game.Room) {
                assert.Equal(t, 0, len(r.Players))
            },
        },
        {
            name: "leave without cookie",
            setupReq: func() *http.Request {
                return httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
            },
            wantStatus: http.StatusUnauthorized,
            checkRoom: func(t *testing.T, r *game.Room) {
                assert.Equal(t, 1, len(r.Players))
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := tt.setupReq()
            w := httptest.NewRecorder()
            
            // Add Chi context
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("code", room.Code)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            
            // Execute
            h.LeaveRoom(w, req)
            
            // Verify
            assert.Equal(t, tt.wantStatus, w.Code)
            
            if tt.wantRedirect != "" {
                assert.Equal(t, tt.wantRedirect, w.Header().Get("Location"))
            }
            
            // Check room state
            updatedRoom, _ := store.GetRoom(room.Code)
            tt.checkRoom(t, updatedRoom)
        })
    }
}

func TestConcurrentJoins(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    
    room, _ := store.CreateRoom()
    
    // Simulate concurrent joins
    var wg sync.WaitGroup
    results := make([]int, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            
            url := fmt.Sprintf("/room/%s?name=Player%d", room.Code, idx)
            req := httptest.NewRequest("GET", url, nil)
            w := httptest.NewRecorder()
            
            // Add Chi context
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("code", room.Code)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            
            h.JoinRoom(w, req)
            results[idx] = w.Code
        }(i)
    }
    
    wg.Wait()
    
    // Verify results
    successCount := 0
    for _, status := range results {
        if status == http.StatusOK {
            successCount++
        }
    }
    
    // Should allow up to MaxPlayers
    assert.LessOrEqual(t, successCount, room.MaxPlayers)
    
    // Check final room state
    finalRoom, _ := store.GetRoom(room.Code)
    assert.LessOrEqual(t, len(finalRoom.Players), room.MaxPlayers)
}
```

### Day 8-9: SSE Edge Cases

#### SSE Error Scenarios
```go
func TestSSEErrorScenarios(t *testing.T) {
    t.Run("disconnection during stream", func(t *testing.T) {
        store := store.NewMemoryStore()
        h := New(store)
        
        room, _ := store.CreateRoom()
        player := game.NewPlayer("p1", "Alice", "s1")
        room.AddPlayer(player)
        store.UpdateRoom(room)
        
        // Create cancellable context
        ctx, cancel := context.WithCancel(context.Background())
        
        req := httptest.NewRequest("GET", "/sse/lobby/"+room.Code, nil)
        req = req.WithContext(ctx)
        req.AddCookie(&http.Cookie{
            Name:  "player_" + room.Code,
            Value: player.ID,
        })
        
        w := httptest.NewRecorder()
        
        // Start SSE in goroutine
        done := make(chan bool)
        go func() {
            h.StreamLobby(w, req)
            done <- true
        }()
        
        // Wait for initial message
        time.Sleep(50 * time.Millisecond)
        
        // Cancel context (simulate disconnect)
        cancel()
        
        // Verify handler exits cleanly
        select {
        case <-done:
            // Good, handler exited
        case <-time.After(1 * time.Second):
            t.Error("Handler didn't exit after context cancel")
        }
    })
    
    t.Run("room deleted during stream", func(t *testing.T) {
        // Test graceful handling when room is deleted mid-stream
        // Implementation depends on your cleanup strategy
    })
}
```

### Day 10: E2E Infrastructure Setup

#### Install Playwright
```bash
cd nix/app
mkdir e2e
cd e2e
npm init -y
npm install --save-dev @playwright/test typescript
npx playwright install
```

#### Create Basic E2E Test
```typescript
// e2e/tests/smoke.spec.ts
import { test, expect } from '@playwright/test';

test('application loads', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/Treacherest/);
    await expect(page.locator('h1')).toContainText('Create or Join');
});
```

## Coverage Monitoring Script

Create a script to track progress:

```bash
#!/bin/bash
# coverage-monitor.sh

echo "Running coverage analysis..."
cd nix/app

# Run tests with coverage
go test -coverprofile=coverage.out ./... > /dev/null 2>&1

# Generate report
echo -e "\nCoverage by Package:"
echo "===================="
go test -cover ./... | grep -E "coverage:|FAIL" | sort

# Total coverage
echo -e "\nTotal Coverage:"
echo "=============="
go tool cover -func=coverage.out | grep total | awk '{print $3}'

# Check if we met goal
TOTAL=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$TOTAL >= 80" | bc -l) )); then
    echo -e "\n✅ Coverage goal achieved!"
else
    echo -e "\n❌ Coverage below 80% target"
    echo "   Need to improve by: $(echo "80 - $TOTAL" | bc)%"
fi

# Generate HTML report
go tool cover -html=coverage.out -o build/coverage/coverage.html
echo -e "\nDetailed report: build/coverage/coverage.html"
```

## Daily Checklist

### Before Starting
- [ ] Run existing tests to ensure nothing broken
- [ ] Check current coverage for package being worked on
- [ ] Review coverage HTML report for gaps

### While Coding
- [ ] Write test first (TDD)
- [ ] Run tests frequently
- [ ] Check coverage after each test added
- [ ] Commit after each meaningful test addition

### End of Day
- [ ] Run full test suite
- [ ] Generate coverage report
- [ ] Document any blockers or issues
- [ ] Push all test code

## Success Metrics

### Week 1 Goals
- [ ] cmd/server coverage ≥ 80%
- [ ] internal/views coverage ≥ 70%
- [ ] All tests run in < 10 seconds
- [ ] No flaky tests

### Week 2 Goals
- [ ] internal/handlers coverage ≥ 85%
- [ ] E2E infrastructure operational
- [ ] At least 5 E2E scenarios covered
- [ ] Total project coverage ≥ 80%

## Troubleshooting

### Common Issues

1. **Chi Route Parameters Not Working**
   ```go
   // Always add Chi context in tests
   rctx := chi.NewRouteContext()
   rctx.URLParams.Add("paramName", "paramValue")
   req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
   ```

2. **Template Rendering Errors**
   ```go
   // Use require.NoError to catch template issues
   err := component.Render(context.Background(), buf)
   require.NoError(t, err, "Template should render without error")
   ```

3. **SSE Test Timing**
   ```go
   // Use channels instead of sleep
   messageReceived := make(chan bool)
   // ... in handler
   messageReceived <- true
   // ... in test
   select {
   case <-messageReceived:
       // Continue test
   case <-time.After(1 * time.Second):
       t.Fatal("Timeout waiting for SSE message")
   }
   ```

## Summary

This implementation guide provides:
1. **Day-by-day plan** to achieve coverage goals
2. **Concrete test examples** for each package
3. **Common patterns** to follow
4. **Monitoring tools** to track progress
5. **Troubleshooting tips** for common issues

Following this guide should achieve >80% test coverage within 2 weeks while establishing sustainable testing practices for future development.