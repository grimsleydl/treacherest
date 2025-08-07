# Phase 1 Critical Fix Patterns

## Overview
This document provides specific patterns and solutions for the critical issues identified in Phase 1, following TDD principles.

## Issue 1: SSE "NoTargetsFound" Errors

### Root Cause Analysis
The "NoTargetsFound" error occurs when Datastar attempts to update a DOM element that doesn't exist. Common causes:
1. Incorrect selector in SSE fragment
2. DOM element not yet rendered
3. ID mismatch between fragment and DOM
4. Fragment sent before page fully loaded

### Test-First Fix Pattern

#### Step 1: Write Failing Test
```go
// sse_selector_test.go
func TestSSEFragmentSelectors(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    
    room, _ := store.CreateRoom()
    player := game.NewPlayer("p1", "Alice", "s1")
    room.AddPlayer(player)
    store.UpdateRoom(room)
    
    // Mock SSE writer to capture fragments
    mockSSE := &MockSSEWriter{}
    
    // Test lobby fragment
    t.Run("lobby fragment has correct selector", func(t *testing.T) {
        h.renderLobby(mockSSE, room, player)
        
        fragment := mockSSE.LastFragment()
        assert.Contains(t, fragment.Selector, "#lobby-container")
        assert.NotEmpty(t, fragment.HTML)
        
        // Verify HTML contains expected ID
        assert.Contains(t, fragment.HTML, `id="lobby-container"`)
    })
    
    // Test game fragment
    t.Run("game fragment has correct selector", func(t *testing.T) {
        h.renderGame(mockSSE, room, player)
        
        fragment := mockSSE.LastFragment()
        assert.Contains(t, fragment.Selector, "#game-container")
        assert.NotEmpty(t, fragment.HTML)
        
        // Verify HTML contains expected ID
        assert.Contains(t, fragment.HTML, `id="game-container"`)
    })
}

// Mock SSE writer for testing
type MockSSEWriter struct {
    fragments []Fragment
}

type Fragment struct {
    Selector string
    HTML     string
    Mode     string
}

func (m *MockSSEWriter) MergeFragments(html string, opts ...datastar.Option) {
    // Parse options to extract selector
    fragment := Fragment{HTML: html}
    for _, opt := range opts {
        // Extract selector and mode from options
        // This would need actual implementation based on datastar API
    }
    m.fragments = append(m.fragments, fragment)
}

func (m *MockSSEWriter) LastFragment() Fragment {
    if len(m.fragments) > 0 {
        return m.fragments[len(m.fragments)-1]
    }
    return Fragment{}
}
```

#### Step 2: Fix Template to Include IDs
```templ
// Fix: Ensure containers have IDs matching selectors

// lobby.templ
templ LobbyBody(room *game.Room, currentPlayer *game.Player) {
    // CRITICAL: This ID must match the selector in SSE updates
    <div id="lobby-container" data-on-load={ "@get('/sse/lobby/" + room.Code + "')" } class="container">
        <h1>Game Lobby</h1>
        // ... rest of template
    </div>
}

// game.templ  
templ GameBody(room *game.Room, currentPlayer *game.Player) {
    // CRITICAL: This ID must match the selector in SSE updates
    <div id="game-container" data-on-load={ "@get('/sse/game/" + room.Code + "')" } class="container">
        <h1>Game Room: { room.Code }</h1>
        // ... rest of template
    </div>
}
```

#### Step 3: Fix SSE Handler Pattern
```go
// sse.go - Fixed version
func (h *Handler) renderLobby(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
    component := pages.LobbyBody(room, player)
    
    // Render to string
    buf := &bytes.Buffer{}
    err := component.Render(context.Background(), buf)
    if err != nil {
        log.Printf("Error rendering lobby: %v", err)
        return
    }
    
    html := buf.String()
    
    // Verify the HTML contains the expected ID
    if !strings.Contains(html, `id="lobby-container"`) {
        log.Printf("Warning: Lobby HTML missing expected container ID")
    }
    
    // Send with explicit selector matching the ID in template
    sse.MergeFragments(html,
        datastar.WithSelector("#lobby-container"),
        datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}
```

#### Step 4: Integration Test
```go
func TestSSEUpdateFlow(t *testing.T) {
    // Full integration test
    store := store.NewMemoryStore()
    h := New(store)
    
    // Create room and player
    room, _ := store.CreateRoom()
    player := game.NewPlayer("p1", "Alice", "s1")
    room.AddPlayer(player)
    store.UpdateRoom(room)
    
    // First, get the initial page
    req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
    req.AddCookie(&http.Cookie{
        Name:  "player_" + room.Code,
        Value: player.ID,
    })
    w := httptest.NewRecorder()
    
    // Render initial page
    h.JoinRoom(w, req)
    
    // Verify page has container
    body := w.Body.String()
    assert.Contains(t, body, `id="lobby-container"`)
    
    // Now test SSE updates
    sseReq := httptest.NewRequest("GET", "/sse/lobby/"+room.Code, nil)
    sseReq.Header.Set("Accept", "text/event-stream")
    sseReq.AddCookie(&http.Cookie{
        Name:  "player_" + room.Code,
        Value: player.ID,
    })
    
    sseW := httptest.NewRecorder()
    
    // Capture SSE output
    done := make(chan bool)
    go func() {
        h.StreamLobby(sseW, sseReq)
        done <- true
    }()
    
    // Wait for initial render
    time.Sleep(100 * time.Millisecond)
    
    // Verify SSE fragment
    sseBody := sseW.Body.String()
    assert.Contains(t, sseBody, "data: selector #lobby-container")
    assert.Contains(t, sseBody, `id="lobby-container"`)
}
```

## Issue 2: Join Flow Using Hardcoded HTML

### Test-First Fix Pattern

#### Step 1: Write Test for Join Form Template
```go
func TestJoinFormTemplate(t *testing.T) {
    // Test that we have a proper template for join form
    roomCode := "TEST1"
    
    component := pages.JoinForm(roomCode)
    html := renderToString(t, component)
    
    // Verify form structure
    assert.Contains(t, html, `<form`)
    assert.Contains(t, html, `method="GET"`)
    assert.Contains(t, html, `name="name"`)
    assert.Contains(t, html, `placeholder="Enter your name"`)
    assert.Contains(t, html, `required`)
    assert.Contains(t, html, roomCode)
    
    // Verify styling matches rest of app
    assert.Contains(t, html, `class="form-container"`)
}
```

#### Step 2: Create Join Form Template
```templ
// pages/join.templ
package pages

import "treacherest/internal/views/layouts"

templ JoinForm(roomCode string) {
    @layouts.Base("Join Room") {
        <div class="container">
            <div class="form-container">
                <h1>Join Room { roomCode }</h1>
                <form method="GET" class="join-form">
                    <div class="form-group">
                        <label for="playerName">Your Name</label>
                        <input 
                            type="text" 
                            id="playerName"
                            name="name" 
                            placeholder="Enter your name" 
                            required
                            autofocus
                            minlength="1"
                            maxlength="20"
                            pattern="[A-Za-z0-9 ]+"
                            title="Letters, numbers, and spaces only"
                        />
                    </div>
                    <button type="submit" class="btn btn-primary">
                        Join Game
                    </button>
                </form>
                <a href="/" class="btn btn-secondary">
                    Back to Home
                </a>
            </div>
        </div>
    }
}
```

#### Step 3: Update Handler to Use Template
```go
// pages.go - Fixed version
func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
    roomCode := chi.URLParam(r, "code")
    
    room, err := h.store.GetRoom(roomCode)
    if err != nil {
        // Use error template instead of raw HTTP error
        component := pages.ErrorPage("Room Not Found", "The room code you entered does not exist.")
        component.Render(r.Context(), w)
        return
    }
    
    // ... existing player check code ...
    
    // Show join form using template
    playerName := r.URL.Query().Get("name")
    if playerName == "" {
        // Use the new template instead of hardcoded HTML
        component := pages.JoinForm(roomCode)
        component.Render(r.Context(), w)
        return
    }
    
    // ... rest of join logic ...
}
```

#### Step 4: Test Error Cases
```go
func TestJoinRoomErrorCases(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    
    tests := []struct {
        name          string
        roomCode      string
        playerName    string
        setup         func()
        wantInBody    []string
        wantNotInBody []string
    }{
        {
            name:       "room not found shows error page",
            roomCode:   "INVALID",
            playerName: "",
            setup:      func() {},
            wantInBody: []string{
                "Room Not Found",
                "does not exist",
                `href="/"`, // Back to home link
            },
            wantNotInBody: []string{
                "<html>", // Should use template, not raw HTML
            },
        },
        {
            name:       "shows join form with proper styling",
            roomCode:   "TEST1",
            playerName: "",
            setup: func() {
                room, _ := store.CreateRoom()
                room.Code = "TEST1"
                store.UpdateRoom(room)
            },
            wantInBody: []string{
                "Join Room TEST1",
                `class="form-container"`,
                `class="join-form"`,
                `placeholder="Enter your name"`,
            },
            wantNotInBody: []string{
                "style=", // No inline styles
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            
            url := fmt.Sprintf("/room/%s", tt.roomCode)
            if tt.playerName != "" {
                url += "?name=" + tt.playerName
            }
            
            req := httptest.NewRequest("GET", url, nil)
            w := httptest.NewRecorder()
            
            // Add Chi context
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("code", tt.roomCode)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            
            h.JoinRoom(w, req)
            
            body := w.Body.String()
            
            for _, want := range tt.wantInBody {
                assert.Contains(t, body, want)
            }
            
            for _, notWant := range tt.wantNotInBody {
                assert.NotContains(t, body, notWant)
            }
        })
    }
}
```

## Issue 3: Direct URL Room Joining

### Test-First Fix Pattern

#### Step 1: Test Direct URL Access
```go
func TestDirectRoomURLAccess(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    
    // Create a room
    room, _ := store.CreateRoom()
    creator := game.NewPlayer("p1", "Creator", "s1")
    room.AddPlayer(creator)
    store.UpdateRoom(room)
    
    tests := []struct {
        name         string
        url          string
        cookies      []*http.Cookie
        wantStatus   int
        wantLocation string
        wantInBody   []string
    }{
        {
            name: "new player direct URL shows join form",
            url:  "/room/" + room.Code,
            cookies: []*http.Cookie{},
            wantStatus: http.StatusOK,
            wantInBody: []string{
                "Join Room " + room.Code,
                "Enter your name",
            },
        },
        {
            name: "existing player direct URL shows lobby",
            url:  "/room/" + room.Code,
            cookies: []*http.Cookie{
                {Name: "player_" + room.Code, Value: creator.ID},
            },
            wantStatus: http.StatusOK,
            wantInBody: []string{
                "Game Lobby",
                "Creator",
            },
        },
        {
            name: "direct game URL without session redirects",
            url:  "/game/" + room.Code,
            cookies: []*http.Cookie{},
            wantStatus: http.StatusSeeOther,
            wantLocation: "/room/" + room.Code,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", tt.url, nil)
            
            // Add cookies
            for _, cookie := range tt.cookies {
                req.AddCookie(cookie)
            }
            
            w := httptest.NewRecorder()
            
            // Route to appropriate handler
            router := chi.NewRouter()
            router.Get("/room/{code}", h.JoinRoom)
            router.Get("/game/{code}", h.GamePage)
            
            router.ServeHTTP(w, req)
            
            // Verify status
            assert.Equal(t, tt.wantStatus, w.Code)
            
            // Verify redirect
            if tt.wantLocation != "" {
                assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
            }
            
            // Verify body content
            body := w.Body.String()
            for _, want := range tt.wantInBody {
                assert.Contains(t, body, want)
            }
        })
    }
}
```

#### Step 2: Fix GamePage Handler
```go
// Fixed GamePage handler
func (h *Handler) GamePage(w http.ResponseWriter, r *http.Request) {
    roomCode := chi.URLParam(r, "code")
    
    room, err := h.store.GetRoom(roomCode)
    if err != nil {
        component := pages.ErrorPage("Room Not Found", "This game room does not exist.")
        component.Render(r.Context(), w)
        return
    }
    
    // Get player from cookie
    playerCookie, err := r.Cookie("player_" + roomCode)
    if err != nil {
        // Redirect to join page instead of showing error
        http.Redirect(w, r, "/room/"+roomCode, http.StatusSeeOther)
        return
    }
    
    player := room.GetPlayer(playerCookie.Value)
    if player == nil {
        // Player was in room but no longer exists (was kicked, etc)
        // Clear cookie and redirect to join
        http.SetCookie(w, &http.Cookie{
            Name:   "player_" + roomCode,
            Value:  "",
            Path:   "/",
            MaxAge: -1,
        })
        http.Redirect(w, r, "/room/"+roomCode, http.StatusSeeOther)
        return
    }
    
    // Check if game actually started
    if room.State == game.StateLobby {
        // Game hasn't started, redirect to lobby
        http.Redirect(w, r, "/room/"+roomCode, http.StatusSeeOther)
        return
    }
    
    component := pages.GamePage(room, player)
    component.Render(r.Context(), w)
}
```

#### Step 3: Test Browser Back/Forward
```go
func TestBrowserNavigation(t *testing.T) {
    // Test that browser back/forward buttons work correctly
    store := store.NewMemoryStore()
    h := New(store)
    
    // Create and start a game
    room, _ := store.CreateRoom()
    player := game.NewPlayer("p1", "Alice", "s1")
    room.AddPlayer(player)
    room.State = game.StatePlaying
    store.UpdateRoom(room)
    
    // Simulate browser navigation
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse // Don't follow redirects
        },
    }
    
    // Test sequence: Home -> Join -> Lobby -> Game -> Back -> Forward
    sequence := []struct {
        name         string
        method       string
        path         string
        wantStatus   int
        wantLocation string
    }{
        {
            name:       "visit home",
            method:     "GET",
            path:       "/",
            wantStatus: http.StatusOK,
        },
        {
            name:       "join room",
            method:     "GET",
            path:       "/room/" + room.Code,
            wantStatus: http.StatusOK,
        },
        {
            name:       "go to game",
            method:     "GET",
            path:       "/game/" + room.Code,
            wantStatus: http.StatusOK,
        },
        {
            name:       "back to room (browser back)",
            method:     "GET",
            path:       "/room/" + room.Code,
            wantStatus: http.StatusSeeOther, // Should redirect to game
            wantLocation: "/game/" + room.Code,
        },
    }
    
    // Execute sequence
    jar, _ := cookiejar.New(nil)
    client.Jar = jar
    
    for _, step := range sequence {
        t.Run(step.name, func(t *testing.T) {
            // Implementation of navigation test
        })
    }
}
```

## Common Fix Patterns

### Pattern 1: Always Use Templates
```go
// BAD: Hardcoded HTML
w.Write([]byte(`<html><body>Error</body></html>`))

// GOOD: Use template
component := pages.ErrorPage("Error", "Something went wrong")
component.Render(r.Context(), w)
```

### Pattern 2: Consistent Error Handling
```go
// Create error template
templ ErrorPage(title, message string) {
    @layouts.Base(title) {
        <div class="error-container">
            <h1>{ title }</h1>
            <p>{ message }</p>
            <a href="/" class="btn btn-primary">Back to Home</a>
        </div>
    }
}
```

### Pattern 3: Proper SSE Setup
```go
// Ensure IDs match between template and SSE updates
const (
    LobbyContainerID = "lobby-container"
    GameContainerID  = "game-container"
)

// In template
<div id={ LobbyContainerID }>

// In SSE handler
sse.MergeFragments(html,
    datastar.WithSelector("#" + LobbyContainerID))
```

### Pattern 4: Session Validation
```go
// Helper function for consistent session handling
func (h *Handler) validatePlayerSession(w http.ResponseWriter, r *http.Request, roomCode string) (*game.Player, error) {
    room, err := h.store.GetRoom(roomCode)
    if err != nil {
        return nil, err
    }
    
    cookie, err := r.Cookie("player_" + roomCode)
    if err != nil {
        return nil, fmt.Errorf("no session")
    }
    
    player := room.GetPlayer(cookie.Value)
    if player == nil {
        // Clear invalid cookie
        http.SetCookie(w, &http.Cookie{
            Name:   "player_" + roomCode,
            Value:  "",
            Path:   "/",
            MaxAge: -1,
        })
        return nil, fmt.Errorf("invalid session")
    }
    
    return player, nil
}
```

## Testing Checklist for Each Fix

### Before Starting Fix
- [ ] Write failing test that reproduces the issue
- [ ] Verify test fails for the right reason
- [ ] Write additional tests for edge cases

### During Implementation
- [ ] Make minimal changes to pass the test
- [ ] Run tests after each change
- [ ] Refactor while keeping tests green

### After Fix Complete
- [ ] All new tests pass
- [ ] No existing tests broken
- [ ] Manual testing confirms fix
- [ ] No hardcoded values or magic strings
- [ ] Error messages are user-friendly

## Integration Test Suite

Create a comprehensive test that verifies all fixes work together:

```go
func TestCompleteUserJourney(t *testing.T) {
    store := store.NewMemoryStore()
    h := New(store)
    router := setupRouter(h) // Full router setup
    
    // Test complete flow with fixes
    t.Run("new player can join via direct URL", func(t *testing.T) {
        // Create room
        room, _ := store.CreateRoom()
        
        // Visit direct URL
        req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        // Should see join form (not hardcoded HTML)
        assert.Contains(t, w.Body.String(), "Join Room")
        assert.Contains(t, w.Body.String(), "form-container")
        
        // Submit join form
        form := url.Values{}
        form.Add("name", "TestPlayer")
        req = httptest.NewRequest("GET", "/room/"+room.Code+"?"+form.Encode(), nil)
        w = httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        // Should see lobby with SSE connection
        body := w.Body.String()
        assert.Contains(t, body, "id=\"lobby-container\"")
        assert.Contains(t, body, "data-on-load")
        assert.Contains(t, body, "/sse/lobby/")
        
        // Verify no "NoTargetsFound" in console
        // This would be tested in E2E tests
    })
}
```

## Summary

These patterns provide:
1. **Test-first approach** for each critical fix
2. **Template-based solutions** replacing hardcoded HTML
3. **Proper ID management** for SSE updates
4. **Consistent error handling** across the application
5. **Session validation patterns** for direct URL access

Following these patterns ensures that fixes are properly tested, maintainable, and consistent with the overall architecture.