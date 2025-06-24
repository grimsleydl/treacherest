package handlers

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"treacherest/internal/game"
	"treacherest/internal/views/pages"
)

// Note: The following template functions need to be created in the views/pages package:
// - pages.HostLanding() - Host mode landing page
// - pages.HostDashboardLobby(room, player) - Host dashboard for lobby state
// - pages.HostDashboardCountdown(room, player) - Host dashboard for countdown state
// - pages.HostDashboardPlaying(room, player) - Host dashboard for playing state
// - pages.HostDashboardEnded(room, player) - Host dashboard for ended state

// HostLanding renders the host mode landing page
func (h *Handler) HostLanding(w http.ResponseWriter, r *http.Request) {
	component := pages.HostLanding()
	component.Render(r.Context(), w)
}

// CreateRoomAsHost creates a new room with the creator as host
func (h *Handler) CreateRoomAsHost(w http.ResponseWriter, r *http.Request) {
	hostName := r.FormValue("hostName")
	if hostName == "" {
		http.Error(w, "Host name is required", http.StatusBadRequest)
		return
	}

	// Create room
	room, err := h.store.CreateRoom()
	if err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	// Create host player
	sessionID := getOrCreateSession(w, r)
	player := game.NewPlayer(generatePlayerID(), hostName, sessionID)
	player.IsHost = true // Mark this player as the host

	// Add player to room
	room.AddPlayer(player)
	
	// Store host ID in room metadata (we'll need to add this field to Room struct)
	// For now, we'll use a cookie to track the host
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

	// Store host status in a separate cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "host_" + room.Code,
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 1 day
	})

	log.Printf("🎮 Host %s created room %s", hostName, room.Code)

	// Redirect to host dashboard
	http.Redirect(w, r, "/host/dashboard/"+room.Code, http.StatusSeeOther)
}

// HostDashboard shows the host dashboard for managing the game
func (h *Handler) HostDashboard(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the user is the host
	hostCookie, err := r.Cookie("host_" + roomCode)
	if err != nil || hostCookie.Value != "true" {
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

	// Render host dashboard based on game state
	switch room.State {
	case game.StateLobby:
		component := pages.HostDashboardLobby(room, player)
		err = component.Render(r.Context(), w)
	case game.StateCountdown:
		component := pages.HostDashboardCountdown(room, player)
		err = component.Render(r.Context(), w)
	case game.StatePlaying:
		component := pages.HostDashboardPlaying(room, player)
		err = component.Render(r.Context(), w)
	case game.StateEnded:
		component := pages.HostDashboardEnded(room, player)
		err = component.Render(r.Context(), w)
	default:
		component := pages.HostDashboardLobby(room, player)
		err = component.Render(r.Context(), w)
	}

	if err != nil {
		log.Printf("❌ Error rendering host dashboard: %v", err)
		http.Error(w, "Failed to render host dashboard", http.StatusInternalServerError)
		return
	}
}