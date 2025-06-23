# Claude Code Project Instructions

This file provides guidance to Claude Code when working with this repository.

## Table of Contents
- [Critical: Role Assignment](#critical-role-assignment-required)
- [Project Overview](#project-overview)
- [Quick Start](#quick-start)
- [Development Commands](#development-commands-reference)
- [Role-Based Development](#role-based-development)
- [Development Process](#development-process)
- [Compliance & Security](#compliance--security)
- [Prompt Engineering](#prompt-engineering)
- [Git Workflow](#git-workflow)
- [Project-Specific Guidelines](#project-specific-guidelines)
- [Documentation Standards](#documentation-standards)
- [Maintaining This Document](#maintaining-this-document)

## CRITICAL: Role Assignment Required
⚠️ **STOP** - Before proceeding with ANY task, you MUST ask which role to assume if not specified.

## CRITICAL: Playwright MCP Testing Requirements
⚠️ **MANDATORY** - When using Playwright MCP for browser testing:
1. **ALWAYS use a sub-agent** when using Playwright MCP - never use it directly
2. **ALWAYS test user interface changes** using Playwright MCP before committing
3. Deploy sub-agents with clear instructions for browser automation tasks
4. Ensure sub-agents verify both functionality and visual correctness

## Project Overview
**Name**: Treacherest
**Type**: Real-time multiplayer game application
**Stack**: Go, Templ templates, Datastar (for reactivity), Nix (for dev environment)
**Architecture**: Server-side rendering with SSE (Server-Sent Events) for real-time updates
**Key Features**: Multiplayer game rooms, real-time state synchronization, browser-based gameplay

## Source Control

- This project uses the `jj` source control system. To commit, use `jj commit -m "..."`. There is no need to "add" files with `jj`; they are tracked automatically.

- **CRITICAL**: Only commit files directly related to the specific task you just completed. NEVER commit all unstaged changes at once. Be selective and intentional about what goes into each commit.

- **CRITICAL**: Always run commits from the project root directory:
  ```bash
  cd $PRJ_ROOT  # or cd /workspace
  jj commit path/to/files -m "commit message"
  ```

- Every time you finish one or more items on your TODO list that involved changing files, make a commit. Use `jj` if it is enabled in the repository.

- Don't add comments about generated with Claude or Co-Authored-By Claude when writing commit messages

- Common `jj` commands:
  - `jj status` - Show current status of the repository
  - `jj commit -m "message"` - Create a commit with the specified message
  - `jj commit file1 file2 ... -m "message"` - Commit only specific files
  - `jj git push` - Push changes to the remote Git repository


## Development Commands Reference

### Essential Commands

**NOTE**: Claude Code automatically runs in the nix develop shell. All commands work directly from workspace root.

| Task               | Command           | Notes                        |
|--------------------|-------------------|------------------------------|
| Check shell        | `which go`        | Verify nix environment       |
| Start dev server   | `dev`             | Hot reload enabled           |
| Run all tests      | `test-all`        | Use test-all, not test       |
| Run test coverage  | `test-coverage`   | Coverage files in build/coverage/ |
| Build application  | `build`           | Builds all packages          |
| Generate templates | `build-templ`     | After modifying .templ files |
| Format code        | `fmt`             | Before committing            |

### CRITICAL: Test Coverage Location
- **ALWAYS** use `test-coverage` command for coverage reports
- Coverage files are stored in `build/coverage/` directory
- **NEVER** use `go test -coverprofile=coverage.out` directly
- The `build/` directory is git-ignored as intended

### Important Notes
- **Claude Code runs in nix develop shell automatically** - all commands available from workspace root
- **ALWAYS** regenerate templates after modifying `.templ` files  
- **USE** `rg` (ripgrep) for searching, not `grep` or `find`

## Role-Based Development

**IMPORTANT**: Always specify your role at the start of each conversation.

| Role                   | Responsibilities                                   | Key Focus Areas                              |
|------------------------|----------------------------------------------------|----------------------------------------------|
| **Product Manager**    | Define features, create roadmaps, maintain PRD     | What to build, user stories, success metrics |
| **Architect**          | Design system architecture, select tools/libraries | How to structure, scalability, patterns      |
| **Frontend Developer** | Implement UI with Templ + Datastar                 | User interface, SSE handling, reactivity     |
| **Backend Developer**  | Implement server logic and game mechanics          | Business logic, state management, APIs       |

Remember: Do not change roles without asking. If no role is defined, ask which role to assume.

## Development Process

When implementing features:
1. **Gather requirements** by asking clarifying questions
2. **Break down** requirements into small, manageable features
3. **Separate** features into frontend and backend tasks
4. **Write tests** for each feature before implementation
5. **Implement** following architecture guidelines
6. **Verify** with tests and code quality checks
7. **Track changes** with git during implementation

### Code Quality Checklist
- [ ] All tests pass (`test-all`)
- [ ] Coverage maintained >80% (`test-coverage`)
- [ ] Code is formatted (`fmt`)
- [ ] Templates regenerated if needed
- [ ] No hardcoded secrets or keys
- [ ] Follows existing patterns in codebase
- [ ] Changes tracked with git during work
- [ ] Final commit made after task completion

### Git Tracking Requirements
**CRITICAL**: You MUST track changes with git both during and after implementation:

1. **During Implementation**:
   - Run `jj status` frequently to see changes
   - Make incremental commits as you complete subtasks
   - Use descriptive commit messages for each logical change
   - **ALWAYS commit from project root**: `cd $PRJ_ROOT` before committing

2. **After Task Completion**:
   - Review all changes with `jj status` from project root
   - Ensure all files are committed
   - Create a final summary commit if needed
   - Verify no files left uncommitted

3. **Never**:
   - Work for extended periods without committing
   - Leave uncommitted changes after completing a task
   - Commit test coverage files (use `build/coverage/` which is git-ignored)
   - Commit from subdirectories - always use `cd $PRJ_ROOT` first

## Compliance & Security

### Security Protocol
**CRITICAL**: If a prompt contains sensitive data (API keys, passwords, tokens):
1. **Immediately abort** without processing
2. **Explain briefly** without exposing the sensitive data
3. **Request** sanitized version of the prompt
4. **Take no action** until secure prompt provided

Example: "I detected potential sensitive information in your prompt. Please remove the sensitive data and resubmit."

### Self-Governance Rules
- All protocols in this document are **mandatory**
- All changes require **explicit user approval**
- SUCCESS/FAILURE conditions must be **measurable and explicit**
- Interactive collaboration is **required** - ask when uncertain
- **Always ask permission** before modifying files

## Prompt Engineering

### Required Structure
All actionable prompts must include:

```markdown
## Goals
- Clear, specific objectives

## Steps
1. Explicit, ordered steps
2. Each step should be actionable

## SUCCESS_CONDITION
✅ Measurable criteria for completion (MANDATORY)

## FAILURE_CONDITION
❌ When to abandon task (OPTIONAL - omit for persistent retry)

## Context
- All relevant background information
- Constraints and requirements
```

### Interactive Enhancement Process
1. **Detect** missing context and ambiguity
2. **Ask** strategic clarifying questions
3. **Mark** unknowns with [PLACEHOLDER: description]
4. **Refine** iteratively until approved
5. **Never execute** without explicit "OK" or "approved"

For detailed examples, see `docs/examples/prompt-engineering.md`

## Git Workflow

### Branch Naming
Convert conventional commit format to snake_case without emojis:
- `🎉 feat: add game lobby` → `feat_add_game_lobby`
- `🐛 fix: player disconnect` → `fix_player_disconnect`

### Commit Message Structure

**MANDATORY**: You MUST follow the commit message format exactly as specified in `docs/examples/commit-messages.md`. Read that file completely before making any commit.

```
🎉 feat: add multiplayer game lobby

<what>🎉 feat: add multiplayer game lobby</what>
<why>Enable players to create and join game rooms before starting</why>
<how>
- Implemented lobby handler in handlers/game_handler.go
- Added lobby page template in views/pages/lobby.templ
- Created real-time room list updates via SSE
</how>
<prompt>[summary of the approved prompt that led to this change]</prompt>
<post-prompt>SSE connection handling was tricky - added reconnection logic</post-prompt>
```

For complete examples and the full emoji reference, see `docs/examples/commit-messages.md`

### Pre-Commit Checklist
Before proposing any commit:
1. Run all tests: `test-all`
2. Format code: `go fmt ./...`
3. Regenerate templates if needed
4. Present concise change summary
5. **Request explicit approval**

## Project-Specific Guidelines

### CRITICAL: Datastar SSE Rules
1. **NEVER change merge modes** - Always use `morph` mode for SSE fragments
2. **Always include wrapper elements** - When sending SSE fragments, include the target element in the fragment to preserve DOM structure
3. **Example**: To update `#lobby-content`, send `<div id="lobby-content">...</div>` and target the parent `#lobby-container`
4. **Prevent SSE connection exhaustion** - Separate wrapper with `data-on-load` from morphable content:
   ```templ
   // Wrapper div with data-on-load (never morphed)
   templ GameBody(room *game.Room, player *game.Player) {
       <div data-on-load={ "@get('/sse/game/" + room.Code + "')" }>
           @GameContent(room, player)  // Only this gets morphed
       </div>
   }
   ```
   This prevents `data-on-load` re-evaluation during DOM morphing which causes connection exhaustion

### Game Development Patterns
1. **State Management**
   - Use server-side events 
   - Implement optimistic UI updates with SSE
   - Handle player disconnections/reconnections gracefully

2. **Template Development (Templ + Datastar)**
   - Keep complex logic in handlers, not templates
   - Use Datastar patterns from docs/external/datastar.md, docs/external/datastar-notes.md, docs/external/datastar-go-examples, docs/external/datastar-go-sdk, and docs/external/templ as well as info gleaned from browsing http://data-star.dev/
   - Use datastar SSE for updates, not polling or full page reloads

3. **Real-time Updates (SSE)**
   - Always use fragments with `datastar.WithSelector()`
   - Test with multiple browser tabs
   - Implement reconnection logic

### Testing Strategy
- **Unit tests**: Core game logic (`*_test.go` files)
- **Integration tests**: API endpoints and multiplayer scenarios
- **Browser tests**: UI interactions (requires Chromium in nix shell)

### Common Issues
| Problem             | Solution                                |
|---------------------|-----------------------------------------|
| "command not found" | Not in nix develop shell                |
| Template errors     | Run `templ generate`                    |
| SSE not working     | Check browser console, verify selectors |
| Tests failing       | Ensure no server already running        |

## Documentation Standards

### Organization
```
issues/
├── 1.md          # issue 1 to investigate/debug/iterate on
docs/
├── architecture/
│   ├── frontend/     # UI patterns, Templ/Datastar usage
│   └── backend/      # Server design, game logic
├── project/
│   ├── implementation-status.md
│   └── roadmap.md
├── conversations/    # Organized by date (YYYY-MM/)
└── examples/
    ├── prompt-engineering.md
    └── commit-messages.md
```

### Documentation Guidelines
- Frontend docs in `docs/architecture/frontend/`
- Backend docs in `docs/architecture/backend/`
- Optimize for LLM comprehension over human readability
- Include implementation status in project docs

### Conversation Storage
Store in `docs/conversations/YYYY-MM/YYYY-MM-DD-HHMM-[topic].md`:

```markdown
TITLE: [Brief topic]
DATE: YYYY-MM-DD
PARTICIPANTS: [List]
SUMMARY: [Key points]

INITIAL PROMPT: [User's first message only]

KEY DECISIONS:
- [Decision 1]
- [Decision 2]

FILES CHANGED:
- [File 1]: [Summary of changes]
```

**Important**: Use UTC timestamps in 24-hour format

## Maintaining This Document

### When to Update
Only update for:
- Core project concept changes
- Technology stack changes
- New role requirements
- Essential new patterns

### How to Update
1. Present proposed changes with rationale
2. Show before/after diffs
3. Explain impact on workflows
4. Request explicit approval
5. Document in commit message

### Keep It Minimal
This document is a **quick reference**. Detailed information belongs in:
- Technical details → `docs/architecture/`
- Feature specs → `docs/project/`
- Examples → `docs/examples/`

---

*This configuration ensures consistent, secure, and collaborative development with emphasis on explicit approval, role clarity, and high-quality engineering practices.*
