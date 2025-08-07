package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// IntegrationTestHelper provides utilities for integration testing
type IntegrationTestHelper struct {
	t       *testing.T
	handler *Handler
	router  *chi.Mux
	store   *store.MemoryStore
}

// NewIntegrationTestHelper creates a new integration test helper
func NewIntegrationTestHelper(t *testing.T) *IntegrationTestHelper {
	gameStore := store.NewMemoryStore()
	h := New(gameStore, createMockCardService())

	// Set up router with all routes
	r := chi.NewRouter()
	r.Get("/", h.Home)
	r.Post("/room/new", h.CreateRoom)
	r.Get("/room/{code}", h.JoinRoom)
	r.Post("/room/{code}/start", h.StartGame)
	r.Post("/room/{code}/leave", h.LeaveRoom)
	r.Get("/game/{code}", h.GamePage)
	r.Get("/sse/lobby/{code}", h.StreamLobby)
	r.Get("/sse/game/{code}", h.StreamGame)

	return &IntegrationTestHelper{
		t:       t,
		handler: h,
		router:  r,
		store:   gameStore,
	}
}

// CreateRoom creates a new room and returns the room code
func (h *IntegrationTestHelper) CreateRoom(playerName string) (string, *http.Cookie) {
	form := url.Values{"playerName": {playerName}}
	req := httptest.NewRequest("POST", "/room/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		h.t.Fatalf("Expected redirect, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/room/") {
		h.t.Fatalf("Expected redirect to room, got %s", location)
	}

	roomCode := strings.TrimPrefix(location, "/room/")

	// Get player cookie
	var playerCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "player_"+roomCode {
			playerCookie = cookie
			break
		}
	}

	if playerCookie == nil {
		h.t.Fatal("No player cookie set")
	}

	return roomCode, playerCookie
}

// JoinRoom joins an existing room
func (h *IntegrationTestHelper) JoinRoom(roomCode, playerName string) *http.Cookie {
	req := httptest.NewRequest("GET", "/room/"+roomCode+"?name="+url.QueryEscape(playerName), nil)
	w := httptest.NewRecorder()

	h.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		h.t.Fatalf("Expected OK, got %d", w.Code)
	}

	// Get player cookie
	var playerCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "player_"+roomCode {
			playerCookie = cookie
			break
		}
	}

	if playerCookie == nil {
		h.t.Fatal("No player cookie set")
	}

	return playerCookie
}

// StartGame starts a game in the given room
func (h *IntegrationTestHelper) StartGame(roomCode string, playerCookie *http.Cookie) {
	req := httptest.NewRequest("POST", "/room/"+roomCode+"/start", nil)
	req.AddCookie(playerCookie)
	w := httptest.NewRecorder()

	h.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		h.t.Fatalf("Expected OK, got %d", w.Code)
	}
}

// LeaveRoom leaves a room
func (h *IntegrationTestHelper) LeaveRoom(roomCode string, playerCookie *http.Cookie) {
	req := httptest.NewRequest("POST", "/room/"+roomCode+"/leave", nil)
	req.AddCookie(playerCookie)
	w := httptest.NewRecorder()

	h.router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		h.t.Fatalf("Expected redirect, got %d", w.Code)
	}
}

// SSEClient simulates an SSE client connection
type SSEClient struct {
	t        *testing.T
	ctx      context.Context
	cancel   context.CancelFunc
	events   chan string
	errors   chan error
	response *httptest.ResponseRecorder
}

// NewSSEClient creates a new SSE client for testing
func (h *IntegrationTestHelper) NewSSEClient(path string, playerCookie *http.Cookie) *SSEClient {
	ctx, cancel := context.WithCancel(context.Background())
	client := &SSEClient{
		t:      h.t,
		ctx:    ctx,
		cancel: cancel,
		events: make(chan string, 10),
		errors: make(chan error, 1),
	}

	// Create request with context
	req := httptest.NewRequest("GET", path, nil).WithContext(ctx)
	req.Header.Set("Accept", "text/event-stream")
	if playerCookie != nil {
		req.AddCookie(playerCookie)
	}

	// Create a custom response writer that captures SSE data
	w := &sseResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		client:           client,
	}

	// Start the SSE handler in a goroutine
	go func() {
		h.router.ServeHTTP(w, req)
	}()

	// Give the connection time to establish
	time.Sleep(100 * time.Millisecond)

	return client
}

// Close closes the SSE client
func (c *SSEClient) Close() {
	c.cancel()
}

// WaitForEvent waits for an SSE event with timeout
func (c *SSEClient) WaitForEvent(timeout time.Duration) (string, error) {
	select {
	case event := <-c.events:
		return event, nil
	case err := <-c.errors:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for event")
	}
}

// Custom response writer for SSE testing
type sseResponseWriter struct {
	*httptest.ResponseRecorder
	client *SSEClient
	mu     sync.Mutex
}

func (w *sseResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Capture the data
	n, err := w.ResponseRecorder.Write(data)

	// Parse SSE events - looking specifically for HTML fragments
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: fragments ") {
			// Extract the HTML fragment
			htmlFragment := strings.TrimPrefix(line, "data: fragments ")
			if htmlFragment != "" {
				select {
				case w.client.events <- htmlFragment:
				default:
					// Buffer full, skip
				}
			}
		}
	}

	return n, err
}

func (w *sseResponseWriter) Flush() {
	// SSE requires flushing - no-op for test recorder
}

// Helper to get room from store
func (h *IntegrationTestHelper) GetRoom(code string) *game.Room {
	room, err := h.store.GetRoom(code)
	if err != nil {
		h.t.Fatalf("Failed to get room: %v", err)
	}
	return room
}

// TestFullGameFlow tests a complete game flow
func TestFullGameFlow(t *testing.T) {
	helper := NewIntegrationTestHelper(t)

	// Create room
	roomCode, hostCookie := helper.CreateRoom("Host Player")

	// Start SSE connection for host
	hostSSE := helper.NewSSEClient("/sse/lobby/"+roomCode, hostCookie)
	defer hostSSE.Close()

	// Note: No initial lobby state is sent - SSE only sends updates on events
	// This is intentional as the page already has the correct content

	// Join 3 more players
	var playerCookies []*http.Cookie
	for i := 2; i <= 4; i++ {
		playerName := fmt.Sprintf("Player %d", i)
		cookie := helper.JoinRoom(roomCode, playerName)
		playerCookies = append(playerCookies, cookie)

		// Wait for join event
		event, err := hostSSE.WaitForEvent(2 * time.Second)
		if err != nil {
			t.Fatalf("Failed to get join event: %v", err)
		}
		if !strings.Contains(event, playerName) {
			t.Errorf("Expected %s in lobby update", playerName)
		}
	}

	// Verify room has 4 players
	room := helper.GetRoom(roomCode)
	if len(room.Players) != 4 {
		t.Errorf("Expected 4 players, got %d", len(room.Players))
	}

	// Start game
	helper.StartGame(roomCode, hostCookie)

	// Wait for countdown
	time.Sleep(100 * time.Millisecond)

	// Verify game state changed
	room = helper.GetRoom(roomCode)
	if room.State != game.StateCountdown {
		t.Errorf("Expected countdown state, got %s", room.State)
	}
}

// TestConcurrentJoins tests multiple players joining simultaneously
func TestConcurrentJoins(t *testing.T) {
	helper := NewIntegrationTestHelper(t)

	// Create room
	roomCode, _ := helper.CreateRoom("Host")

	// Join 7 players concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 7)

	for i := 2; i <= 8; i++ {
		wg.Add(1)
		go func(playerNum int) {
			defer wg.Done()

			playerName := fmt.Sprintf("Player %d", playerNum)
			req := httptest.NewRequest("GET", "/room/"+roomCode+"?name="+url.QueryEscape(playerName), nil)
			w := httptest.NewRecorder()

			helper.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("player %d got status %d", playerNum, w.Code)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify room has exactly 8 players (max capacity)
	room := helper.GetRoom(roomCode)
	if len(room.Players) != 8 {
		t.Errorf("Expected 8 players, got %d", len(room.Players))
	}
}

// TestPlayerReconnection tests player reconnection with same session
func TestPlayerReconnection(t *testing.T) {
	helper := NewIntegrationTestHelper(t)

	// Create room and join
	roomCode, hostCookie := helper.CreateRoom("Host")

	// Get player ID from room
	room := helper.GetRoom(roomCode)
	var playerID string
	for id := range room.Players {
		playerID = id
		break
	}

	// Simulate disconnect by leaving
	helper.LeaveRoom(roomCode, hostCookie)

	// Verify player removed
	room = helper.GetRoom(roomCode)
	if len(room.Players) != 0 {
		t.Error("Expected player to be removed")
	}

	// Rejoin with same session cookie
	req := httptest.NewRequest("GET", "/room/"+roomCode+"?name=Host", nil)
	req.AddCookie(hostCookie)
	w := httptest.NewRecorder()

	helper.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected OK on rejoin, got %d", w.Code)
	}

	// Verify player restored with same ID
	room = helper.GetRoom(roomCode)
	if len(room.Players) != 1 {
		t.Error("Expected player to rejoin")
	}

	// Should have same player ID
	for id := range room.Players {
		if id != playerID {
			t.Error("Expected same player ID on reconnection")
		}
	}
}

// TestSSEReconnection tests SSE reconnection behavior
func TestSSEReconnection(t *testing.T) {
	helper := NewIntegrationTestHelper(t)

	// Create room
	roomCode, hostCookie := helper.CreateRoom("Host")

	// Start first SSE connection
	sse1 := helper.NewSSEClient("/sse/lobby/"+roomCode, hostCookie)

	// No initial event is sent for lobby SSE - this is intentional
	// Close first connection
	sse1.Close()

	// Start second SSE connection (reconnection)
	sse2 := helper.NewSSEClient("/sse/lobby/"+roomCode, hostCookie)
	defer sse2.Close()

	// Join another player to trigger an event
	_ = helper.JoinRoom(roomCode, "Player 2")

	// Should receive the join event
	event, err := sse2.WaitForEvent(2 * time.Second)
	if err != nil {
		t.Fatalf("Failed to get join event: %v", err)
	}

	if !strings.Contains(event, "Player 2") {
		t.Error("Expected player join event")
	}
}
