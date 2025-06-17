package handlers

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
	"net/http"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"
)

// StreamLobby streams lobby updates
func (h *Handler) StreamLobby(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Create SSE connection
	sse := datastar.NewSSE(w, r)

	// Subscribe to events
	events := h.eventBus.Subscribe(roomCode)
	defer h.eventBus.Unsubscribe(roomCode, events)

	// Send initial render
	h.renderLobby(sse, room, player)

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-events:
			switch event.Type {
			case "player_joined", "player_left":
				// Re-render lobby
				room, _ = h.store.GetRoom(roomCode)
				h.renderLobby(sse, room, player)
			case "game_started":
				// Redirect to game page
				sse.ExecuteScript("window.location.href = '/game/" + roomCode + "'")
			}
		}
	}
}

// StreamGame streams game updates
func (h *Handler) StreamGame(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Create SSE connection
	sse := datastar.NewSSE(w, r)

	// Subscribe to events
	events := h.eventBus.Subscribe(roomCode)
	defer h.eventBus.Unsubscribe(roomCode, events)

	// Send initial render
	h.renderGame(sse, room, player)

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			return
		case <-events:
			// Re-render on any event
			room, _ = h.store.GetRoom(roomCode)
			player = room.GetPlayer(player.ID) // Refresh player data
			h.renderGame(sse, room, player)
		}
	}
}

// renderLobby renders the lobby body
func (h *Handler) renderLobby(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	component := pages.LobbyBody(room, player)

	// Render to string
	html := renderToString(component)

	// Send as fragment with morph mode
	sse.MergeFragments(html, datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}

// renderGame renders the game body
func (h *Handler) renderGame(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	component := pages.GameBody(room, player)

	// Render to string
	html := renderToString(component)

	// Send as fragment with morph mode
	sse.MergeFragments(html, datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}

// renderToString renders a templ component to string
func renderToString(component templ.Component) string {
	buf := &bytes.Buffer{}
	component.Render(context.Background(), buf)
	return buf.String()
}
