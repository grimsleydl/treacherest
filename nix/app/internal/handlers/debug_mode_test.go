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
	room.OperatorSessionID = "session-operator"
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/clear", nil)
	addPlayerSessionCookiesForTest(req, room, player)
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
	room.OperatorSessionID = host.SessionID
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/clear", nil)
	addPlayerSessionCookiesForTest(req, room, host)
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

func TestDebugModeRoutes_DebugClearAllowsPlayingRoomOperatorWhenEnabled(t *testing.T) {
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
	operator := game.NewPlayer("operator-1", "Operator", "session-operator")
	if err := room.AddPlayer(operator); err != nil {
		t.Fatalf("add operator: %v", err)
	}
	room.OperatorSessionID = operator.SessionID
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/clear", nil)
	addPlayerSessionCookiesForTest(req, room, operator)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected playing room operator debug clear to succeed, got %d body %q", w.Code, w.Body.String())
	}
	if gameStore.RoomExists(room.Code) {
		t.Fatalf("playing room operator debug clear did not delete room %s", room.Code)
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
	room.OperatorSessionID = host.SessionID
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	addPlayerSessionCookiesForTest(req, room, host)
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
	room.OperatorSessionID = "session-operator"
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	addPlayerSessionCookiesForTest(req, room, player)
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
	room.OperatorSessionID = host.SessionID
	gameStore.UpdateRoom(room)

	startReq := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-with-debug-players", nil)
	addPlayerSessionCookiesForTest(startReq, room, host)
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

func TestDebugModeRoutes_StartAsIsUnderfilledCoupIncludesKingWithoutDebugPlayers(t *testing.T) {
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
	player1 := game.NewPlayer("player-1", "Player 1", "session-1")
	player2 := game.NewPlayer("player-2", "Player 2", "session-2")
	if err := room.AddPlayer(player1); err != nil {
		t.Fatalf("add player 1: %v", err)
	}
	if err := room.AddPlayer(player2); err != nil {
		t.Fatalf("add player 2: %v", err)
	}
	room.OperatorSessionID = host.SessionID
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-as-is", nil)
	addPlayerSessionCookiesForTest(req, room, host)
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
	if updatedRoom.DebugStartMode != game.DebugStartModeAsIs {
		t.Fatalf("expected debug start mode %q, got %q", game.DebugStartModeAsIs, updatedRoom.DebugStartMode)
	}
	activePlayers := updatedRoom.GetActivePlayers()
	if len(activePlayers) != 2 {
		t.Fatalf("expected Start As-Is to keep 2 active players, got %d", len(activePlayers))
	}
	allowedRoles := map[game.RoleType]bool{
		game.RoleKing:        true,
		game.RoleBlueKnight:  true,
		game.RoleBlackKnight: true,
		game.RoleRedKnight:   true,
		game.RoleGreenKnight: true,
	}
	kingCount := 0
	for _, player := range activePlayers {
		if player.IsDebug {
			t.Fatalf("Start As-Is should not create debug players, got %s", player.Name)
		}
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		roleType := player.Role.GetRoleType()
		if !allowedRoles[roleType] {
			t.Fatalf("expected role %s to come from Coup 5 preset pool", roleType)
		}
		if roleType == game.RoleKing {
			kingCount++
			if !player.RoleRevealed || !player.FaceUp {
				t.Fatalf("expected Start As-Is King %s to start revealed and face up", player.Name)
			}
		}
	}
	if kingCount != 1 {
		t.Fatalf("expected exactly one King in Start As-Is Coup assignment, got %d", kingCount)
	}
}

func TestDebugModeRoutes_DisabledRouterDoesNotExposeStartAsIs(t *testing.T) {
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

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-as-is", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected debug route to be absent with 404, got %d body %q", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Debug endpoints") {
		t.Fatalf("expected route absence, got debug handler response %q", w.Body.String())
	}
}

func TestDebugModeRoutes_StartAsIsRejectsNonHost(t *testing.T) {
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
	room.OperatorSessionID = "session-operator"
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/debug/start-as-is", nil)
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected non-host Start As-Is to be rejected with 403, got %d body %q", w.Code, w.Body.String())
	}
	updatedRoom, err := gameStore.GetRoom(room.Code)
	if err != nil {
		t.Fatalf("get updated room: %v", err)
	}
	if updatedRoom.State != game.StateLobby {
		t.Fatalf("expected room to remain in lobby, got %s", updatedRoom.State)
	}
	if updatedRoom.DebugStartMode != game.DebugStartModeNone {
		t.Fatalf("expected debug start mode to remain unset, got %q", updatedRoom.DebugStartMode)
	}
	if player.Role != nil {
		t.Fatalf("expected non-host Start As-Is rejection not to assign a role")
	}
}

func TestDebugModeRoutes_ViewAsPlayerRendersSelectedPlayerUIPerspectiveReadOnly(t *testing.T) {
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
	room.State = game.StatePlaying
	host := game.NewPlayer("host-1", "Host", "session-host")
	host.IsHost = true
	if err := room.AddPlayer(host); err != nil {
		t.Fatalf("add host: %v", err)
	}
	kingRole := mockHandlerCoupCard(1001, "King")
	kingRole.Rulings = []string{"Private information: Blue Knights: Blue Player"}
	king := game.NewPlayer("king", "King Player", "session-king")
	king.Role = kingRole
	king.RoleRevealed = true
	blueRole := mockHandlerCoupCard(1002, "Blue Knight")
	blueRole.Rulings = []string{"Private information: King: King Player"}
	blue := game.NewPlayer("blue", "Blue Player", "session-blue")
	blue.Role = blueRole
	blackRole := mockHandlerCoupCard(1003, "Black Knight")
	blackRole.Rulings = []string{"Private information: Red Knight: Red Player"}
	black := game.NewPlayer("black", "Black Player", "session-black")
	black.Role = blackRole
	red := game.NewPlayer("red", "Red Player", "session-red")
	red.Role = mockHandlerCoupCard(1004, "Red Knight")
	if err := room.AddPlayer(king); err != nil {
		t.Fatalf("add king: %v", err)
	}
	if err := room.AddPlayer(blue); err != nil {
		t.Fatalf("add blue: %v", err)
	}
	if err := room.AddPlayer(black); err != nil {
		t.Fatalf("add black: %v", err)
	}
	if err := room.AddPlayer(red); err != nil {
		t.Fatalf("add red: %v", err)
	}
	room.OperatorSessionID = host.SessionID
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("GET", "/room/"+room.Code+"/debug/view-as/"+blue.ID, nil)
	addPlayerSessionCookiesForTest(req, room, host)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body %q", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="host-dashboard-content"`) {
		t.Fatalf("expected View As Player to morph main room content, got %q", body)
	}
	if !strings.Contains(body, `id="debug-view-as-perspective"`) || !strings.Contains(body, "inert") {
		t.Fatalf("expected selected perspective to be rendered read-only with inert wrapper, got %q", body)
	}
	if !strings.Contains(body, `id="game-container"`) {
		t.Fatalf("expected selected player's normal game UI, got %q", body)
	}
	if !strings.Contains(body, "View As Player: Blue Player") {
		t.Fatalf("expected selected player heading, got %q", body)
	}
	if !strings.Contains(body, "Blue Knight") || !strings.Contains(body, "Test Blue Knight Card") {
		t.Fatalf("expected selected player's role card UI, got %q", body)
	}
	if !strings.Contains(body, "Private information: King: King Player") {
		t.Fatalf("expected selected player private info, got %q", body)
	}
	if strings.Contains(body, "Private information: Red Knight: Red Player") {
		t.Fatalf("view-as leaked another player's private info: %q", body)
	}
	if !strings.Contains(body, "Royal Guard") || !strings.Contains(body, "Call Inquisition") {
		t.Fatalf("expected selected player's normal Coup controls in read-only perspective, got %q", body)
	}
	if !strings.Contains(body, "Read Only: yes") {
		t.Fatalf("expected read-only marker, got %q", body)
	}
	for _, forbidden := range []string{"Record Reveal", "Record Elimination"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("view-as should not expose host control %q in %q", forbidden, body)
		}
	}
}

func TestDebugModeRoutes_ViewAsPlayerSwitchesPerspectiveWithoutLeakingPreviousSelection(t *testing.T) {
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
	room.State = game.StatePlaying
	host := game.NewPlayer("host-1", "Host", "session-host")
	host.IsHost = true
	blueRole := mockHandlerCoupCard(1002, "Blue Knight")
	blueRole.Rulings = []string{"Private information: Blue-only clue"}
	blue := game.NewPlayer("blue", "Blue Player", "session-blue")
	blue.Role = blueRole
	debugRole := mockHandlerCoupCard(1005, "Green Knight")
	debugRole.Rulings = []string{"Private information: Debug-only clue"}
	debugPlayer := game.NewPlayer("debug-1", "Debug Player 1", "debug-session")
	debugPlayer.IsDebug = true
	debugPlayer.Role = debugRole
	for _, player := range []*game.Player{host, blue, debugPlayer} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("add player %s: %v", player.Name, err)
		}
	}
	room.OperatorSessionID = host.SessionID
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("GET", "/room/"+room.Code+"/debug/view-as/"+blue.ID, nil)
	addPlayerSessionCookiesForTest(req, room, host)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	blueBody := w.Body.String()
	if w.Code != http.StatusOK {
		t.Fatalf("expected blue perspective status 200, got %d body %q", w.Code, blueBody)
	}
	if !strings.Contains(blueBody, "Private information: Blue-only clue") {
		t.Fatalf("expected blue private info, got %q", blueBody)
	}
	if strings.Contains(blueBody, "Private information: Debug-only clue") {
		t.Fatalf("blue perspective leaked debug player private info: %q", blueBody)
	}

	req = httptest.NewRequest("GET", "/room/"+room.Code+"/debug/view-as/"+debugPlayer.ID, nil)
	addPlayerSessionCookiesForTest(req, room, host)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	debugBody := w.Body.String()
	if w.Code != http.StatusOK {
		t.Fatalf("expected debug player perspective status 200, got %d body %q", w.Code, debugBody)
	}
	if !strings.Contains(debugBody, "View As Player: Debug Player 1") {
		t.Fatalf("expected debug player heading, got %q", debugBody)
	}
	if !strings.Contains(debugBody, "Private information: Debug-only clue") {
		t.Fatalf("expected debug player private info, got %q", debugBody)
	}
	if strings.Contains(debugBody, "Private information: Blue-only clue") {
		t.Fatalf("debug player perspective leaked previous selected player private info: %q", debugBody)
	}
}

func TestDebugModeRoutes_DisabledRouterDoesNotExposeViewAsPlayer(t *testing.T) {
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
	player := game.NewPlayer("player-1", "Player 1", "session-1")
	if err := room.AddPlayer(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("GET", "/room/"+room.Code+"/debug/view-as/"+player.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected debug route to be absent with 404, got %d body %q", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Debug endpoints") {
		t.Fatalf("expected route absence, got debug handler response %q", w.Body.String())
	}
}

func TestDebugModeRoutes_ViewAsPlayerRejectsNonHost(t *testing.T) {
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
	viewer := game.NewPlayer("viewer", "Viewer", "session-viewer")
	target := game.NewPlayer("target", "Target", "session-target")
	target.Role = mockHandlerCoupCard(1001, "King")
	if err := room.AddPlayer(viewer); err != nil {
		t.Fatalf("add viewer: %v", err)
	}
	if err := room.AddPlayer(target); err != nil {
		t.Fatalf("add target: %v", err)
	}
	room.OperatorSessionID = "session-operator"
	gameStore.UpdateRoom(room)

	req := httptest.NewRequest("GET", "/room/"+room.Code+"/debug/view-as/"+target.ID, nil)
	addPlayerSessionCookiesForTest(req, room, viewer)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected non-host View As Player to be rejected with 403, got %d body %q", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Role: King") {
		t.Fatalf("non-host rejection leaked selected player role: %q", w.Body.String())
	}
}
