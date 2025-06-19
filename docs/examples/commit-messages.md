# Commit Message Examples

This document provides examples of properly formatted commit messages with XML metadata for the Treacherest project.

## Basic Structure

```
[emoji] [type]: [description]

<what>[Full commit message with emoji]</what>
<why>[Business/technical rationale]</why>
<how>[Implementation details]</how>
<prompt>[The approved prompt]</prompt>
<post-prompt>[Lessons learned]</post-prompt>
```

## Commit Type Reference

| Type     | Emoji | Description             | Example                                |
|----------|-------|-------------------------|----------------------------------------|
| feat     | ✨    | New feature             | `✨ feat: add player profile page`     |
| fix      | 🐛    | Bug fix                 | `🐛 fix: resolve SSE connection drops` |
| docs     | 📚    | Documentation           | `📚 docs: update API endpoints`        |
| style    | 💎    | Code style (formatting) | `💎 style: format with go fmt`         |
| refactor | ♻️     | Code refactoring        | `♻️ refactor: extract game logic`       |
| test     | 🧪    | Add/update tests        | `🧪 test: add multiplayer scenarios`   |
| chore    | 🔧    | Maintenance tasks       | `🔧 chore: update dependencies`        |
| perf     | ⚡    | Performance improvement | `⚡ perf: optimize game state updates` |

## Example 1: Feature Addition

```
🎉 feat: add real-time player count display

<what>🎉 feat: add real-time player count display</what>
<why>Players want to see how many others are online before joining games. This improves engagement by showing an active community.</why>
<how>
- Added PlayerCountHandler in handlers/game_handler.go
- Created SSE endpoint /api/player-count
- Added data-signal="playerCount" to lobby template
- Implemented 5-second update interval
- Used atomic operations for thread-safe counting
Commands: templ generate, go test ./internal/handlers
</how>
<prompt>
# Add Online Player Count

## Goals
1. Show real-time count of online players in lobby
2. Update automatically without page refresh

## Steps
1. Create player tracking in session manager
2. Add SSE endpoint for player count
3. Update lobby template with counter
4. Test with multiple connections

## SUCCESS_CONDITION
✅ Player count updates within 5 seconds of connect/disconnect
</prompt>
<post-prompt>Discovered that gorilla/websocket sessions don't trigger disconnect immediately - added heartbeat mechanism for accurate counts.</post-prompt>
```

## Example 2: Bug Fix

```
🐛 fix: prevent duplicate player names in game rooms

<what>🐛 fix: prevent duplicate player names in game rooms</what>
<why>Players could join with same name causing confusion in game UI and breaking game logic that assumed unique names.</why>
<how>
- Added name validation in JoinGame method
- Implemented case-insensitive comparison
- Return error "Name already taken" to client
- Added integration test TestDuplicatePlayerNames
- Updated error handling in game_room.templ
</how>
<prompt>
# Fix Duplicate Player Names

## Goals
1. Prevent players from using same name in a game room
2. Show clear error message when name is taken

## Steps
1. Add validation in game service
2. Handle error in frontend
3. Add test coverage

## SUCCESS_CONDITION
✅ Duplicate names rejected with user-friendly error
</prompt>
<post-prompt>Initially only checked exact matches, but players used case variations. Fixed with strings.EqualFold().</post-prompt>
```

## Example 3: Performance Improvement

```
⚡ perf: optimize game state serialization

<what>⚡ perf: optimize game state serialization</what>
<why>Game state updates were taking 150ms+ for large games, causing noticeable lag. Need sub-50ms for smooth gameplay.</why>
<how>
- Replaced json.Marshal with custom serializer
- Pre-allocated buffers based on game size
- Used sync.Pool for buffer reuse
- Cached unchanged portions of state
- Benchmark results: 150ms → 35ms (77% improvement)
Commands: go test -bench=. ./internal/models
</how>
<prompt>
# Optimize Game State Performance

## Goals
1. Reduce game state serialization time
2. Maintain backward compatibility

## Steps
1. Profile current performance
2. Implement optimizations
3. Benchmark improvements
4. Ensure no regression in functionality

## SUCCESS_CONDITION
✅ Serialization under 50ms for 10-player games
</prompt>
<post-prompt>sync.Pool provided biggest win. Consider protobuf for next iteration if JSON becomes bottleneck again.</post-prompt>
```

## Example 4: Refactoring

```
♻️ refactor: extract SSE handling into middleware

<what>♻️ refactor: extract SSE handling into middleware</what>
<why>SSE setup code was duplicated across 5 handlers. Centralization reduces bugs and makes adding new SSE endpoints easier.</why>
<how>
- Created middleware/sse.go with SSEMiddleware function
- Extracted common headers and error handling
- Updated all handlers to use middleware
- Added context for clean shutdown
- No functional changes, all tests still pass
</how>
<prompt>
# Refactor SSE Code

## Goals
1. Eliminate SSE setup duplication
2. Standardize error handling
3. Maintain current functionality

## Steps
1. Identify common SSE patterns
2. Create reusable middleware
3. Update existing handlers
4. Verify no regression

## SUCCESS_CONDITION
✅ All SSE endpoints use middleware AND existing tests pass
</prompt>
<post-prompt>Middleware pattern worked well. Consider similar approach for WebSocket handlers.</post-prompt>
```

## Example 5: Documentation Update

```
📚 docs: add multiplayer architecture guide

<what>📚 docs: add multiplayer architecture guide</what>
<why>New developers struggle to understand how game state synchronization works. Clear documentation reduces onboarding time.</why>
<how>
- Created docs/architecture/backend/multiplayer-sync.md
- Included sequence diagrams for game flow
- Added code examples from actual implementation
- Documented edge cases and error handling
- Cross-referenced with frontend SSE handling
</how>
<prompt>
# Document Multiplayer Architecture

## Goals
1. Explain game state synchronization
2. Help new developers understand the system

## Steps
1. Document high-level architecture
2. Create sequence diagrams
3. Add code examples
4. Explain error handling

## SUCCESS_CONDITION
✅ Complete guide covering all multiplayer aspects
</prompt>
<post-prompt>Mermaid diagrams render well in markdown. Should add more diagrams to other docs.</post-prompt>
```

## Example 6: Test Addition

```
🧪 test: add browser tests for game disconnection

<what>🧪 test: add browser tests for game disconnection</what>
<why>Users reported lost game state on reconnection. Need automated tests to prevent regression.</why>
<how>
- Created TestPlayerDisconnectRecovery in browser tests
- Simulates network interruption with chromedp
- Verifies state restoration after reconnect
- Tests both voluntary and involuntary disconnects
- Added helper functions for connection manipulation
Runtime: 45 seconds per test
</how>
<prompt>
# Add Disconnection Tests

## Goals
1. Test game state recovery after disconnect
2. Cover voluntary and network failure scenarios

## Steps
1. Setup browser test infrastructure
2. Implement disconnect simulation
3. Verify state restoration
4. Test edge cases

## SUCCESS_CONDITION
✅ Tests reliably reproduce and verify disconnect scenarios
</prompt>
<post-prompt>Chrome DevTools Protocol made network simulation easy. Consider adding more failure scenario tests.</post-prompt>
```

## Best Practices

### DO:
- Keep `<what>` concise and descriptive
- Explain the "why" from user/business perspective
- Include specific technical details in `<how>`
- List actual commands run
- Share valuable learnings in `<post-prompt>`
- **ONLY commit files related to the specific change** (use `jj commit file1 file2 -m "message"`)
- Review `jj status` before committing to ensure you're being selective

### DON'T:
- Don't include file line numbers (they change)
- Don't be vague about changes ("various files")
- Don't skip the business rationale
- Don't forget to include the original prompt
- Don't make the commit message too long
- **NEVER commit all unstaged changes blindly**
- Don't mix unrelated changes in one commit

## Branch Naming from Commits

Remember to create branch names by converting the commit message:

| Commit Message                    | Branch Name                   |
|-----------------------------------|-------------------------------|
| `🎉 feat: add game lobby`         | `feat_add_game_lobby`         |
| `🐛 fix: player disconnect`       | `fix_player_disconnect`       |
| `♻️ refactor: extract auth logic`  | `refactor_extract_auth_logic` |
| `⚡ perf: optimize render loop`   | `perf_optimize_render_loop`   |

## Special Cases

### Multi-Part Changes
When a change touches many areas, group logically in `<how>`:

```
<how>
Backend Changes:
- Updated game service logic
- Added new API endpoints
- Modified database schema

Frontend Changes:
- New templates for game UI
- Updated JavaScript handlers
- Added CSS animations

Tests:
- Unit tests for service layer
- Integration tests for API
- Browser tests for UI
</how>
```

### Breaking Changes
Always highlight breaking changes prominently:

```
<what>💥 breaking: change game ID format to UUID</what>
<why>Integer IDs were predictable, allowing game enumeration attacks. UUIDs provide better security.</why>
```

### Hotfixes
For urgent production fixes:

```
<what>🚑 hotfix: restore game saving functionality</what>
<why>Critical bug preventing any games from being saved. Users losing progress.</why>
```
