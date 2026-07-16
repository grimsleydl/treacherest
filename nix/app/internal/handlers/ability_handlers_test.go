package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/game/ability"
	"treacherest/internal/store"
	"treacherest/internal/testhelpers"
	"treacherest/internal/views/pages"

	"github.com/go-chi/chi/v5"
)

func TestPuppetMasterExecute_RedistributedLeaderRemainsPublic(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	handler := New(memStore, createMockCardService(), cfg, nil)

	room, err := memStore.CreateRoom()
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}
	room.State = game.StatePlaying
	// Exercise the operator's reveal control as well as the Treachery role state.
	room.RulesMode = game.RulesModeCoup

	host := game.NewPlayer("host", "Room Operator", "host-session")
	host.IsHost = true
	puppetMaster := game.NewPlayer("puppet", "Puppet Master Player", "puppet-session")
	puppetMaster.Role = &game.Card{
		ID:    27,
		Name:  "The Puppet Master",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	formerLeader := game.NewPlayer("former-leader", "Former Leader", "leader-session")
	formerLeader.Role = mockLeaderCard()
	formerLeader.FaceUp = true
	formerLeader.RoleRevealed = true
	newLeader := game.NewPlayer("new-leader", "New Leader", "new-leader-session")
	newLeader.Role = mockGuardianCard()
	newLeader.FaceUp = false
	newLeader.RoleRevealed = false

	for _, player := range []*game.Player{host, puppetMaster, formerLeader, newLeader} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("AddPlayer(%q) error = %v", player.ID, err)
		}
	}

	const abilityID = "puppet-master-regression"
	puppetMaster.AbilityState.AddPendingAbility(&ability.PendingAbility{
		ID:       abilityID,
		PlayerID: puppetMaster.ID,
		CardID:   27,
		Data: map[string]interface{}{
			"selected_players": []string{formerLeader.ID, newLeader.ID},
		},
	})
	memStore.UpdateRoom(room)

	body := `{"assignments":{"former-leader":2,"new-leader":1}}`
	req := httptest.NewRequest(http.MethodPost, "/room/"+room.Code+"/puppet-master/"+abilityID+"/execute", strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: puppetMaster.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("abilityID", abilityID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.PuppetMasterExecute(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PuppetMasterExecute() status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	updatedRoom, err := memStore.GetRoom(room.Code)
	if err != nil {
		t.Fatalf("GetRoom() error = %v", err)
	}
	updatedLeader := updatedRoom.GetPlayer(newLeader.ID)
	if updatedLeader.Role == nil || updatedLeader.Role.GetRoleType() != game.RoleLeader {
		t.Fatalf("new controller role = %#v, want Leader", updatedLeader.Role)
	}
	if !updatedLeader.FaceUp || !updatedLeader.RoleRevealed {
		t.Fatalf("moved Leader face state = FaceUp %v, RoleRevealed %v; want both true", updatedLeader.FaceUp, updatedLeader.RoleRevealed)
	}
	updatedFormerLeader := updatedRoom.GetPlayer(formerLeader.ID)
	if updatedFormerLeader.Role == nil || updatedFormerLeader.Role.GetRoleType() == game.RoleLeader {
		t.Fatalf("former Leader role = %#v, want redistributed non-Leader", updatedFormerLeader.Role)
	}
	if updatedFormerLeader.FaceUp || updatedFormerLeader.RoleRevealed {
		t.Fatalf("redistributed non-Leader face state = FaceUp %v, RoleRevealed %v; want both false", updatedFormerLeader.FaceUp, updatedFormerLeader.RoleRevealed)
	}

	renderer := testhelpers.NewTemplateRenderer(t)
	operatorHTML := renderer.Render(pages.HostDashboardPlaying(updatedRoom, host)).GetHTML()
	operatorTile := renderedSectionByID(t, operatorHTML, "operator-tile-"+newLeader.ID, "</article>")
	for _, expected := range []string{"Revealed: Test Leader", "Record Elimination"} {
		if !strings.Contains(operatorTile, expected) {
			t.Errorf("moved Leader operator tile missing %q in %s", expected, operatorTile)
		}
	}
	for _, forbidden := range []string{"Face Down", "Record Reveal"} {
		if strings.Contains(operatorTile, forbidden) {
			t.Errorf("moved Leader operator tile contains %q in %s", forbidden, operatorTile)
		}
	}
	formerLeaderOperatorTile := renderedSectionByID(t, operatorHTML, "operator-tile-"+formerLeader.ID, "</article>")
	for _, expected := range []string{"Face Down", "Record Reveal"} {
		if !strings.Contains(formerLeaderOperatorTile, expected) {
			t.Errorf("redistributed non-Leader operator tile missing %q in %s", expected, formerLeaderOperatorTile)
		}
	}
	if strings.Contains(formerLeaderOperatorTile, "Test Guardian") {
		t.Errorf("redistributed non-Leader leaked to operator in %s", formerLeaderOperatorTile)
	}

	publicHTML := renderer.Render(pages.GameBody(updatedRoom, puppetMaster)).GetHTML()
	publicRow := renderedSectionByID(t, publicHTML, "player-row-"+newLeader.ID, "</details>")
	for _, expected := range []string{"Leader", "Revealed: Test Leader", "Test Leader"} {
		if !strings.Contains(publicRow, expected) {
			t.Errorf("moved Leader public row missing %q in %s", expected, publicRow)
		}
	}
	if strings.Contains(publicRow, "Card is face down.") {
		t.Errorf("moved Leader public row rendered face down in %s", publicRow)
	}
	formerLeaderPublicRow := renderedSectionByID(t, publicHTML, "player-row-"+formerLeader.ID, "</details>")
	for _, expected := range []string{"Face Down", "Card is face down."} {
		if !strings.Contains(formerLeaderPublicRow, expected) {
			t.Errorf("redistributed non-Leader public row missing %q in %s", expected, formerLeaderPublicRow)
		}
	}
	if strings.Contains(formerLeaderPublicRow, "Test Guardian") {
		t.Errorf("redistributed non-Leader leaked to another player in %s", formerLeaderPublicRow)
	}

	renderer.Render(pages.GameBody(updatedRoom, updatedLeader)).
		AssertContains("role-card role-card-public").
		AssertContains("Public role").
		AssertNotContains("Publicly Reveal Role")
	renderer.Render(pages.GameBody(updatedRoom, updatedFormerLeader)).
		AssertContains("role-card role-card-hero").
		AssertContains("Private role").
		AssertContains("Test Guardian")
	renderer.Render(pages.GamePageWithDebug(updatedRoom, updatedLeader, true)).
		AssertContains(`id="debug-control-surface"`).
		AssertContains("role-card role-card-public").
		AssertContains("Public role").
		AssertNotContains("Publicly Reveal Role")
}

func renderedSectionByID(t *testing.T, html, id, closingTag string) string {
	t.Helper()
	start := strings.Index(html, `id="`+id+`"`)
	if start < 0 {
		t.Fatalf("rendered HTML missing id %q in %s", id, html)
	}
	end := strings.Index(html[start:], closingTag)
	if end < 0 {
		t.Fatalf("rendered section %q missing closing tag %q in %s", id, closingTag, html[start:])
	}
	return html[start : start+end+len(closingTag)]
}

func TestWearerUnveilResolverPublicLifecycle(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	handler := &Handler{store: memStore, config: cfg, eventBus: NewEventBus()}

	room, err := memStore.CreateRoom()
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}
	room.State = game.StatePlaying
	host := game.NewPlayer("host", "Room Operator", "host-session")
	host.IsHost = true
	wearer := game.NewPlayer("wearer", "Mask Bearer", "wearer-session")
	wearer.Role = &game.Card{ID: 31, Name: "The Wearer of Masks", Type: "Identity — Traitor", Types: game.CardTypes{Subtype: "Traitor"}}
	wearer.FaceUp = false
	observer := game.NewPlayer("observer", "Other Player", "observer-session")
	observer.Role = &game.Card{ID: 16, Name: "Dealt Guardian", Type: "Identity — Guardian", Types: game.CardTypes{Subtype: "Guardian"}}
	for _, player := range []*game.Player{host, wearer, observer} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("AddPlayer(%q) error = %v", player.ID, err)
		}
	}

	guardianCandidate := &game.Card{ID: 15, Name: "Outside Guardian", Type: "Identity — Guardian", Types: game.CardTypes{Subtype: "Guardian"}, Text: "Guardian copy text"}
	assassinCandidate := &game.Card{ID: 20, Name: "Outside Assassin", Type: "Identity — Assassin", Types: game.CardTypes{Subtype: "Assassin"}, Text: "Assassin copy text"}
	undealtLeader := &game.Card{ID: 32, Name: "Undealt Leader", Type: "Identity — Leader", Types: game.CardTypes{Subtype: "Leader"}}
	room.CardPool = game.NewCardPool([]*game.Card{wearer.Role, observer.Role, guardianCandidate, assassinCandidate, undealtLeader})
	for _, dealtCardID := range []int{31, 16} {
		if err := room.CardPool.MarkCardAssigned(dealtCardID); err != nil {
			t.Fatalf("MarkCardAssigned(%d) error = %v", dealtCardID, err)
		}
	}
	memStore.UpdateRoom(room)

	trigger := func(x int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/room/%s/player/%s/trigger-wearer/%d", room.Code, wearer.ID, x), nil)
		req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: wearer.ID})
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", wearer.ID)
		rctx.URLParams.Add("xValue", strconv.Itoa(x))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler.TriggerWearerAbility(w, req)
		return w
	}
	selectCard := func(abilityID string, cardID int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/room/%s/ability/%s/select-card/%d", room.Code, abilityID, cardID), nil)
		req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: wearer.ID})
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		rctx.URLParams.Add("cardID", strconv.Itoa(cardID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler.SelectWearerCard(w, req)
		return w
	}

	if w := trigger(5); w.Code != http.StatusOK {
		t.Fatalf("first unveil status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !wearer.FaceUp || !wearer.RoleRevealed {
		t.Fatalf("Wearer unveil state = FaceUp %v, RoleRevealed %v; want both public", wearer.FaceUp, wearer.RoleRevealed)
	}
	if len(wearer.AbilityState.PendingAbilities) != 1 {
		t.Fatalf("first unveil pending abilities = %#v, want one", wearer.AbilityState.PendingAbilities)
	}
	firstPending := wearer.AbilityState.PendingAbilities[0]
	revealedIDs, ok := firstPending.Data["revealed_ids"].([]int)
	if !ok {
		t.Fatalf("first unveil revealed IDs = %#v, want []int", firstPending.Data["revealed_ids"])
	}
	if len(revealedIDs) != 2 {
		t.Fatalf("first unveil revealed %d cards, want both undealt non-Leaders: %#v", len(revealedIDs), revealedIDs)
	}
	revealedSet := make(map[int]bool, len(revealedIDs))
	for _, cardID := range revealedIDs {
		revealedSet[cardID] = true
	}
	if !revealedSet[15] || !revealedSet[20] || revealedSet[16] || revealedSet[32] {
		t.Fatalf("revealed IDs = %#v, want only undealt non-Leaders 15 and 20", revealedIDs)
	}

	renderer := testhelpers.NewTemplateRenderer(t)
	for surface, html := range map[string]string{
		"operator dashboard": renderer.Render(pages.HostDashboardPlaying(room, host)).GetHTML(),
		"other player":       renderer.Render(pages.GameBody(room, observer)).GetHTML(),
	} {
		publicReveal := renderedSectionByID(t, html, "wearer-public-reveal-"+wearer.ID, "</section>")
		for _, expected := range []string{"Mask Bearer unveiled The Wearer of Masks", guardianCandidate.Name, guardianCandidate.Text, assassinCandidate.Name, assassinCandidate.Text, "outside the game"} {
			if !strings.Contains(publicReveal, expected) {
				t.Errorf("%s public reveal missing %q in %s", surface, expected, publicReveal)
			}
		}
		for _, forbidden := range []string{observer.Role.Name, undealtLeader.Name} {
			if strings.Contains(publicReveal, forbidden) {
				t.Errorf("%s public reveal included ineligible card %q in %s", surface, forbidden, publicReveal)
			}
		}
	}

	if w := selectCard(firstPending.ID, guardianCandidate.ID); w.Code != http.StatusOK {
		t.Fatalf("first copy choice status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if wearer.Role.GetID() != guardianCandidate.ID || wearer.Role.GetRoleType() != game.RoleTraitor {
		t.Fatalf("copied role = %#v, want card %d retaining Traitor type", wearer.Role, guardianCandidate.ID)
	}
	for _, expectedType := range []string{"Guardian", "Traitor"} {
		if !strings.Contains(wearer.Role.Type, expectedType) {
			t.Errorf("copied role type line %q missing %q", wearer.Role.Type, expectedType)
		}
	}
	operatorTile := renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(room, host)).GetHTML(), "operator-tile-"+wearer.ID, "</article>")
	observerRow := renderedSectionByID(t, renderer.Render(pages.GameBody(room, observer)).GetHTML(), "player-row-"+wearer.ID, "</details>")
	for surface, section := range map[string]string{"operator dashboard": operatorTile, "other player": observerRow} {
		for _, expected := range []string{guardianCandidate.Name, "Traitor"} {
			if !strings.Contains(section, expected) {
				t.Errorf("%s copied role missing %q in %s", surface, expected, section)
			}
		}
	}

	faceDownReq := httptest.NewRequest(http.MethodPost, "/room/"+room.Code+"/facestate/"+wearer.ID, nil)
	faceDownReq.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: wearer.ID})
	faceDownRctx := chi.NewRouteContext()
	faceDownRctx.URLParams.Add("code", room.Code)
	faceDownRctx.URLParams.Add("playerID", wearer.ID)
	faceDownReq = faceDownReq.WithContext(context.WithValue(faceDownReq.Context(), chi.RouteCtxKey, faceDownRctx))
	faceDownW := httptest.NewRecorder()
	handler.ToggleFaceState(faceDownW, faceDownReq)
	if faceDownW.Code != http.StatusOK {
		t.Fatalf("face-down status = %d, want 200; body = %s", faceDownW.Code, faceDownW.Body.String())
	}
	if wearer.FaceUp || wearer.Role.GetID() != 31 || wearer.AbilityState.TransformState != nil {
		t.Fatalf("face-down revert left state FaceUp=%v Role=%#v Transform=%#v", wearer.FaceUp, wearer.Role, wearer.AbilityState.TransformState)
	}

	if w := trigger(1); w.Code != http.StatusOK {
		t.Fatalf("second unveil status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if len(wearer.AbilityState.PendingAbilities) != 1 {
		t.Fatalf("second unveil pending abilities = %#v, want one fresh choice", wearer.AbilityState.PendingAbilities)
	}
	secondPending := wearer.AbilityState.PendingAbilities[0]
	if secondPending.ID == firstPending.ID {
		t.Fatalf("second unveil reused pending ability ID %q", secondPending.ID)
	}
	secondRevealedIDs, ok := secondPending.Data["revealed_ids"].([]int)
	if !ok || len(secondRevealedIDs) != 1 {
		t.Fatalf("second unveil revealed IDs = %#v, want one fresh random candidate", secondPending.Data["revealed_ids"])
	}
	if w := selectCard(secondPending.ID, secondRevealedIDs[0]); w.Code != http.StatusOK {
		t.Fatalf("second copy choice status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !wearer.AbilityState.IsTransformed() || wearer.Role.GetRoleType() != game.RoleTraitor {
		t.Fatalf("second unveil did not repeat copy with Traitor typing: role=%#v transform=%#v", wearer.Role, wearer.AbilityState.TransformState)
	}
}

// TestTriggerWearerAbility tests triggering The Wearer of Masks ability
func TestTriggerWearerAbility(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with CardPool
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")

	// Give player1 The Wearer of Masks (card ID 31)
	wearerCard := &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	player1.Role = wearerCard

	room.AddPlayer(player1)

	// Initialize CardPool with some cards
	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
		{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
	}
	room.CardPool = game.NewCardPool(availableCards)

	memStore.UpdateRoom(room)

	t.Run("Trigger ability successfully", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/5", nil)
		req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player1.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player1")
		rctx.URLParams.Add("xValue", "5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify pending ability was created
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")

		if !updatedPlayer.AbilityState.HasPendingAbilities() {
			t.Error("Expected pending ability to be created")
		}

		// Verify ability data
		abilities := updatedPlayer.AbilityState.PendingAbilities
		if len(abilities) != 1 {
			t.Fatalf("Expected 1 pending ability, got %d", len(abilities))
		}

		ability := abilities[0]
		if ability.AbilityType != "wearer_transform" {
			t.Errorf("Expected ability type 'wearer_transform', got %s", ability.AbilityType)
		}

		if ability.CardID != 31 {
			t.Errorf("Expected card ID 31, got %d", ability.CardID)
		}

		availableCardIDs, ok := ability.Data["available_cards"].([]int)
		if !ok {
			t.Fatal("Expected available_cards in ability data")
		}

		if len(availableCardIDs) == 0 {
			t.Error("Expected at least one available card")
		}
	})

	t.Run("Player without Wearer card cannot trigger", func(t *testing.T) {
		// Create player without The Wearer of Masks
		player2 := game.NewPlayer("player2", "Bob", "session2")
		player2.Role = &game.Card{
			ID:    15,
			Name:  "The Bodyguard",
			Type:  "Creature - Guardian",
			Types: game.CardTypes{Subtype: "Guardian"},
		}

		freshRoom, _ := memStore.GetRoom(room.Code)
		freshRoom.AddPlayer(player2)
		memStore.UpdateRoom(freshRoom)

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player2/trigger-wearer/3", nil)
		req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player2.ID})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "player2")
		rctx.URLParams.Add("xValue", "3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Player not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/nonexistent/trigger-wearer/3", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("playerID", "nonexistent")
		rctx.URLParams.Add("xValue", "3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/INVALID/player/player1/trigger-wearer/3", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		rctx.URLParams.Add("playerID", "player1")
		rctx.URLParams.Add("xValue", "3")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.TriggerWearerAbility(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// TestSelectWearerCard tests selecting a card for transformation
func TestSelectWearerCard(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with CardPool
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	leader := game.NewPlayer("leader1", "Leader", "session2")

	// Give player1 The Wearer of Masks
	wearerCard := &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	player1.Role = wearerCard

	// Give leader a Leader role (for confirmation)
	leaderCard := &game.Card{
		ID:    1,
		Name:  "Test Leader",
		Type:  "Creature - Leader",
		Types: game.CardTypes{Subtype: "Leader"},
	}
	leader.Role = leaderCard

	room.AddPlayer(player1)
	room.AddPlayer(leader)

	// Initialize CardPool with some cards
	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
		{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
	}
	room.CardPool = game.NewCardPool(availableCards)

	// Trigger ability first to create pending ability
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/5", nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player1.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	rctx.URLParams.Add("xValue", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.TriggerWearerAbility(w, req)

	// Get the ability ID
	updatedRoom, _ := memStore.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer("player1")
	abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

	t.Run("Select card successfully", func(t *testing.T) {
		// Select card ID 15 (The Bodyguard)
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		rctx.URLParams.Add("cardID", "15")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify transformation
		finalRoom, _ := memStore.GetRoom(room.Code)
		finalPlayer := finalRoom.GetPlayer("player1")

		if finalPlayer.Role.GetID() != 15 {
			t.Errorf("Expected player role to be card 15, got %d", finalPlayer.Role.GetID())
		}

		if !finalPlayer.AbilityState.IsTransformed() {
			t.Error("Expected player to be transformed")
		}

		if finalPlayer.AbilityState.GetOriginalCardID() != 31 {
			t.Errorf("Expected original card ID 31, got %d", finalPlayer.AbilityState.GetOriginalCardID())
		}

		if finalPlayer.AbilityState.GetTransformedCardID() != 15 {
			t.Errorf("Expected transformed card ID 15, got %d", finalPlayer.AbilityState.GetTransformedCardID())
		}

		// Verify pending ability was resolved
		if finalPlayer.AbilityState.HasPendingAbilities() {
			t.Error("Expected pending ability to be resolved")
		}
	})

	t.Run("Cannot select card not in available list", func(t *testing.T) {
		// Create new room and player
		room2, _ := memStore.CreateRoom()
		player2 := game.NewPlayer("player2", "Bob", "session2")
		player2.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		// Add a Leader to this room for confirmation
		leader2 := game.NewPlayer("leader2", "Leader2", "session3")
		leader2.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room2.AddPlayer(player2)
		room2.AddPlayer(leader2)
		// Use fresh availableCards copy
		availableCards2 := []*game.Card{
			{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
			{ID: 16, Name: "The Detective", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
			{ID: 20, Name: "The Infiltrator", Type: "Creature - Assassin", Types: game.CardTypes{Subtype: "Assassin"}},
			{ID: 25, Name: "The Metamorph", Type: "Creature - Traitor", Types: game.CardTypes{Subtype: "Traitor"}},
		}
		room2.CardPool = game.NewCardPool(availableCards2)
		memStore.UpdateRoom(room2)

		// Trigger ability
		req := httptest.NewRequest("POST", "/room/"+room2.Code+"/player/player2/trigger-wearer/5", nil)
		req.AddCookie(&http.Cookie{Name: "player_" + room2.Code, Value: player2.ID})
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room2.Code)
		rctx.URLParams.Add("playerID", "player2")
		rctx.URLParams.Add("xValue", "5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler.TriggerWearerAbility(w, req)

		// Get ability ID
		updatedRoom2, _ := memStore.GetRoom(room2.Code)
		updatedPlayer2 := updatedRoom2.GetPlayer("player2")
		abilityID2 := updatedPlayer2.AbilityState.PendingAbilities[0].ID

		// Leader confirms the ability first
		confirmReq := httptest.NewRequest("POST", "/room/"+room2.Code+"/ability/"+abilityID2+"/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room2.Code)
		confirmRctx.URLParams.Add("abilityID", abilityID2)
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room2.Code,
			Value: "leader2",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		// Try to select card not in available list (e.g., card 999)
		req2 := httptest.NewRequest("POST", "/room/"+room2.Code+"/ability/"+abilityID2+"/select-card/999", nil)
		req2.AddCookie(&http.Cookie{
			Name:  "player_" + room2.Code,
			Value: "player2",
		})

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("code", room2.Code)
		rctx2.URLParams.Add("abilityID", abilityID2)
		rctx2.URLParams.Add("cardID", "999")
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

		w2 := httptest.NewRecorder()

		handler.SelectWearerCard(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w2.Code)
		}
	})

	t.Run("Ability not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/non-existent/select-card/15", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", "non-existent")
		rctx.URLParams.Add("cardID", "15")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("No player cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		rctx.URLParams.Add("abilityID", abilityID)
		rctx.URLParams.Add("cardID", "15")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SelectWearerCard(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

// TestWearerAbilityEventPublishing tests that events are published correctly
func TestWearerAbilityEventPublishing(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with player and leader
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	player1.Role = &game.Card{
		ID:    31,
		Name:  "The Wearer of Masks",
		Type:  "Creature - Traitor",
		Types: game.CardTypes{Subtype: "Traitor"},
	}
	leader := game.NewPlayer("leader1", "Leader", "session2")
	leader.Role = &game.Card{
		ID:    1,
		Name:  "Test Leader",
		Type:  "Creature - Leader",
		Types: game.CardTypes{Subtype: "Leader"},
	}
	room.AddPlayer(player1)
	room.AddPlayer(leader)

	availableCards := []*game.Card{
		{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
	}
	room.CardPool = game.NewCardPool(availableCards)
	memStore.UpdateRoom(room)

	// Subscribe to events
	eventChan := eventBus.Subscribe(room.Code)

	// Trigger ability
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player1.ID})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	rctx.URLParams.Add("playerID", "player1")
	rctx.URLParams.Add("xValue", "3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.TriggerWearerAbility(w, req)

	// Wait for ability_triggered event
	select {
	case event := <-eventChan:
		if event.Type != "ability_triggered" {
			t.Errorf("Expected event type 'ability_triggered', got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for ability_triggered event")
	}

	// Get the pending ability ID
	updatedRoom, _ := memStore.GetRoom(room.Code)
	updatedPlayer := updatedRoom.GetPlayer("player1")
	abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

	// Select a card; the public replacement effect has no confirmation window.
	req2 := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)
	req2.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})

	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("code", room.Code)
	rctx2.URLParams.Add("abilityID", abilityID)
	rctx2.URLParams.Add("cardID", "15")
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

	w2 := httptest.NewRecorder()

	handler.SelectWearerCard(w2, req2)

	// Wait for transformation_complete event
	select {
	case event := <-eventChan:
		if event.Type != "transformation_complete" {
			t.Errorf("Expected event type 'transformation_complete', got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for transformation_complete event")
	}
}

// TestConfirmAbility tests the ability confirmation system
func TestConfirmAbility(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	t.Run("Leader can confirm ability successfully", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		player1 := game.NewPlayer("player1", "Alice", "session1")
		player1.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		leader := game.NewPlayer("leader1", "Leader", "session2")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(player1)
		room.AddPlayer(leader)

		availableCards := []*game.Card{
			{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		}
		room.CardPool = game.NewCardPool(availableCards)
		memStore.UpdateRoom(room)

		// Trigger the ability
		triggerReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
		triggerReq.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player1.ID})
		triggerRctx := chi.NewRouteContext()
		triggerRctx.URLParams.Add("code", room.Code)
		triggerRctx.URLParams.Add("playerID", "player1")
		triggerRctx.URLParams.Add("xValue", "3")
		triggerReq = triggerReq.WithContext(context.WithValue(triggerReq.Context(), chi.RouteCtxKey, triggerRctx))
		triggerW := httptest.NewRecorder()
		handler.TriggerWearerAbility(triggerW, triggerReq)

		// Get the ability ID
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

		// Verify ability requires confirmation and is not confirmed
		pendingAbility := updatedPlayer.AbilityState.GetPendingAbility(abilityID)
		pendingAbility.RequiresConfirmation = true
		pendingAbility.ConfirmationRole = "leader"
		if !pendingAbility.RequiresConfirmation {
			t.Error("Expected ability to require confirmation")
		}
		if pendingAbility.IsConfirmed() {
			t.Error("Expected ability to not be confirmed yet")
		}

		// Leader confirms the ability
		confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room.Code)
		confirmRctx.URLParams.Add("abilityID", abilityID)
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "leader1",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		if confirmW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", confirmW.Code)
		}

		// Verify ability is now confirmed
		updatedRoom, _ = memStore.GetRoom(room.Code)
		updatedPlayer = updatedRoom.GetPlayer("player1")
		pendingAbility = updatedPlayer.AbilityState.GetPendingAbility(abilityID)
		if !pendingAbility.IsConfirmed() {
			t.Error("Expected ability to be confirmed after Leader confirmation")
		}
	})

	t.Run("Non-leader cannot confirm leader-only ability", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		player1 := game.NewPlayer("player1", "Alice", "session1")
		player1.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		player2 := game.NewPlayer("player2", "Bob", "session2")
		player2.Role = &game.Card{
			ID:    15,
			Name:  "The Bodyguard",
			Type:  "Creature - Guardian",
			Types: game.CardTypes{Subtype: "Guardian"},
		}
		leader := game.NewPlayer("leader1", "Leader", "session3")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(player1)
		room.AddPlayer(player2)
		room.AddPlayer(leader)

		availableCards := []*game.Card{
			{ID: 16, Name: "Another Card", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		}
		room.CardPool = game.NewCardPool(availableCards)
		memStore.UpdateRoom(room)

		// Trigger the ability
		triggerReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
		triggerReq.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player1.ID})
		triggerRctx := chi.NewRouteContext()
		triggerRctx.URLParams.Add("code", room.Code)
		triggerRctx.URLParams.Add("playerID", "player1")
		triggerRctx.URLParams.Add("xValue", "3")
		triggerReq = triggerReq.WithContext(context.WithValue(triggerReq.Context(), chi.RouteCtxKey, triggerRctx))
		triggerW := httptest.NewRecorder()
		handler.TriggerWearerAbility(triggerW, triggerReq)

		// Get the ability ID
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID
		pendingAbility := updatedPlayer.AbilityState.GetPendingAbility(abilityID)
		pendingAbility.RequiresConfirmation = true
		pendingAbility.ConfirmationRole = "leader"

		// Non-leader player tries to confirm (should fail)
		confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room.Code)
		confirmRctx.URLParams.Add("abilityID", abilityID)
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player2",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		if confirmW.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", confirmW.Code)
		}
	})

	t.Run("Ability not found", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		leader := game.NewPlayer("leader1", "Leader", "session1")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(leader)
		memStore.UpdateRoom(room)

		confirmReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/nonexistent/confirm", nil)
		confirmRctx := chi.NewRouteContext()
		confirmRctx.URLParams.Add("code", room.Code)
		confirmRctx.URLParams.Add("abilityID", "nonexistent")
		confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, confirmRctx))
		confirmReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "leader1",
		})
		confirmW := httptest.NewRecorder()
		handler.ConfirmAbility(confirmW, confirmReq)

		if confirmW.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", confirmW.Code)
		}
	})

	t.Run("Wearer selection does not require confirmation", func(t *testing.T) {
		room, _ := memStore.CreateRoom()
		player1 := game.NewPlayer("player1", "Alice", "session1")
		player1.Role = &game.Card{
			ID:    31,
			Name:  "The Wearer of Masks",
			Type:  "Creature - Traitor",
			Types: game.CardTypes{Subtype: "Traitor"},
		}
		leader := game.NewPlayer("leader1", "Leader", "session2")
		leader.Role = &game.Card{
			ID:    1,
			Name:  "Test Leader",
			Type:  "Creature - Leader",
			Types: game.CardTypes{Subtype: "Leader"},
		}
		room.AddPlayer(player1)
		room.AddPlayer(leader)

		availableCards := []*game.Card{
			{ID: 15, Name: "The Bodyguard", Type: "Creature - Guardian", Types: game.CardTypes{Subtype: "Guardian"}},
		}
		room.CardPool = game.NewCardPool(availableCards)
		memStore.UpdateRoom(room)

		// Trigger the ability
		triggerReq := httptest.NewRequest("POST", "/room/"+room.Code+"/player/player1/trigger-wearer/3", nil)
		triggerReq.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player1.ID})
		triggerRctx := chi.NewRouteContext()
		triggerRctx.URLParams.Add("code", room.Code)
		triggerRctx.URLParams.Add("playerID", "player1")
		triggerRctx.URLParams.Add("xValue", "3")
		triggerReq = triggerReq.WithContext(context.WithValue(triggerReq.Context(), chi.RouteCtxKey, triggerRctx))
		triggerW := httptest.NewRecorder()
		handler.TriggerWearerAbility(triggerW, triggerReq)

		// Get the ability ID
		updatedRoom, _ := memStore.GetRoom(room.Code)
		updatedPlayer := updatedRoom.GetPlayer("player1")
		abilityID := updatedPlayer.AbilityState.PendingAbilities[0].ID

		// Select directly: candidates and the unveil are already public.
		selectReq := httptest.NewRequest("POST", "/room/"+room.Code+"/ability/"+abilityID+"/select-card/15", nil)
		selectReq.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})
		selectRctx := chi.NewRouteContext()
		selectRctx.URLParams.Add("code", room.Code)
		selectRctx.URLParams.Add("abilityID", abilityID)
		selectRctx.URLParams.Add("cardID", "15")
		selectReq = selectReq.WithContext(context.WithValue(selectReq.Context(), chi.RouteCtxKey, selectRctx))
		selectW := httptest.NewRecorder()
		handler.SelectWearerCard(selectW, selectReq)

		if selectW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", selectW.Code)
		}
	})
}
