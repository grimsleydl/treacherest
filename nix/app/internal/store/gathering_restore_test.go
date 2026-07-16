package store

import (
	"testing"

	"treacherest"
	"treacherest/internal/config"
	"treacherest/internal/game"
)

func TestRegisterRestoredRoomPreservesGatheringChecklist(t *testing.T) {
	cardService, err := game.NewCardService(treacherest.TreacheryCardsJSON, treacherest.CardImagesFS)
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore(config.DefaultConfig())
	store.SetCardService(cardService)

	var catalogCard *game.Card
	for _, card := range cardService.GetAllCards() {
		if card.ID == game.TheGatheringCardID {
			catalogCard = card
			break
		}
	}
	if catalogCard == nil {
		t.Fatal("The Gathering missing from card catalog")
	}
	role := game.NewDealtRoleCard(catalogCard, 3)
	owner := game.NewPlayer("owner", "Alice", "owner-session")
	owner.Role = role
	owner.RoleRevealed = true
	owner.FaceUp = true
	room := &game.Room{
		Code:    "GATHR",
		State:   game.StatePlaying,
		Players: map[string]*game.Player{owner.ID: owner},
	}
	if _, err := room.ChooseGatheringMode(owner.ID, game.GatheringModeRed); err != nil {
		t.Fatal(err)
	}
	wantChecklist := role.GatheringChecklist

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
	if gotRole.GatheringChecklist != wantChecklist {
		t.Fatal("restore discarded the Gathering checklist")
	}
	if !gotRole.GatheringChecklist.HasChosen(game.GatheringModeRed) || len(gotRole.GatheringChecklist.Choices) != 1 {
		t.Fatalf("restored Gathering checklist = %+v", gotRole.GatheringChecklist)
	}
	if gotRole.Text != catalogCard.Text {
		t.Fatal("restore did not refresh immutable catalog card data")
	}
}
