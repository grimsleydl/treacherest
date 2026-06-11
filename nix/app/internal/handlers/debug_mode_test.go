package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"
)

func TestDebugModeRoutes_DisabledRouterDoesNotExposeDebugClear(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = false
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg, nil)
	router := SetupRouter(h, cfg, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	room, err := gameStore.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/clear", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected debug route to be absent with 404, got %d body %q", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Debug endpoints") {
		t.Fatalf("expected route absence, got debug handler response %q", w.Body.String())
	}
}

func TestDebugModeRoutes_DebugClearRejectsNonHost(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg, nil)
	router := SetupRouter(h, cfg, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	room, err := gameStore.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	player := game.NewPlayer("player-1", "Player 1", "session-1")
	if err := room.AddPlayer(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/clear", nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player.ID})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected non-host debug clear to be rejected with 403, got %d body %q", w.Code, w.Body.String())
	}
	if !gameStore.RoomExists(room.Code) {
		t.Fatalf("non-host debug clear deleted room %s", room.Code)
	}
}

func TestDebugModeRoutes_DebugClearAllowsHostWhenEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.DebugModeEnabled = true
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg, nil)
	router := SetupRouter(h, cfg, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	room, err := gameStore.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	host := game.NewPlayer("host-1", "Host", "session-host")
	host.IsHost = true
	if err := room.AddPlayer(host); err != nil {
		t.Fatalf("add host: %v", err)
	}
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/clear", nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected host debug clear to succeed, got %d body %q", w.Code, w.Body.String())
	}
	if gameStore.RoomExists(room.Code) {
		t.Fatalf("host debug clear did not delete room %s", room.Code)
	}
	if !strings.Contains(w.Body.String(), `"status": "cleared"`) && !strings.Contains(w.Body.String(), `"status":"cleared"`) {
		t.Fatalf("expected cleared JSON response, got %q", w.Body.String())
	}
}
