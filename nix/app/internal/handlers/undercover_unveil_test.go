package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"treacherest/internal/game"

	"github.com/go-chi/chi/v5"
)

const undercoverAdvisoryText = "The app can only verify the unveiled-identity condition."

func TestGetUnveilModal_UndercoverAdvisoryTracksOtherPublicReveal(t *testing.T) {
	for _, card := range []struct {
		id   int
		name string
	}{
		{id: 17, name: "The Supplier"},
		{id: 18, name: "The Warlock"},
	} {
		t.Run(card.name, func(t *testing.T) {
			h := newTestHandler()
			room, owner, other := setupUndercoverRoom(t, h, card.id, card.name)

			body := getUndercoverUnveilModal(t, h, room, owner)
			assertUndercoverUnveilFlow(t, body, room, owner)
			if !strings.Contains(body, `id="undercover-unveil-advisory"`) {
				t.Fatalf("expected Undercover advisory when no other identity is revealed: %s", body)
			}
			for _, expected := range []string{
				"No other identity has been unveiled yet.",
				"another player attacked the Leader this game (table-adjudicated)",
				undercoverAdvisoryText,
			} {
				if !strings.Contains(body, expected) {
					t.Fatalf("expected Undercover advisory text %q: %s", expected, body)
				}
			}

			other.RoleRevealed = true
			h.store.UpdateRoom(room)

			body = getUndercoverUnveilModal(t, h, room, owner)
			// Assert the containing owner flow exists before checking that only
			// the advisory cleared, so the absence assertion cannot be vacuous.
			assertUndercoverUnveilFlow(t, body, room, owner)
			if strings.Contains(body, `id="undercover-unveil-advisory"`) || strings.Contains(body, undercoverAdvisoryText) {
				t.Fatalf("expected Undercover advisory to clear after another public reveal: %s", body)
			}
		})
	}
}

func TestUnveilPlayer_UndercoverNeverBlocked(t *testing.T) {
	for _, cardID := range []int{17, 18} {
		for _, otherRevealed := range []bool{false, true} {
			name := fmt.Sprintf("card_%d/other_revealed_%t", cardID, otherRevealed)
			t.Run(name, func(t *testing.T) {
				h := newTestHandler()
				room, owner, other := setupUndercoverRoom(t, h, cardID, "Undercover Guardian")
				other.RoleRevealed = otherRevealed
				h.store.UpdateRoom(room)

				req := undercoverRequest(http.MethodPost, "/room/"+room.Code+"/unveil/"+owner.ID, room, owner)
				w := httptest.NewRecorder()
				h.UnveilPlayer(w, req)

				if w.Result().StatusCode != http.StatusOK {
					t.Fatalf("expected Undercover unveil to remain available, got %d: %s", w.Result().StatusCode, w.Body.String())
				}
				updatedRoom, err := h.store.GetRoom(room.Code)
				if err != nil {
					t.Fatalf("get updated room: %v", err)
				}
				updatedOwner := updatedRoom.GetPlayer(owner.ID)
				if !updatedOwner.FaceUp || !updatedOwner.RoleRevealed {
					t.Fatalf("expected Undercover unveil to function, got FaceUp=%v RoleRevealed=%v", updatedOwner.FaceUp, updatedOwner.RoleRevealed)
				}
			})
		}
	}
}

func setupUndercoverRoom(t *testing.T, h *Handler, cardID int, cardName string) (*game.Room, *game.Player, *game.Player) {
	t.Helper()

	room, err := h.store.CreateRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.State = game.StatePlaying

	owner := game.NewPlayer("owner", "Owner", "session-owner")
	owner.Role = undercoverCard(cardID, cardName)
	owner.FaceUp = false
	owner.RoleRevealed = false
	other := game.NewPlayer("other", "Other Player", "session-other")
	other.Role = undercoverCard(2, "Other Identity")
	other.FaceUp = false
	other.RoleRevealed = false

	if err := room.AddPlayer(owner); err != nil {
		t.Fatalf("add owner: %v", err)
	}
	if err := room.AddPlayer(other); err != nil {
		t.Fatalf("add other player: %v", err)
	}
	h.store.UpdateRoom(room)
	return room, owner, other
}

func undercoverCard(id int, name string) *game.Card {
	return &game.Card{
		ID:   id,
		Name: name,
		Types: game.CardTypes{
			Supertype: "Identity",
			Subtype:   "Guardian",
		},
	}
}

func getUndercoverUnveilModal(t *testing.T, h *Handler, room *game.Room, owner *game.Player) string {
	t.Helper()

	req := undercoverRequest(http.MethodGet, "/room/"+room.Code+"/unveil-modal/"+owner.ID, room, owner)
	w := httptest.NewRecorder()
	h.GetUnveilModal(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected unveil modal status 200, got %d: %s", w.Result().StatusCode, w.Body.String())
	}
	return w.Body.String()
}

func undercoverRequest(method, path string, room *game.Room, owner *game.Player) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	addPlayerSessionCookiesForTest(req, room, owner)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", owner.ID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func assertUndercoverUnveilFlow(t *testing.T, body string, room *game.Room, owner *game.Player) {
	t.Helper()

	for _, expected := range []string{
		`id="undercover-unveil-flow"`,
		`id="undercover-unveil-confirm"`,
		"/room/" + room.Code + "/unveil/" + owner.ID,
		"Confirm Unveil",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected owner Undercover unveil flow detail %q: %s", expected, body)
		}
	}
}
