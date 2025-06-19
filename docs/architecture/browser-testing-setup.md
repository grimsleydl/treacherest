# Browser/E2E Testing Setup for Treacherest

## Overview
This document provides a complete setup guide for browser-based end-to-end testing using Playwright, integrated with the Nix development environment and the existing Go test suite.

## Why Playwright?
- **Multi-browser support**: Test on Chromium, Firefox, and WebKit
- **Built-in mobile emulation**: Test responsive design
- **Network condition simulation**: Test SSE reconnection
- **Parallel execution**: Fast test runs
- **Great debugging tools**: Screenshots, videos, trace viewer
- **Nix package available**: Easy integration

## Nix Integration

### 1. Update Nix Flake
```nix
# In flake.nix, add to buildInputs:
{
  buildInputs = with pkgs; [
    # ... existing packages
    playwright
    nodejs_20  # For running Playwright
  ];
}
```

### 2. Add E2E Test Commands
```nix
# In nix/local/shells.nix, add aliases:
{
  shellHook = ''
    # ... existing aliases
    alias test-e2e='cd nix/app && npm run test:e2e'
    alias test-e2e-ui='cd nix/app && npm run test:e2e:ui'
    alias test-e2e-debug='cd nix/app && npm run test:e2e:debug'
  '';
}
```

## Project Structure

```
nix/app/
├── e2e/
│   ├── playwright.config.ts
│   ├── package.json
│   ├── tests/
│   │   ├── 01-room-creation.spec.ts
│   │   ├── 02-player-joining.spec.ts
│   │   ├── 03-game-flow.spec.ts
│   │   ├── 04-error-handling.spec.ts
│   │   └── 05-mobile-experience.spec.ts
│   ├── fixtures/
│   │   ├── game-fixture.ts
│   │   └── player-fixture.ts
│   └── utils/
│       ├── test-helpers.ts
│       └── selectors.ts
```

## Configuration Files

### 1. E2E package.json
```json
{
  "name": "treacherest-e2e",
  "version": "1.0.0",
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:debug": "playwright test --debug",
    "test:e2e:headed": "playwright test --headed",
    "test:e2e:report": "playwright show-report"
  },
  "devDependencies": {
    "@playwright/test": "^1.40.0",
    "typescript": "^5.3.0"
  }
}
```

### 2. Playwright Configuration
```typescript
// e2e/playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['junit', { outputFile: 'test-results/junit.xml' }],
    ['list'],
  ],
  
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:8080',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000,
  },

  projects: [
    {
      name: 'Desktop Chrome',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'Desktop Firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  webServer: {
    command: 'cd .. && dev',
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
});
```

## Test Fixtures

### 1. Game Fixture
```typescript
// e2e/fixtures/game-fixture.ts
import { test as base, Page } from '@playwright/test';

type GameFixture = {
  createRoom: (playerName: string) => Promise<string>;
  joinRoom: (page: Page, roomCode: string, playerName: string) => Promise<void>;
  waitForSSEConnection: (page: Page) => Promise<void>;
};

export const test = base.extend<GameFixture>({
  createRoom: async ({ page }, use) => {
    const createRoom = async (playerName: string): Promise<string> => {
      await page.goto('/');
      await page.fill('input[name="playerName"]', playerName);
      await page.click('button[type="submit"]');
      await page.waitForURL(/\/room\/[A-Z0-9]{5}/);
      return page.url().split('/').pop()!;
    };
    await use(createRoom);
  },

  joinRoom: async ({}, use) => {
    const joinRoom = async (page: Page, roomCode: string, playerName: string) => {
      await page.goto(`/room/${roomCode}`);
      await page.fill('input[name="name"]', playerName);
      await page.click('button[type="submit"]');
      await page.waitForSelector('.player-list');
    };
    await use(joinRoom);
  },

  waitForSSEConnection: async ({}, use) => {
    const waitForSSE = async (page: Page) => {
      await page.waitForFunction(() => {
        // Check if SSE connection is established
        return window.EventSource && 
               Array.from(document.querySelectorAll('[data-on-load]')).length > 0;
      });
    };
    await use(waitForSSE);
  },
});
```

### 2. Player Fixture
```typescript
// e2e/fixtures/player-fixture.ts
import { Browser, BrowserContext, Page } from '@playwright/test';

export class Player {
  context: BrowserContext;
  page: Page;
  name: string;

  constructor(context: BrowserContext, page: Page, name: string) {
    this.context = context;
    this.page = page;
    this.name = name;
  }

  static async create(browser: Browser, name: string): Promise<Player> {
    const context = await browser.newContext();
    const page = await context.newPage();
    return new Player(context, page, name);
  }

  async cleanup() {
    await this.context.close();
  }
}

export async function createPlayers(
  browser: Browser, 
  names: string[]
): Promise<Player[]> {
  return Promise.all(names.map(name => Player.create(browser, name)));
}
```

## Test Helpers

### 1. Selectors
```typescript
// e2e/utils/selectors.ts
export const selectors = {
  // Forms
  playerNameInput: 'input[name="playerName"]',
  joinNameInput: 'input[name="name"]',
  submitButton: 'button[type="submit"]',
  
  // Room
  roomCode: '.room-code',
  playerList: '.player-list',
  playerItem: '.player',
  playerCount: '.player-count',
  
  // Game controls
  startGameButton: 'button:has-text("Start Game")',
  leaveRoomButton: 'button:has-text("Leave Room")',
  
  // Game state
  gameContainer: '#game-container',
  lobbyContainer: '#lobby-container',
  connectionStatus: '.connection-status',
  
  // Messages
  errorMessage: '.error-message',
  waitingMessage: ':has-text("Need at least 4 players")',
};
```

### 2. Test Helpers
```typescript
// e2e/utils/test-helpers.ts
import { Page, expect } from '@playwright/test';
import { selectors } from './selectors';

export async function verifyPlayerInList(
  page: Page, 
  playerName: string
): Promise<void> {
  await expect(page.locator(selectors.playerList))
    .toContainText(playerName);
}

export async function getRoomCode(page: Page): Promise<string> {
  const roomCodeElement = await page.locator(selectors.roomCode);
  return await roomCodeElement.textContent() || '';
}

export async function waitForPlayerCount(
  page: Page, 
  count: number
): Promise<void> {
  await expect(page.locator(selectors.playerCount))
    .toContainText(`${count}/`);
}

export async function simulateNetworkOutage(page: Page): Promise<void> {
  await page.context().setOffline(true);
}

export async function restoreNetwork(page: Page): Promise<void> {
  await page.context().setOffline(false);
}
```

## Test Scenarios

### 1. Room Creation Test
```typescript
// e2e/tests/01-room-creation.spec.ts
import { test, expect } from '../fixtures/game-fixture';
import { selectors } from '../utils/selectors';

test.describe('Room Creation', () => {
  test('player can create a new room', async ({ page, createRoom }) => {
    const roomCode = await createRoom('Alice');
    
    // Verify room code displayed
    await expect(page.locator(selectors.roomCode))
      .toHaveText(roomCode);
    
    // Verify player in list
    await expect(page.locator(selectors.playerList))
      .toContainText('Alice');
    
    // Verify SSE connection established
    await expect(page.locator(selectors.lobbyContainer))
      .toHaveAttribute('data-on-load', `/sse/lobby/${roomCode}`);
  });

  test('room code is unique', async ({ browser }) => {
    const roomCodes = new Set<string>();
    
    // Create multiple rooms
    for (let i = 0; i < 10; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();
      
      await page.goto('/');
      await page.fill(selectors.playerNameInput, `Player${i}`);
      await page.click(selectors.submitButton);
      
      await page.waitForURL(/\/room\/[A-Z0-9]{5}/);
      const code = page.url().split('/').pop()!;
      
      expect(roomCodes.has(code)).toBe(false);
      roomCodes.add(code);
      
      await context.close();
    }
  });
});
```

### 2. Multiplayer Flow Test
```typescript
// e2e/tests/03-game-flow.spec.ts
import { test, expect } from '../fixtures/game-fixture';
import { createPlayers } from '../fixtures/player-fixture';
import { selectors } from '../utils/selectors';
import { verifyPlayerInList, waitForPlayerCount } from '../utils/test-helpers';

test.describe('Multiplayer Game Flow', () => {
  test('four players can complete a game', async ({ browser, createRoom }) => {
    const players = await createPlayers(browser, ['Alice', 'Bob', 'Charlie', 'David']);
    
    try {
      // Player 1 creates room
      const roomCode = await createRoom.call(
        { page: players[0].page }, 
        players[0].name
      );
      
      // Other players join
      for (let i = 1; i < 4; i++) {
        await players[i].page.goto(`/room/${roomCode}`);
        await players[i].page.fill(selectors.joinNameInput, players[i].name);
        await players[i].page.click(selectors.submitButton);
      }
      
      // Verify all players see each other
      for (const player of players) {
        for (const otherPlayer of players) {
          await verifyPlayerInList(player.page, otherPlayer.name);
        }
        await waitForPlayerCount(player.page, 4);
      }
      
      // Start game
      await players[0].page.click(selectors.startGameButton);
      
      // Verify all redirect to game
      for (const player of players) {
        await expect(player.page).toHaveURL(`/game/${roomCode}`);
        await expect(player.page.locator(selectors.gameContainer)).toBeVisible();
      }
      
    } finally {
      // Cleanup
      await Promise.all(players.map(p => p.cleanup()));
    }
  });
});
```

### 3. SSE Reconnection Test
```typescript
// e2e/tests/04-error-handling.spec.ts
import { test, expect } from '../fixtures/game-fixture';
import { selectors } from '../utils/selectors';
import { simulateNetworkOutage, restoreNetwork } from '../utils/test-helpers';

test.describe('Error Handling', () => {
  test('handles SSE disconnection and reconnection', async ({ page, createRoom }) => {
    const roomCode = await createRoom('TestPlayer');
    
    // Verify initial connection
    await expect(page.locator(selectors.connectionStatus))
      .toContainText('Connected');
    
    // Simulate network outage
    await simulateNetworkOutage(page);
    
    // Verify disconnection detected
    await expect(page.locator(selectors.connectionStatus))
      .toContainText('Reconnecting', { timeout: 5000 });
    
    // Restore network
    await restoreNetwork(page);
    
    // Verify reconnection
    await expect(page.locator(selectors.connectionStatus))
      .toContainText('Connected', { timeout: 10000 });
    
    // Verify functionality restored
    await page.click(selectors.leaveRoomButton);
    await expect(page).toHaveURL('/');
  });
});
```

### 4. Mobile Experience Test
```typescript
// e2e/tests/05-mobile-experience.spec.ts
import { test, expect, devices } from '@playwright/test';
import { selectors } from '../utils/selectors';

test.use({ ...devices['iPhone 12'] });

test.describe('Mobile Experience', () => {
  test('touch interactions work correctly', async ({ page }) => {
    await page.goto('/');
    
    // Test touch on input
    await page.tap(selectors.playerNameInput);
    await page.fill(selectors.playerNameInput, 'MobilePlayer');
    
    // Test touch on button
    await page.tap(selectors.submitButton);
    
    // Verify navigation
    await expect(page).toHaveURL(/\/room\/[A-Z0-9]{5}/);
    
    // Test viewport
    const viewport = page.viewportSize();
    expect(viewport?.width).toBeLessThan(400);
    
    // Verify responsive layout
    await expect(page.locator('.mobile-menu')).toBeVisible();
  });
});
```

## CI Integration

### 1. GitHub Actions Workflow
```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: cachix/install-nix-action@v22
      
      - name: Install Playwright Browsers
        run: |
          nix develop --command npx playwright install --with-deps
      
      - name: Run E2E Tests
        run: |
          nix develop --command test-e2e
      
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: nix/app/e2e/playwright-report/
          retention-days: 30
```

### 2. Local Pre-commit Hook
```yaml
# In lefthook.yml, add:
pre-commit:
  commands:
    e2e-critical:
      run: cd nix/app && npm run test:e2e -- --grep "@critical"
```

## Debugging Tips

### 1. Interactive Debugging
```bash
# Run with UI mode for visual debugging
test-e2e-ui

# Run in headed mode to see browser
npx playwright test --headed

# Debug specific test
npx playwright test 03-game-flow.spec.ts --debug
```

### 2. Trace Viewer
```bash
# After test failure, view trace
npx playwright show-trace trace.zip
```

### 3. VS Code Integration
```json
// .vscode/settings.json
{
  "playwright.reuseBrowser": true,
  "playwright.showTrace": true
}
```

## Best Practices

### 1. Test Organization
- Group related tests in describe blocks
- Use clear, descriptive test names
- Tag critical tests with @critical
- Keep tests independent

### 2. Selectors Strategy
- Use data-testid for critical elements
- Prefer semantic selectors (role, text)
- Avoid CSS selectors that might change
- Create reusable selector constants

### 3. Waiting Strategies
```typescript
// BAD: Fixed timeouts
await page.waitForTimeout(1000);

// GOOD: Wait for specific conditions
await page.waitForSelector('.player-list');
await expect(page.locator('.status')).toHaveText('Ready');
```

### 4. Test Data Management
```typescript
// Create unique test data
const testId = Date.now();
const playerName = `TestPlayer_${testId}`;
```

## Performance Considerations

### 1. Parallel Execution
- Tests run in parallel by default
- Use test.describe.serial() for dependent tests
- Isolate tests using different room codes

### 2. Resource Management
```typescript
test.afterEach(async ({ page }) => {
  // Clean up any created resources
  if (page.url().includes('/room/')) {
    await page.click('button:has-text("Leave Room")');
  }
});
```

### 3. Optimization Tips
- Reuse browser contexts when possible
- Minimize page navigations
- Use page.goto() with waitUntil: 'networkidle'
- Cache static assets in CI

## Summary

This E2E testing setup provides:
1. **Complete Playwright integration** with Nix
2. **Comprehensive test scenarios** covering all user flows
3. **Mobile testing support** for responsive design
4. **CI/CD integration** with artifact storage
5. **Debugging tools** for quick troubleshooting
6. **Performance optimizations** for fast test runs

With this setup, you can confidently test all user-facing functionality and ensure a high-quality multiplayer gaming experience across all devices and browsers.