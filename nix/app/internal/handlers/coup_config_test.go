package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

func TestUpdateCoupPreset(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetEightChaos))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupPreset(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupPreset != game.CoupPresetEightChaos {
		t.Fatalf("expected Coup preset %q, got %q", game.CoupPresetEightChaos, updatedRoom.CoupPreset)
	}
	if updatedRoom.CoupRoleCountsCustom {
		t.Fatal("expected preset selection to leave Coup role counts in preset mode")
	}
	if updatedRoom.CoupRoleCounts[game.RoleWasteland] != 1 {
		t.Fatalf("expected preset selection to seed Wasteland count 1, got counts %v", updatedRoom.CoupRoleCounts)
	}

	body := w.Body.String()
	if !strings.Contains(body, "8 players, chaos") {
		t.Errorf("expected updated Coup preset label in response, got: %s", body)
	}
	if !strings.Contains(body, "Wasteland Knight") {
		t.Errorf("expected updated Coup role summary in response, got: %s", body)
	}
}

func TestUpdateCoupPresetFromHostDashboardPatchesHostSurface(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	host := game.NewPlayer("host", "Host", "host-session")
	host.IsHost = true
	room.AddPlayer(host)
	markRoomOperatorForTest(room, host)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetSix))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, host)
	req.AddCookie(&http.Cookie{Name: "host_" + room.Code, Value: "true"})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupPreset(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "host-dashboard-container") {
		t.Fatalf("expected host dashboard patch in response, got: %s", body)
	}
	if strings.Contains(body, "lobby-content") {
		t.Fatalf("host dashboard preset update should not target lobby content, got: %s", body)
	}
	if !strings.Contains(body, "2 Black Knights") {
		t.Fatalf("expected host response to include updated 6-player role counts, got: %s", body)
	}
}

func TestUpdateCoupPlayerCountIncrementsPresetAndRoleCounts(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	staleCounts, _ := game.CoupRoleCountsForPreset(game.CoupPresetFive)
	room.CoupRoleCounts = staleCounts
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-player-count/increment", nil)
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupPreset != game.CoupPresetSix {
		t.Fatalf("expected increment to select %q, got %q", game.CoupPresetSix, updatedRoom.CoupPreset)
	}
	if updatedRoom.CoupRoleCountsCustom {
		t.Fatal("expected player-count change to reset to preset role counts")
	}
	counts := game.CoupRoleCountsForRoom(updatedRoom)
	if counts[game.RoleBlackKnight] != 2 {
		t.Fatalf("expected 6-player role counts after increment, got %v", counts)
	}

	body := w.Body.String()
	if !strings.Contains(body, "6 players") {
		t.Fatalf("expected response to render 6-player setup, got: %s", body)
	}
	if !strings.Contains(body, `name="blackKnight" value="2"`) {
		t.Fatalf("expected response to render updated Black Knight count, got: %s", body)
	}
}

func TestUpdateCoupRoleCountsSetsCustomCounts(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("king", "1")
	form.Add("blueKnight", "2")
	form.Add("blackKnight", "0")
	form.Add("redKnight", "1")
	form.Add("greenKnight", "1")
	form.Add("wastelandKnight", "0")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-role-counts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupRoleCounts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if !updatedRoom.CoupRoleCountsCustom {
		t.Fatal("expected count edit to switch Coup setup to custom role counts")
	}
	want := game.CoupRoleCounts{
		game.RoleKing:        1,
		game.RoleBlueKnight:  2,
		game.RoleBlackKnight: 0,
		game.RoleRedKnight:   1,
		game.RoleGreenKnight: 1,
		game.RoleWasteland:   0,
	}
	for role, wantCount := range want {
		if updatedRoom.CoupRoleCounts[role] != wantCount {
			t.Fatalf("expected %s count %d, got %d in %v", role, wantCount, updatedRoom.CoupRoleCounts[role], updatedRoom.CoupRoleCounts)
		}
	}

	body := w.Body.String()
	if !strings.Contains(body, "Custom role counts") {
		t.Errorf("expected custom role count state in response, got: %s", body)
	}
}

func TestUpdateCoupRoleCountsSetsUnsafeOverride(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("king", "0")
	form.Add("blueKnight", "2")
	form.Add("blackKnight", "2")
	form.Add("redKnight", "0")
	form.Add("greenKnight", "1")
	form.Add("wastelandKnight", "0")
	form.Add("unsafeRoleCounts", "on")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-role-counts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.UpdateCoupRoleCounts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if !updatedRoom.CoupAllowUnsafeRoleCounts {
		t.Fatal("expected unsafe role count override to be persisted")
	}
	body := w.Body.String()
	if !strings.Contains(body, "game is probably broken") {
		t.Errorf("expected unsafe override warning copy in response, got: %s", body)
	}
}

func TestSetupRouter_RoutesCoupRoleCountUpdates(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("king", "1")
	form.Add("blueKnight", "1")
	form.Add("blackKnight", "1")
	form.Add("redKnight", "1")
	form.Add("greenKnight", "1")
	form.Add("wastelandKnight", "0")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-role-counts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if !updatedRoom.CoupRoleCountsCustom {
		t.Fatal("expected route to set custom Coup role counts")
	}
}

func TestUpdateCoupPresetRejectsFirstActivePlayerWhoIsNotRoomOperator(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.CoupPreset = game.CoupPresetFive
	firstPlayer := game.NewPlayer("p1", "First Player", "session-first")
	operator := game.NewPlayer("p2", "Operator", "session-operator")
	room.AddPlayer(firstPlayer)
	room.AddPlayer(operator)
	room.OperatorSessionID = operator.SessionID
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetEightChaos))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: firstPlayer.SessionID,
	})
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: firstPlayer.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupPreset(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupPreset != game.CoupPresetFive {
		t.Fatalf("expected Coup preset to remain %q, got %q", game.CoupPresetFive, updatedRoom.CoupPreset)
	}
}

func TestSetupRouter_RoutesCoupPresetUpdates(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetSix))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupPreset != game.CoupPresetSix {
		t.Fatalf("expected Coup preset %q, got %q", game.CoupPresetSix, updatedRoom.CoupPreset)
	}
}

func TestUpdateCoupInfoPolicy(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("kingToBlue", string(game.CoupKingGetsBlueCandidates))
	form.Add("redToBlack", string(game.CoupRedKnowsOneBlack))
	form.Add("blackToRed", string(game.CoupBlackToRedAll))
	form.Add("blackNetwork", string(game.CoupBlackNetworkAll))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-info", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupInfoPolicy(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	want := game.CoupInformationPolicy{
		KingToBlue:   game.CoupKingGetsBlueCandidates,
		RedToBlack:   game.CoupRedKnowsOneBlack,
		BlackToRed:   game.CoupBlackToRedAll,
		BlackNetwork: game.CoupBlackNetworkAll,
	}
	if updatedRoom.CoupInfoPolicy != want {
		t.Fatalf("expected Coup info policy %+v, got %+v", want, updatedRoom.CoupInfoPolicy)
	}
}

func TestSetupRouter_RoutesCoupInfoPolicyUpdates(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("blackToRed", string(game.CoupBlackToRedAll))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-info", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupInfoPolicy.BlackToRed != game.CoupBlackToRedAll {
		t.Fatalf("expected Black-to-Red policy %q, got %q", game.CoupBlackToRedAll, updatedRoom.CoupInfoPolicy.BlackToRed)
	}
}

func TestUpdateCoupRoyalGuardSettings(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("blockerLimit", "1")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-royal-guard", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupRoyalGuardSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupRoyalGuardBlockerLimit != 1 {
		t.Fatalf("expected Royal Guard blocker limit 1, got %d", updatedRoom.CoupRoyalGuardBlockerLimit)
	}
}

func TestSetupRouter_RoutesCoupRoyalGuardSettings(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("blockerLimit", "3")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-royal-guard", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupRoyalGuardBlockerLimit != 3 {
		t.Fatalf("expected Royal Guard blocker limit 3, got %d", updatedRoom.CoupRoyalGuardBlockerLimit)
	}
}

func TestUpdateCoupInquisitionSettings(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("resultPolicy", string(game.CoupInquisitionResultPrivate))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-inquisition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateCoupInquisitionSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupInquisitionResultPolicy != game.CoupInquisitionResultPrivate {
		t.Fatalf("expected private Inquisition result policy, got %q", updatedRoom.CoupInquisitionResultPolicy)
	}
}

func TestSetupRouter_RoutesCoupInquisitionSettings(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	markRoomOperatorForTest(room, player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("resultPolicy", string(game.CoupInquisitionResultPrivate))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-inquisition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, player)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupInquisitionResultPolicy != game.CoupInquisitionResultPrivate {
		t.Fatalf("expected private Inquisition result policy, got %q", updatedRoom.CoupInquisitionResultPolicy)
	}
}
