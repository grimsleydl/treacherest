package store

import (
	"testing"

	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
)

func TestRegisterRestoredRoomPreservesDebtCountersOnIdentityCard(t *testing.T) {
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore(config.DefaultConfig())
	store.SetCardService(cardService)

	var collectorCard, targetCard *game.Card
	for _, card := range cardService.GetAllCards() {
		switch card.ID {
		case game.TheDebtCollectorCardID:
			collectorCard = card
		case 1:
			targetCard = card
		}
	}
	if collectorCard == nil || targetCard == nil {
		t.Fatal("required cards missing from catalog")
	}

	collector := game.NewPlayer("collector", "Alice", "collector-session")
	collector.Role = game.NewDealtRoleCard(collectorCard, 3)
	collector.FaceUp = true
	collector.RoleRevealed = true
	target := game.NewPlayer("target", "Bob", "target-session")
	target.Role = game.NewDealtRoleCard(targetCard, 3)
	target.FaceUp = false
	target.RoleRevealed = false
	room := &game.Room{
		Code:  "DEBTR",
		State: game.StatePlaying,
		Players: map[string]*game.Player{
			collector.ID: collector,
			target.ID:    target,
		},
	}
	if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
		t.Fatal(err)
	}
	wantState := target.Role.DebtCounters

	if err := store.RegisterRestoredRoom(room); err != nil {
		t.Fatal(err)
	}
	restored, err := store.GetRoom(room.Code)
	if err != nil {
		t.Fatal(err)
	}
	gotRole := restored.GetPlayer(target.ID).Role
	if gotRole == targetCard {
		t.Fatal("restore attached mutable debt state to the shared catalog card")
	}
	if gotRole.DebtCounters != wantState {
		t.Fatal("restore discarded the identity-card debt counter state")
	}
	if gotRole.DebtCounters.Count != 1 || len(gotRole.DebtCounters.Placements) != 1 {
		t.Fatalf("restored debt state = %+v", gotRole.DebtCounters)
	}
	if gotRole.Name != targetCard.Name {
		t.Fatal("restore did not refresh immutable catalog card data")
	}
}
