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

func TestActivateLeaderCounterOwnerEndpoint(t *testing.T) {
	tests := []struct {
		id   int
		name string
	}{
		{game.HerSeedbornHighnessCardID, "Her Seedborn Highness"},
		{game.TheLichQueenCardID, "The Lich Queen"},
		{game.TheQueenOfLightCardID, "The Queen of Light"},
		{game.TheVoidTyrantCardID, "The Void Tyrant"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			room, _ := h.store.CreateRoom()
			room.State = game.StatePlaying
			owner := game.NewPlayer("owner", "Alice", "owner-session")
			owner.Role = game.NewDealtRoleCard(&game.Card{ID: tt.id, Name: tt.name, Types: game.CardTypes{Subtype: "Leader"}}, 6)
			owner.RoleRevealed = true
			owner.FaceUp = true
			if err := room.AddPlayer(owner); err != nil {
				t.Fatal(err)
			}
			h.store.UpdateRoom(room)

			events := h.eventBus.Subscribe(room.Code)
			defer h.eventBus.Unsubscribe(room.Code, events)
			request := leaderCounterRequest(room, owner, "activate")
			response := httptest.NewRecorder()
			h.ActivateLeaderCounter(response, request)

			if response.Code != http.StatusNoContent {
				t.Fatalf("activation status = %d, want %d: %s", response.Code, http.StatusNoContent, response.Body.String())
			}
			if got := owner.Role.LeaderCounterPool.Remaining; got != owner.Role.LeaderCounterPool.Initial-1 {
				t.Fatalf("remaining counters = %d, want %d", got, owner.Role.LeaderCounterPool.Initial-1)
			}
			if len(owner.Role.LeaderCounterPool.Activations) != 1 {
				t.Fatal("activation endpoint did not persist its public event")
			}
			select {
			case event := <-events:
				if event.Type != "leader_counter_activated" {
					t.Fatalf("event type = %q", event.Type)
				}
			default:
				t.Fatal("activation endpoint did not publish a public room update")
			}
		})
	}
}

func TestLeaderCounterEndpointEnforcesOwnerAndOncePerTurnReset(t *testing.T) {
	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.State = game.StatePlaying
	owner := game.NewPlayer("owner", "Alice", "owner-session")
	owner.Role = game.NewDealtRoleCard(&game.Card{ID: game.HerSeedbornHighnessCardID, Name: "Her Seedborn Highness", Types: game.CardTypes{Subtype: "Leader"}}, 6)
	owner.RoleRevealed = true
	owner.FaceUp = true
	other := game.NewPlayer("other", "Bob", "other-session")
	_ = room.AddPlayer(owner)
	_ = room.AddPlayer(other)
	h.store.UpdateRoom(room)

	first := httptest.NewRecorder()
	h.ActivateLeaderCounter(first, leaderCounterRequest(room, owner, "activate"))
	second := httptest.NewRecorder()
	h.ActivateLeaderCounter(second, leaderCounterRequest(room, owner, "activate"))
	if second.Code != http.StatusConflict {
		t.Fatalf("same-turn activation status = %d, want %d", second.Code, http.StatusConflict)
	}

	reset := httptest.NewRecorder()
	h.ResetLeaderCounterTurn(reset, leaderCounterRequest(room, owner, "reset-turn"))
	if reset.Code != http.StatusNoContent {
		t.Fatalf("turn reset status = %d, want %d: %s", reset.Code, http.StatusNoContent, reset.Body.String())
	}
	afterReset := httptest.NewRecorder()
	h.ActivateLeaderCounter(afterReset, leaderCounterRequest(room, owner, "activate"))
	if afterReset.Code != http.StatusNoContent {
		t.Fatalf("activation after reset status = %d, want %d", afterReset.Code, http.StatusNoContent)
	}

	foreign := httptest.NewRecorder()
	h.ActivateLeaderCounter(foreign, leaderCounterRequestAs(room, owner, other, "activate"))
	if foreign.Code != http.StatusForbidden {
		t.Fatalf("foreign activation status = %d, want %d", foreign.Code, http.StatusForbidden)
	}
}

func leaderCounterRequest(room *game.Room, owner *game.Player, action string) *http.Request {
	return leaderCounterRequestAs(room, owner, owner, action)
}

func leaderCounterRequestAs(room *game.Room, owner, actor *game.Player, action string) *http.Request {
	path := fmt.Sprintf("/room/%s/player/%s/leader-counter/%s", room.Code, owner.ID, action)
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: actor.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", owner.ID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
