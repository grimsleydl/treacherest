# Phase 1: Specific Issues Identified

## Critical Issues Found

### 1. Hardcoded HTML in Join Form
**File**: `/workspace/nix/app/internal/handlers/pages.go`
**Lines**: 90-103
**Issue**: Raw HTML written directly with `w.Write()` instead of using templ components

```go
w.Write([]byte(`
    <html>
    <body style="background: #1a1a2e; color: #eee; font-family: sans-serif;">
        <div style="max-width: 400px; margin: 100px auto; text-align: center;">
            <h1>Join Room ` + roomCode + `</h1>
            <form method="GET">
                <input type="text" name="name" placeholder="Enter your name" required 
                       style="padding: 10px; font-size: 16px; width: 200px;">
                <button type="submit" style="padding: 10px 20px; font-size: 16px;">Join</button>
            </form>
        </div>
    </body>
    </html>
`))
```

**Fix Required**: Create proper templ component for join form

### 2. SSE Implementation Gaps

#### Missing Test Coverage:
1. **SSE Message Format Tests**: No tests verify the actual datastar format
2. **Concurrent Connection Tests**: No load testing for multiple SSE clients
3. **Reconnection Tests**: No tests for client reconnection scenarios
4. **Selector Verification**: No tests ensuring selectors match HTML IDs

#### Current Test Limitations:
- Tests skip renderLobby/renderGame selector verification (lines 336-345)
- No tests for concurrent broadcasts
- No tests for memory leaks in event subscriptions
- No performance benchmarks

### 3. Event Bus Potential Issues
**File**: `/workspace/nix/app/internal/handlers/handlers.go`
**Concern**: No visible cleanup or connection limits

Potential issues:
- Memory leaks from uncleaned subscriptions
- No connection limits
- No heartbeat/keepalive mechanism
- No event replay for reconnections

### 4. Missing Error Handling
Several areas lack proper error handling:
- SSE write failures not handled
- Context cancellation during rendering
- Panic recovery in SSE handlers
- Template rendering errors

## Non-Critical but Important

### 1. Performance Baselines Missing
No benchmarks exist for:
- Player join operations
- SSE broadcast latency
- Memory usage per connection
- Maximum concurrent connections

### 2. Browser Test Infrastructure
Currently no E2E tests for:
- Join form interaction
- SSE real-time updates in browser
- Multi-tab scenarios
- Network disconnection handling

### 3. Session Management
Current implementation uses simple cookies but lacks:
- Session expiration handling
- Secure flag on cookies (for HTTPS)
- Proper session invalidation

## Verification Needed

### 1. Datastar Integration
Need to verify:
- Correct selector format (#lobby-container, #game-container)
- Merge mode settings (morph vs replace)
- Fragment wrapping requirements
- Script execution format

### 2. Template Rendering
Check if:
- Templates properly escape user input
- IDs match expected selectors
- Fragments are properly structured

## Quick Wins

1. **Fix hardcoded HTML** - Simple template creation
2. **Add SSE format tests** - Verify datastar compatibility
3. **Add benchmark suite** - Establish baselines
4. **Add connection cleanup** - Prevent memory leaks

## Risk Areas

1. **SSE Scalability** - No current limits or throttling
2. **Memory Leaks** - Event subscriptions may not clean up
3. **Race Conditions** - Concurrent access to rooms/players
4. **Browser Compatibility** - SSE reconnection behavior varies

---

This analysis provides specific technical details for implementing the Phase 1 plan.