# Feature Implementation Plan: Custom Role Configurations

This plan outlines the necessary steps to implement a system allowing game hosts to customize the roles used in a game. This involves enabling/disabling roles from a master pool and setting the exact count for each role.

---

### **Phase 1: Discovery and Codebase Analysis**

**Objective:** To understand the existing code structure for game settings, roles, and the game creation process.

**Actions:**

1.  **Locate the Master Role Definition:**
    *   Search the codebase for where the game's roles are defined.
    *   **Goal:** Identify the file(s) that act as the "master pool" of all available roles

2.  **Identify Game Settings and Lobby State:**
    *   Find the data structures that manage the settings for a game lobby or session. 
    *   **Goal:** Understand how game settings are currently stored (in-memory, database) and what object/struct needs to be modified to include the new custom role settings.

3.  **Find the Game Setup UI:**
    *   Locate the UI component(s) where the game host configures a new game.
    *   **Goal:** Pinpoint the exact file to modify for the new UI controls.

4.  **Analyze Role Assignment Logic:**
    *   Find the backend logic that starts a game and assigns roles to players.
    *   **Goal:** Understand the current role assignment algorithm (e.g., random assignment from a fixed list) that will be replaced.

---

### **Phase 2: Data Model and State Management**

**Objective:** To define and integrate the data structure for storing the custom role configuration.

**Actions:**

1.  **Modify the Game Settings Model:**
    *   Based on the findings in Phase 1, create or extend the way game settings are stored. Perhaps we can read from a JSON during the build process
    *   **Proposed Structure:** Add a key, `roleConfig`, structured as follows:
        ```json
        "roleConfig": {
          "enabledRoles": {
            "The Augur": true,
            "The Supplier": true,
            "The Madwoman": true,
            "The Twin Princesses": true,
            "The Chaos Bringer": false,
            ...
          },
          "roleCounts": {
            "Leader": 1,
            "Guardian": 1,
            "Assassin": 3,
            "Traitor": 1
          }
        }
        ```
    *   This structure cleanly separates which roles are available in the UI from the actual count being used in the game.

2.  **Integrate State Management:**
    *   Ensure this new role config is initialized with sensible defaults when a lobby is created.
    *   Wire up the state management so that changes from the UI (Phase 3) update this object and are communicated between the client and server.

---

### **Phase 3: UI/Frontend Implementation**

**Objective:** To build an intuitive UI for the game host to manage the role settings.

**Actions:**

1.  **Create the "Role Configuration" Component:**
    *   In the game setup view identified in Phase 1, add a new section for "Custom Role Settings".
    *   This section should only be visible/editable by the game host.

2.  **Implement Role Pool Management (Enable/Disable):**
    *   Fetch the master list of roles (from Phase 1).
    *   Render a list of all roles, each with a checkbox or toggle switch.

3.  **Implement Role Count Selection:**
    *   Next to each role in the list, display a number input, or a `-/+` button combination.
    *   This input controls the value for that role in the role config for that game.
    *   The input for a role should be disabled or hidden if that role is disabled in the pool.

4.  **Add UI Feedback and Validation:**
    *   Display a running total of selected roles (e.g., "7 / 8 roles selected").
    *   Display a clear error message to the user explaining why they cannot start the game

---

### **Phase 4: Backend and Game Logic Integration**

**Objective:** To make the game engine use the new custom configuration when assigning roles.

**Actions:**

1.  **Modify the game start function:**
    *   When the game start is called, it must now receive the role config from the game settings.
    *   **Crucially, it must perform server-side validation** to ensure the configuration is valid (e.g., total role pool count matches player count), rejecting the request if not. Never trust the client.

2.  **Confirm role pools work with role assignment logic:**
    *   The existing role assignment logic must be replaced with the following process:
        1.  Create a pool of roles based on the enabled roles, number of players, and number of each role required
        2.  Perform a shuffle on this list to ensure randomness.
        3.  Iterate through the list of players in the lobby and assign them a role from the shuffled list.
        4.  Store the assigned roles in the game state as usual.

---

### **Phase 5: Testing**

**Objective:** To ensure the feature is robust, bug-free, and works as expected.

**Actions:**

1.  **Unit Tests:**
    *   Write tests for the new role assignment logic to confirm it correctly generates and assigns roles from a sample `roleConfig`.
    *   Write tests for the server-side validation logic to ensure it correctly rejects invalid configurations.
2.  **Integration Tests:**
    *   Test the full flow: host changes settings -> settings are saved -> game starts -> roles are correctly assigned.
3.  **Playwright-MCP QA:**
    *   Test the UI for usability and edge cases (e.g., what happens if a player leaves after roles are configured?).
    *   Verify that disabling a role correctly removes it from the game.

---

### **Potential Issues and Risks**

1.  **Validation Complexity:**
    *   **Risk:** The plan assumes `total roles == total players`. Some games have complex balancing rules (e.g., "must have at least one evil role," "ratio of X to Y must be Z").
    *   **Mitigation:** This plan covers the core request. The agent should flag if it discovers complex balancing rules in the code, as the validation logic would need to be more sophisticated. The PRD should be clarified if such rules are needed.

2.  **UI/UX Clutter:**
    *   **Risk:** If there are many roles, the configuration screen could become cluttered and overwhelming.
    *   **Mitigation:** For the initial implementation, a simple list is acceptable. A future iteration could involve categories, search bars, or pre-set templates for role configurations.

3.  **State Synchronization:**
    *   **Risk:** Bugs can occur if the host's local settings get out of sync with the server's state, especially in a laggy network environment.
    *   **Mitigation:** Implement a clear, single source of truth (the server's lobby state). The client UI should reflect this state and send update events, but the server state is always authoritative. Everything should be performed via Datastar and SSE, with no extraneous Javascript or client-side stuff.

#### IMPLEMENTATION PLAN:  Custom Role Configurations
 Implementation Plan:

Overview

Implement a comprehensive role customization system that allows:
- Server-side YAML configuration for game settings (max players, available roles)
- Per-room role configurations stored in memory
- Support for any number of players with flexible role distributions
- Predefined and custom role setups

Phase 1: Server Configuration System

1.1 Create YAML Configuration Structure

- Create config/server.yaml with server-wide settings:
server:
  maxPlayersPerRoom: 20
  minPlayersPerRoom: 1
  roomCodeLength: 5
  roomTimeout: 24h

roles:
  # Master pool of all available roles
  available:
    leader:
      displayName: "Leader"
      category: "Leader"
      minCount: 1
      maxCount: 1
      alwaysRevealed: true
    guardian:
      displayName: "Guardian"
      category: "Guardian"
      minCount: 0
      maxCount: 10
    assassin:
      displayName: "Assassin"
      category: "Assassin"
      minCount: 0
      maxCount: 10
    traitor:
      displayName: "Traitor"
      category: "Traitor"
      minCount: 0
      maxCount: 10

  # Predefined distributions
  presets:
    standard:
      name: "Standard"
      description: "Balanced gameplay"
      distributions:
        3: {leader: 1, guardian: 1, traitor: 1}
        4: {leader: 1, guardian: 2, traitor: 1}
        5: {leader: 1, guardian: 2, assassin: 1, traitor: 1}
        # ... up to 20 players
    assassination:
      name: "Assassination Heavy"
      description: "More assassins"
      distributions:
        5: {leader: 1, guardian: 1, assassin: 2, traitor: 1}

1.2 Create Configuration Loader

- Create internal/config/config.go:
  - ServerConfig struct with all settings
  - LoadConfig() function to read YAML
  - Validation logic for configurations
  - Default values if config missing

1.3 Update Server Initialization

- Modify cmd/server/main.go to load config on startup
- Pass config to store and handlers
- Use config.MaxPlayersPerRoom instead of hardcoded value

Phase 2: Per-Room Role Configuration

2.1 Extend Room Structure

- Update game/room.go:
type RoleConfiguration struct {
    PresetName    string           // e.g., "standard", "custom"
    EnabledRoles  map[string]bool  // Which roles from master pool
    RoleCounts    map[string]int   // Exact counts per role
    MinPlayers    int
    MaxPlayers    int
}

type Room struct {
    // ... existing fields
    RoleConfig *RoleConfiguration
}

2.2 Create Role Configuration Service

- Create game/role_config_service.go:
  - ValidateRoleConfig() - ensures configuration is valid
  - GetDistributionForPlayerCount() - returns role distribution
  - Support for flexible player counts (not just matching roles)

Phase 3: Update Role Assignment Logic

3.1 Refactor getRoleDistribution()

- Update game/roles.go:
  - Use room's RoleConfig instead of hardcoded switch
  - Support any number of players
  - Handle cases where roles < players (allow multiple of same role)
  - Handle cases where roles > players (some roles unused)

3.2 Update CardService Usage

- Ensure GetRandomCards() can handle larger counts
- Add validation for available cards vs requested

Phase 4: UI Implementation

4.1 Room Creation Form

- Update views/pages/home.templ:
  - Add role preset selector
  - Add "Custom" option that reveals role configuration
  - Use Datastar for dynamic updates

4.2 Create Role Configuration Component

- Create views/components/role_config.templ:
  - List all available roles with toggles
  - Number inputs for role counts
  - Real-time validation feedback
  - Player count display

4.3 Update Lobby Page

- Display selected role configuration
- Show min/max players based on config
- Allow host to modify before game start

Phase 5: Backend Integration

5.1 Update CreateRoom Handler

- Modify handlers/pages.go:
  - Accept role configuration in POST
  - Validate against server config
  - Store in room

5.2 Add Configuration Endpoints

- Create new endpoints:
  - GET /api/role-presets - return available presets
  - POST /room/{code}/role-config - update room config
  - GET /room/{code}/role-config - get current config

5.3 Update SSE Updates

- Include role configuration in lobby SSE
- Send updates when host changes config

Phase 6: Validation & Edge Cases

6.1 Server-Side Validation

- Validate role counts match player range
- Ensure at least 1 leader
- Check against server max players
- Validate enabled roles exist

6.2 Dynamic Player Count Support

- Allow games with more players than unique roles
- Distribute roles intelligently (e.g., more guardians)
- Handle edge cases (1 player, 20+ players)

Phase 7: Testing

7.1 Unit Tests

- Test configuration loading
- Test role distribution algorithms
- Test validation logic

7.2 Integration Tests

- Test full flow with Playwright
- Test various player counts
- Test preset and custom configs

7.3 Edge Case Tests

- Single player games
- Maximum player games
- Invalid configurations

Key Implementation Details

1. YAML Config Loading: Use gopkg.in/yaml.v3 (already in dependencies)
2. Backward Compatibility: Default to current hardcoded values if no config
3. Flexibility: Support any player count, not just matching role counts
4. Server Authority: All validation server-side, never trust client
5. SSE Updates: Use existing SSE infrastructure for real-time updates

Success Criteria

- ✅ Server boots with YAML configuration
- ✅ Hosts can select role presets or customize
- ✅ Games support 1-20 players with flexible distributions
- ✅ Real-time UI updates via SSE
- ✅ All existing tests pass
- ✅ New tests cover role configuration
