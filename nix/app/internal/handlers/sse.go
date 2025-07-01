package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/views/components"
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
			case "role_config_updated":
				// Role config was updated - just send the updated role config component
				log.Printf("🎯 Role config updated for room %s", roomCode)
				component := components.RoleConfigurationNew(room, h.config, h.cardService)
				html := renderToString(component)
				sse.MergeFragments(html,
					datastar.WithSelector("#role-config"),
					datastar.WithMergeMode(datastar.FragmentMergeModeMorph))
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
	component := pages.LobbyContent(room, player, h.config, h.cardService)

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

// StreamHost streams host dashboard updates
func (h *Handler) StreamHost(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	log.Printf("📡 SSE connection established for host dashboard %s", roomCode)

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		log.Printf("📡 SSE requested for non-existent room: %s", roomCode)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the user is the host
	hostCookie, err := r.Cookie("host_" + roomCode)
	if err != nil || hostCookie.Value != "true" {
		log.Printf("📡 Unauthorized host SSE attempt for room: %s", roomCode)
		http.Error(w, "Unauthorized - Host access only", http.StatusUnauthorized)
		return
	}

	// Get host player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Host player not found", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Host player not found in room", http.StatusUnauthorized)
		return
	}

	// Create SSE connection
	sse := datastar.NewSSE(w, r)

	// Generate and send QR code once
	baseURL := getBaseURL(r)
	qrURL := fmt.Sprintf("%s/room/%s", baseURL, roomCode)

	qrCode, err := generateQRCode(qrURL)
	if err != nil {
		log.Printf("❌ Failed to generate QR code for room %s: %v", roomCode, err)
	} else if qrCode != "" {
		// Send QR code as a signal
		qrDataURI := fmt.Sprintf("data:image/png;base64,%s", qrCode)
		signals := map[string]string{
			"qrCode": qrDataURI,
		}
		if err := sse.MarshalAndMergeSignals(signals); err != nil {
			log.Printf("❌ Failed to send QR code signal for room %s: %v", roomCode, err)
		} else {
			log.Printf("📱 Sent QR code for room %s", roomCode)
		}
	}

	// Send initial player list
	h.renderHostDashboard(sse, room, player)

	// Subscribe to events
	events := h.eventBus.Subscribe(roomCode)
	defer h.eventBus.Unsubscribe(roomCode, events)

	log.Printf("📡 Host SSE connection ready for room %s, waiting for events", roomCode)

	// Set up a heartbeat to detect stale connections
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			log.Printf("📡 Host SSE context cancelled for room %s", roomCode)
			return
		case <-heartbeat.C:
			// Check if room still exists
			_, err := h.store.GetRoom(roomCode)
			if err != nil {
				log.Printf("📡 Heartbeat: Room %s no longer exists, closing host SSE", roomCode)
				return
			}
			// Send SSE comment heartbeat to keep connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case event := <-events:
			log.Printf("📡 Host SSE event received for %s: %s", roomCode, event.Type)

			switch event.Type {
			case "player_joined", "player_left", "player_updated", "role_config_updated":
				// Re-render host dashboard for player changes or role config updates
				room, _ = h.store.GetRoom(roomCode)
				if room.State == game.StateLobby {
					// Refresh player reference in case it was updated
					player = room.GetPlayer(player.ID)
					if player == nil {
						// Host was removed, close SSE connection gracefully
						log.Printf("📡 Host no longer in room %s, closing SSE", roomCode)
						return
					}
					h.renderHostDashboard(sse, room, player)
				}
			case "game_started":
				// Update dashboard to show countdown state
				room, _ = h.store.GetRoom(roomCode)
				h.renderHostDashboard(sse, room, player)
			case "countdown_update":
				// Update countdown display
				room, _ = h.store.GetRoom(roomCode)
				h.renderHostDashboard(sse, room, player)
			case "game_playing":
				// Update dashboard to show game state
				room, _ = h.store.GetRoom(roomCode)
				h.renderHostDashboard(sse, room, player)
			case "game_ended":
				// Update dashboard to show ended state
				room, _ = h.store.GetRoom(roomCode)
				h.renderHostDashboard(sse, room, player)
			default:
				log.Printf("📡 Unknown event type %s for room %s in host SSE", event.Type, roomCode)
			}
		}
	}
}

// renderHostDashboard renders the host dashboard content based on game state
func (h *Handler) renderHostDashboard(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	var component templ.Component

	// Choose the appropriate template based on game state
	switch room.State {
	case game.StateLobby:
		component = pages.HostDashboardContent(room, player, h.config, h.cardService)
	case game.StateCountdown:
		component = pages.HostDashboardCountdown(room, player)
	case game.StatePlaying:
		component = pages.HostDashboardPlaying(room, player)
	case game.StateEnded:
		component = pages.HostDashboardEnded(room, player)
	default:
		component = pages.HostDashboardContent(room, player, h.config, h.cardService)
	}

	// Render to string
	html := renderToString(component)

	// Wrap content in the dashboard container structure to preserve DOM hierarchy during morph
	wrappedHTML := fmt.Sprintf(`<div id="host-dashboard-container" class="host-dashboard"><div id="host-dashboard-content">%s</div></div>`, html)

	log.Printf("🎨 Rendering host dashboard for room %s in state %s", room.Code, room.State)

	// Send fragment with full container structure
	sse.MergeFragments(wrappedHTML,
		datastar.WithSelector("#host-dashboard-container"),
		datastar.WithMergeMode(datastar.FragmentMergeModeMorph))

	log.Printf("✅ Sent host dashboard update for room %s", room.Code)
}

// generateQRCode generates a QR code for the given URL and returns it as base64 encoded PNG
func generateQRCode(url string) (string, error) {
	// Create QR code with medium error correction level
	qrc, err := qrcode.NewWith(url,
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium),
		qrcode.WithEncodingMode(qrcode.EncModeByte),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	// Create a temporary file
	tmpFile := fmt.Sprintf("/tmp/qr_%d.png", time.Now().UnixNano())

	// Create a writer with appropriate options
	w, err := standard.New(tmpFile,
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		standard.WithQRWidth(8), // 8 pixels per module
	)
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %w", err)
	}

	// Save the QR code to the file
	if err := qrc.Save(w); err != nil {
		return "", fmt.Errorf("failed to save QR code: %w", err)
	}

	// Read the file back
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to read QR code file: %w", err)
	}

	// Clean up the temporary file
	os.Remove(tmpFile)

	// Encode the PNG data as base64
	encoded := base64.StdEncoding.EncodeToString(data)

	return encoded, nil
}

// getBaseURL constructs the base URL from the request
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	// Check for X-Forwarded-Proto header (common in reverse proxy setups)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	// Get host from request
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}
