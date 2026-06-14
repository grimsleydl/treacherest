package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"treacherest/internal/game"
)

func TestHandler_Home(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.Home(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected HTML content type, got %s", contentType)
	}

	// Verify some content was rendered
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestHandler_CreateRoom(t *testing.T) {
	t.Run("creates room successfully", func(t *testing.T) {
		h := newTestHandler()

		form := url.Values{}
		form.Add("playerName", "Test Player")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", resp.StatusCode)
		}

		// Verify redirect location
		location := resp.Header.Get("Location")
		if !strings.HasPrefix(location, "/room/") {
			t.Errorf("expected redirect to /room/*, got %s", location)
		}

		// Extract room code from location
		roomCode := strings.TrimPrefix(location, "/room/")
		if len(roomCode) != 5 {
			t.Errorf("expected 5 character room code, got %s", roomCode)
		}

		// Verify room was created in store
		room, err := h.store.GetRoom(roomCode)
		if err != nil {
			t.Fatalf("room not found in store: %v", err)
		}

		// Verify player was added
		if len(room.Players) != 1 {
			t.Errorf("expected 1 player in room, got %d", len(room.Players))
		}

		// Verify cookies were set
		cookies := resp.Cookies()
		var sessionCookie, playerCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "session" {
				sessionCookie = c
			} else if c.Name == "player_"+roomCode {
				playerCookie = c
			}
		}

		if sessionCookie == nil {
			t.Error("session cookie not set")
		}
		if playerCookie == nil {
			t.Error("player cookie not set")
		}
		if sessionCookie != nil && !room.IsOperatorSession(sessionCookie.Value) {
			t.Error("creator session should be the room operator")
		}
	})

	t.Run("creates coup rules mode room", func(t *testing.T) {
		h := newTestHandler()

		form := url.Values{}
		form.Add("playerName", "Test Player")
		form.Add("rulesMode", "coup")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", resp.StatusCode)
		}

		roomCode := strings.TrimPrefix(resp.Header.Get("Location"), "/room/")
		room, err := h.store.GetRoom(roomCode)
		if err != nil {
			t.Fatalf("room not found in store: %v", err)
		}

		if room.RulesMode != game.RulesModeCoup {
			t.Errorf("expected rules mode %q, got %q", game.RulesModeCoup, room.RulesMode)
		}
	})

	t.Run("rejects invalid rules mode", func(t *testing.T) {
		h := newTestHandler()

		form := url.Values{}
		form.Add("playerName", "Test Player")
		form.Add("rulesMode", "bogus")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if location != "" {
			t.Errorf("expected no redirect location, got %s", location)
		}
	})

	t.Run("generates random name when player name is empty", func(t *testing.T) {
		h := newTestHandler()

		form := url.Values{}
		form.Add("playerName", "")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", resp.StatusCode)
		}

		// Extract room code from location
		location := resp.Header.Get("Location")
		roomCode := strings.TrimPrefix(location, "/room/")

		// Verify room was created with player having generated name
		room, err := h.store.GetRoom(roomCode)
		if err != nil {
			t.Fatalf("room not found in store: %v", err)
		}

		if len(room.Players) != 1 {
			t.Errorf("expected 1 player in room, got %d", len(room.Players))
		}

		// Get the first player from the map
		var player *game.Player
		for _, p := range room.Players {
			player = p
			break
		}
		if len(player.Name) != 5 {
			t.Errorf("expected generated name to be 5 characters, got %d", len(player.Name))
		}

		// Check all characters are lowercase letters
		for _, ch := range player.Name {
			if ch < 'a' || ch > 'z' {
				t.Errorf("expected all lowercase letters, got '%c' in name '%s'", ch, player.Name)
			}
		}
	})

	t.Run("preserves existing session", func(t *testing.T) {
		h := newTestHandler()

		existingSession := "existing-session-123"

		form := url.Values{}
		form.Add("playerName", "Test Player")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: existingSession,
		})
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		// Should not create new session cookie
		cookies := w.Result().Cookies()
		for _, c := range cookies {
			if c.Name == "session" {
				t.Error("should not create new session cookie when one exists")
			}
		}

		// Verify player has the existing session ID
		location := w.Result().Header.Get("Location")
		roomCode := strings.TrimPrefix(location, "/room/")
		room, _ := h.store.GetRoom(roomCode)

		for _, player := range room.Players {
			if player.SessionID != existingSession {
				t.Errorf("expected player to have session %s, got %s", existingSession, player.SessionID)
			}
		}
		if !room.IsOperatorSession(existingSession) {
			t.Error("existing creator session should be the room operator")
		}
	})

	t.Run("records host-only creator session as room operator", func(t *testing.T) {
		h := newTestHandler()

		existingSession := "host-only-session-123"

		form := url.Values{}
		form.Add("playerName", "Host Player")
		form.Add("hostOnly", "true")

		req := httptest.NewRequest("POST", "/create-room", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: existingSession,
		})
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", w.Code)
		}

		location := w.Result().Header.Get("Location")
		roomCode := strings.TrimPrefix(location, "/room/")
		room, err := h.store.GetRoom(roomCode)
		if err != nil {
			t.Fatalf("room not found in store: %v", err)
		}
		if !room.IsOperatorSession(existingSession) {
			t.Error("host-only creator session should be the room operator")
		}

		var foundHost bool
		for _, player := range room.Players {
			if player.Name == "Host Player" && player.IsHost {
				foundHost = true
				break
			}
		}
		if !foundHost {
			t.Error("expected host-only creator to be marked as Host")
		}
	})
}

func TestHandler_JoinRoom(t *testing.T) {
	t.Run("renders join form when no name provided", func(t *testing.T) {
		h := newTestHandler()

		// Create a room first
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Create a router to properly handle URL params
		router := chi.NewRouter()
		router.Get("/room/{code}", h.JoinRoom)

		req := httptest.NewRequest("GET", "/room/"+roomCode, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		// Verify the Templ template is rendered
		if !strings.Contains(body, "Join Game Room") {
			t.Error("expected join form title")
		}
		if !strings.Contains(body, roomCode) {
			t.Error("expected room code to be displayed")
		}
		if !strings.Contains(body, `name="player_name"`) {
			t.Error("expected player_name input field")
		}
		if !strings.Contains(body, "Enter your name") {
			t.Error("expected placeholder text")
		}
	})

	t.Run("shows join form even with host cookie", func(t *testing.T) {
		h := newTestHandler()

		// Create a room first
		room, _ := h.store.CreateRoom()
		roomCode := room.Code

		// Create a router to properly handle URL params
		router := chi.NewRouter()
		router.Get("/room/{code}", h.JoinRoom)

		// Try to access room with host cookie (but no player cookie)
		req := httptest.NewRequest("GET", "/room/"+roomCode, nil)
		// Add the host cookie
		req.AddCookie(&http.Cookie{
			Name:  "host_" + roomCode,
			Value: "true",
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Should still show join form since no player cookie
		body := w.Body.String()
		if !strings.Contains(body, "Join Game Room") {
			t.Error("expected join form to be shown")
		}
		if !strings.Contains(body, roomCode) {
			t.Error("expected room code in join form")
		}
	})

	t.Run("renders selected rules mode for existing player", func(t *testing.T) {
		h := newTestHandler()

		room, _ := h.store.CreateRoom()
		room.RulesMode = game.RulesModeCoup
		player := game.NewPlayer("p1", "Coup Player", "session1")
		room.AddPlayer(player)
		h.store.UpdateRoom(room)

		router := chi.NewRouter()
		router.Get("/room/{code}", h.JoinRoom)

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Coup") {
			t.Error("expected Coup rules mode")
		}
		for _, forbidden := range []string{
			`id="coup-preset-form"`,
			`id="coup-role-counts-form"`,
			`/config/coup-`,
			`Unsafe Role Count Override`,
			`Start Game`,
			`/start`,
		} {
			if strings.Contains(body, forbidden) {
				t.Errorf("non-operator lobby rendered forbidden management DOM %q", forbidden)
			}
		}
	})

	t.Run("playing room operator sees operator dashboard before game start", func(t *testing.T) {
		h := newTestHandler()

		room, _ := h.store.CreateRoom()
		room.RulesMode = game.RulesModeCoup
		room.CoupPreset = game.CoupPresetFive
		operator := game.NewPlayer("operator", "Playing Operator", "session-operator")
		operator.IsHost = false
		room.AddPlayer(operator)
		room.OperatorSessionID = operator.SessionID
		h.store.UpdateRoom(room)

		router := chi.NewRouter()
		router.Get("/room/{code}", h.JoinRoom)

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: operator.ID})
		req.AddCookie(&http.Cookie{Name: "session", Value: operator.SessionID})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `id="host-dashboard-container"`) {
			t.Fatalf("expected operator dashboard for playing room operator, got %q", body)
		}
		if strings.Contains(body, `id="player-lobby"`) {
			t.Fatalf("playing room operator should not receive player lobby setup surface: %q", body)
		}
	})

	t.Run("legacy host flag does not grant non-operator lobby management controls", func(t *testing.T) {
		h := newTestHandler()

		room, _ := h.store.CreateRoom()
		room.RulesMode = game.RulesModeCoup
		room.CoupPreset = game.CoupPresetFive
		staleHost := game.NewPlayer("host-flag", "Stale Host", "session-stale-host")
		staleHost.IsHost = true
		operator := game.NewPlayer("operator", "Room Operator", "session-operator")
		room.AddPlayer(staleHost)
		room.AddPlayer(operator)
		room.OperatorSessionID = operator.SessionID
		h.store.UpdateRoom(room)

		router := chi.NewRouter()
		router.Get("/room/{code}", h.JoinRoom)

		req := httptest.NewRequest("GET", "/room/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: staleHost.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, `id="player-lobby"`) {
			t.Fatalf("expected player-safe lobby, got %q", body)
		}
		for _, forbidden := range []string{
			`id="coup-preset-form"`,
			`id="coup-role-counts-form"`,
			`id="coup-info-form"`,
			`id="coup-royal-guard-form"`,
			`id="coup-inquisition-settings-form"`,
			`/config/coup-`,
			`Unsafe Role Count Override`,
			`Start Game`,
			`/start`,
			`id="app-operator-chip"`,
		} {
			if strings.Contains(body, forbidden) {
				t.Fatalf("non-operator lobby rendered forbidden management DOM %q in %q", forbidden, body)
			}
		}
	})
}

func TestHandler_JoinRoomPost(t *testing.T) {
	t.Run("successfully joins room via POST", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		// Join via POST
		formData := "room_code=" + room.Code + "&player_name=Test+Player"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.JoinRoomPost(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", w.Code)
		}

		// Check redirect location
		location := w.Header().Get("Location")
		expectedLocation := "/room/" + room.Code
		if location != expectedLocation {
			t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
		}

		// Verify player was added
		room, _ = h.store.GetRoom(room.Code)
		if len(room.Players) != 1 {
			t.Errorf("expected 1 player, got %d", len(room.Players))
		}

		// Verify player cookie was set
		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "player_"+room.Code {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected player cookie to be set")
		}
	})

	t.Run("preserves host status when joining with host cookie", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		// Join via POST with host cookie
		formData := "room_code=" + room.Code + "&player_name=Host+Player"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Add host cookie
		req.AddCookie(&http.Cookie{
			Name:  "host_" + room.Code,
			Value: "true",
		})
		w := httptest.NewRecorder()

		h.JoinRoomPost(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", w.Code)
		}

		// Verify player was added as host
		room, _ = h.store.GetRoom(room.Code)
		var foundHost bool
		for _, player := range room.Players {
			if player.Name == "Host Player" && player.IsHost {
				foundHost = true
				break
			}
		}
		if !foundHost {
			t.Error("expected player to be marked as host")
		}
	})

	t.Run("rejects duplicate names in same room", func(t *testing.T) {
		h := newTestHandler()

		// Create a room and add first player
		room, _ := h.store.CreateRoom()

		// First player joins
		formData := "room_code=" + room.Code + "&player_name=Alice"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.JoinRoomPost(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected status 303 for first player, got %d", w.Code)
		}

		// Second player tries to join with same name
		formData2 := "room_code=" + room.Code + "&player_name=Alice"
		req2 := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData2))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w2 := httptest.NewRecorder()

		h.JoinRoomPost(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for duplicate name, got %d", w2.Code)
		}

		// Check error message
		body := w2.Body.String()
		if !strings.Contains(body, "a player with that name already exists") {
			t.Errorf("expected duplicate name error message, got: %s", body)
		}

		// Verify only one player in room
		room, _ = h.store.GetRoom(room.Code)
		if len(room.Players) != 1 {
			t.Errorf("expected 1 player in room, got %d", len(room.Players))
		}
	})

	t.Run("rejects duplicate names case-insensitive", func(t *testing.T) {
		h := newTestHandler()

		// Create a room and add first player
		room, _ := h.store.CreateRoom()

		// First player joins with lowercase
		formData := "room_code=" + room.Code + "&player_name=alice"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.JoinRoomPost(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected status 303 for first player, got %d", w.Code)
		}

		// Second player tries to join with uppercase
		formData2 := "room_code=" + room.Code + "&player_name=ALICE"
		req2 := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData2))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w2 := httptest.NewRecorder()

		h.JoinRoomPost(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for duplicate name (case-insensitive), got %d", w2.Code)
		}

		// Check error message
		body := w2.Body.String()
		if !strings.Contains(body, "a player with that name already exists") {
			t.Errorf("expected duplicate name error message, got: %s", body)
		}
	})

	t.Run("allows same name in different rooms", func(t *testing.T) {
		h := newTestHandler()

		// Create two rooms
		room1, _ := h.store.CreateRoom()
		room2, _ := h.store.CreateRoom()

		// Join first room
		formData1 := "room_code=" + room1.Code + "&player_name=Alice"
		req1 := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData1))
		req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w1 := httptest.NewRecorder()

		h.JoinRoomPost(w1, req1)

		if w1.Code != http.StatusSeeOther {
			t.Errorf("expected status 303 for first room, got %d", w1.Code)
		}

		// Join second room with same name
		formData2 := "room_code=" + room2.Code + "&player_name=Alice"
		req2 := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData2))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w2 := httptest.NewRecorder()

		h.JoinRoomPost(w2, req2)

		if w2.Code != http.StatusSeeOther {
			t.Errorf("expected status 303 for second room, got %d", w2.Code)
		}

		// Verify both rooms have a player named Alice
		room1, _ = h.store.GetRoom(room1.Code)
		room2, _ = h.store.GetRoom(room2.Code)

		var found1, found2 bool
		for _, p := range room1.Players {
			if p.Name == "Alice" {
				found1 = true
			}
		}
		for _, p := range room2.Players {
			if p.Name == "Alice" {
				found2 = true
			}
		}

		if !found1 {
			t.Error("expected Alice in room 1")
		}
		if !found2 {
			t.Error("expected Alice in room 2")
		}
	})
}

func TestHandler_GamePage(t *testing.T) {
	t.Run("shows game page for player in game", func(t *testing.T) {
		h := newTestHandler()

		// Create a room with a player
		room, _ := h.store.CreateRoom()
		player := game.NewPlayer("p1", "Test Player", "session1")
		player.Role = mockGuardianCard()
		room.AddPlayer(player)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		// Create a router to handle URL params
		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		// Verify game page content - check for data-star or data- attributes
		if !strings.Contains(body, "data-") {
			t.Error("expected data attributes in game page")
		}
		// Verify it's HTML content
		if len(body) == 0 {
			t.Error("expected non-empty game page body")
		}
	})

	t.Run("does not render debug surface for non-operator player when debug enabled", func(t *testing.T) {
		h := newTestHandler()
		h.config.Server.DebugModeEnabled = true

		room, _ := h.store.CreateRoom()
		operator := game.NewPlayer("operator", "Operator", "session-operator")
		player := game.NewPlayer("p1", "Test Player", "session1")
		player.Role = mockGuardianCard()
		room.OperatorSessionID = operator.SessionID
		room.AddPlayer(operator)
		room.AddPlayer(player)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: player.ID,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if strings.Contains(body, `id="debug-control-surface"`) {
			t.Fatal("non-operator player page should not render debug control surface")
		}
		if strings.Contains(body, "Debug Control Surface") || strings.Contains(body, "Debug Mode active") {
			t.Fatal("non-operator player page should not render debug panel copy")
		}
	})

	t.Run("renders full debug surface for playing room operator when debug enabled", func(t *testing.T) {
		h := newTestHandler()
		h.config.Server.DebugModeEnabled = true

		room, _ := h.store.CreateRoom()
		operator := game.NewPlayer("p1", "Operator", "session-operator")
		operator.Role = mockGuardianCard()
		room.OperatorSessionID = operator.SessionID
		room.AddPlayer(operator)
		room.State = game.StatePlaying
		h.store.UpdateRoom(room)

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		addPlayerSessionCookiesForTest(req, room, operator)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Result().StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, `id="debug-control-surface"`) {
			t.Fatal("playing room operator should see debug control surface")
		}
		if !strings.Contains(body, `id="debug-clear"`) {
			t.Fatal("playing room operator should see debug controls")
		}
		if strings.Contains(body, "Host-only controls are unavailable") {
			t.Fatal("playing room operator should not see constrained non-host debug copy")
		}
	})

	t.Run("returns 404 for non-existent room", func(t *testing.T) {
		h := newTestHandler()

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/XXXXX", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when no player cookie", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 401 when player not in room", func(t *testing.T) {
		h := newTestHandler()

		// Create a room
		room, _ := h.store.CreateRoom()

		router := chi.NewRouter()
		router.Get("/game/{code}", h.GamePage)

		req := httptest.NewRequest("GET", "/game/"+room.Code, nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "nonexistent-player",
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})
}
