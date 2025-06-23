package handlers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
	"log"
	"net/http"
	"strings"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"
)

// StreamLobby streams lobby updates
func (h *Handler) StreamLobby(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	log.Printf("📡 SSE connection established for lobby %s", roomCode)

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("📡 SSE requested for non-existent room: %s", roomCode)
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

	// Don't send initial render - page already has correct content
	// SSE will only send updates when events occur
	log.Printf("📡 SSE connection ready for room %s, waiting for events", roomCode)

	// Set up a heartbeat to detect stale connections
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			log.Printf("📡 Lobby SSE context cancelled for room %s", roomCode)
			return
		case <-heartbeat.C:
			// Check if room still exists
			_, err := h.store.GetRoom(roomCode)
			if err != nil {
				log.Printf("📡 Heartbeat: Room %s no longer exists, closing SSE", roomCode)
				return
			}
			// Send SSE comment heartbeat to keep connection alive
			// Comments don't trigger any client-side processing
			fmt.Fprintf(w, ": heartbeat\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case event := <-events:
			log.Printf("📡 SSE event received for %s: %s", roomCode, event.Type)

			switch event.Type {
			case "player_joined", "player_left", "player_updated":
				// Re-render lobby only if still in lobby state
				room, _ = h.store.GetRoom(roomCode)
				if room.State == game.StateLobby {
					// Refresh player reference in case it was updated
					player = room.GetPlayer(player.ID)
					if player == nil {
						// Player was removed, close SSE connection gracefully
						log.Printf("📡 Player no longer in room %s, closing SSE", roomCode)
						return
					}
					h.renderLobby(sse, room, player)
				} else {
					log.Printf("🎮 Lobby event received but room %s not in lobby state, closing SSE", roomCode)
					return
				}
			case "game_started":
				// Redirect to game page when game starts
				log.Printf("🎮 Game started - redirecting to game page for room %s", roomCode)
				sse.ExecuteScript("window.location.href = '/game/" + roomCode + "'")
				// Flush immediately to ensure redirect is sent
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				return // Close the lobby SSE connection
			case "countdown_update", "game_playing":
				// These events happen after game has started
				// Players should already be on the game page, so just close this lobby connection
				log.Printf("🎮 Game event '%s' received in lobby SSE - closing connection for room %s", event.Type, roomCode)
				return
			default:
				log.Printf("📡 Unknown event type %s for room %s in lobby SSE", event.Type, roomCode)
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

	// If joining during countdown, calculate actual remaining time
	if room.State == game.StateCountdown {
		// Calculate how much time has passed since countdown started
		elapsed := time.Since(room.StartedAt)
		originalCountdown := 5 // seconds
		actualRemaining := originalCountdown - int(elapsed.Seconds())

		// Update the room with actual remaining time
		if actualRemaining > 0 {
			room.CountdownRemaining = actualRemaining
			log.Printf("📡 Browser connected during countdown for room %s, actual remaining: %d seconds", roomCode, actualRemaining)
		} else {
			// Countdown should have finished, transition to playing
			room.State = game.StatePlaying
			room.CountdownRemaining = 0
			room.LeaderRevealed = true
			log.Printf("📡 Browser connected after countdown finished for room %s, showing game state", roomCode)
		}

		// Re-render with updated state
		h.renderGame(sse, room, player)
	}

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

// renderLobby renders the lobby content (without SSE trigger)
func (h *Handler) renderLobby(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	// Only render lobby if room is in lobby state
	if room.State != game.StateLobby {
		log.Printf("🚫 Attempted to render lobby for room %s in state %s", room.Code, room.State)
		return
	}

	log.Printf("🎨 Rendering lobby content for room %s with %d players", room.Code, len(room.Players))
	component := pages.LobbyContent(room, player)

	// Render to string
	html := renderToString(component)

	// Wrap content in the full container structure to preserve DOM hierarchy during morph
	wrappedHTML := `<div id="lobby-container" class="container"><div id="lobby-content">` + html + `</div></div>`

	log.Printf("📝 Rendered lobby HTML length: %d chars", len(wrappedHTML))
	
	// Enhanced debug logging to understand the exact HTML being sent
	log.Printf("🔍 Raw LobbyContent HTML (first 150 chars): %.150s...", html)
	log.Printf("🔍 Wrapped HTML structure check - has lobby-container: %v", strings.Contains(wrappedHTML, `id="lobby-container"`))
	log.Printf("🔍 Wrapped HTML structure check - has lobby-content: %v", strings.Contains(wrappedHTML, `id="lobby-content"`))

	// Debug: show first 200 chars of rendered HTML
	if len(wrappedHTML) > 200 {
		log.Printf("📝 HTML preview: %s...", wrappedHTML[:200])
	} else {
		log.Printf("📝 Full HTML: %s", wrappedHTML)
	}

	// Send fragment with full container structure
	// Target the container and morph will preserve the entire DOM hierarchy
	sse.MergeFragments(wrappedHTML,
		datastar.WithSelector("#lobby-container"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
	log.Printf("✅ Sent lobby fragment update for room %s", room.Code)
}

// renderGame renders the game content (without wrapper to prevent re-triggering data-on-load)
func (h *Handler) renderGame(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	component := pages.GameContent(room, player)

	// Render to string
	html := renderToString(component)

	// Send as fragment with morph mode and explicit selector
	sse.MergeFragments(html,
		datastar.WithSelector("#game-container"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
}

// renderToString renders a templ component to string
func renderToString(component templ.Component) string {
	buf := &bytes.Buffer{}
	component.Render(context.Background(), buf)
	return buf.String()
}
