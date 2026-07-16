package game

import (
	"errors"
	"testing"
)

func TestGatheringChecklistInitializesOnDealtCardAndRecordsOnlyPublicChoice(t *testing.T) {
	catalogCard := testGatheringCard()
	dealt := NewDealtRoleCard(catalogCard, 4)
	if dealt == catalogCard {
		t.Fatal("dealt Gathering reused the shared catalog card")
	}
	if dealt.GatheringChecklist == nil {
		t.Fatal("dealt Gathering has no card-attached checklist")
	}
	if catalogCard.GatheringChecklist != nil {
		t.Fatal("Gathering checklist leaked onto the shared catalog card")
	}

	room, owner := testGatheringRoom()
	choice, err := room.ChooseGatheringMode(owner.ID, GatheringModeWhite)
	if err != nil {
		t.Fatal(err)
	}
	if choice.Number != 1 || choice.Mode != GatheringModeWhite || choice.PlayerName != owner.Name {
		t.Fatalf("choice = %+v", choice)
	}
	if !owner.Role.GatheringChecklist.HasChosen(GatheringModeWhite) {
		t.Fatal("chosen mode was not permanently marked")
	}
	if owner.IsEliminated || !owner.FaceUp || !owner.RoleRevealed {
		t.Fatal("recording a choice mutated table-adjudicated game state")
	}

	if _, err := room.ChooseGatheringMode(owner.ID, GatheringModeWhite); !errors.Is(err, ErrGatheringModeChosen) {
		t.Fatalf("duplicate choice error = %v, want %v", err, ErrGatheringModeChosen)
	}
	if _, err := room.ChooseGatheringMode(owner.ID, GatheringMode("purple")); !errors.Is(err, ErrGatheringModeInvalid) {
		t.Fatalf("invalid choice error = %v, want %v", err, ErrGatheringModeInvalid)
	}
	if got := len(owner.Role.GatheringChecklist.Choices); got != 1 {
		t.Fatalf("rejected choices changed checklist length to %d", got)
	}
}

func TestGatheringModeDefinitionsMatchPrintedCard(t *testing.T) {
	want := []GatheringModeDefinition{
		{Mode: GatheringModeWhite, Color: "White", Effect: "Create four 1/1 white Soldier creature tokens."},
		{Mode: GatheringModeBlue, Color: "Blue", Effect: "Scry 4, then draw two cards."},
		{Mode: GatheringModeBlack, Color: "Black", Effect: "Return target creature card from your graveyard to the battlefield."},
		{Mode: GatheringModeRed, Color: "Red", Effect: "The Gathering deals 4 damage divided as you choose among any number of targets."},
		{Mode: GatheringModeGreen, Color: "Green", Effect: "Destroy target noncreature permanent."},
	}
	got := GatheringModeDefinitions()
	if len(got) != len(want) {
		t.Fatalf("mode count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("mode %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestGatheringChecklistSurvivesFaceChangesAndEveryIdentityCardTransferPath(t *testing.T) {
	assertState := func(t *testing.T, destination *Player, card *Card, checklist *GatheringChecklist) {
		t.Helper()
		if destination.Role != card {
			t.Fatal("destination did not receive the same Gathering card instance")
		}
		if destination.Role.GatheringChecklist != checklist || !checklist.HasChosen(GatheringModeBlue) {
			t.Fatalf("Gathering checklist did not follow the card: %+v", destination.Role.GatheringChecklist)
		}
		if !destination.FaceUp || !destination.RoleRevealed {
			t.Fatal("transferred Gathering Leader was not forced face up and public")
		}
	}
	newTransferRoom := func(t *testing.T) (*Room, *Player, *Player, *Card, *GatheringChecklist) {
		t.Helper()
		room, source := testGatheringRoom()
		destination := NewPlayer("destination", "Destination", "destination-session")
		destination.Role = NewDealtRoleCard(&Card{ID: 7, Name: "Destination Role", Types: CardTypes{Subtype: "Guardian"}}, 3)
		room.Players[destination.ID] = destination
		if _, err := room.ChooseGatheringMode(source.ID, GatheringModeBlue); err != nil {
			t.Fatal(err)
		}
		return room, source, destination, source.Role, source.Role.GatheringChecklist
	}

	t.Run("face transitions", func(t *testing.T) {
		room, owner := testGatheringRoom()
		if _, err := room.ChooseGatheringMode(owner.ID, GatheringModeBlue); err != nil {
			t.Fatal(err)
		}
		card, checklist := owner.Role, owner.Role.GatheringChecklist
		owner.FaceUp = false
		owner.RoleRevealed = false
		if owner.Role != card || owner.Role.GatheringChecklist != checklist || !checklist.HasChosen(GatheringModeBlue) {
			t.Fatal("face-down transition discarded the Gathering checklist")
		}
		owner.FaceUp = true
		owner.RoleRevealed = true
		assertState(t, owner, card, checklist)
	})

	t.Run("TransferRole", func(t *testing.T) {
		room, source, destination, card, checklist := newTransferRoom(t)
		if err := room.TransferRole(source, destination, true); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, checklist)
	})

	t.Run("SwapRoles", func(t *testing.T) {
		room, source, destination, card, checklist := newTransferRoom(t)
		if err := room.SwapRoles(source, destination, true); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, checklist)
	})

	t.Run("StealRole", func(t *testing.T) {
		room, source, destination, card, checklist := newTransferRoom(t)
		source.IsEliminated = true
		if err := room.StealRole(destination, source, true); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, checklist)
	})

	t.Run("Puppet Master redistribution", func(t *testing.T) {
		room, source, destination, card, checklist := newTransferRoom(t)
		otherCard := destination.Role
		if err := room.RedistributeRoles(map[string]*Card{
			source.ID:      otherCard,
			destination.ID: card,
		}); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, checklist)
	})
}

func TestGatheringChecklistSurvivesEncryptedBackupRoundTrip(t *testing.T) {
	room, owner := testGatheringRoom()
	if _, err := room.ChooseGatheringMode(owner.ID, GatheringModeGreen); err != nil {
		t.Fatal(err)
	}
	service, err := NewBackupService(testEncryptionKey(), true)
	if err != nil {
		t.Fatal(err)
	}
	backup, err := service.CreateBackup(room)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := service.RestoreBackup(backup, room.Code)
	if err != nil {
		t.Fatal(err)
	}
	role := restored.GetPlayer(owner.ID).Role
	if role == nil || role.GatheringChecklist == nil || !role.GatheringChecklist.HasChosen(GatheringModeGreen) {
		t.Fatalf("restored Gathering checklist = %#v", role)
	}
}

func TestGatheringIsSupportedAndUnsupportedMechanismIsEmpty(t *testing.T) {
	if !IsCardSupported(TheGatheringCardID) {
		t.Fatal("The Gathering remains excluded from deals")
	}
	if len(unsupportedCardIDs) != 0 {
		t.Fatalf("unsupported card set has %d entries, want empty: %+v", len(unsupportedCardIDs), unsupportedCardIDs)
	}
}

func testGatheringRoom() (*Room, *Player) {
	owner := NewPlayer("owner", "Alice", "owner-session")
	owner.Role = NewDealtRoleCard(testGatheringCard(), 3)
	owner.FaceUp = true
	owner.RoleRevealed = true
	room := &Room{
		Code:    "GATHR",
		State:   StatePlaying,
		Players: map[string]*Player{owner.ID: owner},
	}
	return room, owner
}

func testGatheringCard() *Card {
	return &Card{
		ID: TheGatheringCardID, Name: "The Gathering", Type: "Identity — Leader", Types: CardTypes{Subtype: "Leader"},
	}
}
