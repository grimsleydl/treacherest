# Phase 1: Critical Fixes Implementation Plan

## Overview
This plan addresses the critical fixes needed before implementing new features. The focus is on:
1. Fixing hardcoded HTML in handlers
2. Ensuring SSE real-time updates work correctly
3. Adding comprehensive test coverage using TDD approach
4. Establishing performance baselines

## Current State Analysis
- **Coverage**: 80.9% overall (Phase 0 complete)
- **Critical Issue**: Hardcoded HTML in `pages.go` line 90-103 (join form)
- **SSE Status**: Implementation exists but needs comprehensive testing
- **Test Infrastructure**: Solid foundation from Phase 0

## Implementation Tasks

### Task 1: Fix Hardcoded HTML Join Form (Priority: CRITICAL)
**Location**: `/workspace/nix/app/internal/handlers/pages.go` lines 90-103
**Assignee**: Frontend Developer Role

#### Steps:
1. **Create Join Form Template** (30 minutes)
   ```
   File: /workspace/nix/app/internal/views/pages/join.templ
   - Create templ component for join form
   - Match existing UI styling patterns
   - Include proper form validation attributes
   ```

2. **Write Tests First** (45 minutes)
   ```
   File: /workspace/nix/app/internal/views/pages/join_test.go
   - Test form renders with room code
   - Test input field requirements
   - Test form submission path
   - Test error message display capability
   ```

3. **Replace Hardcoded HTML** (15 minutes)
   ```
   File: /workspace/nix/app/internal/handlers/pages.go
   - Import new join template
   - Replace w.Write() with component.Render()
   - Ensure proper error handling
   ```

4. **Integration Test** (30 minutes)
   ```
   File: /workspace/nix/app/internal/handlers/pages_test.go
   - Add test case for join form display
   - Verify form submission behavior
   - Test edge cases (empty name, special chars)
   ```

**Success Criteria**:
- ✅ No hardcoded HTML in handlers
- ✅ Join form uses templ component
- ✅ All existing tests still pass
- ✅ New tests achieve 100% coverage for join flow

---

### Task 2: SSE Connection Testing & Fixes (Priority: CRITICAL)
**Location**: `/workspace/nix/app/internal/handlers/sse.go`
**Assignee**: Backend Developer Role

#### Steps:
1. **Write SSE Message Format Tests** (45 minutes)
   ```
   File: /workspace/nix/app/internal/handlers/sse_message_test.go
   - Test fragment format with selectors
   - Test merge mode settings
   - Test script execution format
   - Verify datastar compatibility
   ```

2. **Test Concurrent Connections** (1 hour)
   ```
   File: /workspace/nix/app/internal/handlers/sse_concurrent_test.go
   - Multiple clients connecting simultaneously
   - Broadcast to all clients
   - Handle client disconnection
   - Verify no goroutine leaks
   ```

3. **Test Reconnection Logic** (45 minutes)
   ```
   File: /workspace/nix/app/internal/handlers/sse_reconnect_test.go
   - Client reconnection after network drop
   - State consistency after reconnect
   - Event replay if needed
   - Session persistence
   ```

4. **Fix Any Issues Found** (30 minutes)
   - Add missing error handling
   - Ensure proper cleanup on disconnect
   - Add connection timeouts if needed
   - Verify selector matching

**Success Criteria**:
- ✅ SSE messages follow datastar format
- ✅ No "NoTargetsFound" errors
- ✅ Handles 50+ concurrent connections
- ✅ Graceful reconnection handling
- ✅ No goroutine or memory leaks

---

### Task 3: Join Flow End-to-End Testing (Priority: HIGH)
**Location**: Multiple files
**Assignee**: Frontend Developer Role (can run parallel with Task 2)

#### Steps:
1. **Browser Test Setup** (30 minutes)
   ```
   File: /workspace/nix/app/e2e/join_flow_test.go
   - Setup chromium driver
   - Helper functions for form interaction
   - Screenshot on failure
   ```

2. **Test Direct URL Join** (45 minutes)
   ```
   - Navigate to /room/{code}
   - Fill join form
   - Submit and verify redirect
   - Check player appears in lobby
   ```

3. **Test Join Errors** (45 minutes)
   ```
   - Invalid room code
   - Duplicate player name
   - Room full scenario
   - Game already started
   ```

4. **Test Cookie Persistence** (30 minutes)
   ```
   - Join room
   - Close browser
   - Revisit URL
   - Verify still in room
   ```

**Success Criteria**:
- ✅ All join paths work correctly
- ✅ Error messages display properly
- ✅ Cookie persistence works
- ✅ No UI glitches or race conditions

---

### Task 4: Performance Baseline Establishment (Priority: MEDIUM)
**Location**: New benchmark files
**Assignee**: Backend Developer Role (after Task 2)

#### Steps:
1. **Create Benchmark Suite** (45 minutes)
   ```
   File: /workspace/nix/app/internal/benchmarks/baseline_test.go
   - Player join operation
   - SSE broadcast latency
   - Room creation time
   - Concurrent player limits
   ```

2. **Memory Usage Profiling** (30 minutes)
   ```
   - Profile memory per player
   - Profile memory per room
   - Identify any leaks
   - Document baseline usage
   ```

3. **Load Testing** (45 minutes)
   ```
   - Max players per room
   - Max concurrent rooms
   - SSE connection limits
   - CPU usage patterns
   ```

4. **Document Results** (15 minutes)
   ```
   File: /workspace/docs/architecture/performance-baselines.md
   - Current metrics
   - Bottlenecks identified
   - Recommended limits
   ```

**Success Criteria**:
- ✅ Join operation <10ms
- ✅ SSE broadcast <50ms for 10 players
- ✅ Support 100+ concurrent players
- ✅ No memory leaks detected
- ✅ Baseline documented for regression testing

---

## Execution Timeline

### Parallel Execution Plan:
- **Stream 1** (Frontend Dev): Task 1 → Task 3
- **Stream 2** (Backend Dev): Task 2 → Task 4

### Dependencies:
- Task 3 depends on Task 1 completion
- Task 4 should start after Task 2
- All tasks independent enough for parallel work

### Total Estimated Time:
- Stream 1: ~5 hours (Task 1: 2h, Task 3: 3h)
- Stream 2: ~5 hours (Task 2: 3h, Task 4: 2h)
- **Total elapsed time with parallel execution: ~5 hours**

---

## Risk Mitigation

### Potential Issues:
1. **Templ generation errors**: Run `build-templ` after any .templ changes
2. **SSE selector mismatch**: Verify HTML IDs match datastar selectors
3. **Test flakiness**: Add proper waits in browser tests
4. **Performance regression**: Run benchmarks before/after changes

### Rollback Plan:
- All changes in separate commits
- Can revert individual fixes if needed
- Existing tests ensure no regression

---

## Definition of Done

### Phase 1 Complete When:
- [ ] No hardcoded HTML in any handler
- [ ] Join form uses proper templ component
- [ ] SSE tests cover all scenarios
- [ ] Browser tests pass consistently
- [ ] Performance baselines documented
- [ ] All tests run in <30 seconds
- [ ] Coverage maintained >80%
- [ ] No new linting errors

### Next Steps:
After Phase 1 completion, proceed to Phase 2 (Game Mechanics) with confidence that:
- Infrastructure is solid
- Real-time updates work reliably
- Test patterns are established
- Performance baselines exist for comparison

---

## Commands Reference

```bash
# Run after .templ changes
build-templ

# Run all tests
test-all

# Run coverage
test-coverage

# Run benchmarks
go test -bench=. ./internal/benchmarks/...

# Format code
fmt

# Start dev server for manual testing
dev
```

---

*This plan enables parallel execution while maintaining code quality through TDD approach.*