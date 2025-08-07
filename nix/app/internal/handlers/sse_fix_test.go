package handlers

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSEConnectionPersistence verifies SSE connections don't timeout and no duplicates are created
func TestSSEConnectionPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create handler and server
	h := newTestHandler()
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Launch browser in headless mode for CI
	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	// Monitor console errors
	setupConsoleMonitor := func(page *rod.Page, playerName string) {
		page.MustEval(`(playerName) => {
			window.playerName = playerName;
			window.errorCount = 0;
			window.sseConnections = 0;
			
			// Monitor errors
			window.addEventListener('error', (e) => {
				console.error(playerName + ' ERROR:', e);
				window.errorCount++;
			});
			
			// Monitor SSE connections
			const originalEventSource = window.EventSource;
			window.EventSource = function(url) {
				window.sseConnections++;
				console.log(playerName + ' SSE Connection #' + window.sseConnections + ' to:', url);
				return new originalEventSource(url);
			};
		}`, playerName)
	}

	t.Run("no duplicate SSE connections on player joins", func(t *testing.T) {
		// Player 1 creates room
		page1 := browser.MustPage()
		defer page1.MustClose()

		setupConsoleMonitor(page1, "Player1")

		page1.MustNavigate(ts.URL)
		page1.MustElement("input[name='name']").MustInput("Player1")
		page1.MustElement("button[type='submit']").MustClick()
		page1.MustWaitLoad()

		roomCode := page1.MustInfo().URL[len(ts.URL+"/room/"):]

		// Check initial SSE connections
		time.Sleep(1 * time.Second)
		sseCount1 := page1.MustEval(`() => window.sseConnections`).Int()
		assert.Equal(t, 1, sseCount1, "Player 1 should have exactly 1 SSE connection")

		// Player 2 joins
		page2 := browser.MustPage()
		defer page2.MustClose()

		setupConsoleMonitor(page2, "Player2")

		page2.MustNavigate(ts.URL + "/room/" + roomCode)
		page2.MustElement("input[name='name']").MustInput("Player2")
		page2.MustElement("button[type='submit']").MustClick()

		// Wait for SSE update
		time.Sleep(2 * time.Second)

		// Check SSE connections haven't increased
		sseCount1After := page1.MustEval(`() => window.sseConnections`).Int()
		assert.Equal(t, 1, int(sseCount1After), "Player 1 should still have exactly 1 SSE connection after Player 2 joins")

		// Player 3 joins
		page3 := browser.MustPage()
		defer page3.MustClose()

		setupConsoleMonitor(page3, "Player3")

		page3.MustNavigate(ts.URL + "/room/" + roomCode)
		page3.MustElement("input[name='name']").MustInput("Player3")
		page3.MustElement("button[type='submit']").MustClick()

		// Wait for SSE update
		time.Sleep(2 * time.Second)

		// Final check - all players should have exactly 1 SSE connection
		for i, page := range []*rod.Page{page1, page2, page3} {
			sseCount := page.MustEval(`() => window.sseConnections`).Int()
			errorCount := page.MustEval(`() => window.errorCount`).Int()

			assert.Equal(t, 1, sseCount, fmt.Sprintf("Player %d should have exactly 1 SSE connection", i+1))
			assert.Equal(t, 0, errorCount, fmt.Sprintf("Player %d should have no errors", i+1))

			// Verify all players are visible
			playerList, _ := page.MustElement(".player-list").Text()
			assert.Contains(t, playerList, "Player1")
			assert.Contains(t, playerList, "Player2")
			assert.Contains(t, playerList, "Player3")
		}
	})
}

// TestSSEConnectionSurvives70Seconds verifies connections don't timeout after 60 seconds
func TestSSEConnectionSurvives70Seconds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running browser test in short mode")
	}

	h := newTestHandler()
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()
	defer page.MustClose()

	// Set up monitoring
	page.MustEval(`() => {
		window.lastUpdate = Date.now();
		window.updateCount = 0;
		window.connectionClosed = false;
		
		// Monitor for any DOM updates
		const observer = new MutationObserver(() => {
			window.lastUpdate = Date.now();
			window.updateCount++;
		});
		
		observer.observe(document.body, {
			childList: true,
			subtree: true,
			characterData: true
		});
	}`)

	// Create room
	page.MustNavigate(ts.URL)
	page.MustElement("input[name='name']").MustInput("TestPlayer")
	page.MustElement("button[type='submit']").MustClick()
	page.MustWaitLoad()

	startTime := time.Now()
	fmt.Println("Starting 70-second SSE connection test...")

	// Monitor for 70 seconds (past the old 60-second timeout)
	checkInterval := 10 * time.Second
	for time.Since(startTime) < 70*time.Second {
		time.Sleep(checkInterval)

		// Check if page is still functional
		bodyHTML, _ := page.MustElement("body").HTML()
		if strings.TrimSpace(bodyHTML) == "" {
			t.Fatalf("Page went white after %v", time.Since(startTime))
		}

		// Check DOM structure is intact
		hasContainer := page.MustEval(`() => !!document.getElementById('lobby-container')`).Bool()
		hasContent := page.MustEval(`() => !!document.getElementById('lobby-content')`).Bool()

		if !hasContainer || !hasContent {
			t.Fatalf("Lost DOM structure after %v", time.Since(startTime))
		}

		fmt.Printf("✓ Connection alive at %v\n", time.Since(startTime).Round(time.Second))
	}

	// Verify player list is still visible
	playerList, err := page.MustElement(".player-list").Text()
	require.NoError(t, err)
	assert.Contains(t, playerList, "TestPlayer")

	fmt.Println("✓ SSE connection survived 70 seconds!")
}

// TestNoTargetsFoundError verifies the NoTargetsFound error is fixed
func TestNoTargetsFoundError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	h := newTestHandler()
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	// Capture console errors
	captureErrors := func(page *rod.Page) {
		page.MustEval(`() => {
			window.capturedErrors = [];
			window.addEventListener('error', (e) => {
				window.capturedErrors.push(e.message);
			});
			// Also capture unhandled promise rejections
			window.addEventListener('unhandledrejection', (e) => {
				window.capturedErrors.push(e.reason.toString());
			});
		}`)
	}

	// Create 3 players and verify no NoTargetsFound errors
	pages := make([]*rod.Page, 3)
	for i := 0; i < 3; i++ {
		page := browser.MustPage()
		pages[i] = page
		defer page.MustClose()

		captureErrors(page)

		if i == 0 {
			// First player creates room
			page.MustNavigate(ts.URL)
			page.MustElement("input[name='name']").MustInput(fmt.Sprintf("Player%d", i+1))
			page.MustElement("button[type='submit']").MustClick()
			page.MustWaitLoad()
		} else {
			// Other players join
			roomCode := pages[0].MustInfo().URL[len(ts.URL+"/room/"):]
			page.MustNavigate(ts.URL + "/room/" + roomCode)
			page.MustElement("input[name='name']").MustInput(fmt.Sprintf("Player%d", i+1))
			page.MustElement("button[type='submit']").MustClick()
		}

		// Wait for SSE to settle
		time.Sleep(2 * time.Second)
	}

	// Check for NoTargetsFound errors
	for i, page := range pages {
		errors := page.MustEval(`() => window.capturedErrors`).Arr()

		hasNoTargetsFound := false
		for _, err := range errors {
			errStr := fmt.Sprintf("%v", err)
			if strings.Contains(errStr, "NoTargetsFound") {
				hasNoTargetsFound = true
				t.Logf("Player %d has NoTargetsFound error: %s", i+1, errStr)
			}
		}

		assert.False(t, hasNoTargetsFound, fmt.Sprintf("Player %d should not have NoTargetsFound errors", i+1))
	}
}

// TestNoSSEConnectionExhaustionDuringCountdown verifies only one SSE connection per game page during countdown
func TestNoSSEConnectionExhaustionDuringCountdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create handler and server
	h := newTestHandler()
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Launch browser in headless mode for CI
	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	// Monitor SSE connections
	setupSSEMonitor := func(page *rod.Page, playerName string) {
		page.MustEval(`(playerName) => {
			window.playerName = playerName;
			window.sseConnections = 0;
			window.sseUrls = [];
			
			// Override EventSource to track connections
			const originalEventSource = window.EventSource;
			window.EventSource = function(url) {
				window.sseConnections++;
				window.sseUrls.push(url);
				console.log(playerName + ' SSE Connection #' + window.sseConnections + ' to:', url);
				
				// Call original constructor
				const es = new originalEventSource(url);
				
				// Log when connection closes
				const originalClose = es.close.bind(es);
				es.close = function() {
					console.log(playerName + ' SSE Connection closed:', url);
					originalClose();
				};
				
				return es;
			};
		}`, playerName)
	}

	t.Run("no duplicate SSE connections during game countdown", func(t *testing.T) {
		// Player 1 creates room
		page1 := browser.MustPage()
		defer page1.MustClose()

		setupSSEMonitor(page1, "Player1")

		page1.MustNavigate(ts.URL)
		page1.MustElement("input[name='name']").MustInput("Player1")
		page1.MustElement("button[type='submit']").MustClick()
		page1.MustWaitLoad()

		roomCode := page1.MustInfo().URL[len(ts.URL+"/room/"):]

		// Player 2 joins
		page2 := browser.MustPage()
		defer page2.MustClose()

		setupSSEMonitor(page2, "Player2")

		page2.MustNavigate(ts.URL + "/room/" + roomCode)
		page2.MustElement("input[name='name']").MustInput("Player2")
		page2.MustElement("button[type='submit']").MustClick()

		// Wait for lobby to stabilize
		time.Sleep(2 * time.Second)

		// Check initial lobby SSE connections
		lobbySSE1 := page1.MustEval(`() => window.sseConnections`).Int()
		lobbySSE2 := page2.MustEval(`() => window.sseConnections`).Int()
		assert.Equal(t, 1, lobbySSE1, "Player 1 should have exactly 1 SSE connection in lobby")
		assert.Equal(t, 1, lobbySSE2, "Player 2 should have exactly 1 SSE connection in lobby")

		// Start the game
		page1.MustElement("button").MustClick() // Start Game button

		// Wait for redirect to game page
		time.Sleep(2 * time.Second)

		// Both players should now be on game page
		assert.Contains(t, page1.MustInfo().URL, "/game/", "Player 1 should be on game page")
		assert.Contains(t, page2.MustInfo().URL, "/game/", "Player 2 should be on game page")

		// Get initial game SSE connection counts
		gameSSEStart1 := page1.MustEval(`() => window.sseConnections`).Int()
		gameSSEStart2 := page2.MustEval(`() => window.sseConnections`).Int()

		// Monitor SSE connections during the 5-second countdown
		startTime := time.Now()
		maxConnections1 := gameSSEStart1
		maxConnections2 := gameSSEStart2

		for time.Since(startTime) < 6*time.Second {
			time.Sleep(500 * time.Millisecond)

			current1 := page1.MustEval(`() => window.sseConnections`).Int()
			current2 := page2.MustEval(`() => window.sseConnections`).Int()

			if current1 > maxConnections1 {
				maxConnections1 = current1
			}
			if current2 > maxConnections2 {
				maxConnections2 = current2
			}

			// Check countdown is updating
			countdownText1, _ := page1.MustElement(".countdown h1").Text()

			t.Logf("Time: %v, P1 connections: %d, P2 connections: %d, Countdown: %s",
				time.Since(startTime).Round(time.Millisecond),
				current1, current2, countdownText1)
		}

		// Assert no duplicate connections were created during countdown
		assert.Equal(t, gameSSEStart1, maxConnections1,
			"Player 1 should not create new SSE connections during countdown")
		assert.Equal(t, gameSSEStart2, maxConnections2,
			"Player 2 should not create new SSE connections during countdown")

		// Wait for countdown to finish and roles to be revealed
		time.Sleep(2 * time.Second)

		// Verify game is showing roles
		roleCard1 := page1.MustHas(".role-card")
		roleCard2 := page2.MustHas(".role-card")
		assert.True(t, roleCard1, "Player 1 should see role card after countdown")
		assert.True(t, roleCard2, "Player 2 should see role card after countdown")

		// Final check - still no extra connections
		finalSSE1 := page1.MustEval(`() => window.sseConnections`).Int()
		finalSSE2 := page2.MustEval(`() => window.sseConnections`).Int()

		// Should have exactly 2 connections total: 1 for lobby, 1 for game
		assert.Equal(t, 2, finalSSE1, "Player 1 should have exactly 2 total SSE connections (lobby + game)")
		assert.Equal(t, 2, finalSSE2, "Player 2 should have exactly 2 total SSE connections (lobby + game)")

		// Log all SSE URLs for debugging
		urls1 := page1.MustEval(`() => window.sseUrls`).Arr()
		urls2 := page2.MustEval(`() => window.sseUrls`).Arr()
		t.Logf("Player 1 SSE URLs: %v", urls1)
		t.Logf("Player 2 SSE URLs: %v", urls2)
	})
}

// TestPlayerReconnectionDuringCountdown verifies correct countdown handling on reconnection
func TestPlayerReconnectionDuringCountdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	h := newTestHandler()
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	// Player 1 creates room and starts game
	page1 := browser.MustPage()
	defer page1.MustClose()

	page1.MustNavigate(ts.URL)
	page1.MustElement("input[name='name']").MustInput("Player1")
	page1.MustElement("button[type='submit']").MustClick()
	page1.MustWaitLoad()

	roomCode := page1.MustInfo().URL[len(ts.URL+"/room/"):]

	// Start the game
	page1.MustElement("button").MustClick()
	time.Sleep(2 * time.Second) // Wait for redirect

	// Wait 2 seconds into countdown
	time.Sleep(2 * time.Second)

	// Player 2 joins mid-countdown
	page2 := browser.MustPage()
	defer page2.MustClose()

	page2.MustNavigate(ts.URL + "/game/" + roomCode)

	// Check that Player 2 sees the correct remaining countdown (not 5)
	time.Sleep(1 * time.Second)
	countdownText, _ := page2.MustElement(".countdown h1").Text()

	// Should show 1 or 0 seconds remaining, not 5
	assert.NotContains(t, countdownText, "5", "New player should see actual remaining countdown, not initial 5 seconds")
	assert.Contains(t, countdownText, "Revealing roles in", "Should show countdown message")

	// Wait for countdown to finish
	time.Sleep(3 * time.Second)

	// Both players should see roles
	roleCard1 := page1.MustHas(".role-card")
	roleCard2 := page2.MustHas(".role-card")
	assert.True(t, roleCard1, "Player 1 should see role card")
	assert.True(t, roleCard2, "Player 2 should see role card after joining mid-countdown")
}
