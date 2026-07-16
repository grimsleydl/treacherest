package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar-go/datastar"
)

// streamGameBackupFixture builds a room with a recipient and a second player
// whose role is hidden, then returns the recipient's game SSE body.
func streamGameBackupFixture(t *testing.T, h *Handler) (string, *game.Player) {
	t.Helper()

	room, err := h.store.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	recipient := game.NewPlayer("p1", "Recipient", "session-1")
	recipient.Role = mockGuardianCard()
	if err := room.AddPlayer(recipient); err != nil {
		t.Fatalf("add recipient: %v", err)
	}
	hiddenPlayer := game.NewPlayer("p2", "Hidden Player", "session-2")
	hiddenPlayer.Role = &game.Card{ID: 90210, Name: "Other Player Secret Role"}
	hiddenPlayer.RoleRevealed = false
	hiddenPlayer.FaceUp = false
	if err := room.AddPlayer(hiddenPlayer); err != nil {
		t.Fatalf("add hidden player: %v", err)
	}
	room.State = game.StatePlaying
	h.store.UpdateRoom(room)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/sse/game/"+room.Code, nil).WithContext(ctx)
	req.Header.Set("Accept", "text/event-stream")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: recipient.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	done := make(chan bool)
	go func() {
		h.StreamGame(w, req)
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StreamGame did not finish in time")
	}

	body := w.Body.String()
	if !strings.Contains(body, "game-container") {
		t.Fatalf("expected initial game render on stream, got %q", body)
	}
	return body, hiddenPlayer
}

func TestHandler_StreamGameDoesNotExposeHiddenRolesInPlaintextBackup(t *testing.T) {
	h := newTestHandler()
	backupService, err := game.NewBackupService("", false)
	if err != nil {
		t.Fatalf("new backup service: %v", err)
	}
	h.backupService = backupService

	body, hiddenPlayer := streamGameBackupFixture(t, h)
	if strings.Contains(body, `"_stateBackup"`) {
		t.Fatalf("plaintext state backup signal must not be delivered, got %q", body)
	}
	if strings.Contains(body, hiddenPlayer.Role.Name) {
		t.Fatalf("game stream exposed another player's unrevealed role name %q", hiddenPlayer.Role.Name)
	}
	if strings.Contains(body, `\"id\":90210`) {
		t.Fatal("game stream exposed another player's unrevealed role ID")
	}
}

func TestHandler_StreamGameEmitsEncryptedStateBackup(t *testing.T) {
	h := newTestHandler()
	backupService, err := game.NewBackupService(strings.Repeat("ab", 32), true)
	if err != nil {
		t.Fatalf("new backup service: %v", err)
	}
	h.backupService = backupService

	body, hiddenPlayer := streamGameBackupFixture(t, h)
	if !strings.Contains(body, `"_stateBackup"`) {
		t.Fatalf("expected encrypted state backup signal on game stream, got %q", body)
	}
	if strings.Contains(body, hiddenPlayer.Role.Name) {
		t.Fatalf("encrypted backup stream exposed another player's unrevealed role name %q", hiddenPlayer.Role.Name)
	}
}

func TestGameSyncPillState(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		lastSeen time.Time
		want     string
	}{
		{name: "recent heartbeat is live", lastSeen: now.Add(-5 * time.Second), want: "live"},
		{name: "delayed heartbeat is reconnecting", lastSeen: now.Add(-25 * time.Second), want: "reconnecting"},
		{name: "old heartbeat is stale", lastSeen: now.Add(-50 * time.Second), want: "stale"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gameSyncPillState(now, tt.lastSeen)
			if got != tt.want {
				t.Fatalf("gameSyncPillState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandler_StreamLobby(t *testing.T) {
	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := newTestHandler()

		req := httptest.NewRequest("GET", "/room/XXXXX/lobby", nil)
		req.Header.Set("Accept", "text/event-stream")

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "XXXXX")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StreamLobby(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when no player cookie", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req.Header.Set("Accept", "text/event-stream")

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StreamLobby(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when player not in room", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "nonexistent-player",
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StreamLobby(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("streams lobby updates successfully", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)

		// Create a context with timeout to prevent hanging
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Run in goroutine since it blocks
		done := make(chan bool)
		go func() {
			h.StreamLobby(w, req)
			done <- true
		}()

		// Wait for handler to finish or timeout
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("StreamLobby did not finish in time")
		}

		// Check response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify SSE headers
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			t.Errorf("expected SSE content type, got %s", contentType)
		}

		body := w.Body.String()
		if !strings.Contains(body, `"canStartGame"`) {
			t.Fatalf("expected lobby SSE to send validation state, got %q", body)
		}
		if strings.Contains(body, `"qrCode"`) || strings.Contains(body, "data:image/png;base64,") {
			t.Fatalf("expected lobby QR to use the static image endpoint instead of SSE signals, got %q", body)
		}
	})

	t.Run("sends game_started event", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/lobby", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Run handler in goroutine
		go func() {
			h.StreamLobby(w, req)
		}()

		// Give handler time to subscribe
		time.Sleep(50 * time.Millisecond)

		// Send game_started event
		h.eventBus.Publish(Event{
			Type:     "game_started",
			RoomCode: room.Code,
		})

		// Wait for context to timeout
		<-ctx.Done()

		// Check if redirect script was sent
		body := w.Body.String()
		if !strings.Contains(body, "window.location.href") {
			t.Error("expected redirect script in response")
		}
		if !strings.Contains(body, "/game/"+room.Code) {
			t.Error("expected redirect to game page")
		}
	})
}

func TestHandler_StreamGame(t *testing.T) {
	t.Run("streams game updates successfully", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		player.Role = mockGuardianCard()
		room.AddPlayer(player)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/game", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Run in goroutine
		done := make(chan bool)
		go func() {
			h.StreamGame(w, req)
			done <- true
		}()

		// Wait for handler to finish
		select {
		case <-done:
			// Handler finished
		case <-time.After(200 * time.Millisecond):
			t.Error("StreamGame did not finish in time")
		}

		// Check response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify SSE data was sent
		body := w.Body.String()
		if !strings.Contains(body, "data:") {
			t.Error("expected SSE data in response body")
		}
	})

	t.Run("updates on event", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")
		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/room/"+room.Code+"/game", nil)
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Run handler in goroutine
		go func() {
			h.StreamGame(w, req)
		}()

		// Give handler time to subscribe and send initial render
		time.Sleep(50 * time.Millisecond)

		// Record initial response size
		initialSize := w.Body.Len()

		// Send an event to trigger update
		h.eventBus.Publish(Event{
			Type:     "role_revealed",
			RoomCode: room.Code,
			Data:     map[string]interface{}{"playerID": "p2"},
		})

		// Give time for update to be sent
		time.Sleep(50 * time.Millisecond)

		// Verify additional data was sent
		if w.Body.Len() <= initialSize {
			t.Error("expected additional SSE data after event")
		}

		// Cancel context to stop handler
		cancel()
	})
}

func TestRenderToString(t *testing.T) {
	// Test with a simple templ component
	// We'll create a minimal component that implements templ.Component
	component := templTestComponent{content: "<div>Test Content</div>"}

	result := renderToString(component)

	if result != "<div>Test Content</div>" {
		t.Errorf("expected '<div>Test Content</div>', got %s", result)
	}
}

// templTestComponent is a test implementation of templ.Component
type templTestComponent struct {
	content string
}

func (t templTestComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(t.content))
	return err
}

func TestSSEHelpers(t *testing.T) {
	t.Run("renderLobby sends correct selector", func(t *testing.T) {
		h := newTestHandler()

		// Create test data
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player)

		// Create a test writer and request
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "text/event-stream")

		// Create SSE generator
		sse := datastar.NewSSE(w, req)

		// Call renderLobby
		h.renderLobby(sse, room, player)

		// Verify response contains expected data
		body := w.Body.String()
		if !strings.Contains(body, "data:") {
			t.Error("expected SSE data in response")
		}
		if !strings.Contains(body, "lobby-content") {
			t.Error("expected lobby-content selector in response")
		}
	})

	t.Run("renderGame sends correct selector", func(t *testing.T) {
		h := newTestHandler()

		// Create test data
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		player.Role = mockGuardianCard()
		room.AddPlayer(player)
		room.State = game.StatePlaying

		// Create a test writer and request
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "text/event-stream")

		// Create SSE generator
		sse := datastar.NewSSE(w, req)

		// Call renderGame
		h.renderGame(sse, room, player)

		// Verify response contains expected data
		body := w.Body.String()
		if !strings.Contains(body, "data:") {
			t.Error("expected SSE data in response")
		}
		if !strings.Contains(body, "game-container") {
			t.Error("expected game-container selector in response")
		}
		for _, zoneID := range []string{"zone-status", "zone-privy", "zone-notices", "zone-actions", "zone-roster"} {
			if !strings.Contains(body, zoneID) {
				t.Errorf("expected %s in game patch response", zoneID)
			}
		}
		if strings.Contains(body, "data-init") {
			t.Errorf("game patch must not reinitialize data-init wrapper: %s", body)
		}
		if strings.Contains(body, "/sse/game/") {
			t.Errorf("game patch must not contain game SSE init action: %s", body)
		}
	})

	t.Run("patchSyncPill targets sync pill without game sse wrapper", func(t *testing.T) {
		h := newTestHandler()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "text/event-stream")
		sse := datastar.NewSSE(w, req)

		h.patchSyncPill(sse, "live")

		body := w.Body.String()
		if !strings.Contains(body, "#sync-pill") {
			t.Fatalf("expected sync pill selector in response, got %q", body)
		}
		if !strings.Contains(body, `role="status"`) {
			t.Fatalf("expected sync pill status role in response, got %q", body)
		}
		if strings.Contains(body, "data-init") {
			t.Fatalf("sync pill heartbeat patch must not reinitialize game SSE wrapper, got %q", body)
		}
		if strings.Contains(body, "/sse/game/") {
			t.Fatalf("sync pill heartbeat patch must not include game SSE init action, got %q", body)
		}
	})
}
