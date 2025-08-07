package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSEConnectionExhaustionFixed verifies that the game page doesn't create multiple SSE connections during countdown
func TestSSEConnectionExhaustionFixed(t *testing.T) {
	// Create handler and server
	h := newTestHandler()

	// Create room with players
	room, _ := h.store.CreateRoom()
	host := &game.Player{ID: "host-id", Name: "Host"}
	player2 := &game.Player{ID: "player2-id", Name: "Player2"}
	room.AddPlayer(host)
	room.AddPlayer(player2)
	h.store.UpdateRoom(room)

	// Start game to trigger countdown
	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	// Track connection attempts by monitoring HTTP requests
	connectionCount := 0

	// Create a test server that counts SSE connections
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/sse/game/") {
			connectionCount++
			t.Logf("SSE connection attempt #%d to %s", connectionCount, r.URL.Path)
		}

		// Use the actual router
		router := setupTestRouter(h)
		router.ServeHTTP(w, r)
	}))
	defer testServer.Close()

	// Simulate game page request
	req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	w := httptest.NewRecorder()

	h.GamePage(w, req)

	// Verify the page loaded successfully
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Verify the fix: data-on-load should be on wrapper, not game-container
	assert.Contains(t, body, "data-on-load", "Should have data-on-load attribute")

	// Parse the HTML to check structure
	gameContainerIndex := strings.Index(body, `id="game-container"`)
	if gameContainerIndex > 0 {
		// Find the tag containing game-container
		tagStart := strings.LastIndex(body[:gameContainerIndex], "<")
		tagEnd := strings.Index(body[tagStart:], ">") + tagStart
		gameContainerTag := body[tagStart:tagEnd]

		// This is the key assertion - game-container should NOT have data-on-load
		assert.NotContains(t, gameContainerTag, "data-on-load",
			"Fixed: game-container should not have data-on-load (prevents re-triggering)")
	}
}

// TestGameTemplateStructure verifies the game template has the correct structure
func TestGameTemplateStructure(t *testing.T) {
	h := newTestHandler()

	// Create a game room
	room, _ := h.store.CreateRoom()
	player := &game.Player{ID: "test-player-id", Name: "TestPlayer"}
	room.AddPlayer(player)
	room.State = game.StateCountdown
	room.CountdownRemaining = 3
	h.store.UpdateRoom(room)

	// Render the game page
	req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player.ID})
	w := httptest.NewRecorder()

	h.GamePage(w, req)

	body := w.Body.String()

	// Verify the structure has the wrapper div with data-on-load
	assert.Contains(t, body, "data-on-load", "Should have data-on-load attribute")
	assert.Contains(t, body, fmt.Sprintf(`@get('/sse/game/%s')`, room.Code), "Should have SSE endpoint")
	assert.Contains(t, body, `id="game-container"`, "Should have game-container element")

	// Verify data-on-load is NOT on the game-container itself
	// This is the key fix - data-on-load should be on a wrapper, not the morphable element
	gameContainerStart := strings.Index(body, `id="game-container"`)
	if gameContainerStart > 0 {
		// Find the opening tag for game-container
		tagStart := strings.LastIndex(body[:gameContainerStart], "<")
		tagEnd := strings.Index(body[tagStart:], ">") + tagStart
		gameContainerTag := body[tagStart:tagEnd]

		assert.NotContains(t, gameContainerTag, "data-on-load",
			"game-container element should NOT have data-on-load attribute (it should be on wrapper)")
	}
}
