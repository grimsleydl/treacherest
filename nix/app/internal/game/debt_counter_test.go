package game

import (
	"errors"
	"testing"
)

func TestNewDealtRoleCardClonesEveryIdentityCardForDebtCounters(t *testing.T) {
	catalogCard := &Card{ID: 7, Name: "Catalog Secret", Types: CardTypes{Subtype: "Guardian"}}
	dealt := NewDealtRoleCard(catalogCard, 4)
	if dealt == catalogCard {
		t.Fatal("ordinary dealt role reused the shared catalog card")
	}
	dealt.DebtCounters = &DebtCounterState{Count: 2}
	if catalogCard.DebtCounters != nil {
		t.Fatal("debt counter state leaked onto the shared catalog card")
	}
}

func TestDebtCollectorPlacesPublicCounterOnOtherFaceDownIdentity(t *testing.T) {
	room, collector, target := testDebtCounterRoom()
	if target.FaceUp || target.RoleRevealed {
		t.Fatal("test target must begin fully hidden")
	}

	placement, err := room.PlaceDebtCounter(collector.ID, target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if target.Role.DebtCounters == nil || target.Role.DebtCounters.Count != 1 {
		t.Fatalf("target debt state = %+v, want one counter", target.Role.DebtCounters)
	}
	if placement.TargetPlayerName != target.Name || placement.Before != 0 || placement.After != 1 {
		t.Fatalf("placement record = %+v", placement)
	}
	if target.FaceUp || target.RoleRevealed {
		t.Fatal("placing a visible counter disclosed or changed the hidden role")
	}

	if _, err := room.PlaceDebtCounter(collector.ID, collector.ID); !errors.Is(err, ErrDebtCounterTarget) {
		t.Fatalf("self-target error = %v, want %v", err, ErrDebtCounterTarget)
	}
	nonCollector := NewPlayer("other", "Other", "other-session")
	nonCollector.Role = NewDealtRoleCard(&Card{ID: 8, Name: "Other Role"}, 3)
	room.Players[nonCollector.ID] = nonCollector
	if _, err := room.PlaceDebtCounter(nonCollector.ID, target.ID); !errors.Is(err, ErrDebtCollectorUnavailable) {
		t.Fatalf("non-collector placement error = %v, want %v", err, ErrDebtCollectorUnavailable)
	}
}

func TestDebtCountersPersistAcrossFaceTransitionsAndEveryTransferPath(t *testing.T) {
	assertState := func(t *testing.T, destination *Player, card *Card, state *DebtCounterState) {
		t.Helper()
		if destination.Role != card {
			t.Fatal("destination did not receive the same identity-card instance")
		}
		if destination.Role.DebtCounters != state || state.Count != 1 || len(state.Placements) != 1 {
			t.Fatalf("debt state did not follow the card: %+v", destination.Role.DebtCounters)
		}
	}
	newTransferRoom := func(t *testing.T) (*Room, *Player, *Player, *Card, *DebtCounterState) {
		t.Helper()
		room, collector, source := testDebtCounterRoom()
		destination := NewPlayer("destination", "Destination", "destination-session")
		destination.Role = NewDealtRoleCard(&Card{ID: 9, Name: "Destination Role", Types: CardTypes{Subtype: "Assassin"}}, 3)
		destination.FaceUp = false
		room.Players[destination.ID] = destination
		if _, err := room.PlaceDebtCounter(collector.ID, source.ID); err != nil {
			t.Fatal(err)
		}
		return room, source, destination, source.Role, source.Role.DebtCounters
	}

	t.Run("unveil then re-hide", func(t *testing.T) {
		room, collector, target := testDebtCounterRoom()
		if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
			t.Fatal(err)
		}
		card, state := target.Role, target.Role.DebtCounters
		target.FaceUp = true
		target.RoleRevealed = true
		assertState(t, target, card, state)
		target.FaceUp = false
		assertState(t, target, card, state)
	})

	t.Run("TransferRole", func(t *testing.T) {
		room, source, destination, card, state := newTransferRoom(t)
		if err := room.TransferRole(source, destination, true); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, state)
	})

	t.Run("SwapRoles", func(t *testing.T) {
		room, source, destination, card, state := newTransferRoom(t)
		if err := room.SwapRoles(source, destination, true); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, state)
	})

	t.Run("StealRole", func(t *testing.T) {
		room, source, destination, card, state := newTransferRoom(t)
		source.IsEliminated = true
		if err := room.StealRole(destination, source, true); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, state)
	})

	t.Run("Puppet Master redistribution", func(t *testing.T) {
		room, source, destination, card, state := newTransferRoom(t)
		otherCard := destination.Role
		if err := room.RedistributeRoles(map[string]*Card{
			source.ID:      otherCard,
			destination.ID: card,
		}); err != nil {
			t.Fatal(err)
		}
		assertState(t, destination, card, state)
	})
}

func TestDebtCounterUpkeepAndLossAreRecordedAdvisoriesOnly(t *testing.T) {
	room, collector, target := testDebtCounterRoom()
	if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
		t.Fatal(err)
	}

	upkeep, err := room.RecordDebtCounterUpkeep(collector.ID)
	if err != nil {
		t.Fatal(err)
	}
	if upkeep.TotalLifeGain != 2 || len(upkeep.Assessments) != 1 || upkeep.Assessments[0].Count != 2 {
		t.Fatalf("upkeep advisory = %+v", upkeep)
	}
	if target.IsEliminated {
		t.Fatal("recording upkeep changed unrelated table state")
	}

	losses := room.RecordDebtCounterLossAdvisories(target.ID)
	if len(losses) != 1 || losses[0].DrawCount != 2 || losses[0].LostPlayerName != target.Name {
		t.Fatalf("loss advisories = %+v", losses)
	}
	collectorState := collector.Role.DebtCounters
	if len(collectorState.UpkeepAdvisories) != 1 || len(collectorState.LossAdvisories) != 1 {
		t.Fatalf("collector advisory log = %+v", collectorState)
	}
	if target.Role.DebtCounters.Count != 2 {
		t.Fatal("advisory recording changed the physical debt counter count")
	}
}

func TestDebtCountersSurviveEncryptedBackupRoundTrip(t *testing.T) {
	room, collector, target := testDebtCounterRoom()
	if _, err := room.PlaceDebtCounter(collector.ID, target.ID); err != nil {
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
	restoredTarget := restored.Players[target.ID]
	if restoredTarget == nil || restoredTarget.Role == nil || restoredTarget.Role.DebtCounters == nil {
		t.Fatalf("restored target debt state = %#v", restoredTarget)
	}
	if restoredTarget.Role.DebtCounters.Count != 1 || len(restoredTarget.Role.DebtCounters.Placements) != 1 {
		t.Fatalf("restored debt state = %+v", restoredTarget.Role.DebtCounters)
	}
	if restoredTarget.FaceUp || restoredTarget.RoleRevealed {
		t.Fatal("backup restore changed hidden-card visibility")
	}
}

func TestDebtCollectorIsSupportedWithoutProliferateIntegration(t *testing.T) {
	if !IsCardSupported(TheDebtCollectorCardID) {
		t.Fatal("The Debt Collector remains excluded from deals")
	}
	if IsCardSupported(54) {
		t.Fatal("The Gathering unexpectedly left the unsupported list")
	}
	// Identity-card debt counters are intentionally absent from the permanent
	// counter/proliferate ability subsystem; their only mutation is the room
	// placement command exercised above.
}

func testDebtCounterRoom() (*Room, *Player, *Player) {
	collector := NewPlayer("collector", "Alice", "collector-session")
	collector.Role = NewDealtRoleCard(&Card{
		ID: TheDebtCollectorCardID, Name: "The Debt Collector", Types: CardTypes{Subtype: "Leader"},
	}, 3)
	collector.FaceUp = true
	collector.RoleRevealed = true
	target := NewPlayer("target", "Bob", "target-session")
	target.Role = NewDealtRoleCard(&Card{
		ID: 7, Name: "Hidden Target Role", Types: CardTypes{Subtype: "Guardian"},
	}, 3)
	target.FaceUp = false
	target.RoleRevealed = false
	room := &Room{
		Code:  "DEBTS",
		State: StatePlaying,
		Players: map[string]*Player{
			collector.ID: collector,
			target.ID:    target,
		},
	}
	return room, collector, target
}
