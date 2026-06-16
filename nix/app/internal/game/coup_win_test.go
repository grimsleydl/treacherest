package game

import (
	"strings"
	"testing"
)

func TestRecordCoupKingFall_DefaultOneBlueHuntSatisfiedWhenOneBlueDead(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue-1", "Blue One", "Blue Knight", true),
		coupWinPlayer("blue-2", "Blue Two", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)

	RecordCoupKingFall(room)

	if !room.CoupGreenEligibleBeforeKingFall {
		t.Fatal("expected one dead Blue Knight to satisfy Green Hunt before King Fall by default")
	}
}

func TestRecordCoupKingFall_AllBluesHuntRequiresEveryBlueDead(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue-1", "Blue One", "Blue Knight", true),
		coupWinPlayer("blue-2", "Blue Two", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupGreenHuntRequirement = CoupGreenHuntAllBlues

	RecordCoupKingFall(room)

	if room.CoupGreenEligibleBeforeKingFall {
		t.Fatal("expected all-Blues Green Hunt to stay unsatisfied while one Blue Knight remains alive")
	}
}

func TestRecordCoupKingFall_BroadAmnestyLocksGreenHuntWhenInquisitionSucceededBeforeKingFall(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupInquisitionAmnesty = CoupInquisitionAmnestyBroad
	room.CoupInquisition = &CoupInquisitionState{Succeeded: true}

	RecordCoupKingFall(room)

	if !room.CoupGreenEligibleBeforeKingFall {
		t.Fatal("expected Broad Amnesty to satisfy Green Hunt before King Fall after successful Inquisition")
	}
}

func TestRecordCoupKingFall_DefaultAmnestyDoesNotLockRedShareAfterInquisition(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupInquisition = &CoupInquisitionState{Succeeded: true}

	RecordCoupKingFall(room)

	if room.CoupGreenEligibleBeforeKingFall {
		t.Fatal("expected default King-side-only Inquisition Amnesty not to satisfy Red-side Green lock")
	}
}

func TestRecordCoupKingFall_FailedInquisitionDoesNotSatisfyBroadAmnesty(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupInquisitionAmnesty = CoupInquisitionAmnestyBroad
	room.CoupInquisition = &CoupInquisitionState{
		Last:      &CoupInquisitionAttempt{Resolved: true, Success: false},
		Succeeded: false,
	}

	RecordCoupKingFall(room)

	if room.CoupGreenEligibleBeforeKingFall {
		t.Fatal("expected failed Inquisition not to satisfy Broad Amnesty before King Fall")
	}
}

func TestCurrentCoupAdvisoryWin_RedSharesGreenThroughBroadAmnestyBeforeKingFall(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupInquisitionAmnesty = CoupInquisitionAmnestyBroad
	room.CoupInquisition = &CoupInquisitionState{Succeeded: true}

	RecordCoupKingFall(room)
	room.GetPlayer("king").MarkEliminated()
	room.GetPlayer("black").MarkEliminated()

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Red advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeRed {
		t.Fatalf("expected Red outcome, got %q", prompt.Outcome)
	}
	if !prompt.GreenShares {
		t.Fatal("expected Green to share Red victory through Broad Amnesty")
	}
	assertCoupWinFact(t, prompt, "Broad Amnesty")
	assertCoupWinFact(t, prompt, "Inquisition")
	assertCoupWinFactAbsent(t, prompt, "eligib")
}

func TestCurrentCoupAdvisoryWin_DefaultAmnestyDoesNotShareRedAfterInquisition(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupInquisition = &CoupInquisitionState{Succeeded: true}

	RecordCoupKingFall(room)
	room.GetPlayer("king").MarkEliminated()
	room.GetPlayer("black").MarkEliminated()

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Red advisory prompt")
	}
	if prompt.GreenShares {
		t.Fatal("expected King-side-only Inquisition Amnesty not to share Red victory")
	}
	assertCoupWinFact(t, prompt, "King-side Inquisition Amnesty does not apply to Red victories")
}

func TestCurrentCoupAdvisoryWin_BroadAmnestyDoesNotRetroactivelyShareRed(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", true),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", true),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupKingFallen = true
	room.CoupGreenEligibleBeforeKingFall = false
	room.CoupInquisitionAmnesty = CoupInquisitionAmnestyBroad
	room.CoupInquisition = &CoupInquisitionState{Succeeded: true}

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Red advisory prompt")
	}
	if prompt.GreenShares {
		t.Fatal("expected Broad Amnesty not to retroactively share Red after King Fall")
	}
	assertCoupWinFact(t, prompt, "Broad Amnesty was not satisfied before King Fall")
}

func TestCurrentCoupAdvisoryWin_DefaultKingSideAmnestySharesKingAfterInquisition(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", true),
		coupWinPlayer("red", "Red Player", "Red Knight", true),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
		coupWinPlayer("waste", "Wasteland Player", "Wasteland Knight", true),
	)
	room.CoupInquisition = &CoupInquisitionState{Succeeded: true}

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected King-side advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeKingSide {
		t.Fatalf("expected King-side outcome, got %q", prompt.Outcome)
	}
	if !prompt.GreenShares {
		t.Fatal("expected eligible Green to share King-side prompt")
	}
	assertCoupWinFact(t, prompt, "King-side Inquisition Amnesty")
	assertCoupWinFactAbsent(t, prompt, "eligib")
}

func TestCurrentCoupAdvisoryWin_KingSideSharesWithGreenWhenHuntIsSatisfied(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue-1", "Blue One", "Blue Knight", true),
		coupWinPlayer("blue-2", "Blue Two", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", true),
		coupWinPlayer("red", "Red Player", "Red Knight", true),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
		coupWinPlayer("waste", "Wasteland Player", "Wasteland Knight", true),
	)

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected King-side advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeKingSide {
		t.Fatalf("expected King-side outcome, got %q", prompt.Outcome)
	}
	if !prompt.GreenShares {
		t.Fatal("expected Green to share King-side victory when Green Hunt is satisfied")
	}
	assertCoupWinFact(t, prompt, "Green Hunt is satisfied")
}

func TestCurrentCoupAdvisoryWin_BlackWhenKingDeadAndRedDead(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", true),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", true),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Black advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeBlack {
		t.Fatalf("expected Black outcome, got %q", prompt.Outcome)
	}
	if prompt.GreenShares {
		t.Fatal("expected Green not to share Black victory")
	}
	assertCoupWinFact(t, prompt, "King has fallen")
	assertCoupWinFact(t, prompt, "Red Knight is eliminated")
}

func TestCurrentCoupAdvisoryWin_RedDoesNotShareGreenWhenBlueDiesAfterKing(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", true),
		coupWinPlayer("blue", "Blue Player", "Blue Knight", true),
		coupWinPlayer("black", "Black Player", "Black Knight", true),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupKingFallen = true
	room.CoupGreenEligibleBeforeKingFall = false

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Red advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeRed {
		t.Fatalf("expected Red outcome, got %q", prompt.Outcome)
	}
	if prompt.GreenShares {
		t.Fatal("expected Green not to share Red victory when eligibility was not locked before King fell")
	}
	assertCoupWinFact(t, prompt, "Green Hunt was not satisfied before King Fall")
}

func TestCurrentCoupAdvisoryWin_RedSharesGreenWhenHuntWasSatisfiedBeforeKingFall(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", true),
		coupWinPlayer("blue-1", "Blue One", "Blue Knight", true),
		coupWinPlayer("blue-2", "Blue Two", "Blue Knight", false),
		coupWinPlayer("black", "Black Player", "Black Knight", true),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupKingFallen = true
	room.CoupGreenEligibleBeforeKingFall = true

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Red advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeRed {
		t.Fatalf("expected Red outcome, got %q", prompt.Outcome)
	}
	if !prompt.GreenShares {
		t.Fatal("expected Green to share Red victory when Green Hunt was satisfied before King Fall")
	}
	assertCoupWinFact(t, prompt, "Green Hunt was satisfied before King Fall")
}

func TestCurrentCoupAdvisoryWin_RedSharesGreenWhenAllBluesHuntSatisfiedBeforeKingFall(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", false),
		coupWinPlayer("blue-1", "Blue One", "Blue Knight", true),
		coupWinPlayer("blue-2", "Blue Two", "Blue Knight", true),
		coupWinPlayer("black", "Black Player", "Black Knight", false),
		coupWinPlayer("red", "Red Player", "Red Knight", false),
		coupWinPlayer("green", "Green Player", "Green Knight", false),
	)
	room.CoupGreenHuntRequirement = CoupGreenHuntAllBlues

	RecordCoupKingFall(room)
	room.GetPlayer("king").MarkEliminated()
	room.GetPlayer("black").MarkEliminated()

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Red advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeRed {
		t.Fatalf("expected Red outcome, got %q", prompt.Outcome)
	}
	if !prompt.GreenShares {
		t.Fatal("expected Green to share Red victory when all-Blues Hunt was satisfied before King Fall")
	}
	assertCoupWinFact(t, prompt, "Green Hunt was satisfied before King Fall")
}

func TestCurrentCoupAdvisoryWin_WastelandSoleSurvivor(t *testing.T) {
	room := newCoupWinRoom(
		coupWinPlayer("king", "King Player", "King", true),
		coupWinPlayer("black", "Black Player", "Black Knight", true),
		coupWinPlayer("red", "Red Player", "Red Knight", true),
		coupWinPlayer("green", "Green Player", "Green Knight", true),
		coupWinPlayer("waste", "Wasteland Player", "Wasteland Knight", false),
	)

	prompt := CurrentCoupAdvisoryWin(room)

	if prompt == nil {
		t.Fatal("expected Wasteland advisory prompt")
	}
	if prompt.Outcome != CoupWinOutcomeWasteland {
		t.Fatalf("expected Wasteland outcome, got %q", prompt.Outcome)
	}
	if prompt.GreenShares {
		t.Fatal("expected Wasteland not to share victory")
	}
	assertCoupWinFact(t, prompt, "sole surviving player")
}

func assertCoupWinFact(t *testing.T, prompt *CoupWinPrompt, want string) {
	t.Helper()

	for _, fact := range prompt.Facts {
		if strings.Contains(fact, want) {
			return
		}
	}
	t.Fatalf("expected prompt facts to contain %q, got %#v", want, prompt.Facts)
}

func assertCoupWinFactAbsent(t *testing.T, prompt *CoupWinPrompt, unwanted string) {
	t.Helper()

	for _, fact := range prompt.Facts {
		if strings.Contains(fact, unwanted) {
			t.Fatalf("expected prompt facts not to contain %q, got %#v", unwanted, prompt.Facts)
		}
	}
}

func newCoupWinRoom(players ...*Player) *Room {
	room := &Room{
		Code:      "COUPWIN",
		State:     StatePlaying,
		RulesMode: RulesModeCoup,
		Players:   make(map[string]*Player),
	}
	for _, player := range players {
		room.Players[player.ID] = player
	}
	return room
}

func coupWinPlayer(id, name, roleName string, eliminated bool) *Player {
	player := &Player{
		ID:           id,
		Name:         name,
		Role:         coupWinCard(roleName),
		RoleRevealed: eliminated,
		FaceUp:       !eliminated,
		IsEliminated: eliminated,
	}
	return player
}

func coupWinCard(name string) *Card {
	return &Card{
		ID:   len(name),
		Name: name,
		Types: CardTypes{
			Supertype: "Coup",
			Subtype:   name,
		},
	}
}
