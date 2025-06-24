package handlers

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLobbySSEMultiplePlayers tests SSE updates work correctly with multiple players
func TestLobbySSEMultiplePlayers(t *testing.T) {
	// Skip if running in CI without browser
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create handler with in-memory store
	h := New(store.NewMemoryStore())

	// Create test server
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Launch browser in headless mode
	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	// Test context with timeout (removed - not used)

	t.Run("multiple players join lobby without DOM errors", func(t *testing.T) {
		// Player 1 creates room
		page1 := browser.MustPage()
		defer page1.MustClose()

		page1.MustNavigate(ts.URL)

		// Create room
		page1.MustElement("input[name='name']").MustInput("Player1")
		page1.MustElement("button[type='submit']").MustClick()

		// Wait for lobby page
		page1.MustWaitLoad()
		assert.Contains(t, page1.MustInfo().URL, "/room/")

		// Extract room code from URL
		roomCode := page1.MustInfo().URL[len(ts.URL+"/room/"):]

		// Verify initial DOM structure
		lobbyContainer := page1.MustElement("#lobby-container")
		assert.NotNil(t, lobbyContainer)

		lobbyContent := page1.MustElement("#lobby-content")
		assert.NotNil(t, lobbyContent)

		// Verify Player 1 is shown
		playerList := page1.MustElement(".player-list")
		assert.Contains(t, playerList.MustText(), "Player1")

		// Player 2 joins
		page2 := browser.MustPage()
		defer page2.MustClose()

		page2.MustNavigate(ts.URL + "/room/" + roomCode)
		page2.MustElement("input[name='name']").MustInput("Player2")
		page2.MustElement("button[type='submit']").MustClick()

		// Wait for SSE update
		time.Sleep(500 * time.Millisecond)

		// Check Player 1's page still has correct DOM structure
		lobbyContent1 := page1.MustElement("#lobby-content")
		assert.NotNil(t, lobbyContent1, "Player 1 should still have #lobby-content after Player 2 joins")

		// Both players should see both names
		playerList1 := page1.MustElement(".player-list")
		assert.Contains(t, playerList1.MustText(), "Player1")
		assert.Contains(t, playerList1.MustText(), "Player2")

		playerList2 := page2.MustElement(".player-list")
		assert.Contains(t, playerList2.MustText(), "Player1")
		assert.Contains(t, playerList2.MustText(), "Player2")

		// Player 3 joins
		page3 := browser.MustPage()
		defer page3.MustClose()

		page3.MustNavigate(ts.URL + "/room/" + roomCode)
		page3.MustElement("input[name='name']").MustInput("Player3")
		page3.MustElement("button[type='submit']").MustClick()

		// Wait for SSE update
		time.Sleep(500 * time.Millisecond)

		// Check all pages still have correct DOM structure
		for i, page := range []*rod.Page{page1, page2, page3} {
			lobbyContent := page.MustElement("#lobby-content")
			assert.NotNil(t, lobbyContent, fmt.Sprintf("Player %d should still have #lobby-content after Player 3 joins", i+1))

			// All players should see all three names
			playerList := page.MustElement(".player-list")
			text := playerList.MustText()
			assert.Contains(t, text, "Player1")
			assert.Contains(t, text, "Player2")
			assert.Contains(t, text, "Player3")
		}

		// Check browser console for errors
		for i, page := range []*rod.Page{page1, page2, page3} {
			// Inject error detection script
			hasErrors, err := page.Eval(`() => {
				// Check for any datastar errors in console
				if (window.datastarErrors && window.datastarErrors.length > 0) {
					return true;
				}
				return false;
			}`)
			require.NoError(t, err)
			hasErrorsBool := hasErrors.Value.Bool()
			assert.False(t, hasErrorsBool, fmt.Sprintf("Player %d browser should not have datastar errors", i+1))
		}
	})
}

// TestLobbyDOMStructurePreservation ensures DOM structure is maintained through SSE updates
func TestLobbyDOMStructurePreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create handler with in-memory store
	h := New(store.NewMemoryStore())

	// Create test server
	router := setupTestRouter(h)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Launch browser
	l := launcher.New().Headless(true)
	browserURL := l.MustLaunch()
	defer l.Kill()

	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()
	defer page.MustClose()

	// Navigate and create room
	page.MustNavigate(ts.URL)
	page.MustElement("input[name='name']").MustInput("TestPlayer")
	page.MustElement("button[type='submit']").MustClick()
	page.MustWaitLoad()

	// Capture initial DOM structure
	initialHTML, err := page.MustElement("#lobby-container").HTML()
	require.NoError(t, err)

	// Verify initial structure has both container and content
	assert.Contains(t, initialHTML, `id="lobby-content"`)
	assert.Contains(t, initialHTML, `data-on-load=`)

	// Trigger an SSE update by having another player join via API
	roomCode := page.MustInfo().URL[len(ts.URL+"/room/"):]

	// Create a second player directly via handler (simulating another browser)
	room, _ := h.store.GetRoom(roomCode)
	room.AddPlayer(&game.Player{ID: "player2", Name: "Player2"})
	h.store.UpdateRoom(room)
	h.eventBus.Publish(Event{
		Type:     "player_joined",
		RoomCode: roomCode,
		Data:     room,
	})

	// Wait for SSE update
	time.Sleep(500 * time.Millisecond)

	// Verify DOM structure is still intact
	updatedHTML, err := page.MustElement("#lobby-container").HTML()
	require.NoError(t, err)

	// Should still have lobby-content div
	assert.Contains(t, updatedHTML, `id="lobby-content"`)
	// Should still have SSE trigger
	assert.Contains(t, updatedHTML, `data-on-load=`)
	// Should show both players
	assert.Contains(t, updatedHTML, "TestPlayer")
	assert.Contains(t, updatedHTML, "Player2")
}
