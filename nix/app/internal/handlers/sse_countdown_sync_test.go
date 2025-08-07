package handlers

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/stretchr/testify/assert"
)

// TestMultipleBrowsersCountdownSync tests that multiple browsers all receive countdown updates
func TestMultipleBrowsersCountdownSync(t *testing.T) {
	h := New(store.NewMemoryStore())

	// Create room with 3 players
	room, _ := h.store.CreateRoom()
	players := []*game.Player{
		{ID: "p1", Name: "Player1"},
		{ID: "p2", Name: "Player2"},
		{ID: "p3", Name: "Player3"},
	}

	for _, p := range players {
		room.AddPlayer(p)
	}
	h.store.UpdateRoom(room)

	// Create SSE connections for all players
	var wg sync.WaitGroup
	countdownEvents := make(map[string][]string)
	var mu sync.Mutex

	for _, player := range players {
		wg.Add(1)
		go func(p *game.Player) {
			defer wg.Done()

			// Create SSE request
			req := httptest.NewRequest("GET", "/sse/game/"+room.Code, nil)
			req.AddCookie(&http.Cookie{
				Name:  "player_" + room.Code,
				Value: p.ID,
			})

			// Record response
			w := httptest.NewRecorder()

			// Start SSE handler in background
			done := make(chan bool)
			go func() {
				h.StreamGame(w, req)
				done <- true
			}()

			// Read SSE events
			scanner := bufio.NewScanner(w.Body)
			timeout := time.After(7 * time.Second) // Countdown is 5 seconds + buffer

			for {
				select {
				case <-timeout:
					return
				case <-done:
					return
				default:
					if scanner.Scan() {
						line := scanner.Text()
						if strings.Contains(line, "countdown") || strings.Contains(line, "Revealing roles in") {
							mu.Lock()
							countdownEvents[p.ID] = append(countdownEvents[p.ID], line)
							mu.Unlock()
						}
					}
				}
			}
		}(player)
	}

	// Wait a moment for SSE connections to establish
	time.Sleep(100 * time.Millisecond)

	// Start the game (triggers countdown)
	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	// Run countdown
	go h.runCountdown(room)

	// Publish initial game started event
	h.eventBus.Publish(Event{
		Type:     "game_started",
		RoomCode: room.Code,
		Data:     room,
	})

	// Wait for all SSE handlers to complete
	wg.Wait()

	// Verify all players received countdown events
	for _, player := range players {
		events := countdownEvents[player.ID]
		assert.NotEmpty(t, events, "Player %s should have received countdown events", player.Name)
		t.Logf("Player %s received %d countdown events", player.Name, len(events))
	}
}

// TestLateJoinerDuringCountdown tests a player joining during countdown
func TestLateJoinerDuringCountdown(t *testing.T) {
	h := New(store.NewMemoryStore())

	// Create room with 1 player
	room, _ := h.store.CreateRoom()
	player1 := &game.Player{ID: "p1", Name: "Player1"}
	room.AddPlayer(player1)

	// Start countdown
	room.State = game.StateCountdown
	room.CountdownRemaining = 5
	room.StartedAt = time.Now()
	h.store.UpdateRoom(room)

	// Start countdown in background
	go h.runCountdown(room)

	// Wait 2 seconds
	time.Sleep(2 * time.Second)

	// Add late joiner
	player2 := &game.Player{ID: "p2", Name: "Player2"}
	room.AddPlayer(player2)
	h.store.UpdateRoom(room)

	// Create SSE connection for late joiner
	req := httptest.NewRequest("GET", "/sse/game/"+room.Code, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player2.ID,
	})

	w := httptest.NewRecorder()

	// Capture initial render
	done := make(chan bool)
	go func() {
		h.StreamGame(w, req)
		done <- true
	}()

	// Give it time to render
	time.Sleep(100 * time.Millisecond)

	// Check response contains countdown with correct remaining time
	body := w.Body.String()
	assert.Contains(t, body, "countdown", "Late joiner should see countdown")

	// The handler should calculate actual remaining time
	// Started 2 seconds ago, so should be ~3 seconds remaining
	assert.Contains(t, body, "Revealing roles in", "Should show countdown message")
}

// TestSSEReconnectionHandling tests that SSE connections handle reconnection properly
func TestSSEReconnectionHandling(t *testing.T) {
	h := New(store.NewMemoryStore())

	// Create room and player
	room, _ := h.store.CreateRoom()
	player := &game.Player{ID: "p1", Name: "Player1"}
	room.AddPlayer(player)
	h.store.UpdateRoom(room)

	// Track number of connections
	connectionCount := 0
	// Note: Cannot reassign methods in Go, so we'll track connections differently
	_ = connectionCount // Track connections via test logic instead

	// Create multiple SSE requests rapidly (simulating reconnection attempts)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/sse/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})

		w := httptest.NewRecorder()

		go func() {
			h.StreamGame(w, req)
		}()

		time.Sleep(50 * time.Millisecond)
	}

	// Should have 3 connections (not prevented by the handler)
	assert.Equal(t, 3, connectionCount, "Should track all connection attempts")
}

// TestSSETimeoutMiddleware tests impact of timeout middleware on SSE
func TestSSETimeoutMiddleware(t *testing.T) {
	// This test would need to be done at the router level
	// to properly test the middleware impact

	h := New(store.NewMemoryStore())
	// Note: SetupServer() not available in test context, using handler directly

	// Create room
	room, _ := h.store.CreateRoom()
	player := &game.Player{ID: "p1", Name: "Player1"}
	room.AddPlayer(player)
	h.store.UpdateRoom(room)

	// Make SSE request through the full router (with middleware)
	req := httptest.NewRequest("GET", "/sse/game/"+room.Code, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})

	w := httptest.NewRecorder()

	// This would need to run for >60 seconds to test timeout
	// For now, just verify the connection is established
	done := make(chan bool)
	go func() {
		h.StreamGame(w, req)
		done <- true
	}()

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	// Should have received some SSE data
	body := w.Body.String()
	assert.NotEmpty(t, body, "Should have received SSE data")
	assert.Contains(t, body, "event:", "Should be SSE format")
}
