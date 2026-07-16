package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/game/ability"
	"treacherest/internal/store"
	"treacherest/internal/testhelpers"
	"treacherest/internal/views/pages"

	"github.com/go-chi/chi/v5"
)

func TestFerrymanAndGatekeeperTableRemembersAfterFaceDown(t *testing.T) {
	roles := []*game.Card{
		{ID: 21, Name: "The Ferryman", Type: "Identity — Guardian", Text: "FERRYMAN ROLE TEXT SENTINEL", Types: game.CardTypes{Subtype: "Guardian"}},
		{ID: 22, Name: "The Gatekeeper", Type: "Identity — Guardian", Text: "GATEKEEPER ROLE TEXT SENTINEL", Types: game.CardTypes{Subtype: "Guardian"}},
	}

	for _, role := range roles {
		role := role
		t.Run(strings.TrimPrefix(role.Name, "The "), func(t *testing.T) {
			cfg := config.DefaultConfig()
			memStore := store.NewMemoryStore(cfg)
			eventBus := NewEventBus()
			handler := &Handler{store: memStore, config: cfg, eventBus: eventBus}
			room, err := memStore.CreateRoom()
			if err != nil {
				t.Fatalf("CreateRoom() error = %v", err)
			}
			room.State = game.StatePlaying

			host := game.NewPlayer("host", "Room Operator", "host-session")
			host.IsHost = true
			owner := game.NewPlayer("owner", role.Name+" Player", "owner-session")
			owner.Role = role
			owner.FaceUp = false
			owner.RoleRevealed = false
			observer := game.NewPlayer("observer", "Other Player", "observer-session")
			observer.Role = &game.Card{ID: 10, Name: "Observer Role", Types: game.CardTypes{Subtype: "Assassin"}}
			observer.FaceUp = false
			observer.RoleRevealed = false
			for _, player := range []*game.Player{host, owner, observer} {
				if err := room.AddPlayer(player); err != nil {
					t.Fatalf("AddPlayer(%q) error = %v", player.ID, err)
				}
			}
			memStore.UpdateRoom(room)

			if ability.HasUnveilRequirements(role.ID) || ability.GetUnveilRequirements(role.ID).InputType != ability.NoInput {
				t.Fatalf("%s should use the generic NoInput unveil path", role.Name)
			}

			unveilW := httptest.NewRecorder()
			handler.UnveilPlayer(unveilW, supportedIdentityRequest(room, owner, "/room/"+room.Code+"/unveil/"+owner.ID, map[string]string{
				"code": room.Code, "playerID": owner.ID,
			}))
			if unveilW.Code != http.StatusOK {
				t.Fatalf("initial %s unveil status = %d, want 200; body = %s", role.Name, unveilW.Code, unveilW.Body.String())
			}
			if !owner.FaceUp || !owner.RoleRevealed {
				t.Fatalf("initial %s unveil state = FaceUp %v, RoleRevealed %v; want both true", role.Name, owner.FaceUp, owner.RoleRevealed)
			}

			faceDownW := httptest.NewRecorder()
			handler.ToggleFaceState(faceDownW, supportedIdentityRequest(room, owner, "/room/"+room.Code+"/facestate/"+owner.ID, map[string]string{
				"code": room.Code, "playerID": owner.ID,
			}))
			if faceDownW.Code != http.StatusOK {
				t.Fatalf("%s face-down status = %d, want 200; body = %s", role.Name, faceDownW.Code, faceDownW.Body.String())
			}
			if owner.FaceUp || !owner.RoleRevealed {
				t.Fatalf("%s face-down state = FaceUp %v, RoleRevealed %v; want false, true", role.Name, owner.FaceUp, owner.RoleRevealed)
			}

			renderer := testhelpers.NewTemplateRenderer(t)
			operatorTile := renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(room, host)).GetHTML(), "operator-tile-"+owner.ID, "</article>")
			supportedIdentityAssertContains(t, "operator dashboard after "+role.Name+" turns face down", operatorTile, "Revealed: "+role.Name)
			supportedIdentityAssertNotContains(t, "operator dashboard after "+role.Name+" turns face down", operatorTile, "Face Down")

			publicRow := renderedSectionByID(t, renderer.Render(pages.GameRosterZone(room, nil)).GetHTML(), "player-row-"+owner.ID, "</details>")
			supportedIdentityAssertContains(t, "public roster after "+role.Name+" turns face down", publicRow, "Revealed: "+role.Name, role.Name, role.Text)

			// The action surface is keyed to FaceUp, not remembered public knowledge.
			faceDownActions := renderer.Render(pages.GameActionsZone(room, owner)).GetHTML()
			supportedIdentityAssertContains(t, role.Name+" face-up-keyed actions", faceDownActions, "/room/"+room.Code+"/unveil/"+owner.ID, "Unveil")

			reUnveilW := httptest.NewRecorder()
			handler.UnveilPlayer(reUnveilW, supportedIdentityRequest(room, owner, "/room/"+room.Code+"/unveil/"+owner.ID, map[string]string{
				"code": room.Code, "playerID": owner.ID,
			}))
			if reUnveilW.Code != http.StatusOK {
				t.Fatalf("%s re-unveil status = %d, want 200; body = %s", role.Name, reUnveilW.Code, reUnveilW.Body.String())
			}
			if !owner.FaceUp || !owner.RoleRevealed {
				t.Fatalf("%s re-unveil state = FaceUp %v, RoleRevealed %v; want both true", role.Name, owner.FaceUp, owner.RoleRevealed)
			}

			reUnveiledOperatorTile := renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(room, host)).GetHTML(), "operator-tile-"+owner.ID, "</article>")
			reUnveiledPublicRow := renderedSectionByID(t, renderer.Render(pages.GameRosterZone(room, nil)).GetHTML(), "player-row-"+owner.ID, "</details>")
			supportedIdentityAssertContains(t, "operator dashboard after "+role.Name+" re-unveils", reUnveiledOperatorTile, "Revealed: "+role.Name)
			supportedIdentityAssertContains(t, "public roster after "+role.Name+" re-unveils", reUnveiledPublicRow, "Revealed: "+role.Name, role.Text)
		})
	}
}

func TestMetamorphStealTransfersIdentityAndPublishesEvent(t *testing.T) {
	t.Run("stolen non-Leader is face-down and owner-only", func(t *testing.T) {
		stolenRole := &game.Card{
			ID: 701, Name: "STOLEN NON-LEADER NAME SENTINEL", Type: "Identity — Secret Sentinel",
			Text: "STOLEN NON-LEADER ORACLE SENTINEL", Types: game.CardTypes{Subtype: "Guardian"},
		}
		harness := newMetamorphStealHarness(t, stolenRole)
		events := map[string]chan Event{
			"player room subscriber":   harness.eventBus.Subscribe(harness.room.Code),
			"operator room subscriber": harness.eventBus.Subscribe(harness.room.Code),
		}

		w := httptest.NewRecorder()
		harness.handler.StealRole(w, supportedIdentityRequest(harness.room, harness.metamorph, "/room/"+harness.room.Code+"/player/"+harness.metamorph.ID+"/steal-role/"+harness.victim.ID, map[string]string{
			"code": harness.room.Code, "playerID": harness.metamorph.ID, "targetPlayerID": harness.victim.ID,
		}))
		if w.Code != http.StatusOK {
			t.Fatalf("Metamorph non-Leader steal status = %d, want 200; body = %s", w.Code, w.Body.String())
		}

		assertMetamorphStealState(t, harness, stolenRole)
		if harness.metamorph.FaceUp || harness.metamorph.RoleRevealed {
			t.Fatalf("stolen non-Leader state = FaceUp %v, RoleRevealed %v; want both false", harness.metamorph.FaceUp, harness.metamorph.RoleRevealed)
		}
		for subscriber, ch := range events {
			assertPublicMetamorphStealEvent(t, subscriber, ch, harness, stolenRole)
		}

		// TODO(PRD Ambiguous Rulings Q1): once the maintainer rules on whether
		// the table sees stolen non-Leader contents before the face-down flip,
		// add the corresponding pre-flip disclosure assertion. Do not assert
		// either direction until that ruling is made.

		renderer := testhelpers.NewTemplateRenderer(t)
		ownerPrivy := renderedSectionByID(t, renderer.Render(pages.GameContent(harness.room, harness.metamorph)).GetHTML(), "zone-privy", "</article>")
		debugOwnerPrivy := renderedSectionByID(t, renderer.Render(pages.GamePageWithDebug(harness.room, harness.metamorph, true)).GetHTML(), "zone-privy", "</article>")
		for surface, section := range map[string]string{
			"Metamorph owner player view":          ownerPrivy,
			"Metamorph owner debug view-as-player": debugOwnerPrivy,
		} {
			supportedIdentityAssertContains(t, surface, section, "Private role", stolenRole.Name, stolenRole.Type, stolenRole.Text)
		}

		observerRow := renderedSectionByID(t, renderer.Render(pages.GameContent(harness.room, harness.observer)).GetHTML(), "player-row-"+harness.metamorph.ID, "</details>")
		debugObserverRow := renderedSectionByID(t, renderer.Render(pages.GamePageWithDebug(harness.room, harness.observer, true)).GetHTML(), "player-row-"+harness.metamorph.ID, "</details>")
		operatorTile := renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(harness.room, harness.host)).GetHTML(), "operator-tile-"+harness.metamorph.ID, "</article>")
		publicRow := renderedSectionByID(t, renderer.Render(pages.GameRosterZone(harness.room, nil)).GetHTML(), "player-row-"+harness.metamorph.ID, "</details>")
		for surface, section := range map[string]string{
			"other player game view":     observerRow,
			"operator dashboard":         operatorTile,
			"debug view-as-other-player": debugObserverRow,
			"public host table state":    publicRow,
		} {
			supportedIdentityAssertContains(t, surface, section, "Face Down")
			supportedIdentityAssertNotContains(t, surface, section, stolenRole.Name, stolenRole.Type, stolenRole.Text)
		}
	})

	t.Run("stolen Leader is forced public", func(t *testing.T) {
		stolenRole := &game.Card{
			ID: 702, Name: "STOLEN LEADER NAME SENTINEL", Type: "Identity — Leader",
			Text: "STOLEN LEADER ORACLE SENTINEL", Types: game.CardTypes{Subtype: "Leader"},
		}
		harness := newMetamorphStealHarness(t, stolenRole)
		events := map[string]chan Event{
			"player room subscriber":   harness.eventBus.Subscribe(harness.room.Code),
			"operator room subscriber": harness.eventBus.Subscribe(harness.room.Code),
		}

		w := httptest.NewRecorder()
		harness.handler.StealRole(w, supportedIdentityRequest(harness.room, harness.metamorph, "/room/"+harness.room.Code+"/player/"+harness.metamorph.ID+"/steal-role/"+harness.victim.ID, map[string]string{
			"code": harness.room.Code, "playerID": harness.metamorph.ID, "targetPlayerID": harness.victim.ID,
		}))
		if w.Code != http.StatusOK {
			t.Fatalf("Metamorph Leader steal status = %d, want 200; body = %s", w.Code, w.Body.String())
		}

		assertMetamorphStealState(t, harness, stolenRole)
		if !harness.metamorph.FaceUp || !harness.metamorph.RoleRevealed {
			t.Fatalf("stolen Leader state = FaceUp %v, RoleRevealed %v; want both true", harness.metamorph.FaceUp, harness.metamorph.RoleRevealed)
		}
		for subscriber, ch := range events {
			assertPublicMetamorphStealEvent(t, subscriber, ch, harness, stolenRole)
		}

		renderer := testhelpers.NewTemplateRenderer(t)
		observerRow := renderedSectionByID(t, renderer.Render(pages.GameContent(harness.room, harness.observer)).GetHTML(), "player-row-"+harness.metamorph.ID, "</details>")
		debugObserverRow := renderedSectionByID(t, renderer.Render(pages.GamePageWithDebug(harness.room, harness.observer, true)).GetHTML(), "player-row-"+harness.metamorph.ID, "</details>")
		publicRow := renderedSectionByID(t, renderer.Render(pages.GameRosterZone(harness.room, nil)).GetHTML(), "player-row-"+harness.metamorph.ID, "</details>")
		for surface, section := range map[string]string{
			"other player game view":     observerRow,
			"debug view-as-other-player": debugObserverRow,
			"public host table state":    publicRow,
		} {
			supportedIdentityAssertContains(t, surface, section, stolenRole.Name, stolenRole.Type, stolenRole.Text)
			supportedIdentityAssertNotContains(t, surface, section, "Card is face down.")
		}

		operatorTile := renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(harness.room, harness.host)).GetHTML(), "operator-tile-"+harness.metamorph.ID, "</article>")
		supportedIdentityAssertContains(t, "operator dashboard", operatorTile, "Revealed: "+stolenRole.Name, "Type: "+stolenRole.Type)
		supportedIdentityAssertNotContains(t, "operator dashboard", operatorTile, "Face Down")

		ownerPrivy := renderedSectionByID(t, renderer.Render(pages.GameContent(harness.room, harness.metamorph)).GetHTML(), "zone-privy", "</article>")
		supportedIdentityAssertContains(t, "Metamorph Leader owner view", ownerPrivy, "Public role", stolenRole.Name, stolenRole.Text)
	})
}

func TestWarShamanUsesPureGenericUnveilLifecycle(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{store: memStore, config: cfg, eventBus: eventBus}
	room, err := memStore.CreateRoom()
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}
	room.State = game.StatePlaying

	host := game.NewPlayer("host", "Room Operator", "host-session")
	host.IsHost = true
	warShaman := game.NewPlayer("war-shaman", "War Shaman Player", "war-shaman-session")
	warShaman.Role = &game.Card{
		ID: 49, Name: "The War Shaman", Type: "Identity — Traitor",
		Text: "WAR SHAMAN ORACLE TEXT SENTINEL", Types: game.CardTypes{Subtype: "Traitor"},
	}
	warShaman.FaceUp = false
	warShaman.RoleRevealed = false
	observer := game.NewPlayer("observer", "Other Player", "observer-session")
	observer.Role = &game.Card{ID: 10, Name: "Observer Role", Types: game.CardTypes{Subtype: "Guardian"}}
	observer.FaceUp = false
	observer.RoleRevealed = false
	for _, player := range []*game.Player{host, warShaman, observer} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("AddPlayer(%q) error = %v", player.ID, err)
		}
	}
	memStore.UpdateRoom(room)

	requirements := ability.GetUnveilRequirements(warShaman.Role.ID)
	if ability.HasUnveilRequirements(warShaman.Role.ID) || requirements.InputType != ability.NoInput || !requirements.SetsFaceUp {
		t.Fatalf("The War Shaman requirements = %#v, want generic NoInput face-up unveil", requirements)
	}

	renderer := testhelpers.NewTemplateRenderer(t)
	ownerPrivy := renderedSectionByID(t, renderer.Render(pages.GameContent(room, warShaman)).GetHTML(), "zone-privy", "</article>")
	debugOwnerPrivy := renderedSectionByID(t, renderer.Render(pages.GamePageWithDebug(room, warShaman, true)).GetHTML(), "zone-privy", "</article>")
	for surface, section := range map[string]string{
		"War Shaman owner player view":          ownerPrivy,
		"War Shaman owner debug view-as-player": debugOwnerPrivy,
	} {
		supportedIdentityAssertContains(t, surface, section, "Private role", warShaman.Role.Name, warShaman.Role.Type, warShaman.Role.Text)
	}

	observerRow := renderedSectionByID(t, renderer.Render(pages.GameContent(room, observer)).GetHTML(), "player-row-"+warShaman.ID, "</details>")
	debugObserverRow := renderedSectionByID(t, renderer.Render(pages.GamePageWithDebug(room, observer, true)).GetHTML(), "player-row-"+warShaman.ID, "</details>")
	operatorTile := renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(room, host)).GetHTML(), "operator-tile-"+warShaman.ID, "</article>")
	publicRow := renderedSectionByID(t, renderer.Render(pages.GameRosterZone(room, nil)).GetHTML(), "player-row-"+warShaman.ID, "</details>")
	for surface, section := range map[string]string{
		"other player game view":     observerRow,
		"operator dashboard":         operatorTile,
		"debug view-as-other-player": debugObserverRow,
		"public host table state":    publicRow,
	} {
		supportedIdentityAssertContains(t, surface+" before War Shaman unveils", section, "Face Down")
		supportedIdentityAssertNotContains(t, surface+" before War Shaman unveils", section, warShaman.Role.Name, warShaman.Role.Type, warShaman.Role.Text)
	}

	beforeWarShaman := *warShaman
	beforeObserver := *observer
	beforeRole := *warShaman.Role
	events := map[string]chan Event{
		"player room subscriber":   eventBus.Subscribe(room.Code),
		"operator room subscriber": eventBus.Subscribe(room.Code),
	}

	w := httptest.NewRecorder()
	handler.UnveilPlayer(w, supportedIdentityRequest(room, warShaman, "/room/"+room.Code+"/unveil/"+warShaman.ID, map[string]string{
		"code": room.Code, "playerID": warShaman.ID,
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("The War Shaman unveil status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	wantWarShaman := beforeWarShaman
	wantWarShaman.FaceUp = true
	wantWarShaman.RoleRevealed = true
	if !reflect.DeepEqual(warShaman, &wantWarShaman) {
		t.Fatalf("The War Shaman unveil mutated player state beyond FaceUp and RoleRevealed\ngot:  %#v\nwant: %#v", warShaman, &wantWarShaman)
	}
	if warShaman.Role != beforeWarShaman.Role || !reflect.DeepEqual(warShaman.Role, &beforeRole) {
		t.Fatalf("The War Shaman unveil replaced or mutated the role card: got %#v, want original %#v", warShaman.Role, &beforeRole)
	}
	if !reflect.DeepEqual(warShaman.AbilityState, ability.NewAbilityState()) {
		t.Fatalf("The War Shaman generic unveil unexpectedly mutated ability state: %#v", warShaman.AbilityState)
	}
	if !reflect.DeepEqual(observer, &beforeObserver) || len(room.Players) != 3 || room.State != game.StatePlaying {
		t.Fatalf("The War Shaman generic unveil unexpectedly mutated unrelated room state")
	}
	for subscriber, ch := range events {
		event := supportedIdentityReceiveEvent(t, subscriber, ch)
		if event.Type != "role_revealed" || event.RoomCode != room.Code || event.Data != room {
			t.Fatalf("%s got unveil event %#v, want room-wide role_revealed for %s", subscriber, event, room.Code)
		}
	}

	observerRow = renderedSectionByID(t, renderer.Render(pages.GameContent(room, observer)).GetHTML(), "player-row-"+warShaman.ID, "</details>")
	debugObserverRow = renderedSectionByID(t, renderer.Render(pages.GamePageWithDebug(room, observer, true)).GetHTML(), "player-row-"+warShaman.ID, "</details>")
	publicRow = renderedSectionByID(t, renderer.Render(pages.GameRosterZone(room, nil)).GetHTML(), "player-row-"+warShaman.ID, "</details>")
	for surface, section := range map[string]string{
		"other player game view":     observerRow,
		"debug view-as-other-player": debugObserverRow,
		"public host table state":    publicRow,
	} {
		supportedIdentityAssertContains(t, surface+" after War Shaman unveils", section, warShaman.Role.Name, warShaman.Role.Type, warShaman.Role.Text)
		supportedIdentityAssertNotContains(t, surface+" after War Shaman unveils", section, "Card is face down.")
	}
	operatorTile = renderedSectionByID(t, renderer.Render(pages.HostDashboardPlaying(room, host)).GetHTML(), "operator-tile-"+warShaman.ID, "</article>")
	supportedIdentityAssertContains(t, "operator dashboard after War Shaman unveils", operatorTile, "Revealed: "+warShaman.Role.Name, "Type: "+warShaman.Role.Type)
	supportedIdentityAssertNotContains(t, "operator dashboard after War Shaman unveils", operatorTile, "Face Down")
}

func TestOperatorDashboardOutsideDebugNeverRendersUnrevealedRole(t *testing.T) {
	host := game.NewPlayer("host", "Room Operator", "host-session")
	host.IsHost = true
	hiddenFaceDown := game.NewPlayer("hidden-down", "Hidden Face-Down Player", "hidden-down-session")
	hiddenFaceDown.Role = &game.Card{
		ID: 801, Name: "HIDDEN FACE-DOWN NAME SENTINEL", Type: "Identity — Hidden Down Sentinel",
		Text: "HIDDEN FACE-DOWN ORACLE SENTINEL", Types: game.CardTypes{Subtype: "Guardian"},
	}
	hiddenFaceDown.FaceUp = false
	hiddenFaceDown.RoleRevealed = false
	hiddenFaceUp := game.NewPlayer("hidden-up", "Hidden Face-Up Player", "hidden-up-session")
	hiddenFaceUp.Role = &game.Card{
		ID: 802, Name: "HIDDEN FACE-UP NAME SENTINEL", Type: "Identity — Hidden Up Sentinel",
		Text: "HIDDEN FACE-UP ORACLE SENTINEL", Types: game.CardTypes{Subtype: "Assassin"},
	}
	hiddenFaceUp.FaceUp = true
	hiddenFaceUp.RoleRevealed = false
	rememberedFaceDown := game.NewPlayer("remembered-down", "Remembered Face-Down Player", "remembered-session")
	rememberedFaceDown.Role = &game.Card{
		ID: 803, Name: "REMEMBERED PUBLIC NAME SENTINEL", Type: "Identity — Remembered Sentinel",
		Text: "REMEMBERED PUBLIC ORACLE SENTINEL", Types: game.CardTypes{Subtype: "Traitor"},
	}
	rememberedFaceDown.FaceUp = false
	rememberedFaceDown.RoleRevealed = true
	room := &game.Room{
		Code: "OPSEC", State: game.StatePlaying,
		Players: map[string]*game.Player{
			host.ID: host, hiddenFaceDown.ID: hiddenFaceDown, hiddenFaceUp.ID: hiddenFaceUp, rememberedFaceDown.ID: rememberedFaceDown,
		},
	}

	html := testhelpers.NewTemplateRenderer(t).Render(pages.HostDashboardPlaying(room, host)).GetHTML()
	for _, hidden := range []*game.Player{hiddenFaceDown, hiddenFaceUp} {
		tile := renderedSectionByID(t, html, "operator-tile-"+hidden.ID, "</article>")
		supportedIdentityAssertContains(t, "non-debug operator tile for "+hidden.Name, tile, hidden.Name, "Face Down")
		supportedIdentityAssertNotContains(t, "non-debug operator tile for "+hidden.Name, tile, hidden.Role.Name, hidden.Role.Type, hidden.Role.Text)
	}

	rememberedTile := renderedSectionByID(t, html, "operator-tile-"+rememberedFaceDown.ID, "</article>")
	supportedIdentityAssertContains(t, "non-debug operator tile for remembered role", rememberedTile, rememberedFaceDown.Name, "Revealed: "+rememberedFaceDown.Role.Name, "Type: "+rememberedFaceDown.Role.Type)
	supportedIdentityAssertNotContains(t, "non-debug operator tile for remembered role", rememberedTile, "Face Down")
}

type metamorphStealHarness struct {
	handler   *Handler
	eventBus  *EventBus
	room      *game.Room
	host      *game.Player
	metamorph *game.Player
	victim    *game.Player
	observer  *game.Player
}

func newMetamorphStealHarness(t *testing.T, stolenRole *game.Card) *metamorphStealHarness {
	t.Helper()
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{store: memStore, config: cfg, eventBus: eventBus}
	room, err := memStore.CreateRoom()
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}
	room.State = game.StatePlaying

	host := game.NewPlayer("host", "Room Operator", "host-session")
	host.IsHost = true
	metamorphCard := &game.Card{
		ID: 25, Name: "The Metamorph", Type: "Identity — Traitor",
		Text: "METAMORPH ORIGINAL ROLE TEXT", Types: game.CardTypes{Subtype: "Traitor"},
	}
	metamorph := game.NewPlayer("metamorph", "Metamorph Player", "metamorph-session")
	metamorph.Role = metamorphCard
	metamorph.FaceUp = true
	metamorph.RoleRevealed = true
	metamorph.AbilityState.ActivateMetamorph()
	victim := game.NewPlayer("victim", "Eliminated Victim", "victim-session")
	victim.Role = stolenRole
	victim.FaceUp = false
	victim.RoleRevealed = false
	victim.IsEliminated = true
	observer := game.NewPlayer("observer", "Other Player", "observer-session")
	observer.Role = &game.Card{ID: 10, Name: "Observer Role", Types: game.CardTypes{Subtype: "Assassin"}}
	observer.FaceUp = false
	observer.RoleRevealed = false
	for _, player := range []*game.Player{host, metamorph, victim, observer} {
		if err := room.AddPlayer(player); err != nil {
			t.Fatalf("AddPlayer(%q) error = %v", player.ID, err)
		}
	}
	room.CardPool = game.NewCardPool([]*game.Card{metamorphCard, stolenRole, observer.Role})
	memStore.UpdateRoom(room)

	return &metamorphStealHarness{
		handler: handler, eventBus: eventBus, room: room, host: host,
		metamorph: metamorph, victim: victim, observer: observer,
	}
}

func assertMetamorphStealState(t *testing.T, harness *metamorphStealHarness, stolenRole *game.Card) {
	t.Helper()
	if harness.victim.Role != nil {
		t.Fatalf("Metamorph victim role = %#v, want nil after steal", harness.victim.Role)
	}
	if harness.victim.FaceUp || harness.victim.RoleRevealed {
		t.Fatalf("Metamorph victim face state = FaceUp %v, RoleRevealed %v; want both false after role removal", harness.victim.FaceUp, harness.victim.RoleRevealed)
	}
	if harness.metamorph.Role != stolenRole {
		t.Fatalf("Metamorph role = %#v, want stolen card pointer %#v", harness.metamorph.Role, stolenRole)
	}
	for _, player := range harness.room.GetPlayers() {
		if player.Role != nil && player.Role.ID == 25 {
			t.Fatalf("The Metamorph remained controlled by %s after the steal", player.Name)
		}
	}

	transform := harness.metamorph.AbilityState.TransformState
	if transform == nil {
		t.Fatal("Metamorph steal did not record StealIdentity TransformState")
	}
	if !transform.IsTransformed || transform.OriginalCardID != 25 || transform.TransformedCardID != stolenRole.ID ||
		transform.ChangeType != ability.RoleChangeSteal || !transform.IsPermanent || !transform.SourceCardRemoved || transform.EndCondition != "never" {
		t.Fatalf("Metamorph StealIdentity state = %#v, want permanent steal from card 25 to card %d", transform, stolenRole.ID)
	}
}

func assertPublicMetamorphStealEvent(t *testing.T, subscriber string, ch chan Event, harness *metamorphStealHarness, stolenRole *game.Card) {
	t.Helper()
	event := supportedIdentityReceiveEvent(t, subscriber, ch)
	if event.Type != "role_stolen" || event.RoomCode != harness.room.Code {
		t.Fatalf("%s got event %#v, want room-wide role_stolen for %s", subscriber, event, harness.room.Code)
	}
	payload, ok := event.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("%s role_stolen payload type = %T, want map", subscriber, event.Data)
	}
	if payload["room"] != harness.room || payload["stealer"] != harness.metamorph || payload["victim"] != harness.victim {
		t.Fatalf("%s role_stolen payload = %#v, want room, Metamorph, and victim", subscriber, payload)
	}
	if harness.metamorph.Role != stolenRole || harness.victim.Role != nil || harness.metamorph.AbilityState.TransformState.OriginalCardID != 25 {
		t.Fatalf("%s role_stolen event did not expose the public transfer outcome", subscriber)
	}
}

func supportedIdentityRequest(room *game.Room, player *game.Player, path string, params map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.AddCookie(&http.Cookie{Name: "player_" + room.Code, Value: player.ID})
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func supportedIdentityReceiveEvent(t *testing.T, subscriber string, ch chan Event) Event {
	t.Helper()
	select {
	case event := <-ch:
		return event
	default:
		t.Fatalf("%s did not receive the room-wide event", subscriber)
		return Event{}
	}
}

func supportedIdentityAssertContains(t *testing.T, surface, html string, expected ...string) {
	t.Helper()
	for _, value := range expected {
		if !strings.Contains(html, value) {
			t.Fatalf("%s missing %q in anchored rendering: %s", surface, value, html)
		}
	}
}

func supportedIdentityAssertNotContains(t *testing.T, surface, html string, forbidden ...string) {
	t.Helper()
	for _, value := range forbidden {
		if strings.Contains(html, value) {
			t.Fatalf("%s leaked or unexpectedly rendered %q in anchored rendering: %s", surface, value, html)
		}
	}
}
