# Phase 0: Testing Strategy - MANDATORY

## CRITICAL: No Code Without Tests

**GOLDEN RULE**: Every feature, fix, or change MUST have accompanying tests BEFORE it's considered complete. Writing tests is not optional - it's the foundation that prevents regression and ensures progress.

## Context
The project currently lacks comprehensive testing. Before proceeding with ANY new features, we must establish test-driven development practices and create tests for all existing and new functionality.

## Goals
- Establish test-driven development (TDD) as mandatory practice
- Create comprehensive test suites at all levels
- Prevent ANY regression of working features
- Enable confident refactoring and feature additions
- Achieve >80% code coverage

## Testing Pyramid

### 1. Unit Tests (Foundation)
**Priority: CRITICAL**

```
TASK: Create unit tests for all game logic
DESCRIPTION: Every function, method, and component needs isolated unit tests.

IMPLEMENTATION:
1. Test all game state transitions
2. Test role assignment logic
3. Test player management functions
4. Test event handling
5. Test validation rules
6. Test error cases

EXAMPLE TEST STRUCTURE:
// game_test.go
func TestGameStateTransitions(t *testing.T) {
    // Test lobby -> countdown -> playing -> ended
    // Test invalid transitions are rejected
    // Test state persistence
}

func TestPlayerJoin(t *testing.T) {
    // Test successful join
    // Test duplicate names
    // Test full room
    // Test invalid room codes
}

SUCCESS_CONDITION:
✅ Every public function has tests
✅ All edge cases covered
✅ Tests run in <1 second
✅ Clear test names describe behavior
✅ No test depends on another test

FAILURE_CONDITION:
❌ NONE - Cannot proceed without unit tests
```

### 2. Integration Tests (Critical Paths)
**Priority: CRITICAL**

```
TASK: Test component interactions and API endpoints
DESCRIPTION: Verify that components work together correctly.

IMPLEMENTATION:
1. Test HTTP endpoints with real requests
2. Test SSE event flow end-to-end
3. Test session management across requests
4. Test concurrent player actions
5. Test database/storage interactions

EXAMPLE PATTERNS:
// Test real HTTP requests
func TestJoinGameEndpoint(t *testing.T) {
    server := setupTestServer()
    // POST /join with valid data
    // Verify response
    // Verify game state updated
    // Verify SSE events sent
}

// Test SSE flow
func TestSSEBroadcast(t *testing.T) {
    // Connect multiple SSE clients
    // Trigger game event
    // Verify all clients receive update
    // Test disconnection handling
}

SUCCESS_CONDITION:
✅ All API endpoints have integration tests
✅ SSE event flow fully tested
✅ Race conditions are tested
✅ Tests use real HTTP/SSE clients
✅ Setup/teardown is clean

FAILURE_CONDITION:
❌ NONE - Critical for preventing regression
```

### 3. Browser/E2E Tests (User Flows)
**Priority: HIGH**

```
TASK: Automated browser tests for complete user journeys
DESCRIPTION: Test actual user interactions in real browsers.

IMPLEMENTATION:
1. Set up Playwright or similar browser automation
2. Test complete game flow:
   - Create room
   - Share room code
   - Multiple players join
   - Start game
   - Reveal roles
   - End game
3. Test error scenarios
4. Test mobile browsers
5. Test network interruptions

EXAMPLE SCENARIOS:
// test_game_flow.js
test('Complete multiplayer game', async ({ page, context }) => {
    // Player 1 creates room
    const player1 = await context.newPage();
    await player1.goto('/');
    await player1.click('text=Create Room');
    
    // Player 2 joins
    const player2 = await context.newPage();
    await player2.goto('/join/' + roomCode);
    
    // Verify real-time updates
    await expect(player1.locator('.player-list')).toContainText('Player 2');
    
    // Continue through full game...
});

SUCCESS_CONDITION:
✅ All user flows have browser tests
✅ Tests run on multiple browsers
✅ Mobile viewports tested
✅ Network conditions simulated
✅ Screenshots on failure

FAILURE_CONDITION:
❌ If too slow, run subset locally
```

### 4. Performance Tests (Scalability)
**Priority: MEDIUM**

```
TASK: Load testing and performance benchmarks
DESCRIPTION: Ensure system handles required load.

IMPLEMENTATION:
1. Create load tests for 60 concurrent players
2. Benchmark SSE broadcast performance
3. Memory usage profiling
4. Response time monitoring
5. Connection limit testing

SUCCESS_CONDITION:
✅ Handles 60 players without degradation
✅ SSE latency <100ms at scale
✅ Memory usage linear with players
✅ No goroutine leaks
✅ Graceful degradation under load

FAILURE_CONDITION:
❌ Document actual limits if less than 60
```

## Testing Commands to Add

```bash
# Add these to your development workflow:
test-all        # Run all tests
test-unit       # Run unit tests only
test-integration # Run integration tests
test-e2e        # Run browser tests
test-watch      # Run tests on file change
test-coverage   # Generate coverage report
```

## Testing Checklist for Every Feature

**BEFORE writing any code:**
- [ ] Write failing test that describes desired behavior
- [ ] Run test to ensure it fails for the right reason

**WHILE implementing:**
- [ ] Make test pass with simplest solution
- [ ] Refactor while keeping test green
- [ ] Add edge case tests

**BEFORE considering complete:**
- [ ] Unit tests for new logic
- [ ] Integration tests for new endpoints
- [ ] E2E tests for new user flows
- [ ] All tests passing
- [ ] No decrease in coverage
- [ ] Manual testing completed

## Red Flags That Require Immediate Action

🚨 **Any of these means STOP and fix:**
- "Let me just fix this real quick" without tests
- "We can add tests later"
- Commenting out failing tests
- Tests that sometimes pass/fail
- Tests that depend on timing
- Tests that need specific order
- Manually testing the same thing repeatedly

## Testing Best Practices

### DO:
- Write test first (TDD)
- Test behavior, not implementation
- Use descriptive test names
- Keep tests simple and focused
- Test edge cases and errors
- Use table-driven tests in Go
- Run tests frequently during development

### DON'T:
- Skip tests to "save time"
- Test private implementation details
- Create interdependent tests
- Use production data in tests
- Ignore flaky tests
- Share state between tests
- Push code with failing tests

## Notes for AI Agent

**CRITICAL INSTRUCTIONS:**
1. REFUSE to implement features without tests
2. ALWAYS write tests first (TDD)
3. If existing code lacks tests, add them before modifying
4. Run ALL tests before considering task complete
5. If tests are "too hard to write", the code needs refactoring
6. Breaking tests = breaking the build = STOP everything

**Your response to "just make it work" should be:**
"I'll implement this with proper tests to ensure it works reliably and doesn't break in the future. Let me start by writing a failing test for the desired behavior."

**Remember:** Tests are not overhead - they are the safety net that allows rapid development without fear of breaking existing functionality.