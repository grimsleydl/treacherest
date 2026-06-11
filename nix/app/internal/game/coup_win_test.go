package game

import (
	"strings"
	"testing"
)

func TestCurrentCoupAdvisoryWin_KingSideSharesWithGreenAfterInquisition(t *testing.T) {
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
	assertCoupWinFact(t, prompt, "Inquisition has succeeded")
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
	assertCoupWinFact(t, prompt, "Green Knight is not eligible")
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
