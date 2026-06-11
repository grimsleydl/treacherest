package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar-go/datastar"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"log"
	"net/http"
	"os"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/views/components"
	"treacherest/internal/views/pages"
)

// StreamLobby streams lobby updates
func (h *Handler) StreamLobby(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	log.Printf("📡 SSE connection established for lobby %s", roomCode)

	// Debug mode: log request details
	if os.Getenv("DEBUG") != "" {
		log.Printf("DEBUG: 📡 SSE request details - User-Agent: %s, RemoteAddr: %s", r.Header.Get("User-Agent"), r.RemoteAddr)
		deadline, hasDeadline := r.Context().Deadline()
		log.Printf("DEBUG: 📡 SSE request timeout context: deadline=%v, hasDeadline=%v", deadline, hasDeadline)
	}

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
	playerID := player.ID // Capture player ID for defer
	events := h.eventBus.Subscribe(roomCode)
	log.Printf("📡 DEBUG: Player %s subscribed to events for room %s", playerID, roomCode)
	defer func() {
		h.eventBus.Unsubscribe(roomCode, events)
		log.Printf("📡 DEBUG: Player %s unsubscribed from room %s", playerID, roomCode)
	}()

	// Don't send initial render - page already has correct content
	// But DO send initial validation state to ensure UI is in sync
	roleService := game.NewRoleConfigService(h.config)
	validationState := room.GetValidationState(roleService)

	err = sse.MarshalAndPatchSignals(map[string]interface{}{
		"canStartGame":      validationState.CanStart,
		"validationMessage": validationState.ValidationMessage,
		"canAutoScale":      validationState.CanAutoScale,
		"autoScaleDetails":  validationState.AutoScaleDetails,
		"requiredRoles":     validationState.RequiredRoles,
		"configuredRoles":   validationState.ConfiguredRoles,
		// Ensure button is not in loading state on initial connect
		"isStarting": false,
		"startError": "",
	})

	if err != nil {
		log.Printf("❌ Failed to send initial validation state: %v", err)
	}

	// Send debug mode signal if debug mode is enabled (for debug panel visibility)
	if h.config.Server.DebugModeEnabled {
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"debugmode": true,
		})
	}

	log.Printf("📡 SSE connection ready for room %s with validation state v%d", roomCode, validationState.Version)

	// Set up a heartbeat to prevent timeouts
	// 15 seconds is well under our 10-minute WriteTimeout
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 SSE context cancelled for room %s - error: %v", roomCode, r.Context().Err())
			}
			log.Printf("📡 Lobby SSE context cancelled for room %s", roomCode)
			return
		case <-heartbeat.C:
			// Check if room still exists
			_, err := h.store.GetRoom(roomCode)
			if err != nil {
				log.Printf("📡 Heartbeat: Room %s no longer exists, closing SSE", roomCode)
				return
			}

			// Debug mode: log detailed heartbeat info
			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 SSE heartbeat for lobby %s - player %s (%s)", roomCode, player.Name, player.ID)
			}

			// Send minimal keepalive comment to prevent timeout
			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 Sending keepalive for room %s", roomCode)
			}

			// Send minimal SSE comment - just colon and newlines
			if _, err := w.Write([]byte(":\n\n")); err != nil {
				if os.Getenv("DEBUG") != "" {
					log.Printf("DEBUG: 📡 Keepalive write error: %v", err)
				}
				log.Printf("📡 Keepalive failed for room %s: %v - closing connection", roomCode, err)
				return
			}

			// Flush to ensure the comment is sent immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 Keepalive sent successfully for room %s", roomCode)
			}
		case event := <-events:
			log.Printf("📡 SSE event received for %s: %s", roomCode, event.Type)

			switch event.Type {
			case "player_joined", "player_left", "player_updated":
				log.Printf("📡 DEBUG: StreamLobby received %s event for room %s, player %s", event.Type, roomCode, player.ID)
				// Re-render lobby only if still in lobby state
				room, _ = h.store.GetRoom(roomCode)
				if room.State == game.StateLobby {
					log.Printf("📡 DEBUG: Room is in lobby state, current players: %d", len(room.Players))
					for pid, p := range room.Players {
						log.Printf("📡 DEBUG:   - Player %s: %s", pid, p.Name)
					}

					// Refresh player reference in case it was updated
					originalPlayerID := player.ID
					player = room.GetPlayer(player.ID)
					if player == nil {
						// Player was removed, close SSE connection gracefully
						log.Printf("📡 Player %s no longer in room %s, closing SSE", originalPlayerID, roomCode)
						return
					}
					// For player events, only send player list update (not the entire lobby)
					renderPlayer := h.effectivePlayerForRender(r, room, player)
					if renderPlayer == nil {
						log.Printf("📡 Effective player no longer in room %s, closing SSE", roomCode)
						return
					}
					log.Printf("📤 DEBUG: Sending player list update to player %s in room %s", renderPlayer.ID, roomCode)
					h.sendPlayerListUpdate(sse, room, renderPlayer)
					log.Printf("📤 DEBUG: Player list update sent successfully to player %s", renderPlayer.ID)
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
				// Role config was updated - send updates appropriately based on player type
				log.Printf("🎯 Role config updated for room %s", roomCode)
				room, _ = h.store.GetRoom(roomCode)

				// Check if current player can control the game
				canControl := !hasHost(room) && player.ID == getFirstPlayerID(room)

				if canControl {
					// Send the role config component only to controlling players
					playerCountDisplay := h.createPlayerCountDisplay(room)
					component := components.RoleConfigurationNew(room, h.config, h.cardService, playerCountDisplay)
					html := renderToString(component)
					sse.PatchElements(html,
						datastar.WithSelector("#role-config"))

					// Also update validation state for controlling players
					roleService := game.NewRoleConfigService(h.config)
					validationState := room.GetValidationState(roleService)

					sse.MarshalAndPatchSignals(map[string]interface{}{
						"canStartGame":      validationState.CanStart,
						"validationMessage": validationState.ValidationMessage,
						"canAutoScale":      validationState.CanAutoScale,
						"autoScaleDetails":  validationState.AutoScaleDetails,
						"requiredRoles":     validationState.RequiredRoles,
						"configuredRoles":   validationState.ConfiguredRoles,
					})
				} else {
					// Non-controlling players don't need role config updates
					log.Printf("📡 Skipping role config update for non-controlling player %s in room %s", player.ID, roomCode)
				}
			case "coup_config_updated":
				log.Printf("🎯 Coup config updated for room %s", roomCode)
				room, _ = h.store.GetRoom(roomCode)
				player = room.GetPlayer(player.ID)
				if player == nil {
					log.Printf("📡 Player no longer in room %s after Coup config update, closing SSE", roomCode)
					return
				}
				renderPlayer := h.effectivePlayerForRender(r, room, player)
				if renderPlayer == nil {
					log.Printf("📡 Effective player no longer in room %s after Coup config update, closing SSE", roomCode)
					return
				}
				h.sendLobbyUpdate(sse, room, renderPlayer)
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
	log.Printf("🎮 Initial render for room %s, state: %s, countdown: %d", roomCode, room.State, room.CountdownRemaining)
	renderPlayer := h.effectivePlayerForRender(r, room, player)
	if renderPlayer == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}
	h.renderGame(sse, room, renderPlayer)

	// Send initial signals including countdown
	signals := map[string]interface{}{
		"countdown": room.CountdownRemaining,
	}
	err = sse.MarshalAndPatchSignals(signals)
	if err != nil {
		log.Printf("❌ Failed to send initial game signals: %v", err)
	}

	// Send initial state backup
	h.emitStateBackup(sse, room)

	// Send debug mode signal if debug mode is enabled (for debug panel visibility)
	if h.config.Server.DebugModeEnabled {
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"debugmode": true,
		})
	}

	// If joining during countdown, calculate actual remaining time
	if room.State == game.StateCountdown {
		// Calculate how much time has passed since countdown started
		elapsed := time.Since(room.StartedAt)
		originalCountdown := 5 // seconds
		actualRemaining := originalCountdown - int(elapsed.Seconds())

		// Update the room with actual remaining time
		if actualRemaining > 0 {
			room.CountdownRemaining = actualRemaining
			h.store.UpdateRoom(room) // Save the updated countdown to store
			log.Printf("📡 Browser connected during countdown for room %s, actual remaining: %d seconds", roomCode, actualRemaining)
		} else {
			// Countdown should have finished, transition to playing
			room.State = game.StatePlaying
			room.CountdownRemaining = 0
			room.LeaderRevealed = true
			h.store.UpdateRoom(room) // Save the updated state to store
			log.Printf("📡 Browser connected after countdown finished for room %s, showing game state", roomCode)
		}

		// Re-render with updated state
		renderPlayer = h.effectivePlayerForRender(r, room, player)
		if renderPlayer == nil {
			return
		}
		h.renderGame(sse, room, renderPlayer)
	}

	// Set up a heartbeat to prevent timeouts
	// 15 seconds is well under our 10-minute WriteTimeout
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	// Track heartbeat count for periodic backup (every 4 heartbeats = 60 seconds)
	heartbeatCount := 0

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			heartbeatCount++

			// Send minimal keepalive comment to prevent timeout
			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 Sending keepalive for game room %s", roomCode)
			}

			// Send minimal SSE comment - just colon and newlines
			if _, err := w.Write([]byte(":\n\n")); err != nil {
				if os.Getenv("DEBUG") != "" {
					log.Printf("DEBUG: 📡 Keepalive write error: %v", err)
				}
				log.Printf("📡 Keepalive failed for game room %s: %v - closing connection", roomCode, err)
				return
			}

			// Flush to ensure the comment is sent immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 Keepalive sent successfully for game room %s", roomCode)
			}

			// Send periodic backup every 60 seconds (4 heartbeats)
			if heartbeatCount%4 == 0 {
				room, _ = h.store.GetRoom(roomCode)
				if room != nil {
					h.emitStateBackup(sse, room)
				}
			}
		case event := <-events:
			log.Printf("📡 SSE event received for game %s: %s", roomCode, event.Type)
			switch event.Type {
			case "countdown_update":
				// Get fresh room data
				room, _ = h.store.GetRoom(roomCode)

				// Send ONLY the countdown signal
				signals := map[string]interface{}{
					"countdown": room.CountdownRemaining,
				}

				err := sse.MarshalAndPatchSignals(signals)
				if err != nil {
					log.Printf("❌ Failed to send countdown signal: %v", err)
				} else {
					log.Printf("⏱️ Sent countdown signal for room %s: %d", roomCode, room.CountdownRemaining)
				}
			case "game_playing":
				// Transition to playing state - render and clear countdown
				room, _ = h.store.GetRoom(roomCode)
				player = room.GetPlayer(player.ID)
				renderPlayer := h.effectivePlayerForRender(r, room, player)
				if renderPlayer == nil {
					return
				}
				h.renderGame(sse, room, renderPlayer)

				// Clear countdown signal
				signals := map[string]interface{}{
					"countdown": 0,
				}
				sse.MarshalAndPatchSignals(signals)
				log.Printf("🎮 Game playing - cleared countdown signal for room %s", roomCode)

				// Emit backup after game state transition
				h.emitStateBackup(sse, room)
			default:
				// All other events need full re-render
				room, _ = h.store.GetRoom(roomCode)
				player = room.GetPlayer(player.ID) // Refresh player data
				renderPlayer := h.effectivePlayerForRender(r, room, player)
				if renderPlayer == nil {
					return
				}
				h.renderGame(sse, room, renderPlayer)

				// Clear any temporary modals (e.g., X input modal for Wearer of Masks)
				h.clearModalContainer(sse)

				// Emit backup after any game state change
				h.emitStateBackup(sse, room)
			}
		}
	}
}

// clearModalContainer clears temporary modals from #modal-container via SSE
// This is needed because #modal-container is outside #game-container and doesn't get
// automatically cleared when game content is morphed
func (h *Handler) clearModalContainer(sse *datastar.ServerSentEventGenerator) {
	sse.PatchElements("",
		datastar.WithSelector("#modal-container"),
		datastar.WithModeInner())
}

// sendPlayerListUpdate sends only the player list card - minimal update for player join/leave
func (h *Handler) sendPlayerListUpdate(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	log.Printf("📤 Sending minimal player list update for room %s", room.Code)

	// Render just the player list card
	component := pages.LobbyPlayerList(room, player)
	html := renderToString(component)

	log.Printf("📝 Player list HTML length: %d chars (was 5MB before!)", len(html))
	log.Printf("[DEBUG] Player list HTML: %s", html)

	// Send fragment targeting the player list card
	sse.PatchElements(html,
		datastar.WithSelector("#player-list-card"))

	log.Printf("✅ Sent minimal player list update for room %s", room.Code)
}

// sendLobbyUpdate sends a consistent lobby update with validation state
// This is the helper function that ensures SSE updates use the same validation logic
func (h *Handler) sendLobbyUpdate(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) error {
	// CRITICAL: Always use GetValidationState for consistency
	roleService := game.NewRoleConfigService(h.config)
	validationState := room.GetValidationState(roleService)

	// First send the HTML fragment
	log.Printf("📤 DEBUG: sendLobbyUpdate called for player %s in room %s", player.ID, room.Code)
	log.Printf("📤 DEBUG: sendLobbyUpdate calling renderLobby for room %s", room.Code)
	h.renderLobby(sse, room, player)

	// Then send the validation signals to keep UI in sync
	err := sse.MarshalAndPatchSignals(map[string]interface{}{
		"canStartGame":      validationState.CanStart,
		"validationMessage": validationState.ValidationMessage,
		"canAutoScale":      validationState.CanAutoScale,
		"autoScaleDetails":  validationState.AutoScaleDetails,
		"requiredRoles":     validationState.RequiredRoles,
		"configuredRoles":   validationState.ConfiguredRoles,
		// Reset error state on updates
		"isStarting": false,
		"startError": "",
	})

	if err != nil {
		log.Printf("❌ Failed to update validation signals: %v", err)
		return err
	}

	log.Printf("✅ Sent lobby update with validation state v%d for room %s", validationState.Version, room.Code)
	return nil
}

// renderLobby renders the lobby content (without SSE trigger)
func (h *Handler) renderLobby(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	// Only render lobby if room is in lobby state
	if room.State != game.StateLobby {
		log.Printf("🚫 Attempted to render lobby for room %s in state %s", room.Code, room.State)
		return
	}

	log.Printf("🎨 Rendering lobby content for room %s with %d players", room.Code, len(room.Players))
	log.Printf("🎨 DEBUG: Room player details:")
	for pid, p := range room.Players {
		log.Printf("🎨 DEBUG:   - Player %s: %s (Host: %v)", pid, p.Name, p.IsHost)
	}
	log.Printf("🎨 DEBUG: Active player count: %d", room.GetActivePlayerCount())

	component := pages.LobbyContent(room, player, h.config, h.cardService)

	// Render to string
	html := renderToString(component)

	log.Printf("📝 Rendered lobby HTML length: %d chars", len(html))

	// Debug: show first 200 chars of rendered HTML
	if len(html) > 200 {
		log.Printf("📝 HTML preview: %s...", html[:200])
	} else {
		log.Printf("📝 Full HTML: %s", html)
	}

	// Send the target wrapper too, so morphing #lobby-content keeps the
	// element that future patches target.
	wrappedHTML := fmt.Sprintf(`<div id="lobby-content">%s</div>`, html)
	log.Printf("📤 DEBUG: Sending fragment with selector #lobby-content, merge mode: morph")
	sse.PatchElements(wrappedHTML,
		datastar.WithSelector("#lobby-content"))
	log.Printf("✅ Sent lobby fragment update for room %s to player %s", room.Code, player.ID)
}

// renderGame renders the game content (without wrapper to prevent re-triggering data-on-load)
func (h *Handler) renderGame(sse *datastar.ServerSentEventGenerator, room *game.Room, player *game.Player) {
	log.Printf("🎨 Rendering game for room %s, state: %s, countdown: %d", room.Code, room.State, room.CountdownRemaining)
	component := pages.GameContent(room, player)

	// Render to string
	html := renderToString(component)

	// Log first 200 chars of rendered HTML for debugging
	if len(html) > 200 {
		log.Printf("🎭 Rendered HTML preview: %s...", html[:200])
	} else {
		log.Printf("🎭 Rendered HTML: %s", html)
	}

	// Send as fragment with morph mode and explicit selector
	sse.PatchElements(html,
		datastar.WithSelector("#game-container"))
}

// renderToString renders a templ component to string
func renderToString(component templ.Component) string {
	buf := &bytes.Buffer{}
	component.Render(context.Background(), buf)
	return buf.String()
}

// emitStateBackup sends an encrypted state backup to the client for localStorage storage
// This is used for recovering game state after Cloud Run instance replacement
func (h *Handler) emitStateBackup(sse *datastar.ServerSentEventGenerator, room *game.Room) {
	if h.backupService == nil {
		return // Backup service not configured
	}

	backup, err := h.backupService.CreateBackup(room)
	if err != nil {
		log.Printf("❌ Failed to create state backup for room %s: %v", room.Code, err)
		return
	}

	// Send backup as a signal - handled by dedicated #backup-handler element's data-effect
	err = sse.MarshalAndPatchSignals(map[string]interface{}{
		"_stateBackup": backup,
	})
	if err != nil {
		log.Printf("❌ Failed to send state backup signal for room %s: %v", room.Code, err)
	} else if os.Getenv("DEBUG") != "" {
		log.Printf("DEBUG: 💾 Sent state backup for room %s (%d bytes)", room.Code, len(backup))
	}
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
		if err := sse.MarshalAndPatchSignals(signals); err != nil {
			log.Printf("❌ Failed to send QR code signal for room %s: %v", roomCode, err)
		} else {
			log.Printf("📱 Sent QR code for room %s", roomCode)
		}
	}

	// Send initial player list
	h.renderHostDashboard(sse, room, player)

	// Send initial validation state for host dashboard
	if room.State == game.StateLobby {
		roleService := game.NewRoleConfigService(h.config)
		validationState := room.GetValidationState(roleService)

		sse.MarshalAndPatchSignals(map[string]interface{}{
			"canStartGame":      validationState.CanStart,
			"validationMessage": validationState.ValidationMessage,
			"canAutoScale":      validationState.CanAutoScale,
			"autoScaleDetails":  validationState.AutoScaleDetails,
			"requiredRoles":     validationState.RequiredRoles,
			"configuredRoles":   validationState.ConfiguredRoles,
		})

		log.Printf("📡 Sent initial validation state for host dashboard: canAutoScale=%v", validationState.CanAutoScale)
	} else if room.State == game.StateCountdown {
		// Send initial countdown signal if joining during countdown
		signals := map[string]interface{}{
			"countdown": room.CountdownRemaining,
		}
		err = sse.MarshalAndPatchSignals(signals)
		if err != nil {
			log.Printf("❌ Failed to send initial countdown signal to host: %v", err)
		} else {
			log.Printf("📡 Sent initial countdown signal to host: %d", room.CountdownRemaining)
		}
	}

	// Send debug mode signal if debug mode is enabled (for debug panel visibility)
	if h.config.Server.DebugModeEnabled {
		sse.MarshalAndPatchSignals(map[string]interface{}{
			"debugmode": true,
		})
	}

	// Subscribe to events
	events := h.eventBus.Subscribe(roomCode)
	defer h.eventBus.Unsubscribe(roomCode, events)

	log.Printf("📡 Host SSE connection ready for room %s, waiting for events (subscriber channel: %p)", roomCode, events)

	// Set up a heartbeat to prevent timeouts
	// 15 seconds is well under our 10-minute WriteTimeout
	heartbeat := time.NewTicker(15 * time.Second)
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

			// Send minimal keepalive comment to prevent timeout
			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 Sending keepalive for host room %s", roomCode)
			}

			// Send minimal SSE comment - just colon and newlines
			if _, err := w.Write([]byte(":\n\n")); err != nil {
				if os.Getenv("DEBUG") != "" {
					log.Printf("DEBUG: 📡 Keepalive write error: %v", err)
				}
				log.Printf("📡 Keepalive failed for host room %s: %v - closing connection", roomCode, err)
				return
			}

			// Flush to ensure the comment is sent immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			if os.Getenv("DEBUG") != "" {
				log.Printf("DEBUG: 📡 Keepalive sent successfully for host room %s", roomCode)
			}
		case event := <-events:
			log.Printf("📡 Host SSE event received for %s: %s", roomCode, event.Type)

			switch event.Type {
			case "player_joined", "player_left", "player_updated", "role_config_updated", "coup_config_updated":
				// Re-render host dashboard for player changes or setup config updates.
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

					// Also send validation state for host dashboard
					roleService := game.NewRoleConfigService(h.config)
					validationState := room.GetValidationState(roleService)

					sse.MarshalAndPatchSignals(map[string]interface{}{
						"canStartGame":      validationState.CanStart,
						"validationMessage": validationState.ValidationMessage,
						"canAutoScale":      validationState.CanAutoScale,
						"autoScaleDetails":  validationState.AutoScaleDetails,
						"requiredRoles":     validationState.RequiredRoles,
						"configuredRoles":   validationState.ConfiguredRoles,
					})
				}
			case "game_started":
				// Update dashboard to show countdown state
				room, _ = h.store.GetRoom(roomCode)
				h.renderHostDashboard(sse, room, player)
			case "countdown_update":
				// Get fresh room data
				room, _ = h.store.GetRoom(roomCode)

				// Send ONLY the countdown signal for the host
				signals := map[string]interface{}{
					"countdown": room.CountdownRemaining,
				}

				err := sse.MarshalAndPatchSignals(signals)
				if err != nil {
					log.Printf("❌ Failed to send countdown signal to host: %v", err)
				} else {
					log.Printf("⏱️ Sent countdown signal to host for room %s: %d", roomCode, room.CountdownRemaining)
				}
			case "game_playing":
				// Update dashboard to show game state
				room, _ = h.store.GetRoom(roomCode)
				h.renderHostDashboard(sse, room, player)

				// Clear countdown signal for host
				signals := map[string]interface{}{
					"countdown": 0,
				}
				sse.MarshalAndPatchSignals(signals)
				log.Printf("🎮 Game playing - cleared countdown signal for host in room %s", roomCode)
			case "role_revealed", "player_eliminated", "coup_win_prompt_rejected":
				room, _ = h.store.GetRoom(roomCode)
				player = room.GetPlayer(player.ID)
				if player == nil {
					log.Printf("📡 Host no longer in room %s, closing SSE", roomCode)
					return
				}
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
	sse.PatchElements(wrappedHTML,
		datastar.WithSelector("#host-dashboard-container"))

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

// Helper function to check if there's a host in the room
func hasHost(room *game.Room) bool {
	for _, player := range room.Players {
		if player.IsHost {
			return true
		}
	}
	return false
}

// Helper function to get the first player ID (room creator)
func getFirstPlayerID(room *game.Room) string {
	var firstPlayer *game.Player
	var firstJoinTime time.Time

	for _, player := range room.Players {
		if !player.IsHost && (firstPlayer == nil || player.JoinedAt.Before(firstJoinTime)) {
			firstPlayer = player
			firstJoinTime = player.JoinedAt
		}
	}

	if firstPlayer != nil {
		return firstPlayer.ID
	}
	return ""
}
