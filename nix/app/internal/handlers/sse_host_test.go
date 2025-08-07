package handlers

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"
)

// TestStreamHost tests the host SSE endpoint
func TestStreamHost(t *testing.T) {
	// Create test handler
	cfg := config.DefaultConfig()
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg)

	// Create a room with host
	room, err := gameStore.CreateRoom()
	require.NoError(t, err)

	host := game.NewPlayer("host-123", "Host", "session-123")
	host.IsHost = true
	room.AddPlayer(host)
	gameStore.UpdateRoom(room)

	// Create request with proper path params
	req := httptest.NewRequest("GET", "/sse/host/"+room.Code, nil)

	// Add cookies
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	req.AddCookie(&http.Cookie{Name: "host_" + room.Code, Value: "true"})

	// Add chi context with URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a channel to capture SSE messages
	sseMessages := make(chan string, 10)
	done := make(chan bool)

	// Start the SSE handler in a goroutine
	go func() {
		h.StreamHost(w, req)
		done <- true
	}()

	// Give it time to send initial messages
	time.Sleep(100 * time.Millisecond)

	// Read the response
	go func() {
		body := w.Body.String()
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "data:") {
				sseMessages <- line
			}
		}
		close(sseMessages)
	}()

	// Wait a bit for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Check we got messages
	messageCount := 0
	hasQRCode := false
	hasFragment := false

	for msg := range sseMessages {
		messageCount++
		if strings.Contains(msg, "qrCode") {
			hasQRCode = true
		}
		if strings.Contains(msg, "host-dashboard-container") {
			hasFragment = true
		}
		if messageCount >= 10 { // Prevent infinite loop
			break
		}
	}

	// Verify we got both QR code signal and dashboard fragment
	assert.True(t, hasQRCode, "Should have received QR code signal")
	assert.True(t, hasFragment, "Should have received dashboard fragment")
}

// TestStreamHostUnauthorized tests unauthorized access to host SSE
func TestStreamHostUnauthorized(t *testing.T) {
	// Create test handler
	cfg := config.DefaultConfig()
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg)

	// Create a room with regular player
	room, err := gameStore.CreateRoom()
	require.NoError(t, err)

	player := game.NewPlayer("player-123", "Player", "session-123")
	room.AddPlayer(player)
	gameStore.UpdateRoom(room)

	tests := []struct {
		name        string
		cookies     map[string]string
		expectError bool
	}{
		{
			name:        "No cookies",
			cookies:     map[string]string{},
			expectError: true,
		},
		{
			name: "No host cookie",
			cookies: map[string]string{
				"player_" + room.Code: player.ID,
			},
			expectError: true,
		},
		{
			name: "Invalid host cookie",
			cookies: map[string]string{
				"player_" + room.Code: player.ID,
				"host_" + room.Code:   "false",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/sse/host/"+room.Code, nil)

			// Add cookies
			for name, value := range tt.cookies {
				req.AddCookie(&http.Cookie{Name: name, Value: value})
			}

			// Add chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", room.Code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			h.StreamHost(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusUnauthorized, w.Code)
			}
		})
	}
}

// TestStreamHostPlayerUpdates tests that host SSE receives player join/leave events
func TestStreamHostPlayerUpdates(t *testing.T) {
	// Create test handler
	cfg := config.DefaultConfig()
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg)

	// Create a room with host
	room, err := gameStore.CreateRoom()
	require.NoError(t, err)

	host := game.NewPlayer("host-123", "Host", "session-123")
	host.IsHost = true
	room.AddPlayer(host)
	gameStore.UpdateRoom(room)

	// Create request
	req := httptest.NewRequest("GET", "/sse/host/"+room.Code, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	req.AddCookie(&http.Cookie{Name: "host_" + room.Code, Value: "true"})

	// Add chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create a context we can cancel
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Start SSE handler
	sseStarted := make(chan bool)
	go func() {
		sseStarted <- true
		h.StreamHost(w, req)
	}()

	// Wait for SSE to start
	<-sseStarted
	time.Sleep(100 * time.Millisecond)

	// Add a new player
	newPlayer := game.NewPlayer("player-456", "New Player", "session-456")
	room.AddPlayer(newPlayer)
	gameStore.UpdateRoom(room)

	// Publish player joined event
	h.eventBus.Publish(Event{
		Type:     "player_joined",
		RoomCode: room.Code,
		Data:     room,
	})

	// Give time for event to be processed
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop the SSE handler
	cancel()
	time.Sleep(100 * time.Millisecond)

	// Check the response
	body := w.Body.String()

	// Should contain player updates
	assert.Contains(t, body, "New Player", "Response should contain new player name")
	assert.Contains(t, body, "1 connected", "Response should show 1 non-host player connected")
}

// TestQRCodeGeneration tests QR code generation
func TestQRCodeGeneration(t *testing.T) {
	url := "http://example.com/room/ABC123"

	qrCode, err := generateQRCode(url)

	// Should succeed
	require.NoError(t, err, "QR code generation should not fail")
	assert.NotEmpty(t, qrCode, "QR code should not be empty")

	// Base64 string should be valid
	// Try to decode it to verify it's valid base64
	decoded, err := base64.StdEncoding.DecodeString(qrCode)
	require.NoError(t, err, "Should be valid base64")
	assert.True(t, len(decoded) > 0, "Decoded data should not be empty")

	// Check if it's a valid PNG by checking PNG header
	// PNG files start with: 137 80 78 71 13 10 26 10
	pngHeader := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	assert.True(t, len(decoded) >= len(pngHeader), "Decoded data should be long enough for PNG header")
	assert.Equal(t, pngHeader, decoded[:len(pngHeader)], "Should have valid PNG header")
}

// TestGetBaseURL tests base URL construction
func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func(*http.Request)
		expected string
	}{
		{
			name: "HTTP request",
			setupReq: func(r *http.Request) {
				r.Host = "example.com"
			},
			expected: "http://example.com",
		},
		{
			name: "HTTPS request",
			setupReq: func(r *http.Request) {
				r.Host = "example.com"
				r.TLS = &tls.ConnectionState{}
			},
			expected: "https://example.com",
		},
		{
			name: "With X-Forwarded-Proto",
			setupReq: func(r *http.Request) {
				r.Host = "example.com"
				r.Header.Set("X-Forwarded-Proto", "https")
			},
			expected: "https://example.com",
		},
		{
			name: "With X-Forwarded-Host",
			setupReq: func(r *http.Request) {
				r.Host = "internal.com"
				r.Header.Set("X-Forwarded-Host", "external.com")
			},
			expected: "http://external.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupReq(req)

			result := getBaseURL(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMarshalAndMergeSignals tests the signal sending for QR code
func TestMarshalAndMergeSignals(t *testing.T) {
	// This test verifies that the signal format is correct
	signals := map[string]string{
		"qrCode": "data:image/png;base64,iVBORw0KGgoAAAANS...",
	}

	// Marshal to JSON to verify format
	data, err := json.Marshal(signals)
	require.NoError(t, err)

	// Should be valid JSON
	var parsed map[string]string
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, signals["qrCode"], parsed["qrCode"])
}
