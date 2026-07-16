package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"treacherest/internal/game"
)

func TestPlaceDebtCounterOwnerEndpointPublishesPublicEvent(t *testing.T) {
	h, room, collector, target := testDebtCounterHandlerRoom(t)
	events := h.eventBus.Subscribe(room.Code)
	defer h.eventBus.Unsubscribe(room.Code, events)

	response := httptest.NewRecorder()
	h.PlaceDebtCounter(response, debtCounterRequest(room, collector, collector, "place", target.ID))
	if response.Code != http.StatusNoContent {
		t.Fatalf("placement status = %d, want %d: %s", response.Code, http.StatusNoContent, response.Body.String())
	}
	if target.Role.DebtCounters == nil || target.Role.DebtCounters.Count != 1 {
		t.Fatalf("target debt state = %+v", target.Role.DebtCounters)
	}
	select {
	case event := <-events:
		if event.Type != "debt_counter_placed" {
			t.Fatalf("event type = %q", event.Type)
		}
		placement, ok := event.Data.(*game.DebtCounterPlacement)
		if !ok || placement.TargetPlayerID != target.ID {
			t.Fatalf("public event data = %#v", event.Data)
		}
	default:
		t.Fatal("placement endpoint did not publish a public room update")
	}

	foreign := httptest.NewRecorder()
	h.PlaceDebtCounter(foreign, debtCounterRequest(room, collector, target, "place", target.ID))
	if foreign.Code != http.StatusForbidden {
		t.Fatalf("foreign placement status = %d, want %d", foreign.Code, http.StatusForbidden)
	}

	self := httptest.NewRecorder()
	h.PlaceDebtCounter(self, debtCounterRequest(room, collector, collector, "place", collector.ID))
	if self.Code != http.StatusConflict {
		t.Fatalf("self placement status = %d, want %d", self.Code, http.StatusConflict)
	}
}

func TestDebtCounterUpkeepAndEliminationHandlersRecordAdvisories(t *testing.T) {
	h, room, collector, target := testDebtCounterHandlerRoom(t)
	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		h.PlaceDebtCounter(response, debtCounterRequest(room, collector, collector, "place", target.ID))
		if response.Code != http.StatusNoContent {
			t.Fatalf("placement %d failed: %d %s", i+1, response.Code, response.Body.String())
		}
	}

	upkeep := httptest.NewRecorder()
	h.RecordDebtCounterUpkeep(upkeep, debtCounterRequest(room, collector, collector, "upkeep", ""))
	if upkeep.Code != http.StatusNoContent {
		t.Fatalf("upkeep status = %d, want %d: %s", upkeep.Code, http.StatusNoContent, upkeep.Body.String())
	}
	if logs := collector.Role.DebtCounters.UpkeepAdvisories; len(logs) != 1 || logs[0].TotalLifeGain != 2 {
		t.Fatalf("upkeep advisories = %+v", logs)
	}

	eliminate := httptest.NewRecorder()
	h.EliminatePlayer(eliminate, eliminateDebtCounterTargetRequest(room, target))
	if eliminate.Code != http.StatusOK {
		t.Fatalf("elimination status = %d, want %d: %s", eliminate.Code, http.StatusOK, eliminate.Body.String())
	}
	if !target.IsEliminated {
		t.Fatal("target was not marked eliminated")
	}
	if logs := collector.Role.DebtCounters.LossAdvisories; len(logs) != 1 || logs[0].DrawCount != 2 {
		t.Fatalf("loss advisories = %+v", logs)
	}
	if target.Role.DebtCounters.Count != 2 {
		t.Fatal("elimination advisory changed the public counter count")
	}
}

func testDebtCounterHandlerRoom(t *testing.T) (*Handler, *game.Room, *game.Player, *game.Player) {
	t.Helper()
	h := newTestHandler()
	room, err := h.store.CreateRoom()
	if err != nil {
		t.Fatal(err)
	}
	room.State = game.StatePlaying
	collector := game.NewPlayer("collector", "Alice", "collector-session")
	collector.Role = game.NewDealtRoleCard(&game.Card{
		ID: game.TheDebtCollectorCardID, Name: "The Debt Collector", Types: game.CardTypes{Subtype: "Leader"},
	}, 3)
	collector.FaceUp = true
	collector.RoleRevealed = true
	target := game.NewPlayer("target", "Bob", "target-session")
	target.Role = game.NewDealtRoleCard(&game.Card{
		ID: 7, Name: "Hidden Target", Types: game.CardTypes{Subtype: "Guardian"},
	}, 3)
	target.FaceUp = false
	target.RoleRevealed = false
	if err := room.AddPlayer(collector); err != nil {
		t.Fatal(err)
	}
	if err := room.AddPlayer(target); err != nil {
		t.Fatal(err)
	}
	h.store.UpdateRoom(room)
	return h, room, collector, target
}

func debtCounterRequest(room *game.Room, collector, actor *game.Player, action, targetID string) *http.Request {
	path := fmt.Sprintf("/room/%s/player/%s/debt-counter/%s", room.Code, collector.ID, action)
	if targetID != "" {
		path += "/" + targetID
	}
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: actor.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", collector.ID)
	if targetID != "" {
		rctx.URLParams.Add("targetID", targetID)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func eliminateDebtCounterTargetRequest(room *game.Room, target *game.Player) *http.Request {
	path := fmt.Sprintf("/room/%s/player/%s/eliminate", room.Code, target.ID)
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: target.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", target.ID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
