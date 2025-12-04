package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"treacherest/internal/game"
	"treacherest/internal/game/ability"

	"github.com/go-chi/chi/v5"
)

// TriggerWearerAbility initiates The Wearer of Masks ability for a player
// The X value (number of cards to reveal) is passed as a URL parameter
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

	// Check if player already has a pending Wearer ability - prevent rerolling
	if player.AbilityState != nil && player.AbilityState.HasPendingAbilities() {
		for _, pending := range player.AbilityState.PendingAbilities {
			if pending.CardID == 31 {
				http.Error(w, "Ability already in progress - cannot reroll", http.StatusConflict)
				return
			}
		}
	}

	// Get X value from URL parameter (mana spent by player)
	xValueStr := chi.URLParam(r, "xValue")
	xValue, err := strconv.Atoi(xValueStr)
	if err != nil || xValue < 0 {
		http.Error(w, "Invalid X value", http.StatusBadRequest)
		return
	}

	// If X is 0, player chose not to reveal any cards - ability resolves with no effect
	if xValue == 0 {
		// Just set the card face up without transformation
		player.FaceUp = true
		player.RoleRevealed = true
		h.store.UpdateRoom(room)

		log.Printf("🎭 Wearer ability for %s in room %s - X=0, no transformation", player.Name, roomCode)

		h.eventBus.Publish(Event{
			Type:     "role_revealed",
			RoomCode: room.Code,
			Data:     room,
		})

		w.WriteHeader(http.StatusOK)
		return
	}

	// Get role options for whether to include Leader cards
	var useAllCards bool = false

	if room.RoleOptionsManager != nil && room.RoleOptionsManager.HasOptions(31) {
		opts := room.RoleOptionsManager.GetOrCreateOptions(31)
		if val, err := opts.GetBoolOption("use_all_cards"); err == nil {
			useAllCards = val
		}
	}

	// X value is now the maxReveal
	maxReveal := xValue

	// Get available cards from CardPool
	if room.CardPool == nil {
		http.Error(w, "Card pool not initialized", http.StatusInternalServerError)
		return
	}

	// Get available cards based on options
	// By default, include all role types EXCEPT Leaders (unless useAllCards is enabled)
	var filterTypes []string
	if !useAllCards {
		// Default: exclude Leaders, include all other role types
		filterTypes = []string{"Guardian", "Assassin", "Traitor"}
	}
	// If useAllCards is true, filterTypes remains empty, which returns ALL available cards

	availableCards := room.CardPool.GetCardsByTypes(filterTypes...)

	// Shuffle the cards to ensure random selection
	rand.Shuffle(len(availableCards), func(i, j int) {
		availableCards[i], availableCards[j] = availableCards[j], availableCards[i]
	})

	// Limit to maxReveal cards (now from a shuffled pool)
	if len(availableCards) > maxReveal {
		availableCards = availableCards[:maxReveal]
	}

	if len(availableCards) == 0 {
		http.Error(w, "No cards available to reveal", http.StatusBadRequest)
		return
	}

	// Create pending ability with confirmation requirement
	// The Leader must confirm they've witnessed the physical card reveal
	// before the player can see their transformation options
	// If there's no Leader in the game, skip confirmation requirement
	leader := room.GetLeader()
	requiresConfirmation := leader != nil

	abilityID := fmt.Sprintf("wearer-%s-%d", playerID, room.CountdownRemaining)
	pendingAbility := &ability.PendingAbility{
		ID:          abilityID,
		PlayerID:    playerID,
		CardID:      31,
		AbilityType: "unveil",
		Data: map[string]interface{}{
			"available_cards": convertCardsToIDs(availableCards),
			"x_value":         xValue,
			"use_all_cards":   useAllCards,
			"player_name":     player.Name,
			"card_name":       player.Role.Name,
		},
		ModalDismissed:       false,
		RequiresConfirmation: requiresConfirmation,
		ConfirmationRole:     "leader", // Leader must confirm they've seen the reveal
		ConfirmedBy:          []string{},
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

	// Verify ability has been confirmed (if required)
	if pendingAbility.RequiresConfirmation && !pendingAbility.IsConfirmed() {
		http.Error(w, "Ability has not been confirmed by the Leader yet", http.StatusForbidden)
		return
	}

	// Get card ID from URL path
	cardIDStr := chi.URLParam(r, "cardID")
	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil || cardID == 0 {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Verify the selected card is in the available cards
	availableCardIDs, ok := pendingAbility.Data["available_cards"].([]int)
	if !ok {
		http.Error(w, "Invalid ability data", http.StatusInternalServerError)
		return
	}

	cardAllowed := false
	for _, availableID := range availableCardIDs {
		if availableID == cardID {
			cardAllowed = true
			break
		}
	}

	if !cardAllowed {
		http.Error(w, "Selected card not in available cards", http.StatusBadRequest)
		return
	}

	// Get the selected card from CardPool
	selectedCard := room.CardPool.GetCardByID(cardID)
	if selectedCard == nil {
		http.Error(w, "Selected card not found in pool", http.StatusNotFound)
		return
	}

	// Determine keep types based on original role
	keepTypes := []string{string(player.Role.GetRoleType())} // Keep original type (e.g., "Traitor")

	// Start transformation
	originalCardID := player.Role.GetID()
	player.AbilityState.StartTransform(originalCardID, cardID, keepTypes, "face_down")

	// Update player's role to the transformed card
	player.Role = selectedCard

	// Resolve the pending ability
	player.AbilityState.ResolvePendingAbility(abilityID)

	h.store.UpdateRoom(room)

	log.Printf("🎭 Player %s transformed from card %d to card %d in room %s", player.Name, originalCardID, cardID, roomCode)

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "transformation_complete",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// ConfirmAbility allows the Leader (or designated role) to confirm they've witnessed
// a player's physical card reveal, enabling the ability to proceed
func (h *Handler) ConfirmAbility(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get the confirming player (should be the Leader for "leader" confirmation role)
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	confirmer := room.GetPlayer(playerCookie.Value)
	if confirmer == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Find the pending ability (could be on any player)
	var pendingAbility *ability.PendingAbility
	var abilityOwner *game.Player
	for _, p := range room.GetPlayers() {
		if p.AbilityState != nil {
			if pa := p.AbilityState.GetPendingAbility(abilityID); pa != nil {
				pendingAbility = pa
				abilityOwner = p
				break
			}
		}
	}

	if pendingAbility == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	// Check if ability requires confirmation
	if !pendingAbility.RequiresConfirmation {
		http.Error(w, "Ability does not require confirmation", http.StatusBadRequest)
		return
	}

	// Verify the confirmer is authorized based on ConfirmationRole
	switch pendingAbility.ConfirmationRole {
	case "leader":
		// Only the Leader can confirm
		if confirmer.Role == nil || confirmer.Role.GetRoleType() != game.RoleLeader {
			http.Error(w, "Only the Leader can confirm this ability", http.StatusForbidden)
			return
		}
	case "any_player":
		// Anyone except the ability owner can confirm
		if confirmer.ID == pendingAbility.PlayerID {
			http.Error(w, "You cannot confirm your own ability", http.StatusForbidden)
			return
		}
	case "all_players":
		// Anyone except the ability owner can contribute to confirmation
		if confirmer.ID == pendingAbility.PlayerID {
			http.Error(w, "You cannot confirm your own ability", http.StatusForbidden)
			return
		}
	default:
		http.Error(w, "Unknown confirmation role", http.StatusInternalServerError)
		return
	}

	// Add confirmation
	pendingAbility.AddConfirmation(confirmer.ID)

	h.store.UpdateRoom(room)

	log.Printf("✅ %s confirmed ability %s for %s in room %s", confirmer.Name, abilityID, abilityOwner.Name, roomCode)

	// Publish event to update all clients
	h.eventBus.Publish(Event{
		Type:     "ability_confirmed",
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

// TriggerMetamorphAbility activates The Metamorph's steal ability when unveiled
// The Metamorph (ID 25): "When The Metamorph is unveiled, until end of turn,
// as an opponent loses the game, you may remove The Metamorph from the game.
// If you do, gain control of that player's identity card and turn it face down if it isn't a Leader."
func (h *Handler) TriggerMetamorphAbility(w http.ResponseWriter, r *http.Request) {
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

	// Verify player has The Metamorph (card ID 25)
	if player.Role == nil || player.Role.GetID() != 25 {
		http.Error(w, "Player does not have The Metamorph", http.StatusBadRequest)
		return
	}

	// Check if already activated
	if player.AbilityState != nil && player.AbilityState.IsMetamorphActive() {
		http.Error(w, "Metamorph ability already active", http.StatusConflict)
		return
	}

	// Activate the Metamorph ability
	if player.AbilityState == nil {
		player.AbilityState = ability.NewAbilityState()
	}
	player.AbilityState.ActivateMetamorph()

	// Set card face up
	player.FaceUp = true
	player.RoleRevealed = true

	h.store.UpdateRoom(room)

	log.Printf("🦎 Metamorph ability activated for %s in room %s", player.Name, roomCode)

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "metamorph_activated",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// StealRole handles The Metamorph stealing an eliminated player's identity
func (h *Handler) StealRole(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")
	targetPlayerID := chi.URLParam(r, "targetPlayerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get the Metamorph player
	player := room.GetPlayer(playerID)
	if player == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Verify player has The Metamorph and ability is active
	if player.Role == nil || player.Role.GetID() != 25 {
		http.Error(w, "Player does not have The Metamorph", http.StatusBadRequest)
		return
	}

	if player.AbilityState == nil || !player.AbilityState.CanUseMetamorph() {
		http.Error(w, "Metamorph ability is not active or already used", http.StatusBadRequest)
		return
	}

	// Get the target (eliminated) player
	targetPlayer := room.GetPlayer(targetPlayerID)
	if targetPlayer == nil {
		http.Error(w, "Target player not found", http.StatusNotFound)
		return
	}

	// Verify target is eliminated
	if !targetPlayer.IsEliminated {
		http.Error(w, "Can only steal roles from eliminated players", http.StatusBadRequest)
		return
	}

	// Verify target still has a role
	if targetPlayer.Role == nil {
		http.Error(w, "Target player has no role to steal", http.StatusBadRequest)
		return
	}

	// Store original card IDs before stealing
	originalCardID := player.Role.GetID()
	stolenCardID := targetPlayer.Role.GetID()

	// Steal the role using StealRole utility
	err = room.StealRole(player, targetPlayer, true) // turnFaceDown=true unless Leader
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to steal role: %v", err), http.StatusInternalServerError)
		return
	}

	// Track the identity theft in TransformState (using the extended abstraction)
	player.AbilityState.StealIdentity(originalCardID, stolenCardID)

	// Mark Metamorph timing window as used
	player.AbilityState.UseMetamorph()

	h.store.UpdateRoom(room)

	log.Printf("🦎 %s stole %s's role (%s) in room %s", player.Name, targetPlayer.Name, player.Role.Name, roomCode)

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "role_stolen",
		RoomCode: room.Code,
		Data: map[string]interface{}{
			"room":    room,
			"stealer": player,
			"victim":  targetPlayer,
		},
	})

	w.WriteHeader(http.StatusOK)
}

// EndMetamorphEffect deactivates The Metamorph's steal ability window
// Called when the player clicks "End Turn" or the effect naturally expires
func (h *Handler) EndMetamorphEffect(w http.ResponseWriter, r *http.Request) {
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

	// Verify player has The Metamorph
	if player.Role == nil || player.Role.GetID() != 25 {
		http.Error(w, "Player does not have The Metamorph", http.StatusBadRequest)
		return
	}

	// Deactivate the ability
	if player.AbilityState != nil {
		player.AbilityState.DeactivateMetamorph()
	}

	h.store.UpdateRoom(room)

	log.Printf("🦎 Metamorph ability ended for %s in room %s", player.Name, roomCode)

	// Publish event
	h.eventBus.Publish(Event{
		Type:     "metamorph_ended",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// EliminatePlayer handles marking a player as eliminated from the game
// This is a manual action - players mark themselves as eliminated when they lose
func (h *Handler) EliminatePlayer(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	playerID := chi.URLParam(r, "playerID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get the requesting player
	playerCookie, err := r.Cookie("player_" + roomCode)
	if err != nil {
		http.Error(w, "Not in room", http.StatusUnauthorized)
		return
	}

	requestingPlayer := room.GetPlayer(playerCookie.Value)
	if requestingPlayer == nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
		return
	}

	// Get the player to be eliminated
	targetPlayer := room.GetPlayer(playerID)
	if targetPlayer == nil {
		http.Error(w, "Target player not found", http.StatusNotFound)
		return
	}

	// Only allow players to eliminate themselves (or hosts to eliminate anyone)
	if requestingPlayer.ID != targetPlayer.ID && !requestingPlayer.IsHost {
		http.Error(w, "You can only eliminate yourself", http.StatusForbidden)
		return
	}

	// Check if already eliminated
	if targetPlayer.IsEliminated {
		http.Error(w, "Player is already eliminated", http.StatusBadRequest)
		return
	}

	// Mark the player as eliminated
	targetPlayer.MarkEliminated()

	h.store.UpdateRoom(room)

	log.Printf("💀 Player %s has been eliminated in room %s", targetPlayer.Name, roomCode)

	// Publish elimination event
	h.eventBus.Publish(Event{
		Type:     "player_eliminated",
		RoomCode: room.Code,
		Data: map[string]interface{}{
			"room":              room,
			"eliminated_player": targetPlayer,
		},
	})

	w.WriteHeader(http.StatusOK)
}

// TriggerPuppetMasterAbility initiates The Puppet Master's ability when unveiled
// The Puppet Master (ID 27): "When The Puppet Master is unveiled, redistribute control
// of any number of other identity cards. Then turn face down each of those cards
// that isn't a Leader."
func (h *Handler) TriggerPuppetMasterAbility(w http.ResponseWriter, r *http.Request) {
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

	// Verify player has The Puppet Master (card ID 27)
	if player.Role == nil || player.Role.GetID() != 27 {
		http.Error(w, "Player does not have The Puppet Master", http.StatusBadRequest)
		return
	}

	// Check if player already has a pending Puppet Master ability
	if player.AbilityState != nil && player.AbilityState.HasPendingAbilities() {
		for _, pending := range player.AbilityState.PendingAbilities {
			if pending.CardID == 27 {
				http.Error(w, "Ability already in progress", http.StatusConflict)
				return
			}
		}
	}

	// Create pending ability with confirmation requirement
	leader := room.GetLeader()
	requiresConfirmation := leader != nil

	abilityID := fmt.Sprintf("puppet-master-%s-%d", playerID, rand.Int())
	pendingAbility := &ability.PendingAbility{
		ID:          abilityID,
		PlayerID:    playerID,
		CardID:      27,
		AbilityType: "puppet_master_redistribution",
		Data: map[string]interface{}{
			"step":             1, // Start at step 1 (player selection)
			"selected_players": []string{},
			"player_name":      player.Name,
			"card_name":        player.Role.Name,
		},
		ModalDismissed:       false,
		RequiresConfirmation: requiresConfirmation,
		ConfirmationRole:     "leader",
		ConfirmedBy:          []string{},
	}

	if player.AbilityState == nil {
		player.AbilityState = ability.NewAbilityState()
	}
	player.AbilityState.AddPendingAbility(pendingAbility)

	// Set card face up
	player.FaceUp = true
	player.RoleRevealed = true

	// Grant the permanent ability to view face-down cards
	player.AbilityState.GrantViewOthersFaceDown()

	h.store.UpdateRoom(room)

	log.Printf("🎭 Puppet Master ability triggered for %s in room %s", player.Name, roomCode)

	h.eventBus.Publish(Event{
		Type:     "ability_triggered",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// PuppetMasterSelectPlayers handles step 1 - selecting players to include in redistribution
func (h *Handler) PuppetMasterSelectPlayers(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get the requesting player
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

	// Verify ability has been confirmed (if required)
	if pendingAbility.RequiresConfirmation && !pendingAbility.IsConfirmed() {
		http.Error(w, "Ability has not been confirmed by the Leader yet", http.StatusForbidden)
		return
	}

	// Parse selected players from form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	selectedPlayers := r.Form["selected_players"]

	// Validate selected players (must be living players, not the Puppet Master)
	validatedPlayers := []string{}
	for _, pID := range selectedPlayers {
		p := room.GetPlayer(pID)
		if p != nil && !p.IsEliminated && p.ID != player.ID && p.Role != nil {
			validatedPlayers = append(validatedPlayers, pID)
		}
	}

	// Need at least 2 players to redistribute
	if len(validatedPlayers) < 2 {
		// If less than 2 players, skip to completion (can't swap with only 1 or 0)
		player.AbilityState.ResolvePendingAbility(abilityID)
		h.store.UpdateRoom(room)

		log.Printf("🎭 Puppet Master redistribution skipped (fewer than 2 players selected) for %s in room %s", player.Name, roomCode)

		h.eventBus.Publish(Event{
			Type:     "ability_resolved",
			RoomCode: room.Code,
			Data:     room,
		})

		w.WriteHeader(http.StatusOK)
		return
	}

	// Store selected players and move to step 2
	pendingAbility.Data["selected_players"] = validatedPlayers
	pendingAbility.Data["step"] = 2

	h.store.UpdateRoom(room)

	log.Printf("🎭 Puppet Master selected %d players for redistribution in room %s", len(validatedPlayers), roomCode)

	h.eventBus.Publish(Event{
		Type:     "ability_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// PuppetMasterSkip handles skipping the redistribution entirely
func (h *Handler) PuppetMasterSkip(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

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

	pendingAbility := player.AbilityState.GetPendingAbility(abilityID)
	if pendingAbility == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	if pendingAbility.PlayerID != player.ID {
		http.Error(w, "Ability does not belong to this player", http.StatusForbidden)
		return
	}

	// Resolve the ability without redistribution
	player.AbilityState.ResolvePendingAbility(abilityID)
	h.store.UpdateRoom(room)

	log.Printf("🎭 Puppet Master redistribution skipped for %s in room %s", player.Name, roomCode)

	h.eventBus.Publish(Event{
		Type:     "ability_resolved",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// PuppetMasterBack goes back to player selection (step 1)
func (h *Handler) PuppetMasterBack(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

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

	pendingAbility := player.AbilityState.GetPendingAbility(abilityID)
	if pendingAbility == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	if pendingAbility.PlayerID != player.ID {
		http.Error(w, "Ability does not belong to this player", http.StatusForbidden)
		return
	}

	// Reset to step 1
	pendingAbility.Data["step"] = 1
	pendingAbility.Data["selected_players"] = []string{}

	h.store.UpdateRoom(room)

	h.eventBus.Publish(Event{
		Type:     "ability_updated",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}

// PuppetMasterExecute performs the actual redistribution
func (h *Handler) PuppetMasterExecute(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	abilityID := chi.URLParam(r, "abilityID")

	room, err := h.store.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

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

	pendingAbility := player.AbilityState.GetPendingAbility(abilityID)
	if pendingAbility == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	if pendingAbility.PlayerID != player.ID {
		http.Error(w, "Ability does not belong to this player", http.StatusForbidden)
		return
	}

	// Parse form data for assignments
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get selected players
	selectedPlayerIDs, ok := pendingAbility.Data["selected_players"].([]string)
	if !ok || len(selectedPlayerIDs) < 2 {
		http.Error(w, "Invalid selected players", http.StatusBadRequest)
		return
	}

	// Build assignment map: playerID -> cardID they should receive
	assignments := make(map[string]int)
	for _, pID := range selectedPlayerIDs {
		cardIDStr := r.FormValue(fmt.Sprintf("assignment_%s", pID))
		cardID, err := strconv.Atoi(cardIDStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid card assignment for player %s", pID), http.StatusBadRequest)
			return
		}
		assignments[pID] = cardID
	}

	// Validate: each card should be assigned exactly once
	cardAssignmentCount := make(map[int]int)
	for _, cardID := range assignments {
		cardAssignmentCount[cardID]++
		if cardAssignmentCount[cardID] > 1 {
			http.Error(w, "Each card can only be assigned to one player", http.StatusBadRequest)
			return
		}
	}

	// Collect current roles for selected players
	playerRoles := make(map[string]*game.Card)
	for _, pID := range selectedPlayerIDs {
		p := room.GetPlayer(pID)
		if p != nil && p.Role != nil {
			playerRoles[pID] = p.Role
		}
	}

	// Execute the redistribution
	for playerID, cardID := range assignments {
		p := room.GetPlayer(playerID)
		if p == nil {
			continue
		}

		// Find the card by ID
		var newRole *game.Card
		for _, originalRole := range playerRoles {
			if originalRole.GetID() == cardID {
				newRole = originalRole
				break
			}
		}

		if newRole == nil {
			continue
		}

		// Assign the new role
		p.Role = newRole

		// Turn face down if not a Leader
		if newRole.GetRoleType() != game.RoleLeader {
			p.FaceUp = false
			p.RoleRevealed = false
		}
	}

	// Resolve the ability
	player.AbilityState.ResolvePendingAbility(abilityID)
	h.store.UpdateRoom(room)

	log.Printf("🎭 Puppet Master redistribution executed by %s in room %s", player.Name, roomCode)

	h.eventBus.Publish(Event{
		Type:     "puppet_master_redistribution",
		RoomCode: room.Code,
		Data:     room,
	})

	w.WriteHeader(http.StatusOK)
}
