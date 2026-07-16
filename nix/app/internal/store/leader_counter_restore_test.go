package store

import (
	"testing"

	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
)

func TestRegisterRestoredRoomPreservesIdentityCardCounterState(t *testing.T) {
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore(config.DefaultConfig())
	store.SetCardService(cardService)

	var catalogCard *game.Card
	for _, card := range cardService.GetAllCards() {
		if card.ID == game.HerSeedbornHighnessCardID {
			catalogCard = card
			break
		}
	}
	if catalogCard == nil {
		t.Fatal("counter Leader missing from card catalog")
	}
	role := game.NewDealtRoleCard(catalogCard, 6)
	owner := game.NewPlayer("owner", "Alice", "owner-session")
	owner.Role = role
	owner.RoleRevealed = true
	owner.FaceUp = true
	room := &game.Room{
		Code:    "RESTO",
		State:   game.StatePlaying,
		Players: map[string]*game.Player{owner.ID: owner},
	}
	if _, err := room.ActivateLeaderCounter(owner.ID); err != nil {
		t.Fatal(err)
	}
	wantPool := role.LeaderCounterPool

	if err := store.RegisterRestoredRoom(room); err != nil {
		t.Fatal(err)
	}
	restored, err := store.GetRoom(room.Code)
	if err != nil {
		t.Fatal(err)
	}
	gotRole := restored.GetPlayer(owner.ID).Role
	if gotRole == catalogCard {
		t.Fatal("restore attached mutable state to the shared catalog card")
	}
	if gotRole.LeaderCounterPool != wantPool {
		t.Fatal("restore discarded the identity-card counter state")
	}
	if gotRole.LeaderCounterPool.Remaining != 5 || len(gotRole.LeaderCounterPool.Activations) != 1 {
		t.Fatalf("restored counter state = %+v", gotRole.LeaderCounterPool)
	}
}
