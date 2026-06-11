package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

func markRoomOperatorForTest(room *game.Room, player *game.Player) {
	room.OperatorSessionID = player.SessionID
}

func addPlayerSessionCookiesForTest(req *http.Request, room *game.Room, player *game.Player) {
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: player.SessionID,
	})
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
}

func TestHandler_StartGame(t *testing.T) {
	t.Run("starts game successfully", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")
		player3 := game.NewPlayer("p3", "Player 3", "session3")
		player4 := game.NewPlayer("p4", "Player 4", "session4")

		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.AddPlayer(player3)
		room.AddPlayer(player4)
		markRoomOperatorForTest(room, player1)
		h.store.UpdateRoom(room)

		// Create request with chi context
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		addPlayerSessionCookiesForTest(req, room, player1)

		// Add chi URL params to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify room state changed
		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateCountdown {
			t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
		}

		// Verify roles were assigned
		players := updatedRoom.GetPlayers()
		rolesAssigned := false
		for _, p := range players {
			if p.Role != nil {
				rolesAssigned = true
				break
			}
		}
		if !rolesAssigned {
			t.Error("no roles were assigned to players")
		}

		// Wait a bit to ensure countdown goroutine starts
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("responds with datastar redirect script", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with 1 player (minimum to start)
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player1)
		markRoomOperatorForTest(room, player1)
		h.store.UpdateRoom(room)

		// Create request with chi context
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		addPlayerSessionCookiesForTest(req, room, player1)

		// Add chi URL params to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Check that response contains SSE redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/game/" + room.Code + "'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
		}

		// Verify room state changed to countdown
		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateCountdown {
			t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
		}
	})

	t.Run("rejects non-operator player start", func(t *testing.T) {
		h := newTestHandler()

		room, _ := h.store.CreateRoom()
		operator := game.NewPlayer("p1", "Operator", "session-operator")
		nonOperator := game.NewPlayer("p2", "Non Operator", "session-player")
		room.OperatorSessionID = operator.SessionID
		room.AddPlayer(operator)
		room.AddPlayer(nonOperator)
		h.store.UpdateRoom(room)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		addPlayerSessionCookiesForTest(req, room, nonOperator)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateLobby {
			t.Fatalf("expected non-operator start to leave room in lobby, got %s", updatedRoom.State)
		}
		if strings.Contains(w.Body.String(), "window.location.href") {
			t.Fatalf("expected no redirect script for non-operator start, got %s", w.Body.String())
		}
	})

	t.Run("rejects unsupported coup player count", func(t *testing.T) {
		h := newTestHandler()

		room, _ := h.store.CreateRoom()
		room.RulesMode = game.RulesModeCoup
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		room.AddPlayer(player1)
		markRoomOperatorForTest(room, player1)
		h.store.UpdateRoom(room)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		addPlayerSessionCookiesForTest(req, room, player1)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Coup preset coup-5 requires exactly 5 active players") {
			t.Errorf("expected response to explain unsupported Coup player count, got: %s", body)
		}
		if strings.Contains(body, "window.location.href") {
			t.Errorf("expected no game redirect for unsupported Coup player count, got: %s", body)
		}

		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateLobby {
			t.Errorf("expected state %s, got %s", game.StateLobby, updatedRoom.State)
		}
		if updatedRoom.RulesMode != game.RulesModeCoup {
			t.Errorf("expected rules mode %q, got %q", game.RulesModeCoup, updatedRoom.RulesMode)
		}
		for _, player := range updatedRoom.GetPlayers() {
			if player.Role != nil {
				t.Errorf("expected no role assignment for Coup room, got %s for %s", player.Role.Name, player.Name)
			}
		}
	})

	t.Run("rejects selected coup preset that does not match player count", func(t *testing.T) {
		h := newTestHandler()

		room, _ := h.store.CreateRoom()
		room.RulesMode = game.RulesModeCoup
		room.CoupPreset = game.CoupPresetSix
		players := []*game.Player{
			game.NewPlayer("p1", "Player 1", "session1"),
			game.NewPlayer("p2", "Player 2", "session2"),
			game.NewPlayer("p3", "Player 3", "session3"),
			game.NewPlayer("p4", "Player 4", "session4"),
			game.NewPlayer("p5", "Player 5", "session5"),
		}
		for _, player := range players {
			room.AddPlayer(player)
		}
		markRoomOperatorForTest(room, players[0])
		h.store.UpdateRoom(room)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		addPlayerSessionCookiesForTest(req, room, players[0])

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Coup preset coup-6 requires exactly 6 active players") {
			t.Errorf("expected response to explain selected Coup preset mismatch, got: %s", body)
		}
		if strings.Contains(body, "window.location.href") {
			t.Errorf("expected no game redirect for mismatched Coup preset, got: %s", body)
		}

		updatedRoom, _ := h.store.GetRoom(room.Code)
		if updatedRoom.State != game.StateLobby {
			t.Errorf("expected state %s, got %s", game.StateLobby, updatedRoom.State)
		}
		for _, player := range updatedRoom.GetPlayers() {
			if player.Role != nil {
				t.Errorf("expected no role assignment for mismatched Coup preset, got %s for %s", player.Role.Name, player.Name)
			}
		}
	})

	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := newTestHandler()

		req := httptest.NewRequest("POST", "/room/XXXXX/start", nil)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "XXXXX")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when player not in room", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
		// No player cookie

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.StartGame(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("returns SSE error when player not found in room", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with 0 players (not enough to start)
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Player 1", "session1")
		// Don't add player to room, but still send request with cookie

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
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

		h.StartGame(w, req)

		// Check for SSE error message
		body := w.Body.String()
		if !strings.Contains(body, "You are not in this room") {
			t.Errorf("expected SSE error message for player not in room, got: %s", body)
		}
	})
}

func TestHandler_StartGame_CoupFivePlayerHappyPath(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.StartGame(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	body := w.Body.String()
	expectedScript := "window.location.href = '/game/" + room.Code + "'"
	if !strings.Contains(body, expectedScript) {
		t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StateCountdown {
		t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
	}
	if updatedRoom.RulesMode != game.RulesModeCoup {
		t.Errorf("expected rules mode %q, got %q", game.RulesModeCoup, updatedRoom.RulesMode)
	}

	gotRoles := make(map[string]int)
	for _, player := range updatedRoom.GetActivePlayers() {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		gotRoles[player.Role.Name]++
		if player.Role.Name == "King" {
			if !player.RoleRevealed {
				t.Error("expected King to be publicly revealed")
			}
			if !player.FaceUp {
				t.Error("expected King to be face up")
			}
		} else {
			if player.RoleRevealed {
				t.Errorf("expected %s to start hidden", player.Role.Name)
			}
			if player.FaceUp {
				t.Errorf("expected %s to start face down", player.Role.Name)
			}
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
		if gotRoles[roleName] != expectedCount {
			t.Errorf("expected %d %s role(s), got %d in %v", expectedCount, roleName, gotRoles[roleName], gotRoles)
		}
	}
	if len(gotRoles) != len(expectedRoles) {
		t.Errorf("expected only Coup roles %v, got %v", expectedRoles, gotRoles)
	}
}

func TestHandler_StartGame_CoupCustomRoleCounts(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	room.CoupRoleCountsCustom = true
	room.CoupRoleCounts = game.CoupRoleCounts{
		game.RoleKing:        1,
		game.RoleBlueKnight:  2,
		game.RoleBlackKnight: 0,
		game.RoleRedKnight:   1,
		game.RoleGreenKnight: 1,
		game.RoleWasteland:   0,
	}
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.StartGame(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	if !strings.Contains(w.Body.String(), "window.location.href = '/game/"+room.Code+"'") {
		t.Fatalf("expected custom Coup start to redirect, got %s", w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StateCountdown {
		t.Fatalf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
	}
	gotRoles := map[game.RoleType]int{}
	for _, player := range updatedRoom.GetActivePlayers() {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		roleType := player.Role.GetRoleType()
		gotRoles[roleType]++
		if roleType == game.RoleKing {
			if !player.RoleRevealed || !player.FaceUp {
				t.Fatalf("expected King %s to start revealed", player.Name)
			}
		} else if player.RoleRevealed || player.FaceUp {
			t.Fatalf("expected non-King %s to start hidden", player.Name)
		}
	}
	for role, wantCount := range room.CoupRoleCounts {
		if gotRoles[role] != wantCount {
			t.Fatalf("expected %d %s role(s), got counts %v", wantCount, role, gotRoles)
		}
	}
}

func TestHandler_StartGame_CoupCustomRoleCountsRejectsStructuralErrors(t *testing.T) {
	tests := []struct {
		name        string
		counts      game.CoupRoleCounts
		wantMessage string
	}{
		{
			name: "total mismatch",
			counts: game.CoupRoleCounts{
				game.RoleKing:        1,
				game.RoleBlueKnight:  1,
				game.RoleBlackKnight: 1,
				game.RoleRedKnight:   1,
				game.RoleGreenKnight: 1,
				game.RoleWasteland:   1,
			},
			wantMessage: "Coup role counts total 6 but there are 5 active players",
		},
		{
			name: "missing King",
			counts: game.CoupRoleCounts{
				game.RoleBlueKnight:  2,
				game.RoleBlackKnight: 1,
				game.RoleRedKnight:   1,
				game.RoleGreenKnight: 1,
			},
			wantMessage: "Coup role counts require exactly one King",
		},
		{
			name: "missing Red",
			counts: game.CoupRoleCounts{
				game.RoleKing:        1,
				game.RoleBlueKnight:  2,
				game.RoleBlackKnight: 1,
				game.RoleGreenKnight: 1,
			},
			wantMessage: "Coup role counts require exactly one Red Knight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			room, _ := h.store.CreateRoom()
			room.RulesMode = game.RulesModeCoup
			room.CoupRoleCountsCustom = true
			room.CoupRoleCounts = tt.counts
			players := []*game.Player{
				game.NewPlayer("p1", "Player 1", "session1"),
				game.NewPlayer("p2", "Player 2", "session2"),
				game.NewPlayer("p3", "Player 3", "session3"),
				game.NewPlayer("p4", "Player 4", "session4"),
				game.NewPlayer("p5", "Player 5", "session5"),
			}
			for _, player := range players {
				room.AddPlayer(player)
			}
			markRoomOperatorForTest(room, players[0])
			h.store.UpdateRoom(room)

			req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
			addPlayerSessionCookiesForTest(req, room, players[0])
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", room.Code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			h.StartGame(w, req)

			body := w.Body.String()
			if !strings.Contains(body, tt.wantMessage) {
				t.Fatalf("expected response to contain %q, got %s", tt.wantMessage, body)
			}
			if strings.Contains(body, "window.location.href") {
				t.Fatalf("expected no redirect for invalid custom Coup counts, got %s", body)
			}
			updatedRoom, _ := h.store.GetRoom(room.Code)
			if updatedRoom.State != game.StateLobby {
				t.Fatalf("expected invalid custom Coup counts to leave lobby state, got %s", updatedRoom.State)
			}
		})
	}
}

func TestHandler_StartGame_CoupUnsafeRoleCountsAllowsMissingKingAndRed(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupRoleCountsCustom = true
	room.CoupAllowUnsafeRoleCounts = true
	room.CoupRoleCounts = game.CoupRoleCounts{
		game.RoleBlueKnight:  2,
		game.RoleBlackKnight: 2,
		game.RoleGreenKnight: 1,
	}
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.StartGame(w, req)

	if !strings.Contains(w.Body.String(), "window.location.href = '/game/"+room.Code+"'") {
		t.Fatalf("expected unsafe custom Coup start to redirect, got %s", w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StateCountdown {
		t.Fatalf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
	}
	gotRoles := map[game.RoleType]int{}
	for _, player := range updatedRoom.GetActivePlayers() {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		gotRoles[player.Role.GetRoleType()]++
		if player.RoleRevealed || player.FaceUp {
			t.Fatalf("expected no role to start revealed when unsafe pool has no King")
		}
	}
	if gotRoles[game.RoleKing] != 0 || gotRoles[game.RoleRedKnight] != 0 {
		t.Fatalf("expected unsafe pool to omit King and Red, got counts %v", gotRoles)
	}
}

func TestHandler_StartGame_CoupUnsafeRoleCountsStillRejectsTotalMismatch(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupRoleCountsCustom = true
	room.CoupAllowUnsafeRoleCounts = true
	room.CoupRoleCounts = game.CoupRoleCounts{
		game.RoleBlueKnight:  2,
		game.RoleBlackKnight: 2,
		game.RoleGreenKnight: 1,
		game.RoleWasteland:   1,
	}
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.StartGame(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Coup role counts total 6 but there are 5 active players") {
		t.Fatalf("expected total mismatch error, got %s", body)
	}
	if strings.Contains(body, "window.location.href") {
		t.Fatalf("expected no redirect for unsafe total mismatch, got %s", body)
	}
}

func TestToggleReveal_CoupHostCanRecordPublicReveal(t *testing.T) {
	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	host := game.NewPlayer("host", "Host", "session-host")
	host.IsHost = true
	target := game.NewPlayer("p1", "Blue Player", "session-blue")
	target.Role = mockHandlerCoupCard(1002, "Blue Knight")
	target.RoleRevealed = false
	target.FaceUp = false

	room.AddPlayer(host)
	room.AddPlayer(target)
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/reveal/"+target.ID, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: host.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", target.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ToggleReveal(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedTarget := updatedRoom.GetPlayer(target.ID)
	if !updatedTarget.RoleRevealed {
		t.Fatal("expected target role to be publicly revealed")
	}
	if !updatedTarget.FaceUp {
		t.Fatal("expected public reveal to turn target role face up")
	}
}

func TestToggleReveal_CoupPublicRevealIsIdempotent(t *testing.T) {
	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	player := game.NewPlayer("p1", "Blue Player", "session-blue")
	player.Role = mockHandlerCoupCard(1002, "Blue Knight")
	player.RoleRevealed = true
	player.FaceUp = true

	room.AddPlayer(player)
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/reveal/"+player.ID, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", player.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ToggleReveal(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer(player.ID)
	if !updatedPlayer.RoleRevealed {
		t.Fatal("expected Coup public reveal endpoint to leave role revealed")
	}
	if !updatedPlayer.FaceUp {
		t.Fatal("expected Coup public reveal endpoint to leave role face up")
	}
}

func TestEliminatePlayer_CoupSelfEliminationRevealsRole(t *testing.T) {
	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	player := game.NewPlayer("p1", "Black Player", "session-black")
	player.Role = mockHandlerCoupCard(1003, "Black Knight")
	player.RoleRevealed = false
	player.FaceUp = false

	room.AddPlayer(player)
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/"+player.ID+"/eliminate", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", player.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.EliminatePlayer(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer(player.ID)
	if !updatedPlayer.IsEliminated {
		t.Fatal("expected player to be eliminated")
	}
	if !updatedPlayer.RoleRevealed {
		t.Fatal("expected eliminated player's role to be publicly revealed")
	}
	if !updatedPlayer.FaceUp {
		t.Fatal("expected eliminated player's role to be face up")
	}
}

func TestUseCoupRoyalGuard_RevealsHiddenBlueKnight(t *testing.T) {
	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	blue := game.NewPlayer("p1", "Blue Player", "session-blue")
	blue.Role = mockHandlerCoupCard(1002, "Blue Knight")
	blue.RoleRevealed = false
	blue.FaceUp = false

	room.AddPlayer(blue)
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/royal-guard/"+blue.ID, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: blue.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", blue.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.UseCoupRoyalGuard(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedBlue := updatedRoom.GetPlayer(blue.ID)
	if !updatedBlue.RoleRevealed {
		t.Fatal("expected Royal Guard to publicly reveal Blue Knight")
	}
	if !updatedBlue.FaceUp {
		t.Fatal("expected Royal Guard to turn Blue Knight face up")
	}
}

func TestSetupRouter_RoutesCoupRoyalGuardAction(t *testing.T) {
	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	blue := game.NewPlayer("p1", "Blue Player", "session-blue")
	blue.Role = mockHandlerCoupCard(1002, "Blue Knight")
	blue.RoleRevealed = false
	blue.FaceUp = false

	room.AddPlayer(blue)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/royal-guard/"+blue.ID, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: blue.ID,
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}
}

func TestHandler_StartGame_CoupSixPlayerPreset(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetSix
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
		game.NewPlayer("p6", "Player 6", "session6"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.StartGame(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	body := w.Body.String()
	expectedScript := "window.location.href = '/game/" + room.Code + "'"
	if !strings.Contains(body, expectedScript) {
		t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StateCountdown {
		t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
	}

	gotRoles := make(map[string]int)
	for _, player := range updatedRoom.GetActivePlayers() {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		gotRoles[player.Role.Name]++
	}

	expectedRoles := map[string]int{
		"King":         1,
		"Blue Knight":  1,
		"Black Knight": 2,
		"Red Knight":   1,
		"Green Knight": 1,
	}
	for roleName, expectedCount := range expectedRoles {
		if gotRoles[roleName] != expectedCount {
			t.Errorf("expected %d %s role(s), got %d in %v", expectedCount, roleName, gotRoles[roleName], gotRoles)
		}
	}
	if len(gotRoles) != len(expectedRoles) {
		t.Errorf("expected only Coup roles %v, got %v", expectedRoles, gotRoles)
	}
}

func TestHandler_StartGame_CoupInformationPolicy(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetSix
	room.CoupInfoPolicy = game.CoupInformationPolicy{
		BlackToRed:   game.CoupBlackToRedAll,
		BlackNetwork: game.CoupBlackNetworkAll,
	}
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
		game.NewPlayer("p6", "Player 6", "session6"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.StartGame(w, req)

	updatedRoom, _ := h.store.GetRoom(room.Code)
	red := findHandlerTestPlayerByRole(t, updatedRoom.GetActivePlayers(), game.RoleRedKnight)
	blacks := findHandlerTestPlayersByRole(t, updatedRoom.GetActivePlayers(), game.RoleBlackKnight)
	if len(blacks) != 2 {
		t.Fatalf("expected 2 Black Knights, got %d", len(blacks))
	}

	for _, black := range blacks {
		if !handlerTestRoleRulingsContain(black.Role, "Red Knight: "+red.Name) {
			t.Fatalf("expected Black %q to know Red %q, got rulings %v", black.Name, red.Name, black.Role.Rulings)
		}
		for _, otherBlack := range blacks {
			if !handlerTestRoleRulingsContain(black.Role, otherBlack.Name) {
				t.Fatalf("expected Black %q to know Black network member %q, got rulings %v", black.Name, otherBlack.Name, black.Role.Rulings)
			}
		}
	}
}

func TestHandler_StartGame_CoupEightPlayerChaosPreset(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetEightChaos
	players := []*game.Player{
		game.NewPlayer("p1", "Player 1", "session1"),
		game.NewPlayer("p2", "Player 2", "session2"),
		game.NewPlayer("p3", "Player 3", "session3"),
		game.NewPlayer("p4", "Player 4", "session4"),
		game.NewPlayer("p5", "Player 5", "session5"),
		game.NewPlayer("p6", "Player 6", "session6"),
		game.NewPlayer("p7", "Player 7", "session7"),
		game.NewPlayer("p8", "Player 8", "session8"),
	}
	for _, player := range players {
		room.AddPlayer(player)
	}
	markRoomOperatorForTest(room, players[0])
	h.store.UpdateRoom(room)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/start", nil)
	addPlayerSessionCookiesForTest(req, room, players[0])

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.StartGame(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	body := w.Body.String()
	expectedScript := "window.location.href = '/game/" + room.Code + "'"
	if !strings.Contains(body, expectedScript) {
		t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StateCountdown {
		t.Errorf("expected state %s, got %s", game.StateCountdown, updatedRoom.State)
	}

	gotRoles := make(map[string]int)
	for _, player := range updatedRoom.GetActivePlayers() {
		if player.Role == nil {
			t.Fatalf("expected player %s to have a role", player.Name)
		}
		gotRoles[player.Role.Name]++
	}

	expectedRoles := map[string]int{
		"King":             1,
		"Blue Knight":      2,
		"Black Knight":     2,
		"Red Knight":       1,
		"Green Knight":     1,
		"Wasteland Knight": 1,
	}
	for roleName, expectedCount := range expectedRoles {
		if gotRoles[roleName] != expectedCount {
			t.Errorf("expected %d %s role(s), got %d in %v", expectedCount, roleName, gotRoles[roleName], gotRoles)
		}
	}
	if len(gotRoles) != len(expectedRoles) {
		t.Errorf("expected only Coup roles %v, got %v", expectedRoles, gotRoles)
	}
}

func TestHandler_LeaveRoom(t *testing.T) {
	t.Run("player leaves room successfully", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with players
		room, _ := h.store.CreateRoom()
		player1 := game.NewPlayer("p1", "Player 1", "session1")
		player2 := game.NewPlayer("p2", "Player 2", "session2")

		room.AddPlayer(player1)
		room.AddPlayer(player2)
		h.store.UpdateRoom(room)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player1.ID,
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify Datastar redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
		}

		// Verify player was removed from room
		updatedRoom, _ := h.store.GetRoom(room.Code)
		if len(updatedRoom.Players) != 1 {
			t.Errorf("expected 1 player remaining, got %d", len(updatedRoom.Players))
		}

		if updatedRoom.GetPlayer(player1.ID) != nil {
			t.Error("player1 should have been removed from room")
		}
	})

	t.Run("redirects to home for non-existent room", func(t *testing.T) {
		h := newTestHandler()

		req := httptest.NewRequest("POST", "/room/XXXXX/leave", nil)
		req.Header.Set("Accept", "text/event-stream")

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "XXXXX")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		// Check for SSE redirect
		body := w.Body.String()
		if !strings.Contains(body, "window.location.href = '/'") {
			t.Errorf("expected SSE redirect to home, got: %s", body)
		}
	})

	t.Run("redirects to home when no player cookie", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		req.Header.Set("Accept", "text/event-stream")
		// No player cookie

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		// Check for SSE redirect
		body := w.Body.String()
		if !strings.Contains(body, "window.location.href = '/'") {
			t.Errorf("expected SSE redirect to home, got: %s", body)
		}
	})

	t.Run("handles player not in room", func(t *testing.T) {
		h := newTestHandler()

		// Create a room without adding the player
		room, _ := h.store.CreateRoom()

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/leave", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "nonexistent-player",
		})

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.LeaveRoom(w, req)

		// Should still redirect even if player wasn't in room
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify Datastar redirect script
		body := w.Body.String()
		expectedScript := "window.location.href = '/'"
		if !strings.Contains(body, expectedScript) {
			t.Errorf("expected response to contain redirect script %q, got: %s", expectedScript, body)
		}
	})
}

func TestHandler_runCountdown(t *testing.T) {
	t.Run("runs countdown and transitions to playing", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()
		room.State = game.StateCountdown
		room.CountdownRemaining = 5
		h.store.UpdateRoom(room)

		// Subscribe to events to verify they're published
		events := h.eventBus.Subscribe(room.Code)
		defer h.eventBus.Unsubscribe(room.Code, events)

		// Run countdown in goroutine
		done := make(chan bool)
		go func() {
			h.runCountdown(room)
			done <- true
		}()

		// Collect events
		var receivedEvents []Event
		timeout := time.After(6 * time.Second)

		collecting := true
		for collecting {
			select {
			case event := <-events:
				receivedEvents = append(receivedEvents, event)
			case <-done:
				// Give a bit more time for final event
				time.Sleep(100 * time.Millisecond)
				collecting = false
			case <-timeout:
				t.Fatal("countdown took too long")
			}
		}

		// Verify countdown events were sent
		countdownEvents := 0
		gamePlayingEvent := false
		for _, event := range receivedEvents {
			if event.Type == "countdown_update" {
				countdownEvents++
			} else if event.Type == "game_playing" {
				gamePlayingEvent = true
			}
		}

		if countdownEvents < 5 {
			t.Errorf("expected at least 5 countdown events, got %d", countdownEvents)
		}

		if !gamePlayingEvent {
			t.Error("expected game_playing event")
		}

		// Verify final room state
		finalRoom, _ := h.store.GetRoom(room.Code)
		if finalRoom.State != game.StatePlaying {
			t.Errorf("expected state %s, got %s", game.StatePlaying, finalRoom.State)
		}

		if finalRoom.CountdownRemaining != 0 {
			t.Errorf("expected countdown 0, got %d", finalRoom.CountdownRemaining)
		}

		if !finalRoom.LeaderRevealed {
			t.Error("expected leader to be revealed")
		}
	})
}

func findHandlerTestPlayerByRole(t *testing.T, players []*game.Player, roleType game.RoleType) *game.Player {
	t.Helper()
	for _, player := range players {
		if player.Role != nil && player.Role.GetRoleType() == roleType {
			return player
		}
	}
	t.Fatalf("expected to find %s in assigned players", roleType)
	return nil
}

func findHandlerTestPlayersByRole(t *testing.T, players []*game.Player, roleType game.RoleType) []*game.Player {
	t.Helper()
	found := make([]*game.Player, 0)
	for _, player := range players {
		if player.Role != nil && player.Role.GetRoleType() == roleType {
			found = append(found, player)
		}
	}
	if len(found) == 0 {
		t.Fatalf("expected to find %s in assigned players", roleType)
	}
	return found
}

func mockHandlerCoupCard(id int, name string) *game.Card {
	return &game.Card{
		ID:   id,
		Name: name,
		Types: game.CardTypes{
			Supertype: "Coup",
			Subtype:   name,
		},
		Text:        "Test " + name + " Card",
		Type:        "Coup Role",
		Rarity:      "Coup",
		Artist:      "Test Artist",
		Base64Image: "data:image/jpeg;base64,test",
	}
}

func handlerTestRoleRulingsContain(card *game.Card, want string) bool {
	if card == nil {
		return false
	}
	for _, ruling := range card.Rulings {
		if strings.Contains(ruling, want) {
			return true
		}
	}
	return false
}
