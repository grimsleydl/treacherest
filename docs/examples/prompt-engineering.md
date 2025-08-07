# Prompt Engineering Examples

This document contains detailed examples of high-quality prompts and the interactive enhancement process.

## Example 1: Complete End-to-End Interaction

### Initial Messy Prompt
```
User: "Add logging to the system"
```

### AI Enhancement Process
```
AI: I need to understand your logging requirements better. Let me ask some clarifying questions:

**Logging Scope:**
- What should be logged? (agent actions, system events, errors, performance metrics)
- Which components need logging? (just agents, CLI, coordination system, all)

**Logging Infrastructure:**
- Log format preference? (JSON, plain text, structured format)
- Log destinations? (console, files, remote service)
- Log levels needed? (debug, info, warn, error, fatal)

**Go-Specific Considerations:**
- Should we use Go's standard log package or a third-party solution?
- Any specific requirements for log messages?

**Operational Requirements:**
- Log rotation needed?
- Performance impact constraints?
- Integration with existing monitoring tools?

Could you provide guidance on these aspects so I can design an appropriate logging system?
```

### User Response
```
User: "JSON format, file-based, use standard library, focus on game events and errors"
```

### Final Enhanced Prompt
```
# Implement JSON-Based Game Logging System

## Goals
1. Add structured JSON logging for game events and errors
2. Use Go's standard library for logging
3. Implement file-based log output with rotation
4. Focus on game-specific events (player actions, game state changes, errors)

## Steps
1. Create custom JSON formatter using standard log package
2. Implement log file management with rotation (10MB files, keep last 5)
3. Add logging to game event handlers
4. Add logging to error paths in game logic
5. Create log level configuration
6. Test logging output and verify JSON structure
7. Measure performance impact

## SUCCESS_CONDITION
✅ Game events and errors are logged in valid JSON format to rotating files AND logs contain sufficient context for debugging AND performance impact is less than 5% overhead

## FAILURE_CONDITION
❌ Performance impact exceeds 10% overhead OR JSON output is malformed

## Context
- Application is a real-time multiplayer game using SSE
- Current logging is minimal (stdout only)
- Need to debug player disconnection issues
- Must not impact game responsiveness
```

## Example 2: Clarifying Questions with Placeholders

### Initial Request
```
User: "Implement player authentication"
```

### AI Response with Placeholders
```
I need to clarify authentication requirements before implementing:

**Authentication Method:**
- Username/password, OAuth, or magic links? [PLACEHOLDER: auth_method]
- Need email verification? [PLACEHOLDER: email_verification]

**Session Management:**
- Session duration? [PLACEHOLDER: session_timeout]
- Remember me functionality? [PLACEHOLDER: remember_me]
- Multiple device support? [PLACEHOLDER: multi_device]

**Security Requirements:**
- Password complexity rules? [PLACEHOLDER: password_rules]
- Rate limiting for login attempts? [PLACEHOLDER: rate_limiting]
- Two-factor authentication? [PLACEHOLDER: 2fa_required]

**Integration Points:**
- Existing user database? [PLACEHOLDER: user_storage]
- Need API endpoints or just web forms? [PLACEHOLDER: auth_endpoints]

Please provide details so I can create a complete implementation plan.
```

## Example 3: Multi-Agent Coordination Prompt

### Well-Structured Prompt
```
# Add Agent Coordination System

## Goals
1. Enable multiple agents to work simultaneously without conflicts
2. Implement resource locking mechanisms
3. Create shared state synchronization
4. Provide coordination visibility through logging

## Steps
1. Design AgentCoordinator architecture
   - Define agent identity structure
   - Design locking mechanism
   - Plan state sharing protocol
2. Implement resource locking
   - File-level locks with timeouts
   - Process-level coordination
   - Deadlock detection
3. Add state synchronization
   - Event-driven updates
   - Conflict resolution
   - State persistence
4. Integrate with CLI
   - Add agent management commands
   - Status reporting
   - Debug commands
5. Test multi-agent scenarios
   - Concurrent file access
   - State consistency
   - Failure recovery

## SUCCESS_CONDITION
✅ Multiple agents can work simultaneously without file conflicts AND shared state remains consistent across all agents AND all coordination tests pass AND deadlocks are detected and resolved

## FAILURE_CONDITION
❌ Agents experience unrecoverable deadlocks OR state corruption occurs during concurrent operations

## Context
- Building a multi-agent development system
- Agents may run in separate processes
- Need to handle agent crashes gracefully
- Performance is less critical than correctness
```

## Example 4: Feature Implementation with Unknown Scope

### Initial Vague Request
```
User: "Add search functionality"
```

### AI Clarification Process
```
I need more context about the search functionality:

**Search Scope:**
- What are we searching? (code, documentation, game data, players?)
- Full-text search or structured queries?
- Real-time search or batch processing?

**Technical Requirements:**
- Expected data volume? [PLACEHOLDER: data_size]
- Performance requirements? (response time) [PLACEHOLDER: perf_requirements]
- Need search indexing? [PLACEHOLDER: indexing_needs]

**User Interface:**
- Where will search appear? (nav bar, dedicated page, API only?)
- Auto-complete needed? [PLACEHOLDER: autocomplete]
- Search filters/facets? [PLACEHOLDER: search_filters]

**Results Handling:**
- How to display results? (list, cards, table)
- Pagination or infinite scroll? [PLACEHOLDER: pagination_style]
- Result ranking/relevance? [PLACEHOLDER: ranking_algorithm]

Could you clarify these points for accurate implementation?
```

## Best Practices for Prompt Enhancement

### 1. Always Include Measurable Success Criteria
Bad: "Make the system faster"
Good: "✅ Response time under 100ms for 95% of requests"

### 2. Break Down Complex Tasks
Bad: "Implement the entire authentication system"
Good: Separate into subtasks - user model, password hashing, session management, etc.

### 3. Provide Context and Constraints
Bad: "Add caching"
Good: "Add Redis caching for game state with 5-minute TTL, must handle cache misses gracefully"

### 4. Use Placeholders for Unknown Information
Bad: Make assumptions about requirements
Good: Mark unknowns with `[PLACEHOLDER: description]` and ask for clarification

### 5. Include Failure Conditions When Appropriate
- For exploratory tasks: May omit FAILURE_CONDITION
- For performance-critical tasks: Always include failure thresholds
- For integration tasks: Include compatibility failure conditions

## Anti-Patterns to Avoid

### 1. Ambiguous Success Conditions
❌ "System should work well"
✅ "All unit tests pass AND integration tests complete in under 30 seconds"

### 2. Missing Context
❌ "Fix the bug"
✅ "Fix player disconnect bug occurring when switching game rooms (see error in logs line 245)"

### 3. Assuming Technical Decisions
❌ "Implement using PostgreSQL" (without confirming)
✅ "Which database should we use? Current options: PostgreSQL, SQLite, or in-memory"

### 4. Skipping the Enhancement Process
❌ Immediately executing vague requests
✅ Always clarify, enhance, and get approval first