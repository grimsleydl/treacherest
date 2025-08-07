import { test, expect, Page } from '@playwright/test';

test.describe('SSE Countdown Synchronization', () => {
  const baseURL = 'http://localhost:8080';
  
  test('multiple browsers should all auto-refresh to countdown when game starts', async ({ browser }) => {
    // Create 3 browser contexts to simulate 3 different players
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();
    const context3 = await browser.newContext();
    
    const page1 = await context1.newPage();
    const page2 = await context2.newPage();
    const page3 = await context3.newPage();
    
    try {
      // Player 1: Create room
      await page1.goto(baseURL);
      await page1.click('text=Create Room');
      await page1.fill('input[name="playerName"]', 'Player1');
      await page1.click('button[type="submit"]');
      
      // Get room code
      await page1.waitForSelector('h2:has-text("Room Code:")');
      const roomCodeElement = await page1.$('h2:has-text("Room Code:")');
      const roomCodeText = await roomCodeElement?.textContent();
      const roomCode = roomCodeText?.split(': ')[1];
      console.log(`Created room: ${roomCode}`);
      
      // Player 2: Join room
      await page2.goto(`${baseURL}/room/${roomCode}`);
      await page2.fill('input[name="name"]', 'Player2');
      await page2.click('button[type="submit"]');
      await page2.waitForSelector('text=Player2');
      
      // Player 3: Join room (this one often has issues)
      await page3.goto(`${baseURL}/room/${roomCode}`);
      await page3.fill('input[name="name"]', 'Player3');
      await page3.click('button[type="submit"]');
      await page3.waitForSelector('text=Player3');
      
      // Wait a moment to ensure all SSE connections are established
      await page1.waitForTimeout(1000);
      
      // Player 1: Start game
      await page1.click('button:has-text("Start Game")');
      
      // All players should redirect to game page and see countdown
      const pages = [
        { page: page1, name: 'Player1' },
        { page: page2, name: 'Player2' },
        { page: page3, name: 'Player3' }
      ];
      
      for (const { page, name } of pages) {
        // Wait for redirect to game page
        await page.waitForURL(`${baseURL}/game/${roomCode}`, { timeout: 5000 });
        
        // Should see countdown
        const countdownVisible = await page.locator('text=Revealing roles in').isVisible();
        console.log(`${name} sees countdown: ${countdownVisible}`);
        expect(countdownVisible).toBe(true);
        
        // Take screenshot for debugging
        await page.screenshot({ path: `test-results/countdown-${name}.png` });
      }
      
      // Wait for countdown to complete
      await page1.waitForTimeout(6000);
      
      // All players should now see their roles
      for (const { page, name } of pages) {
        const roleCardVisible = await page.locator('.role-card').isVisible();
        console.log(`${name} sees role card: ${roleCardVisible}`);
        expect(roleCardVisible).toBe(true);
        
        // Take screenshot of final state
        await page.screenshot({ path: `test-results/role-${name}.png` });
      }
      
    } finally {
      await context1.close();
      await context2.close();
      await context3.close();
    }
  });
  
  test('late joiner during countdown should sync properly', async ({ browser }) => {
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();
    
    const page1 = await context1.newPage();
    const page2 = await context2.newPage();
    
    try {
      // Player 1: Create room and start game
      await page1.goto(baseURL);
      await page1.click('text=Create Room');
      await page1.fill('input[name="playerName"]', 'Player1');
      await page1.click('button[type="submit"]');
      
      const roomCodeElement = await page1.$('h2:has-text("Room Code:")');
      const roomCodeText = await roomCodeElement?.textContent();
      const roomCode = roomCodeText?.split(': ')[1];
      
      await page1.click('button:has-text("Start Game")');
      await page1.waitForURL(`${baseURL}/game/${roomCode}`);
      
      // Wait 2 seconds into countdown
      await page1.waitForTimeout(2000);
      
      // Player 2: Join during countdown
      await page2.goto(`${baseURL}/room/${roomCode}`);
      await page2.fill('input[name="name"]', 'Player2');
      await page2.click('button[type="submit"]');
      
      // Player 2 should be redirected to game page and see countdown
      await page2.waitForURL(`${baseURL}/game/${roomCode}`, { timeout: 5000 });
      
      // Should see countdown with correct remaining time
      const countdownText = await page2.locator('h1:has-text("Revealing roles in")').textContent();
      console.log(`Late joiner sees: ${countdownText}`);
      expect(countdownText).toContain('Revealing roles in');
      
      // Both should transition to game after countdown
      await page1.waitForTimeout(4000);
      
      const player1HasRole = await page1.locator('.role-card').isVisible();
      const player2HasRole = await page2.locator('.role-card').isVisible();
      
      expect(player1HasRole).toBe(true);
      expect(player2HasRole).toBe(true);
      
    } finally {
      await context1.close();
      await context2.close();
    }
  });
  
  test('SSE connection should handle reconnection properly', async ({ page }) => {
    await page.goto(baseURL);
    
    // Monitor network for SSE connections
    const sseConnections: string[] = [];
    const consoleErrors: string[] = [];
    
    page.on('request', request => {
      if (request.url().includes('/sse/')) {
        sseConnections.push(request.url());
        console.log(`SSE connection: ${request.url()}`);
      }
    });
    
    // Monitor console messages for errors
    page.on('console', msg => {
      if (msg.type() === 'error') {
        const error = `Console error: ${msg.text()}`;
        consoleErrors.push(error);
        console.log(error);
      }
    });
    
    // Create and join room
    await page.click('text=Create Room');
    await page.fill('input[name="playerName"]', 'TestPlayer');
    await page.click('button[type="submit"]');
    
    console.log(`Current URL after form submit: ${page.url()}`);
    
    // Wait for redirect to lobby page
    await page.waitForFunction(() => window.location.pathname.includes('/room/'), { timeout: 5000 });
    console.log(`URL after redirect: ${page.url()}`);
    
    // Check if Datastar is loaded
    const datastartLoaded = await page.evaluate(() => {
      return typeof window.datastar !== 'undefined';
    });
    console.log(`Datastar loaded: ${datastartLoaded}`);
    
    // Check for JavaScript errors
    const consoleMessages = await page.evaluate(() => {
      return window.__errors || [];
    });
    console.log(`Console errors: ${JSON.stringify(consoleMessages)}`);
    
    // Take screenshot to see what page we're on
    await page.screenshot({ path: 'test-results/debug-lobby-page.png' });
    
    // Should have one SSE connection to lobby
    // Wait longer for Datastar to load and SSE connection to establish
    await page.waitForTimeout(3000);
    
    console.log(`All SSE connections: ${JSON.stringify(sseConnections)}`);
    const lobbyConnections = sseConnections.filter(url => url.includes('/sse/lobby/'));
    console.log(`Lobby SSE connections: ${JSON.stringify(lobbyConnections)}`);
    
    expect(lobbyConnections).toHaveLength(1);
    
    // Start game
    await page.click('button:has-text("Start Game")');
    
    // Should redirect and create one SSE connection to game
    await page.waitForURL(/\/game\//);
    await page.waitForTimeout(1000);
    
    const gameConnections = sseConnections.filter(url => url.includes('/sse/game/'));
    console.log(`Game SSE connections: ${gameConnections.length}`);
    
    // Should only have ONE game SSE connection (not multiple)
    expect(gameConnections).toHaveLength(1);
  });
});