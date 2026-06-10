# Prioritized Implementation Tasks for Treacherest

## Current State Summary
- **Total Coverage**: 38.9% (Target: >80%)
- **Critical Issue**: Phase 0 testing must be completed before any other work
- **Blocking Pattern**: Phase 0 → Phase 1 → Phase 2 → Phase 3 → Phase 4

## Task Breakdown

### PHASE 0: Testing Infrastructure (HIGHEST PRIORITY - UNBLOCK ALL OTHER WORK)

#### 0.1 Server Package Tests [L] - CRITICAL PATH
**Impact**: Will add ~15-20% to total coverage
**Dependencies**: None
**Parallel Work**: Can be done alongside 0.2

```
Tasks:
1. [M] Write tests for cmd/server/main.go
   - Test main function initialization
   - Test environment variable handling
   - Test graceful shutdown
   - Mock server setup

2. [M] Write tests for cmd/server/server.go 
   - Test SetupServer with different configurations
   - Test middleware chain
   - Test route registration
   - Test static file serving
   - Test error handling

3. [S] Add integration tests for server startup
   - Test port binding
   - Test health check endpoint
   - Test graceful shutdown signals
```

#### 0.2 View/Template Tests [L] - CRITICAL PATH
**Impact**: Will add ~20-25% to total coverage
**Dependencies**: None
**Parallel Work**: Can be done alongside 0.1

```
Tasks:
1. [M] Set up template testing infrastructure
   - Create test helpers for rendering templates
   - Mock Datastar context
   - Create assertion helpers for HTML output

2. [M] Test view components
   - Test Base layout rendering
   - Test Home page component
   - Test LobbyPage and LobbyBody
   - Test GamePage and GameBody
   - Test getRoleClass helper

3. [S] Test Datastar attributes
   - Verify correct data-* attributes
   - Test SSE target selectors
   - Test event bindings
```

#### 0.3 Missing Handler Tests [M] - HIGH PRIORITY
**Impact**: Will add ~5-10% to total coverage
**Dependencies**: None
**Parallel Work**: Can be done independently

```
Tasks:
1. [S] Increase StreamGame coverage to 100%
   - Test error conditions
   - Test client disconnection
   - Test concurrent connections

2. [S] Increase CreateRoom coverage to 100%
   - Test error handling path
   - Test edge cases

3. [S] Add missing action handler tests
   - Test timeout scenarios
   - Test concurrent modifications
```

#### 0.4 Integration Test Suite [L] - CRITICAL
**Impact**: Validates system behavior, not coverage %
**Dependencies**: 0.1, 0.2, 0.3 should be mostly complete
**Parallel Work**: None - requires other tests first

```
Tasks:
1. [M] Create integration test framework
   - Set up test server with real components
   - Create test data fixtures
   - Add request/response helpers

2. [M] Test complete user flows
   - Create room → Join → Start → Play → End
   - Multiple players joining simultaneously
   - Player disconnection and reconnection
   - Session persistence

3. [S] Test SSE integration
   - Multiple SSE connections
   - Event broadcasting
   - Reconnection handling
   - Fragment targeting
```

#### 0.5 E2E Test Infrastructure [XL] - HIGH PRIORITY
**Impact**: Validates real user experience
**Dependencies**: 0.4 complete
**Parallel Work**: Can start setup while 0.4 in progress

```
Tasks:
1. [M] Set up Playwright/similar
   - Install in nix environment
   - Create test configuration
   - Set up CI integration

2. [L] Write critical path E2E tests
   - Home → Create Room → Copy code
   - Join via form submission
   - Join via direct URL
   - Start game flow
   - Role reveal action

3. [M] Mobile browser tests
   - Test touch interactions
   - Test viewport sizes
   - Test orientation changes
```

### PHASE 1: Critical Fixes (BLOCKED BY PHASE 0)

#### 1.1 SSE NoTargetsFound Fix [M] - CRITICAL
**Dependencies**: Phase 0 tests must exist first
**Test First**: Write failing tests before fixing

```
Tasks:
1. [S] Write tests for SSE fragment targeting
   - Test current failing behavior
   - Test expected selector format
   - Test DOM element presence

2. [S] Fix fragment selectors
   - Ensure all IDs match between templates and SSE
   - Use consistent selector patterns
   - Add debug logging

3. [S] Verify fix with integration tests
   - Test real-time updates work
   - No console errors
   - All players see updates
```

#### 1.2 Join Flow Template Fix [S] - HIGH
**Dependencies**: Template tests from 0.2
**Test First**: Write tests for expected behavior

```
Tasks:
1. [S] Write tests for join page rendering
   - Test form structure
   - Test error display
   - Test direct URL handling

2. [S] Replace hardcoded HTML
   - Create proper Templ component
   - Integrate with layout
   - Add proper styling

3. [S] Test all join scenarios
   - Form submission
   - Direct URL
   - Error cases
```

#### 1.3 Performance Baseline [M] - MEDIUM
**Dependencies**: Core tests complete
**Purpose**: Prevent regression

```
Tasks:
1. [S] Create benchmark suite
   - Player join operations
   - SSE broadcast latency
   - Memory usage patterns

2. [S] Document baseline metrics
   - Acceptable response times
   - Memory limits
   - Concurrent connection limits
```

### PHASE 2: Game Mechanics (BLOCKED BY PHASE 1)

#### 2.1 Role Reveal Implementation [L] - CORE FEATURE
**Dependencies**: SSE fixes complete
**Test Coverage**: Required for all new code

```
Tasks:
1. [S] TDD: Write tests for role reveal
   - User can reveal own role
   - Others see revealed roles
   - State persistence

2. [M] Implement backend logic
   - Add reveal action handler
   - Update game state
   - Broadcast events

3. [S] Implement UI components
   - Reveal button
   - Confirmation dialog
   - Visual indicators
```

#### 2.2 Counter System [L] - CORE FEATURE
**Dependencies**: Role reveal complete
**Architecture**: Must follow event-driven pattern

```
Tasks:
1. [M] TDD: Design counter system tests
   - Counter types and limits
   - Modification rules
   - Validation logic

2. [M] Implement counter backend
   - Add to player state
   - Create modification handlers
   - Add validation

3. [S] Create counter UI
   - Display components
   - +/- controls
   - Real-time updates
```

#### 2.3 Win Condition System [L] - CORE FEATURE
**Dependencies**: Counters complete
**Complexity**: Requires PRD clarification

```
Tasks:
1. [M] TDD: Win condition tests
   - All win scenarios
   - Edge cases
   - Tie conditions

2. [M] Implement detection logic
   - Check after each action
   - Handle multiple winners
   - Game end flow

3. [S] Create end game UI
   - Winner announcement
   - Final scores
   - Play again option
```

### PHASE 3: UI Polish (FUTURE)

#### 3.1 MTG Theme Implementation [L]
#### 3.2 Mobile Optimization [XL]
#### 3.3 Animation System [L]
#### 3.4 Sound Effects [M]

### PHASE 4: Advanced Features (FUTURE)

#### 4.1 Spectator Mode [XL]
#### 4.2 Tournament System [XL]
#### 4.3 Statistics Tracking [L]
#### 4.4 Life Counter Tool [M]

## Critical Path Analysis

```
Phase 0 Testing (MUST COMPLETE FIRST):
0.1 Server Tests ────┐
                     ├─→ 0.4 Integration Tests ─→ 0.5 E2E Tests
0.2 Template Tests ──┤
0.3 Handler Tests ───┘

Then Phase 1 Fixes:
1.1 SSE Fix ─────────┐
1.2 Template Fix ────┼─→ Ready for Phase 2
1.3 Benchmarks ──────┘

Then Phase 2 Features:
2.1 Role Reveal ─→ 2.2 Counters ─→ 2.3 Win Conditions ─→ Phase 3
```

## Parallel Work Opportunities

### During Phase 0:
- **Developer A**: Work on 0.1 (Server Tests)
- **Developer B**: Work on 0.2 (Template Tests)
- **Developer C**: Work on 0.3 (Handler Tests) + Start 0.5 setup

### During Phase 1:
- All fixes require serial work due to test-first approach
- Documentation can be updated in parallel

### During Phase 2:
- UI components can be prototyped while backend is built
- Performance testing can run in parallel

## Success Metrics

### Phase 0 Complete When:
- [ ] Total coverage >80%
- [ ] All packages have tests
- [ ] Integration tests pass
- [ ] E2E tests configured
- [ ] No flaky tests

### Phase 1 Complete When:
- [ ] All SSE updates work
- [ ] Join flow uses templates
- [ ] Zero console errors
- [ ] Performance baselines set

### Phase 2 Complete When:
- [ ] Role reveal works
- [ ] Counter system functional
- [ ] Win conditions detected
- [ ] Full game playable

## Next Concrete Steps

1. **IMMEDIATELY**: Start with task 0.1.1 - Write tests for main.go
2. **IN PARALLEL**: Another developer starts 0.2.1 - Template test infrastructure
3. **MEASURE**: Run coverage after each task to track progress
4. **REVIEW**: After each module, ensure tests are comprehensive
5. **DOCUMENT**: Update this task list as items complete

Remember: **NO CODE WITHOUT TESTS** - This is not negotiable!