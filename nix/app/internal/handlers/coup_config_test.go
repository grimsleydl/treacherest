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
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetEightChaos))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
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

	body := w.Body.String()
	if !strings.Contains(body, "8 players, chaos") {
		t.Errorf("expected updated Coup preset label in response, got: %s", body)
	}
	if !strings.Contains(body, "Wasteland Knight") {
		t.Errorf("expected updated Coup role summary in response, got: %s", body)
	}
}

func TestSetupRouter_RoutesCoupPresetUpdates(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	player := game.NewPlayer("p1", "Player 1", "session1")
	room.AddPlayer(player)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetSix))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
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
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("kingToBlue", string(game.CoupKingGetsBlueCandidates))
	form.Add("redToBlack", string(game.CoupRedKnowsOneBlack))
	form.Add("blackToRed", string(game.CoupBlackToRedAll))
	form.Add("blackNetwork", string(game.CoupBlackNetworkAll))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-info", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
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
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("blackToRed", string(game.CoupBlackToRedAll))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-info", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
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
	h.store.UpdateRoom(room)

	form := url.Values{}
	form.Add("blockerLimit", "1")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-royal-guard", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
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
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("blockerLimit", "3")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-royal-guard", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
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
