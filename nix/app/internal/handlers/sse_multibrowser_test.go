package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// TestMultiBrowserSSEScenarios tests SSE behavior with multiple concurrent browser connections
func TestMultiBrowserSSEScenarios(t *testing.T) {
	t.Run("multiple browsers receive game start redirect", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create a room and add players
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Simulate 3 browsers joining
		browsers := make([]*browserClient, 3)
		for i := 0; i < 3; i++ {
			playerName := fmt.Sprintf("Player%d", i+1)
			browsers[i] = joinRoomAsPlayer(t, router, roomCode, playerName)
			defer browsers[i].close()
		}

		// Track redirect events
		redirectCount := atomic.Int32{}
		sseCloseCount := atomic.Int32{}

		// All browsers connect to lobby SSE
		var wg sync.WaitGroup

		for idx, browser := range browsers {
			wg.Add(1)
			go func(b *browserClient, browserIdx int) {
				defer wg.Done()

				// Use custom writer to capture SSE data
				captureWriter := &sseCaptureWriter{
					ResponseRecorder: httptest.NewRecorder(),
					data:             &bytes.Buffer{},
				}

				// Connect to SSE
				req := httptest.NewRequest("GET", "/sse/lobby/"+roomCode, nil)
				req.AddCookie(b.playerCookie)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				req = req.WithContext(ctx)

				// Start SSE handler
				done := make(chan bool)
				go func() {
					router.ServeHTTP(captureWriter, req)
					sseCloseCount.Add(1)
					done <- true
				}()

				// Monitor captured data
				ticker := time.NewTicker(50 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-done:
						// SSE handler finished
						data := captureWriter.data.String()
						t.Logf("Browser %d SSE closed. Total data: %d bytes", browserIdx, len(data))
						if len(data) > 0 {
							t.Logf("Browser %d SSE data sample: %.200s", browserIdx, data)
						}

						// Check for redirect in SSE data format
						if strings.Contains(data, "executeScript") || strings.Contains(data, "window.location") {
							redirectCount.Add(1)
							t.Logf("Browser %d found redirect in SSE data", browserIdx)
						}
						return

					case <-ticker.C:
						// Check periodically
						data := captureWriter.data.String()
						if strings.Contains(data, "executeScript") || strings.Contains(data, "window.location") {
							redirectCount.Add(1)
							t.Logf("Browser %d found redirect during monitoring", browserIdx)
							cancel()
							return
						}

					case <-ctx.Done():
						t.Logf("Browser %d SSE timeout", browserIdx)
						return
					}
				}
			}(browser, idx)
		}

		// Give SSE connections time to establish
		time.Sleep(300 * time.Millisecond)

		// First browser starts the game
		t.Log("Starting game...")
		startReq := httptest.NewRequest("POST", "/action/start/"+roomCode, nil)
		startReq.AddCookie(browsers[0].playerCookie)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Wait for all SSE handlers to process
		wg.Wait()

		t.Logf("SSE connections closed: %d", sseCloseCount.Load())
		t.Logf("Redirects received: %d", redirectCount.Load())

		// This test demonstrates the issue - SSE connections close without sending redirect
		if redirectCount.Load() == 0 {
			t.Log("ISSUE REPRODUCED: No browsers received redirect before SSE connection closed")
			t.Log("This confirms the bug where ExecuteScript is called but data is not flushed before handler returns")
		} else if redirectCount.Load() < 3 {
			t.Logf("PARTIAL ISSUE: Only %d of 3 browsers received redirect", redirectCount.Load())
		}
	})

	t.Run("all browsers receive countdown updates after redirect", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create room and add players
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Join 3 players
		browsers := make([]*browserClient, 3)
		for i := 0; i < 3; i++ {
			playerName := fmt.Sprintf("Player%d", i+1)
			browsers[i] = joinRoomAsPlayer(t, router, roomCode, playerName)
			defer browsers[i].close()
		}

		// Start the game
		startReq := httptest.NewRequest("POST", "/action/start/"+roomCode, nil)
		startReq.AddCookie(browsers[0].playerCookie)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// All browsers connect to game SSE
		var wg sync.WaitGroup
		countdownEvents := make([][]string, 3)

		for idx, browser := range browsers {
			wg.Add(1)
			go func(b *browserClient, browserIdx int) {
				defer wg.Done()

				// Connect to game SSE
				req := httptest.NewRequest("GET", "/sse/game/"+roomCode, nil)
				req.AddCookie(b.playerCookie)
				w := httptest.NewRecorder()

				ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
				defer cancel()
				req = req.WithContext(ctx)

				// Start SSE in goroutine
				done := make(chan bool)
				go func() {
					router.ServeHTTP(w, req)
					done <- true
				}()

				// Monitor for countdown updates
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						body := w.Body.String()
						// Look for countdown numbers in the response
						for i := 1; i <= 5; i++ {
							marker := fmt.Sprintf("Revealing roles in %d...", i)
							if strings.Contains(body, marker) {
								// Check if we already recorded this
								found := false
								for _, event := range countdownEvents[browserIdx] {
									if event == fmt.Sprintf("countdown-%d", i) {
										found = true
										break
									}
								}
								if !found {
									countdownEvents[browserIdx] = append(countdownEvents[browserIdx], fmt.Sprintf("countdown-%d", i))
								}
							}
						}
					case <-ctx.Done():
						return
					}
				}
			}(browser, idx)
		}

		// Wait for countdown to complete
		wg.Wait()

		// Verify all browsers received countdown events
		for i, events := range countdownEvents {
			t.Logf("Browser %d received %d countdown events", i+1, len(events))
			if len(events) < 3 {
				t.Errorf("Browser %d only received %d countdown events, expected at least 3", i+1, len(events))
			}
		}
	})

	t.Run("late-joining browser during countdown", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create room and add initial players
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Join 2 players initially
		browser1 := joinRoomAsPlayer(t, router, roomCode, "Player1")
		browser2 := joinRoomAsPlayer(t, router, roomCode, "Player2")
		defer browser1.close()
		defer browser2.close()

		// Start the game
		startReq := httptest.NewRequest("POST", "/action/start/"+roomCode, nil)
		startReq.AddCookie(browser1.playerCookie)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Wait for countdown to begin
		time.Sleep(2 * time.Second)

		// Late browser tries to join during countdown
		joinReq := httptest.NewRequest("GET", "/room/"+roomCode+"?name=LatePlayer", nil)
		joinW := httptest.NewRecorder()
		router.ServeHTTP(joinW, joinReq)

		// Verify late join is rejected
		if joinW.Code != http.StatusBadRequest {
			t.Errorf("expected late join to be rejected with 400, got %d", joinW.Code)
		}

		body := joinW.Body.String()
		if !strings.Contains(body, "Game already started") {
			t.Error("expected 'Game already started' message for late join")
		}
	})

	t.Run("browser reconnection during game", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create room and join players
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		browser1 := joinRoomAsPlayer(t, router, roomCode, "Player1")
		browser2 := joinRoomAsPlayer(t, router, roomCode, "Player2")
		defer browser1.close()
		defer browser2.close()

		// Start the game
		startReq := httptest.NewRequest("POST", "/action/start/"+roomCode, nil)
		startReq.AddCookie(browser1.playerCookie)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Browser1 connects to game SSE
		req1 := httptest.NewRequest("GET", "/sse/game/"+roomCode, nil)
		req1.AddCookie(browser1.playerCookie)
		w1 := httptest.NewRecorder()

		ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel1()
		req1 = req1.WithContext(ctx1)

		// Start first connection
		done1 := make(chan bool)
		go func() {
			router.ServeHTTP(w1, req1)
			done1 <- true
		}()

		// Wait for initial connection
		time.Sleep(500 * time.Millisecond)

		// Disconnect (simulate network interruption)
		cancel1()
		<-done1

		// Reconnect after 1 second
		time.Sleep(1 * time.Second)

		req2 := httptest.NewRequest("GET", "/sse/game/"+roomCode, nil)
		req2.AddCookie(browser1.playerCookie)
		w2 := httptest.NewRecorder()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		req2 = req2.WithContext(ctx2)

		done2 := make(chan bool)
		go func() {
			router.ServeHTTP(w2, req2)
			done2 <- true
		}()

		// Wait a bit
		time.Sleep(500 * time.Millisecond)
		cancel2()
		<-done2

		// Verify reconnection received game state
		body := w2.Body.String()
		if !strings.Contains(body, "game-container") {
			t.Error("reconnected browser did not receive game state")
		}
	})

	t.Run("concurrent SSE connections stress test", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Join max players (8)
		browsers := make([]*browserClient, 8)
		for i := 0; i < 8; i++ {
			playerName := fmt.Sprintf("Player%d", i+1)
			browsers[i] = joinRoomAsPlayer(t, router, roomCode, playerName)
			defer browsers[i].close()
		}

		// All browsers connect to SSE simultaneously
		var wg sync.WaitGroup
		successCount := atomic.Int32{}

		for _, browser := range browsers {
			wg.Add(1)
			go func(b *browserClient) {
				defer wg.Done()

				req := httptest.NewRequest("GET", "/sse/lobby/"+roomCode, nil)
				req.AddCookie(b.playerCookie)
				w := httptest.NewRecorder()

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				req = req.WithContext(ctx)

				done := make(chan bool)
				go func() {
					router.ServeHTTP(w, req)
					done <- true
				}()

				select {
				case <-done:
					// Check if we got valid SSE response
					if w.Code == http.StatusOK {
						successCount.Add(1)
					}
				case <-ctx.Done():
					successCount.Add(1) // Timeout is ok, connection was established
				}
			}(browser)
		}

		wg.Wait()

		// All connections should succeed
		if successCount.Load() != 8 {
			t.Errorf("expected 8 successful SSE connections, got %d", successCount.Load())
		}
	})

	t.Run("event ordering across multiple browsers", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Join 2 players
		browser1 := joinRoomAsPlayer(t, router, roomCode, "Player1")
		browser2 := joinRoomAsPlayer(t, router, roomCode, "Player2")
		defer browser1.close()
		defer browser2.close()

		// Both connect to lobby SSE
		events1 := make(chan string, 10)
		events2 := make(chan string, 10)

		// Browser 1 SSE
		go func() {
			req := httptest.NewRequest("GET", "/sse/lobby/"+roomCode, nil)
			req.AddCookie(browser1.playerCookie)
			w := httptest.NewRecorder()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			req = req.WithContext(ctx)

			done := make(chan bool)
			go func() {
				router.ServeHTTP(w, req)
				done <- true
			}()

			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			lastLen := 0
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					body := w.Body.String()
					if len(body) > lastLen {
						newContent := body[lastLen:]
						if strings.Contains(newContent, "Player3") {
							events1 <- "player3-joined"
						}
						if strings.Contains(newContent, "window.location.href") {
							events1 <- "redirect"
							cancel()
							return
						}
						lastLen = len(body)
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		// Browser 2 SSE
		go func() {
			req := httptest.NewRequest("GET", "/sse/lobby/"+roomCode, nil)
			req.AddCookie(browser2.playerCookie)
			w := httptest.NewRecorder()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			req = req.WithContext(ctx)

			done := make(chan bool)
			go func() {
				router.ServeHTTP(w, req)
				done <- true
			}()

			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			lastLen := 0
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					body := w.Body.String()
					if len(body) > lastLen {
						newContent := body[lastLen:]
						if strings.Contains(newContent, "Player3") {
							events2 <- "player3-joined"
						}
						if strings.Contains(newContent, "window.location.href") {
							events2 <- "redirect"
							cancel()
							return
						}
						lastLen = len(body)
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		// Wait for SSE to establish
		time.Sleep(200 * time.Millisecond)

		// Player 3 joins
		browser3 := joinRoomAsPlayer(t, router, roomCode, "Player3")
		defer browser3.close()

		// Start game
		time.Sleep(200 * time.Millisecond)
		startReq := httptest.NewRequest("POST", "/action/start/"+roomCode, nil)
		startReq.AddCookie(browser1.playerCookie)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Collect events with timeout
		timeout := time.After(3 * time.Second)
		browser1Events := []string{}
		browser2Events := []string{}

		collecting := true
		for collecting {
			select {
			case e := <-events1:
				browser1Events = append(browser1Events, e)
			case e := <-events2:
				browser2Events = append(browser2Events, e)
			case <-timeout:
				collecting = false
			}
		}

		// Both browsers should see player join and redirect in same order
		t.Logf("Browser1 events: %v", browser1Events)
		t.Logf("Browser2 events: %v", browser2Events)

		// Verify both got the events
		if len(browser1Events) < 2 {
			t.Errorf("Browser1 missing events, got %v", browser1Events)
		}
		if len(browser2Events) < 2 {
			t.Errorf("Browser2 missing events, got %v", browser2Events)
		}
	})

	t.Run("TestMultipleBrowsersReceiveGameStartRedirect", func(t *testing.T) {
		// Setup
		h := New(store.NewMemoryStore(), createMockCardService())
		router := setupSSETestRouter(h)

		// Create room
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Simulate 3 browsers joining
		browsers := make([]*browserClient, 3)
		for i := 0; i < 3; i++ {
			playerName := fmt.Sprintf("Player%d", i+1)
			browsers[i] = joinRoomAsPlayer(t, router, roomCode, playerName)
			defer browsers[i].close()
		}

		// Track results for each browser
		type browserResult struct {
			gotRedirect       bool
			redirectData      string
			connectionClosed  bool
			eventsBeforeClose []string
		}
		results := make([]browserResult, 3)
		var wg sync.WaitGroup

		// All browsers connect to lobby SSE
		for idx, browser := range browsers {
			wg.Add(1)
			go func(b *browserClient, browserIdx int) {
				defer wg.Done()

				// Use custom writer to capture SSE data
				captureWriter := &sseCaptureWriter{
					ResponseRecorder: httptest.NewRecorder(),
					data:             &bytes.Buffer{},
				}

				// Connect to SSE
				req := httptest.NewRequest("GET", "/sse/lobby/"+roomCode, nil)
				req.AddCookie(b.playerCookie)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				req = req.WithContext(ctx)

				// Start SSE handler
				done := make(chan bool)
				go func() {
					router.ServeHTTP(captureWriter, req)
					results[browserIdx].connectionClosed = true
					done <- true
				}()

				// Monitor captured data
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-done:
						// SSE handler finished - capture final data
						captureWriter.mu.Lock()
						data := captureWriter.data.String()
						captureWriter.mu.Unlock()

						// Check for redirect
						if strings.Contains(data, "datastar-execute-script") && strings.Contains(data, "window.location.href") {
							results[browserIdx].gotRedirect = true
							results[browserIdx].redirectData = data
						}

						// Log all data for debugging
						t.Logf("Browser %d SSE closed. Got redirect: %v, Total data length: %d",
							browserIdx, results[browserIdx].gotRedirect, len(data))

						// Log the actual data for debugging
						if len(data) > 0 {
							t.Logf("Browser %d raw SSE data: %q", browserIdx, data)
						}
						return

					case <-ticker.C:
						// Check periodically for events
						captureWriter.mu.Lock()
						data := captureWriter.data.String()
						captureWriter.mu.Unlock()

						// Track any game events we see
						if strings.Contains(data, "game_started") {
							results[browserIdx].eventsBeforeClose = append(results[browserIdx].eventsBeforeClose, "game_started")
						}
						if strings.Contains(data, "countdown_update") {
							results[browserIdx].eventsBeforeClose = append(results[browserIdx].eventsBeforeClose, "countdown_update")
						}

					case <-ctx.Done():
						t.Errorf("Browser %d SSE timed out without closing", browserIdx)
						return
					}
				}
			}(browser, idx)
		}

		// Wait for SSE connections to establish
		time.Sleep(100 * time.Millisecond)

		// Start the game from browser 0
		startReq := httptest.NewRequest("POST", "/action/start/"+roomCode, nil)
		startReq.AddCookie(browsers[0].playerCookie)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Wait for all SSE handlers to complete
		wg.Wait()

		// Verify results
		for i, result := range results {
			t.Logf("Browser %d results: gotRedirect=%v, connectionClosed=%v, events=%v",
				i, result.gotRedirect, result.connectionClosed, result.eventsBeforeClose)

			// All browsers must receive redirect
			if !result.gotRedirect {
				t.Errorf("Browser %d did not receive redirect script", i)
			}

			// All connections must close after redirect
			if !result.connectionClosed {
				t.Errorf("Browser %d SSE connection did not close", i)
			}

			// Verify redirect contains correct URL
			if result.gotRedirect && !strings.Contains(result.redirectData, "/game/"+roomCode) {
				t.Errorf("Browser %d redirect URL incorrect: %s", i, result.redirectData)
			}
		}

		// Additional verification: All browsers should have gotten redirects
		redirectCount := 0
		for _, result := range results {
			if result.gotRedirect {
				redirectCount++
			}
		}

		if redirectCount != 3 {
			t.Errorf("Expected all 3 browsers to receive redirects, but only %d did", redirectCount)
		}
	})
}

// browserClient simulates a browser with cookies
type browserClient struct {
	sessionCookie *http.Cookie
	playerCookie  *http.Cookie
	roomCode      string
}

func (b *browserClient) close() {
	// Cleanup if needed
}

// joinRoomAsPlayer simulates a browser joining a room
func joinRoomAsPlayer(t *testing.T, router *chi.Mux, roomCode, playerName string) *browserClient {
	req := httptest.NewRequest("GET", "/room/"+roomCode+"?name="+playerName, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to join room as %s: status %d", playerName, w.Code)
	}

	client := &browserClient{roomCode: roomCode}

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "session" {
			client.sessionCookie = cookie
		} else if cookie.Name == "player_"+roomCode {
			client.playerCookie = cookie
		}
	}

	if client.playerCookie == nil {
		t.Fatalf("No player cookie received for %s", playerName)
	}

	return client
}

// setupSSETestRouter creates a test router with SSE routes
func setupSSETestRouter(h *Handler) *chi.Mux {
	router := chi.NewRouter()

	// Page routes
	router.Get("/room/{code}", h.JoinRoom)
	router.Get("/game/{code}", h.GamePage)

	// SSE routes
	router.Get("/sse/lobby/{code}", h.StreamLobby)
	router.Get("/sse/game/{code}", h.StreamGame)

	// Action routes
	router.Post("/action/start/{code}", h.StartGame)
	router.Post("/action/leave/{code}", h.LeaveRoom)

	return router
}

// sseCaptureWriter captures SSE output
type sseCaptureWriter struct {
	*httptest.ResponseRecorder
	data *bytes.Buffer
	mu   sync.Mutex
}

func (w *sseCaptureWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Capture to our buffer
	n, err := w.data.Write(b)

	// Also write to ResponseRecorder
	w.ResponseRecorder.Write(b)

	return n, err
}

func (w *sseCaptureWriter) Flush() {
	// SSE requires flushing - no-op for test
}

func init() {
	// Ensure consistent timing in tests
	time.Local = time.UTC
}
