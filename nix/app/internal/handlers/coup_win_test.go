package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

func TestConfirmCoupWinPrompt_ActivePlayerEndsCoupGame(t *testing.T) {
	h := newTestHandler()
	room, actor := setupCoupWinHandlerRoom(t, h)

	req := newCoupWinRequest(room, actor, "/room/"+room.Code+"/coup/win/confirm")
	w := httptest.NewRecorder()

	h.ConfirmCoupWinPrompt(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StateEnded {
		t.Fatalf("expected game to end, got state %s", updatedRoom.State)
	}
	if updatedRoom.CoupWin == nil || updatedRoom.CoupWin.Confirmed == nil {
		t.Fatal("expected confirmed Coup win to be recorded")
	}
	if updatedRoom.CoupWin.Confirmed.Outcome != game.CoupWinOutcomeBlack {
		t.Fatalf("expected Black confirmed win, got %q", updatedRoom.CoupWin.Confirmed.Outcome)
	}
}

func TestRejectCoupWinPrompt_ActivePlayerKeepsGamePlayingAndSuppressesPrompt(t *testing.T) {
	h := newTestHandler()
	room, actor := setupCoupWinHandlerRoom(t, h)

	req := newCoupWinRequest(room, actor, "/room/"+room.Code+"/coup/win/reject")
	w := httptest.NewRecorder()

	h.RejectCoupWinPrompt(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StatePlaying {
		t.Fatalf("expected game to keep playing, got state %s", updatedRoom.State)
	}
	if updatedRoom.CoupWin == nil || updatedRoom.CoupWin.RejectedAdvisoryID == "" {
		t.Fatal("expected rejected advisory id to be recorded")
	}
	if current := game.CurrentCoupAdvisoryWin(updatedRoom); current != nil {
		t.Fatalf("expected rejected current prompt to be suppressed, got %#v", current)
	}
	if detected := game.DetectCoupAdvisoryWin(updatedRoom); detected == nil {
		t.Fatal("expected raw detector still to see the tracked-state candidate")
	}
}

func TestConfirmCoupWinPrompt_EliminatedPlayerCannotConfirm(t *testing.T) {
	h := newTestHandler()
	room, _ := setupCoupWinHandlerRoom(t, h)
	eliminated := room.GetPlayer("red")

	req := newCoupWinRequest(room, eliminated, "/room/"+room.Code+"/coup/win/confirm")
	w := httptest.NewRecorder()

	h.ConfirmCoupWinPrompt(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.State != game.StatePlaying {
		t.Fatalf("expected game to keep playing, got state %s", updatedRoom.State)
	}
}

func TestSetupRouter_RoutesCoupWinPromptDecisions(t *testing.T) {
	h := newTestHandler()
	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	confirmRoom, confirmActor := setupCoupWinHandlerRoom(t, h)
	confirmReq := httptest.NewRequest("POST", "/room/"+confirmRoom.Code+"/coup/win/confirm", nil)
	confirmReq.AddCookie(&http.Cookie{Name: "player_" + confirmRoom.Code, Value: confirmActor.ID})
	confirmW := httptest.NewRecorder()

	router.ServeHTTP(confirmW, confirmReq)

	if confirmW.Code != http.StatusOK {
		t.Fatalf("expected confirm route status 200, got %d: %s", confirmW.Code, confirmW.Body.String())
	}
	updatedConfirmRoom, _ := h.store.GetRoom(confirmRoom.Code)
	if updatedConfirmRoom.State != game.StateEnded {
		t.Fatalf("expected confirm route to end game, got %s", updatedConfirmRoom.State)
	}

	rejectRoom, rejectActor := setupCoupWinHandlerRoom(t, h)
	rejectReq := httptest.NewRequest("POST", "/room/"+rejectRoom.Code+"/coup/win/reject", nil)
	rejectReq.AddCookie(&http.Cookie{Name: "player_" + rejectRoom.Code, Value: rejectActor.ID})
	rejectW := httptest.NewRecorder()

	router.ServeHTTP(rejectW, rejectReq)

	if rejectW.Code != http.StatusOK {
		t.Fatalf("expected reject route status 200, got %d: %s", rejectW.Code, rejectW.Body.String())
	}
	updatedRejectRoom, _ := h.store.GetRoom(rejectRoom.Code)
	if updatedRejectRoom.State != game.StatePlaying {
		t.Fatalf("expected reject route to keep game playing, got %s", updatedRejectRoom.State)
	}
	if updatedRejectRoom.CoupWin == nil || updatedRejectRoom.CoupWin.RejectedAdvisoryID == "" {
		t.Fatal("expected reject route to record rejected advisory")
	}
}

func TestEliminatePlayer_CoupKingFallLocksGreenEligibilityBeforeBlueDies(t *testing.T) {
	h := newTestHandler()
	room, host := setupCoupRedGreenEligibilityRoom(t, h)

	w := httptest.NewRecorder()
	h.EliminatePlayer(w, newCoupEliminateRequest(room, host, "king"))

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedRoom, _ := h.store.GetRoom(room.Code)
	if !updatedRoom.CoupKingFallen {
		t.Fatal("expected King fall to be recorded")
	}
	if updatedRoom.CoupGreenEligibleBeforeKingFall {
		t.Fatal("expected Green not to become Red-eligible while Blue was alive at King fall")
	}

	updatedRoom.GetPlayer("blue").MarkEliminated()
	h.store.UpdateRoom(updatedRoom)
	prompt := game.DetectCoupAdvisoryWin(updatedRoom)
	if prompt == nil || prompt.Outcome != game.CoupWinOutcomeRed {
		t.Fatalf("expected Red advisory prompt, got %#v", prompt)
	}
	if prompt.GreenShares {
		t.Fatal("expected Green not to share Red victory after Blue died post-King-fall")
	}
}

func setupCoupWinHandlerRoom(t *testing.T, h *Handler) (*game.Room, *game.Player) {
	t.Helper()

	room, err := h.store.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	king := game.NewPlayer("king", "King Player", "session-king")
	king.Role = mockHandlerCoupCard(1001, "King")
	king.MarkEliminated()

	black := game.NewPlayer("black", "Black Player", "session-black")
	black.Role = mockHandlerCoupCard(1003, "Black Knight")

	red := game.NewPlayer("red", "Red Player", "session-red")
	red.Role = mockHandlerCoupCard(1004, "Red Knight")
	red.MarkEliminated()

	for _, player := range []*game.Player{king, black, red} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("add player %s: %v", player.ID, err)
		}
	}
	h.store.UpdateRoom(room)
	return room, black
}

func setupCoupRedGreenEligibilityRoom(t *testing.T, h *Handler) (*game.Room, *game.Player) {
	t.Helper()

	room, err := h.store.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.RulesMode = game.RulesModeCoup
	room.State = game.StatePlaying

	host := game.NewPlayer("host", "Host", "session-host")
	host.IsHost = true

	king := game.NewPlayer("king", "King Player", "session-king")
	king.Role = mockHandlerCoupCard(1001, "King")

	blue := game.NewPlayer("blue", "Blue Player", "session-blue")
	blue.Role = mockHandlerCoupCard(1002, "Blue Knight")

	black := game.NewPlayer("black", "Black Player", "session-black")
	black.Role = mockHandlerCoupCard(1003, "Black Knight")
	black.MarkEliminated()

	red := game.NewPlayer("red", "Red Player", "session-red")
	red.Role = mockHandlerCoupCard(1004, "Red Knight")

	green := game.NewPlayer("green", "Green Player", "session-green")
	green.Role = mockHandlerCoupCard(1005, "Green Knight")

	for _, player := range []*game.Player{host, king, blue, black, red, green} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("add player %s: %v", player.ID, err)
		}
	}
	h.store.UpdateRoom(room)
	return room, host
}

func newCoupWinRequest(room *game.Room, player *game.Player, path string) *http.Request {
	req := httptest.NewRequest("POST", path, nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func newCoupEliminateRequest(room *game.Room, player *game.Player, targetID string) *http.Request {
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/"+targetID+"/eliminate", nil)
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: player.ID,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", targetID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
