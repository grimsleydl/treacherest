package handlers

import (
	"fmt"
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

func TestDebugModeRoutes_StartWithDebugPlayersFillsCoupPresetAndAssignsRoles(t *testing.T) {
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
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	host := game.NewPlayer("host-1", "Host", "session-host")
	host.IsHost = true
	if err := room.AddPlayer(host); err != nil {
		t.Fatalf("add host: %v", err)
	}
	realPlayer := game.NewPlayer("player-1", "Real Player", "session-real")
	if err := room.AddPlayer(realPlayer); err != nil {
		t.Fatalf("add real player: %v", err)
	}
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body %q", w.Code, w.Body.String())
	}

	updatedRoom, err := gameStore.GetRoom(room.Code)
	if err != nil {
		t.Fatalf("get updated room: %v", err)
	}
	if updatedRoom.State != game.StateCountdown {
		t.Fatalf("expected room state %s, got %s", game.StateCountdown, updatedRoom.State)
	}
	activePlayers := updatedRoom.GetActivePlayers()
	if len(activePlayers) != 5 {
		t.Fatalf("expected 5 active players after debug fill, got %d", len(activePlayers))
	}

	debugNames := make(map[string]bool)
	roleCounts := make(map[string]int)
	for _, player := range activePlayers {
		if player.Role == nil {
			t.Fatalf("expected player %q to have a role", player.Name)
		}
		roleCounts[player.Role.Name]++
		if player.Name == "Real Player" && player.IsDebug {
			t.Fatalf("real player should not be marked as a debug player")
		}
		if strings.HasPrefix(player.Name, "Debug Player ") {
			debugNames[player.Name] = true
			if !player.IsDebug {
				t.Fatalf("debug player %q should be marked as a debug player", player.Name)
			}
			if player.IsHost {
				t.Fatalf("debug player %q should not be host", player.Name)
			}
		}
	}
	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("Debug Player %d", i)
		if !debugNames[name] {
			t.Fatalf("expected generated debug player %q in active players, got %v", name, debugNames)
		}
	}

	expectedRoles := map[string]int{
		"King":         1,
		"Blue Knight":  1,
		"Black Knight": 1,
		"Red Knight":   1,
		"Green Knight": 1,
	}
	for roleName, expectedCount := range expectedRoles {
		if roleCounts[roleName] != expectedCount {
			t.Fatalf("expected %d %s role(s), got role counts %v", expectedCount, roleName, roleCounts)
		}
	}
}

func TestDebugModeRoutes_DisabledRouterDoesNotExposeStartWithDebugPlayers(t *testing.T) {
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

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected debug route to be absent with 404, got %d body %q", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Debug endpoints") {
		t.Fatalf("expected route absence, got debug handler response %q", w.Body.String())
	}
}

func TestDebugModeRoutes_StartWithDebugPlayersRejectsNonHost(t *testing.T) {
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
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	player := game.NewPlayer("player-1", "Player 1", "session-1")
	if err := room.AddPlayer(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player.ID})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected non-host debug start to be rejected with 403, got %d body %q", w.Code, w.Body.String())
	}
	updatedRoom, err := gameStore.GetRoom(room.Code)
	if err != nil {
		t.Fatalf("get updated room: %v", err)
	}
	if updatedRoom.State != game.StateLobby {
		t.Fatalf("expected room to remain in lobby, got %s", updatedRoom.State)
	}
	if updatedRoom.GetActivePlayerCount() != 1 {
		t.Fatalf("expected no debug players to be added, got %d active players", updatedRoom.GetActivePlayerCount())
	}
}

func TestDebugModeRoutes_DebugPlayersCanBeEliminatedAsActiveSeats(t *testing.T) {
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
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	host := game.NewPlayer("host-1", "Host", "session-host")
	host.IsHost = true
	if err := room.AddPlayer(host); err != nil {
		t.Fatalf("add host: %v", err)
	}
	realPlayer := game.NewPlayer("player-1", "Real Player", "session-real")
	if err := room.AddPlayer(realPlayer); err != nil {
		t.Fatalf("add real player: %v", err)
	}
	gameStore.UpdateRoom(room)

	startReq := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	startReq.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	startW := httptest.NewRecorder()
	router.ServeHTTP(startW, startReq)
	if startW.Code != http.StatusOK {
		t.Fatalf("expected debug start status 200, got %d body %q", startW.Code, startW.Body.String())
	}

	updatedRoom, err := gameStore.GetRoom(room.Code)
	if err != nil {
		t.Fatalf("get updated room: %v", err)
	}
	var debugPlayer *game.Player
	for _, player := range updatedRoom.GetActivePlayers() {
		if player.IsDebug {
			debugPlayer = player
			break
		}
	}
	if debugPlayer == nil {
		t.Fatalf("expected at least one debug player after debug start")
	}

	eliminateReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/"+debugPlayer.ID+"/eliminate", nil)
	eliminateReq.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	eliminateW := httptest.NewRecorder()
	router.ServeHTTP(eliminateW, eliminateReq)

	if eliminateW.Code != http.StatusOK {
		t.Fatalf("expected debug player elimination status 200, got %d body %q", eliminateW.Code, eliminateW.Body.String())
	}
	finalRoom, err := gameStore.GetRoom(room.Code)
	if err != nil {
		t.Fatalf("get final room: %v", err)
	}
	finalDebugPlayer := finalRoom.GetPlayer(debugPlayer.ID)
	if finalDebugPlayer == nil {
		t.Fatalf("expected debug player %s to persist after elimination", debugPlayer.ID)
	}
	if !finalDebugPlayer.IsEliminated {
		t.Fatalf("expected debug player %s to be eliminated", finalDebugPlayer.Name)
	}
	if !finalDebugPlayer.RoleRevealed {
		t.Fatalf("expected eliminated debug player %s role to be revealed", finalDebugPlayer.Name)
	}
}
