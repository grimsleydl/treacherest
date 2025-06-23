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

func TestSSEMorphingIssue(t *testing.T) {
	store := NewInMemoryStore()
	eventBus := NewEventBus()
	h := NewHandler(store, eventBus)
	
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	// Launch browser with devtools
	l := launcher.New().
		Headless(false).
		Devtools(true).
		MustLaunch()
	defer l.Cleanup()

	browser := rod.New().
		ControlURL(l).
		MustConnect().
		SlowMotion(100 * time.Millisecond).
		Trace(true)
	defer browser.MustClose()

	t.Run("test lobby container preservation during morphing", func(t *testing.T) {
		// Player 1 creates room
		page1 := browser.MustPage()
		defer page1.MustClose()

		// Navigate and create room
		page1.MustNavigate(server.URL)
		
		// Enable console logging after navigation
		page1.MustEval(`() => {
			window.addEventListener('error', (e) => {
				console.error('PAGE ERROR:', e.message);
			});
			
			// Monitor mutations
			const observer = new MutationObserver((mutations) => {
				mutations.forEach((mutation) => {
					if (mutation.type === 'childList' && mutation.target.id) {
						console.log('DOM Mutation on', mutation.target.id, '- removed nodes:', mutation.removedNodes.length, 'added nodes:', mutation.addedNodes.length);
					}
				});
			});
			observer.observe(document.body, { childList: true, subtree: true });
		}`)
		page1.MustElement("input[name='playerName']").MustInput("Player 1")
		page1.MustElement("button[type='submit']").MustClick()
		page1.MustWaitLoad()

		// Get room code
		roomCode := strings.TrimSpace(page1.MustElement(".room-code").MustText())
		t.Logf("Room created: %s", roomCode)

		// Verify initial DOM structure
		initialHTML := page1.MustElement("body").MustHTML()
		assert.Contains(t, initialHTML, `id="lobby-container"`, "Initial page should have lobby-container")
		assert.Contains(t, initialHTML, `data-on-load`, "Initial page should have data-on-load")

		// Log initial structure
		lobbyContainer1 := page1.MustElement("#lobby-container")
		t.Logf("Initial lobby-container present: %v", lobbyContainer1 != nil)

		// Player 2 joins
		page2 := browser.MustPage()
		defer page2.MustClose()
		
		page2.MustNavigate(fmt.Sprintf("%s/room/%s?name=Player+2", server.URL, roomCode))
		page2.MustWaitLoad()

		// Wait for SSE update on player 1's page
		time.Sleep(500 * time.Millisecond)

		// Check DOM after first update
		page1.MustEval(`() => {
			const container = document.querySelector('#lobby-container');
			console.log('After player 2 joins - lobby-container exists:', container !== null);
			if (container) {
				console.log('lobby-container HTML:', container.outerHTML.substring(0, 100) + '...');
			}
		}`)

		// Verify lobby-container still exists
		lobbyExists := page1.MustEval(`() => document.querySelector('#lobby-container') !== null`).Bool()
		assert.True(t, lobbyExists, "lobby-container should exist after first SSE update")

		// Get the HTML after player 2 joins
		bodyHTML2 := page1.MustElement("body").MustHTML()
		t.Logf("Has lobby-container after player 2: %v", strings.Contains(bodyHTML2, `id="lobby-container"`))

		// Player 3 joins - this is where the error occurs
		page3 := browser.MustPage()
		defer page3.MustClose()

		t.Log("Player 3 joining...")
		page3.MustNavigate(fmt.Sprintf("%s/room/%s?name=Player+3", server.URL, roomCode))
		page3.MustWaitLoad()

		// Wait for SSE update
		time.Sleep(500 * time.Millisecond)

		// Check for errors in console
		page1.MustEval(`() => {
			const container = document.querySelector('#lobby-container');
			console.log('After player 3 joins - lobby-container exists:', container !== null);
			if (!container) {
				console.error('CRITICAL: lobby-container is missing from DOM!');
			}
		}`)

		// Final verification
		finalLobbyExists := page1.MustEval(`() => document.querySelector('#lobby-container') !== null`).Bool()
		require.True(t, finalLobbyExists, "lobby-container MUST exist after all SSE updates")

		// Verify all players see each other
		players1 := page1.MustElements(".player")
		assert.Len(t, players1, 3, "Player 1 should see 3 players")

		players2 := page2.MustElements(".player") 
		assert.Len(t, players2, 3, "Player 2 should see 3 players")

		players3 := page3.MustElements(".player")
		assert.Len(t, players3, 3, "Player 3 should see 3 players")
	})
}