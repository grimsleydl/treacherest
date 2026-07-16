package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"
)

func TestStreamHostDebtCounterEventRefreshesPublicCountWithoutRoleContents(t *testing.T) {
	cfg := config.DefaultConfig()
	gameStore := store.NewMemoryStore(cfg)
	h := New(gameStore, createMockCardService(), cfg, nil)
	room, err := gameStore.CreateRoom()
	require.NoError(t, err)
	room.State = game.StatePlaying

	host := game.NewPlayer("host", "Host", "host-session")
	host.IsHost = true
	collector := game.NewPlayer("collector", "Alice", "collector-session")
	collector.Role = game.NewDealtRoleCard(&game.Card{
		ID: game.TheDebtCollectorCardID, Name: "The Debt Collector", Types: game.CardTypes{Subtype: "Leader"},
	}, 3)
	collector.FaceUp = true
	collector.RoleRevealed = true
	target := game.NewPlayer("target", "Bob", "target-session")
	target.Role = game.NewDealtRoleCard(&game.Card{
		ID: 777, Name: "SSE SECRET ROLE", Type: "Identity — SSE Secret", Text: "SSE SECRET TEXT", Types: game.CardTypes{Subtype: "Guardian"},
	}, 3)
	target.FaceUp = false
	target.RoleRevealed = false
	room.OperatorSessionID = host.SessionID
	require.NoError(t, room.AddPlayer(host))
	require.NoError(t, room.AddPlayer(collector))
	require.NoError(t, room.AddPlayer(target))
	require.NoError(t, gameStore.UpdateRoom(room))

	req := httptest.NewRequest(http.MethodGet, "/sse/host/"+room.Code, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: host.ID})
	req.AddCookie(&http.Cookie{Name: "host_" + room.Code, Value: "true"})
	req.AddCookie(&http.Cookie{Name: "session", Value: host.SessionID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		h.StreamHost(w, req)
		close(done)
	}()
	require.Eventually(t, func() bool {
		h.eventBus.mu.RLock()
		defer h.eventBus.mu.RUnlock()
		return len(h.eventBus.subscribers[room.Code]) > 0
	}, time.Second, 10*time.Millisecond, "operator SSE did not subscribe")

	placement, err := room.PlaceDebtCounter(collector.ID, target.ID)
	require.NoError(t, err)
	require.NoError(t, gameStore.UpdateRoom(room))
	h.eventBus.Publish(Event{Type: "debt_counter_placed", RoomCode: room.Code, Data: placement})

	require.Eventually(t, func() bool {
		return strings.Contains(w.Body.String(), `data-debt-counter-count="1"`)
	}, time.Second, 10*time.Millisecond, "operator dashboard should refresh immediately with the public debt count")
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("operator SSE did not stop after cancellation")
	}
	body := w.Body.String()
	for _, secret := range []string{"SSE SECRET ROLE", "SSE SECRET TEXT", "Identity — SSE Secret"} {
		require.NotContains(t, body, secret, "operator SSE leaked hidden role contents")
	}
}
