package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

func TestCallCoupInquisition_RevealsBlueAndWaitsForWitness(t *testing.T) {
	h := newTestHandler()
	room, blue, red, _ := setupCoupInquisitionRoom(t, h)

	form := url.Values{}
	form.Add("targetID", red.ID)
	form.Add("currentLife", "39")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/"+blue.ID, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: blue.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", blue.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.CallCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedBlue := updatedRoom.GetPlayer(blue.ID)
	if !updatedBlue.RoleRevealed || !updatedBlue.FaceUp {
		t.Fatal("expected calling Inquisition to publicly reveal Blue")
	}
	updatedRed := updatedRoom.GetPlayer(red.ID)
	if updatedRed.RoleRevealed {
		t.Fatal("expected target result to remain hidden before witness confirmation")
	}
	if updatedRoom.CoupInquisition == nil || updatedRoom.CoupInquisition.Pending == nil {
		t.Fatal("expected pending Inquisition notice")
	}
	if updatedRoom.CoupInquisition.Pending.InquisitorID != blue.ID {
		t.Fatalf("expected pending inquisitor %s, got %s", blue.ID, updatedRoom.CoupInquisition.Pending.InquisitorID)
	}
	if updatedRoom.CoupInquisition.Pending.TargetID != red.ID {
		t.Fatalf("expected pending target %s, got %s", red.ID, updatedRoom.CoupInquisition.Pending.TargetID)
	}
}

func TestCallCoupInquisition_RejectsSecondAttemptBySameBlue(t *testing.T) {
	h := newTestHandler()
	room, blue, red, green := setupCoupInquisitionRoom(t, h)
	callCoupInquisitionForTest(t, h, room, blue, red, 39)
	confirmCoupInquisitionForTest(t, h, room, green)

	form := url.Values{}
	form.Add("targetID", green.ID)
	form.Add("currentLife", "20")
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/"+blue.ID, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: blue.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", blue.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.CallCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 for second Inquisition, got %d: %s", w.Result().StatusCode, w.Body.String())
	}
}

func TestConfirmCoupInquisition_CorrectGuessRevealsRedAndRecordsSuccess(t *testing.T) {
	h := newTestHandler()
	room, blue, red, green := setupCoupInquisitionRoom(t, h)
	callCoupInquisitionForTest(t, h, room, blue, red, 39)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/confirm", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: green.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ConfirmCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedRed := updatedRoom.GetPlayer(red.ID)
	if !updatedRed.RoleRevealed || !updatedRed.FaceUp {
		t.Fatal("expected correct public Inquisition to reveal Red")
	}
	if updatedRoom.CoupInquisition == nil || !updatedRoom.CoupInquisition.Succeeded {
		t.Fatal("expected Inquisition success to be recorded")
	}
	if updatedRoom.CoupInquisition.Pending != nil {
		t.Fatal("expected pending Inquisition to be cleared after confirmation")
	}
	if updatedRoom.CoupInquisition.Last == nil || !updatedRoom.CoupInquisition.Last.Success {
		t.Fatal("expected successful Inquisition result to be recorded")
	}
	if updatedRoom.CoupInquisition.Last.ConfirmedBy != green.ID {
		t.Fatalf("expected witness %s, got %s", green.ID, updatedRoom.CoupInquisition.Last.ConfirmedBy)
	}
}

func TestConfirmCoupInquisition_RejectsBlueWitnessAndKeepsPending(t *testing.T) {
	h := newTestHandler()
	room, blue, red, green := setupCoupInquisitionRoom(t, h)
	secondBlue := game.NewPlayer("blue2", "Second Blue", "session-blue2")
	secondBlue.Role = mockHandlerCoupCard(1002, "Blue Knight")
	secondBlue.RoleRevealed = false
	secondBlue.FaceUp = false
	room.AddPlayer(secondBlue)
	h.store.UpdateRoom(room)

	callCoupInquisitionForTest(t, h, room, blue, red, 39)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/confirm", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: secondBlue.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ConfirmCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403 for Blue witness, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupInquisition == nil || updatedRoom.CoupInquisition.Pending == nil {
		t.Fatal("expected pending Inquisition to remain after rejected Blue witness")
	}
	if updatedRoom.GetPlayer(red.ID).RoleRevealed {
		t.Fatal("expected Red to remain hidden after rejected Blue witness")
	}

	confirmCoupInquisitionForTest(t, h, updatedRoom, green)
}

func TestConfirmCoupInquisition_WrongGuessKeepsTargetHiddenAndRecordsPenalty(t *testing.T) {
	h := newTestHandler()
	room, blue, red, green := setupCoupInquisitionRoom(t, h)
	callCoupInquisitionForTest(t, h, room, blue, green, 39)

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/confirm", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: red.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ConfirmCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	updatedGreen := updatedRoom.GetPlayer(green.ID)
	if updatedGreen.RoleRevealed || updatedGreen.FaceUp {
		t.Fatal("expected wrong Inquisition target to remain hidden")
	}
	if updatedRoom.CoupInquisition == nil || updatedRoom.CoupInquisition.Succeeded {
		t.Fatal("expected wrong Inquisition not to record success")
	}
	if updatedRoom.CoupInquisition.Last == nil {
		t.Fatal("expected failed Inquisition result to be recorded")
	}
	if updatedRoom.CoupInquisition.Last.Success {
		t.Fatal("expected failed Inquisition result")
	}
	if updatedRoom.CoupInquisition.Last.PenaltyLife != 20 {
		t.Fatalf("expected penalty 20, got %d", updatedRoom.CoupInquisition.Last.PenaltyLife)
	}
}

func TestSetupRouter_RoutesCoupInquisitionFlow(t *testing.T) {
	h := newTestHandler()
	room, blue, red, green := setupCoupInquisitionRoom(t, h)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("targetID", red.ID)
	form.Add("currentLife", "39")
	callReq := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/"+blue.ID, strings.NewReader(form.Encode()))
	callReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	callReq.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: blue.ID,
	})
	callW := httptest.NewRecorder()
	router.ServeHTTP(callW, callReq)

	if callW.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected call status 200, got %d: %s", callW.Result().StatusCode, callW.Body.String())
	}

	confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/confirm", nil)
	confirmReq.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: green.ID,
	})
	confirmW := httptest.NewRecorder()
	router.ServeHTTP(confirmW, confirmReq)

	if confirmW.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected confirm status 200, got %d: %s", confirmW.Result().StatusCode, confirmW.Body.String())
	}
}

func callCoupInquisitionForTest(t *testing.T, h *Handler, room *game.Room, blue *game.Player, target *game.Player, currentLife int) {
	t.Helper()

	form := url.Values{}
	form.Add("targetID", target.ID)
	form.Add("currentLife", strconv.Itoa(currentLife))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/"+blue.ID, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: blue.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", blue.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.CallCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected call Inquisition status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}
}

func confirmCoupInquisitionForTest(t *testing.T, h *Handler, room *game.Room, witness *game.Player) {
	t.Helper()

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/coup/inquisition/confirm", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: witness.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ConfirmCoupInquisition(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected confirm Inquisition status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}
}

func setupCoupInquisitionRoom(t *testing.T, h *Handler) (*game.Room, *game.Player, *game.Player, *game.Player) {
	t.Helper()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	blue := game.NewPlayer("blue", "Blue Player", "session-blue")
	blue.Role = mockHandlerCoupCard(1002, "Blue Knight")
	blue.RoleRevealed = false
	blue.FaceUp = false

	red := game.NewPlayer("red", "Red Player", "session-red")
	red.Role = mockHandlerCoupCard(1004, "Red Knight")
	red.RoleRevealed = false
	red.FaceUp = false

	green := game.NewPlayer("green", "Green Player", "session-green")
	green.Role = mockHandlerCoupCard(1005, "Green Knight")
	green.RoleRevealed = false
	green.FaceUp = false

	room.AddPlayer(blue)
	room.AddPlayer(red)
	room.AddPlayer(green)
	h.store.UpdateRoom(room)

	return room, blue, red, green
}
