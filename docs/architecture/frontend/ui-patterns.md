# Treacherest Frontend UI Patterns

## Overview
Treacherest uses server-side rendering with Templ templates and Datastar for reactivity. The frontend follows a "DOM as State" philosophy where the server sends HTML fragments via SSE, and the browser efficiently updates the DOM using Idiomorph.

## Core Frontend Principles

### 1. Server-Side Rendering First
- All UI logic lives on the server
- Templates generate complete HTML fragments
- No client-side state management needed
- Progressive enhancement over SPA complexity

### 2. DOM as State Philosophy
- The DOM itself is the application state
- Server sends new DOM fragments when state changes
- Idiomorph handles efficient DOM diffing
- No virtual DOM or client-side reconciliation

### 3. Datastar Reactivity Model
- Declarative attributes for interactivity
- SSE for real-time updates
- Event handlers via `data-on-*` attributes
- Automatic reconnection and error handling

## Templ Template Patterns

### 1. Page Structure Pattern
```templ
// Page component wraps entire page
templ PageName(data DataType) {
    @layouts.Base("Page Title") {
        @PageBody(data)
    }
}

// Body component for SSE updates
templ PageBody(data DataType) {
    <div id="container-id" data-on-load={"@get('/sse/endpoint')"}>
        // Page content
    </div>
}
```

### 2. Component Composition
```templ
// Reusable components
templ PlayerCard(player *Player) {
    <div class="player-card">
        <h3>{ player.Name }</h3>
        if player.Role != nil && player.Role.Revealed {
            <p>{ player.Role.Name }</p>
        }
    </div>
}

// Using components
templ PlayerList(players []*Player) {
    <div class="player-list">
        for _, player := range players {
            @PlayerCard(player)
        }
    </div>
}
```

### 3. Conditional Rendering
```templ
// State-based UI
templ GameControls(room *Room, isHost bool) {
    switch room.State {
    case StateLobby:
        if isHost && room.CanStart() {
            <button data-on-click={"@post('/room/" + room.Code + "/start')"}>
                Start Game
            </button>
        }
    case StatePlaying:
        @GameActions(room)
    case StateEnded:
        @GameResults(room)
    }
}
```

## Datastar Attribute Patterns

### 1. SSE Connection Pattern
```html
<!-- Auto-connect to SSE on page load -->
<div id="game-container" data-on-load="@get('/sse/game/ABC123')">
    <!-- Content updated via SSE -->
</div>
```

### 2. Action Patterns
```html
<!-- Simple POST action -->
<button data-on-click="@post('/room/ABC123/start')">
    Start Game
</button>

<!-- Form submission -->
<form data-on-submit="@post('/room/new')">
    <input name="playerName" required>
    <button type="submit">Create Room</button>
</form>

<!-- Conditional actions -->
<button 
    data-on-click="@post('/action')"
    data-if="canPerformAction"
>
    Perform Action
</button>
```

### 3. Real-time Update Patterns
```html
<!-- Container with specific ID for targeted updates -->
<div id="player-list">
    <!-- Server sends fragments targeting #player-list -->
</div>

<!-- Multiple update targets -->
<div id="game-state">...</div>
<div id="player-actions">...</div>
<div id="chat-messages">...</div>
```

## SSE Fragment Patterns

### 1. Full Container Updates
```go
// Server sends complete container replacement
func (h *Handler) renderGameState(sse *datastar.ServerSentEventGenerator, room *Room) {
    component := pages.GameBody(room, player)
    html := renderToString(component)
    
    sse.MergeFragments(html,
        datastar.WithSelector("#game-container"),
        datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}
```

### 2. Targeted Updates
```go
// Update specific UI sections
func (h *Handler) updatePlayerList(sse *datastar.ServerSentEventGenerator, players []*Player) {
    component := components.PlayerList(players)
    html := renderToString(component)
    
    sse.MergeFragments(html,
        datastar.WithSelector("#player-list"),
        datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}
```

### 3. Multi-Fragment Updates
```go
// Update multiple UI sections in one SSE message
func (h *Handler) updateGameUI(sse *datastar.ServerSentEventGenerator, game *Game) {
    // Update game state
    stateHTML := renderToString(components.GameState(game))
    sse.MergeFragments(stateHTML,
        datastar.WithSelector("#game-state"))
    
    // Update player actions
    actionsHTML := renderToString(components.PlayerActions(game, player))
    sse.MergeFragments(actionsHTML,
        datastar.WithSelector("#player-actions"))
}
```

## CSS Architecture

### 1. Utility-First Approach
```css
/* Base utilities */
.container { max-width: 1200px; margin: 0 auto; padding: 1rem; }
.card { background: var(--bg-secondary); border-radius: 8px; padding: 1rem; }
.button { padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; }

/* Component-specific */
.player-card { @apply card; margin-bottom: 0.5rem; }
.room-code { font-size: 2rem; font-weight: bold; text-align: center; }
```

### 2. State-Based Styling
```templ
// Dynamic classes based on state
templ PlayerItem(player *Player, isCurrentPlayer bool) {
    <div class={
        "player-item",
        templ.KV("current-player", isCurrentPlayer),
        templ.KV("has-role", player.Role != nil)
    }>
        { player.Name }
    </div>
}
```

### 3. Dark Mode Support
```css
:root {
    --bg-primary: #ffffff;
    --bg-secondary: #f5f5f5;
    --text-primary: #333333;
}

@media (prefers-color-scheme: dark) {
    :root {
        --bg-primary: #1a1a2e;
        --bg-secondary: #16213e;
        --text-primary: #eeeeeee;
    }
}
```

## Mobile-First Patterns

### 1. Responsive Layouts
```templ
templ GameLayout(room *Room) {
    <div class="game-layout">
        <!-- Mobile: Stack vertically -->
        <!-- Desktop: Side-by-side -->
        <div class="game-main">
            @GameBoard(room)
        </div>
        <div class="game-sidebar">
            @PlayerList(room.Players)
            @GameActions(room)
        </div>
    </div>
}
```

### 2. Touch-Friendly Interactions
```css
/* Large touch targets */
.button {
    min-height: 44px;
    min-width: 44px;
    padding: 12px 24px;
}

/* Prevent accidental taps */
.danger-button {
    margin-top: 2rem;
}

/* Visual feedback */
.button:active {
    transform: scale(0.98);
}
```

### 3. Progressive Enhancement
```templ
// Base functionality works without JavaScript
templ JoinForm(roomCode string) {
    <form method="GET" action={"/room/" + roomCode}>
        <input name="playerName" required>
        <button type="submit">Join Game</button>
    </form>
}
```

## Error Handling Patterns

### 1. User-Friendly Error Display
```templ
templ ErrorMessage(err error) {
    <div class="error-message" role="alert">
        <span class="error-icon">⚠️</span>
        <span class="error-text">{ getUserFriendlyError(err) }</span>
    </div>
}
```

### 2. Loading States
```templ
templ LoadingState(message string) {
    <div class="loading-state">
        <div class="spinner"></div>
        if message != "" {
            <p>{ message }</p>
        }
    </div>
}
```

### 3. Connection Status
```templ
templ ConnectionStatus(connected bool) {
    <div class={
        "connection-status",
        templ.KV("connected", connected),
        templ.KV("disconnected", !connected)
    }>
        if connected {
            <span>🟢 Connected</span>
        } else {
            <span>🔴 Reconnecting...</span>
        }
    </div>
}
```

## Performance Patterns

### 1. Efficient Re-renders
- Use specific container IDs for targeted updates
- Avoid re-rendering static content
- Batch related updates together

### 2. Asset Optimization
```html
<!-- Inline critical CSS -->
<style>
    /* Critical path CSS */
</style>

<!-- Defer non-critical CSS -->
<link rel="stylesheet" href="/static/styles.css" media="print" onload="this.media='all'">
```

### 3. Image Loading
```templ
templ PlayerAvatar(player *Player) {
    <img 
        src={"/static/avatars/" + player.AvatarID + ".webp"}
        alt={player.Name + " avatar"}
        loading="lazy"
        width="64"
        height="64"
    />
}
```

## Accessibility Patterns

### 1. Semantic HTML
```templ
templ GameRoom(room *Room) {
    <main>
        <h1>Game Room: { room.Code }</h1>
        <section aria-label="Players">
            @PlayerList(room.Players)
        </section>
        <section aria-label="Game Controls">
            @GameControls(room)
        </section>
    </main>
}
```

### 2. ARIA Attributes
```templ
templ CountdownTimer(seconds int) {
    <div 
        role="timer"
        aria-live="polite"
        aria-label={"Game starting in " + strconv.Itoa(seconds) + " seconds"}
    >
        { strconv.Itoa(seconds) }
    </div>
}
```

### 3. Keyboard Navigation
```html
<!-- Focusable elements in logical order -->
<button tabindex="0">Primary Action</button>
<button tabindex="0">Secondary Action</button>

<!-- Skip links -->
<a href="#main-content" class="skip-link">Skip to main content</a>
```

## Testing Patterns

### 1. Component Testing
```go
func TestPlayerCard(t *testing.T) {
    player := &Player{Name: "Alice", Role: &Role{Name: "Leader"}}
    component := PlayerCard(player)
    
    html := renderToString(component)
    assert.Contains(t, html, "Alice")
    assert.Contains(t, html, "Leader")
}
```

### 2. SSE Update Testing
```go
func TestSSEFragmentUpdate(t *testing.T) {
    // Create test SSE generator
    w := httptest.NewRecorder()
    sse := datastar.NewSSE(w, &http.Request{})
    
    // Send fragment
    h.renderGameState(sse, testRoom)
    
    // Verify SSE message format
    body := w.Body.String()
    assert.Contains(t, body, "event: datastar-merge-fragments")
    assert.Contains(t, body, "data: selector #game-container")
}
```

### 3. Integration Testing
```javascript
// Playwright test example
test('player can join game', async ({ page }) => {
    await page.goto('/');
    await page.fill('input[name="playerName"]', 'Alice');
    await page.click('button[type="submit"]');
    
    // Verify redirect to lobby
    await expect(page).toHaveURL(/\/room\/[A-Z0-9]{5}/);
    await expect(page.locator('.player-name')).toContainText('Alice');
});
```

## Common UI Components

### 1. Room Code Display
```templ
templ RoomCode(code string) {
    <div class="room-code-container">
        <label>Room Code:</label>
        <div class="room-code">{ code }</div>
        <button data-on-click={"navigator.clipboard.writeText('" + code + "')"}>
            Copy
        </button>
    </div>
}
```

### 2. Player Count
```templ
templ PlayerCount(current, max int) {
    <div class="player-count">
        <span class="current">{ strconv.Itoa(current) }</span>
        <span class="separator">/</span>
        <span class="max">{ strconv.Itoa(max) }</span>
        <span class="label">players</span>
    </div>
}
```

### 3. Timer Display
```templ
templ Timer(seconds int) {
    <div class="timer" data-seconds={strconv.Itoa(seconds)}>
        <span class="minutes">{ fmt.Sprintf("%02d", seconds/60) }</span>
        <span class="separator">:</span>
        <span class="seconds">{ fmt.Sprintf("%02d", seconds%60) }</span>
    </div>
}
```

## Best Practices Summary

1. **Keep Templates Simple**: Logic belongs in handlers, not templates
2. **Use Semantic HTML**: Proper elements and ARIA labels
3. **Mobile-First Design**: Start with mobile, enhance for desktop
4. **Progressive Enhancement**: Base functionality without JavaScript
5. **Efficient Updates**: Target specific containers for SSE updates
6. **Accessibility First**: Keyboard navigation, screen reader support
7. **Performance Matters**: Minimize re-renders, optimize assets
8. **Test Everything**: Component, integration, and E2E tests

This architecture provides a robust foundation for building interactive, real-time multiplayer games with excellent performance and user experience.