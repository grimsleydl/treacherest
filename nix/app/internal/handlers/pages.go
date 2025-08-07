package handlers

import (
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"
)

// Home renders the home page
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	component := pages.Home()
	component.Render(r.Context(), w)
}

// CreateRoom creates a new room and redirects to it
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	playerName := r.FormValue("playerName")
	if playerName == "" {
		http.Error(w, "Player name is required", http.StatusBadRequest)
		return
	}

	// Check if creating as host only
	hostOnly := r.FormValue("hostOnly") == "true"

	// Create room
	room, err := h.store.CreateRoom()
	if err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	// Create player
	sessionID := getOrCreateSession(w, r)
	player := game.NewPlayer(generatePlayerID(), playerName, sessionID)

	// Set host flag if requested
	if hostOnly {
		player.IsHost = true
	}

	// Add player to room
	room.AddPlayer(player)
	h.store.UpdateRoom(room)

	// Store player ID in session
	http.SetCookie(w, &http.Cookie{
		Name:     "player_" + room.Code,
		Value:    player.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 1 day
	})

	// If host only, also set a host cookie
	if hostOnly {
		http.SetCookie(w, &http.Cookie{
			Name:     "host_" + room.Code,
			Value:    "true",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400, // 1 day
		})
	}

	// Redirect to room
	http.Redirect(w, r, "/room/"+room.Code, http.StatusSeeOther)
}

// JoinRoom shows the join room page or lobby
func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Check if player is already in room
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err == nil {
		// Player already in room
		player := room.GetPlayer(playerCookie.Value)
		if player != nil {

			// Show appropriate page based on player type and game state
			if player.IsHost {
				// Host sees dashboard view
				var component templ.Component
				switch room.State {
				case game.StateLobby:
					component = pages.HostDashboardLobby(room, player, h.config, h.cardService)
				case game.StateCountdown:
					component = pages.HostDashboardCountdown(room, player)
				case game.StatePlaying:
					component = pages.HostDashboardPlaying(room, player)
				case game.StateEnded:
					component = pages.HostDashboardEnded(room, player)
				default:
					component = pages.HostDashboardLobby(room, player, h.config, h.cardService)
				}
				component.Render(r.Context(), w)
			} else if room.State == game.StateLobby {
				// Regular player sees lobby
				component := pages.LobbyPage(room, player, h.config, h.cardService)
				component.Render(r.Context(), w)
			} else {
				// Regular player in active game
				http.Redirect(w, r, "/game/"+roomCode, http.StatusSeeOther)
			}
			return
		} else {
			// Cookie exists but player not in room - clear the stale cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "player_" + room.Code,
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
		}
	}

	// Check if game already started
	if room.State != game.StateLobby {
		http.Error(w, "Game already started", http.StatusBadRequest)
		return
	}

	// Show join form
	playerName := r.URL.Query().Get("name")
	if playerName == "" {
		// Show join form using Templ template
		component := pages.Join(roomCode, "")
		component.Render(r.Context(), w)
		return
	}

	// Create player and add to room
	sessionID := getOrCreateSession(w, r)

	// Check if we have an existing player ID in the cookie (for reconnection)
	var playerID string
	if playerCookie != nil {
		playerID = playerCookie.Value
	} else {
		playerID = generatePlayerID()
	}

	player := game.NewPlayer(playerID, playerName, sessionID)

	// Check if this player should be marked as a host
	// This happens when they previously created the room as host-only
	if hostCookie, err := r.Cookie("host_" + roomCode); err == nil && hostCookie.Value == "true" {
		player.IsHost = true
	}

	err = room.AddPlayer(player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.store.UpdateRoom(room)

	// Store player ID in session
	http.SetCookie(w, &http.Cookie{
		Name:     "player_" + room.Code,
		Value:    player.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 1 day
	})

	// Notify other players
	h.eventBus.Publish(Event{
		Type:     "player_joined",
		RoomCode: room.Code,
		Data:     room,
	})

	// Show lobby
	log.Printf("🏠 Rendering lobby page for room %s with %d players", room.Code, len(room.Players))
	component := pages.LobbyPage(room, player, h.config, h.cardService)
	err = component.Render(r.Context(), w)
	if err != nil {
		log.Printf("❌ Error rendering lobby page: %v", err)
		http.Error(w, "Failed to render lobby", http.StatusInternalServerError)
		return
	}
	log.Printf("✅ Successfully rendered lobby page for room %s", room.Code)
}

// GamePage shows the active game page
func (h *Handler) GamePage(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get player from cookie
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in game", http.StatusUnauthorized)
		return
	}

	player := room.GetPlayer(playerCookie.Value)
	if player == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Show appropriate view based on player type
	if player.IsHost {
		var component templ.Component
		switch room.State {
		case game.StateCountdown:
			component = pages.HostDashboardCountdownPage(room, player, h.config, h.cardService)
		case game.StatePlaying:
			component = pages.HostDashboardPlayingPage(room, player, h.config, h.cardService)
		case game.StateEnded:
			component = pages.HostDashboardEndedPage(room, player, h.config, h.cardService)
		default:
			// Shouldn't happen, but fallback to playing view
			component = pages.HostDashboardPlayingPage(room, player, h.config, h.cardService)
		}
		component.Render(r.Context(), w)
	} else {
		component := pages.GamePage(room, player)
		component.Render(r.Context(), w)
	}
}
