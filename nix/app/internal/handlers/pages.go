package handlers

import (
	"github.com/go-chi/chi/v5"
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

	// Create room
	room, err := h.store.CreateRoom()
	if err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	// Create player
	sessionID := getOrCreateSession(w, r)
	player := game.NewPlayer(generatePlayerID(), playerName, sessionID)

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
			// Show appropriate page based on game state
			if room.State == game.StateLobby {
				component := pages.LobbyPage(room, player)
				component.Render(r.Context(), w)
			} else {
				http.Redirect(w, r, "/game/"+roomCode, http.StatusSeeOther)
			}
			return
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
	component := pages.LobbyPage(room, player)
	component.Render(r.Context(), w)
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

	component := pages.GamePage(room, player)
	component.Render(r.Context(), w)
}
