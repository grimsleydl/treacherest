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

func TestChooseGatheringModeOwnerEndpointPublishesPublicEvent(t *testing.T) {
	h := newTestHandler()
	room, owner, _ := gatheringHandlerRoom(t, h)
	events := h.eventBus.Subscribe(room.Code)
	defer h.eventBus.Unsubscribe(room.Code, events)

	response := httptest.NewRecorder()
	h.ChooseGatheringMode(response, gatheringRequest(room, owner, owner, game.GatheringModeWhite))
	if response.Code != http.StatusNoContent {
		t.Fatalf("choice status = %d, want %d: %s", response.Code, http.StatusNoContent, response.Body.String())
	}
	if !owner.Role.GatheringChecklist.HasChosen(game.GatheringModeWhite) {
		t.Fatal("choice endpoint did not persist the chosen mode")
	}
	select {
	case event := <-events:
		if event.Type != "gathering_mode_chosen" {
			t.Fatalf("event type = %q", event.Type)
		}
		choice, ok := event.Data.(*game.GatheringModeChoice)
		if !ok || choice.Mode != game.GatheringModeWhite || choice.PlayerName != owner.Name {
			t.Fatalf("public event data = %#v", event.Data)
		}
	default:
		t.Fatal("choice endpoint did not publish a public room update")
	}
}

func TestChooseGatheringModeEndpointRejectsChosenInvalidAndForeignChoices(t *testing.T) {
	h := newTestHandler()
	room, owner, other := gatheringHandlerRoom(t, h)

	first := httptest.NewRecorder()
	h.ChooseGatheringMode(first, gatheringRequest(room, owner, owner, game.GatheringModeBlue))
	duplicate := httptest.NewRecorder()
	h.ChooseGatheringMode(duplicate, gatheringRequest(room, owner, owner, game.GatheringModeBlue))
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate choice status = %d, want %d", duplicate.Code, http.StatusConflict)
	}

	invalid := httptest.NewRecorder()
	h.ChooseGatheringMode(invalid, gatheringRequest(room, owner, owner, game.GatheringMode("purple")))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid choice status = %d, want %d", invalid.Code, http.StatusBadRequest)
	}

	foreign := httptest.NewRecorder()
	h.ChooseGatheringMode(foreign, gatheringRequest(room, owner, other, game.GatheringModeRed))
	if foreign.Code != http.StatusForbidden {
		t.Fatalf("foreign choice status = %d, want %d", foreign.Code, http.StatusForbidden)
	}
	if got := len(owner.Role.GatheringChecklist.Choices); got != 1 {
		t.Fatalf("rejected requests changed checklist length to %d", got)
	}
}

func gatheringHandlerRoom(t *testing.T, h *Handler) (*game.Room, *game.Player, *game.Player) {
	t.Helper()
	room, err := h.store.CreateRoom()
	if err != nil {
		t.Fatal(err)
	}
	room.State = game.StatePlaying
	owner := game.NewPlayer("owner", "Alice", "owner-session")
	owner.Role = game.NewDealtRoleCard(&game.Card{
		ID: game.TheGatheringCardID, Name: "The Gathering", Types: game.CardTypes{Subtype: "Leader"},
	}, 3)
	owner.FaceUp = true
	owner.RoleRevealed = true
	other := game.NewPlayer("other", "Bob", "other-session")
	if err := room.AddPlayer(owner); err != nil {
		t.Fatal(err)
	}
	if err := room.AddPlayer(other); err != nil {
		t.Fatal(err)
	}
	if err := h.store.UpdateRoom(room); err != nil {
		t.Fatal(err)
	}
	return room, owner, other
}

func gatheringRequest(room *game.Room, owner, actor *game.Player, mode game.GatheringMode) *http.Request {
	path := fmt.Sprintf("/room/%s/player/%s/gathering/choose/%s", room.Code, owner.ID, mode)
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: actor.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", owner.ID)
	rctx.URLParams.Add("mode", string(mode))
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
