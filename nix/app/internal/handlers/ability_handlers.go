package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"treacherest/internal/game"
	"treacherest/internal/game/ability"

	"github.com/go-chi/chi/v5"
)

// TriggerWearerAbility initiates The Wearer of Masks ability for a player
// This is typically called at game start for players with The Wearer of Masks card
func (h *Handler) TriggerWearerAbility(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	player := room.GetPlayer(playerID)
	if player == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Verify player has The Wearer of Masks (card ID 31)
	if player.Role == nil || player.Role.GetID() != 31 {
		http.Error(w, "Player does not have The Wearer of Masks", http.StatusBadRequest)
		return
	}

	// Get role options for The Wearer of Masks
	var maxReveal int = 5 // Default
	var useAllCards bool = false

	if room.RoleOptionsManager != nil && room.RoleOptionsManager.HasOptions(31) {
		opts := room.RoleOptionsManager.GetOrCreateOptions(31)
		if val, err := opts.GetIntOption("max_reveal"); err == nil {
			maxReveal = val
		}
		if val, err := opts.GetBoolOption("use_all_cards"); err == nil {
			useAllCards = val
		}
	}

	// Get available cards from CardPool
	if room.CardPool == nil {
		http.Error(w, "Card pool not initialized", http.StatusInternalServerError)
		return
	}

	// Filter cards based on options
	var filterTypes []string
	if !useAllCards {
		// Exclude Leaders by default
		filterTypes = []string{"Guardian", "Assassin", "Traitor"}
	}

	availableCards := room.CardPool.GetCardsByTypes(filterTypes...)

	// Limit to maxReveal cards
	if len(availableCards) > maxReveal {
		availableCards = availableCards[:maxReveal]
	}

	if len(availableCards) == 0 {
		http.Error(w, "No cards available to reveal", http.StatusBadRequest)
		return
	}

	// Create pending ability
	abilityID := fmt.Sprintf("wearer-%s-%d", playerID, room.CountdownRemaining)
	pendingAbility := &ability.PendingAbility{
		ID:          abilityID,
		PlayerID:    playerID,
		CardID:      31,
		AbilityType: "unveil",
		Data: map[string]interface{}{
			"available_cards": convertCardsToIDs(availableCards),
			"max_reveal":      maxReveal,
			"use_all_cards":   useAllCards,
		},
		ModalDismissed: false,
	}

	player.AbilityState.AddPendingAbility(pendingAbility)

	h.store.UpdateRoom(room)

	log.Printf("🎭 Triggered Wearer ability for %s in room %s (revealed %d cards)", player.Name, roomCode, len(availableCards))

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "ability_triggered",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// SelectWearerCard handles card selection for The Wearer of Masks ability
func (h *Handler) SelectWearerCard(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify the requesting player is in the room
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

	// Get the pending ability
	pendingAbility := player.AbilityState.GetPendingAbility(abilityID)
	if pendingAbility == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	// Verify ability belongs to this player
	if pendingAbility.PlayerID != player.ID {
		http.Error(w, "Ability does not belong to this player", http.StatusForbidden)
		return
	}

	// Parse request body to get selected card ID
	var req struct {
		CardID int `json:"card_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CardID == 0 {
		http.Error(w, "Card ID is required", http.StatusBadRequest)
		return
	}

	// Verify the selected card is in the available cards
	availableCardIDs, ok := pendingAbility.Data["available_cards"].([]int)
	if !ok {
		http.Error(w, "Invalid ability data", http.StatusInternalServerError)
		return
	}

	cardAllowed := false
	for _, cardID := range availableCardIDs {
		if cardID == req.CardID {
			cardAllowed = true
			break
		}
	}

	if !cardAllowed {
		http.Error(w, "Selected card not in available cards", http.StatusBadRequest)
		return
	}

	// Get the selected card from CardPool
	selectedCard := room.CardPool.GetCardByID(req.CardID)
	if selectedCard == nil {
		http.Error(w, "Selected card not found in pool", http.StatusNotFound)
		return
	}

	// Determine keep types based on original role
	keepTypes := []string{string(player.Role.GetRoleType())} // Keep original type (e.g., "Traitor")

	// Start transformation
	originalCardID := player.Role.GetID()
	player.AbilityState.StartTransform(originalCardID, req.CardID, keepTypes, "face_down")

	// Update player's role to the transformed card
	player.Role = selectedCard

	// Resolve the pending ability
	player.AbilityState.ResolvePendingAbility(abilityID)

	h.store.UpdateRoom(room)

	log.Printf("🎭 Player %s transformed from card %d to card %d in room %s", player.Name, originalCardID, req.CardID, roomCode)

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "transformation_complete",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// Helper function to convert cards to card IDs
func convertCardsToIDs(cards []*game.Card) []int {
	ids := make([]int, len(cards))
	for i, card := range cards {
		ids[i] = card.GetID()
	}
	return ids
}
