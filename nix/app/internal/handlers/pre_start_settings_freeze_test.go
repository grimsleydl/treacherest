package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"treacherest/internal/game"
)

func TestPreStartSettingsFreezeRejectsTreacheryPlayerCountAfterLobby(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeTreachery
	room.State = game.StateCountdown
	roleConfig, err := h.roleConfigService.CreateFromPreset("standard", 5)
	if err != nil {
		t.Fatalf("failed to create role config: %v", err)
	}
	room.RoleConfig = roleConfig
	operator := game.NewPlayer("operator", "Playing Operator", "session-operator")
	room.AddPlayer(operator)
	markRoomOperatorForTest(room, operator)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/player-count/increment", nil)
	addPlayerSessionCookiesForTest(req, room, operator)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.RoleConfig.MaxPlayers != 5 {
		t.Fatalf("post-start player-count update mutated room: got %d players", updatedRoom.RoleConfig.MaxPlayers)
	}
	if !containsPreStartSettingsLockedMessage(w.Body.String()) {
		t.Fatalf("expected locked settings message, got: %s", w.Body.String())
	}
}

func containsPreStartSettingsLockedMessage(body string) bool {
	return strings.Contains(body, preStartSettingsLockedMessage)
}

func TestPreStartSettingsFreezeRejectsCoupPresetAfterLobby(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = game.StateCountdown
	room.CoupPreset = game.CoupPresetFive
	counts, ok := game.CoupRoleCountsForPreset(game.CoupPresetFive)
	if !ok {
		t.Fatal("missing five-player Coup preset")
	}
	room.CoupRoleCounts = counts
	operator := game.NewPlayer("operator", "Playing Operator", "session-operator")
	room.AddPlayer(operator)
	markRoomOperatorForTest(room, operator)
	h.store.UpdateRoom(room)

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	form := url.Values{}
	form.Add("preset", string(game.CoupPresetSix))
	req := httptest.NewRequest("POST", "/room/"+room.Code+"/config/coup-preset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addPlayerSessionCookiesForTest(req, room, operator)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupPreset != game.CoupPresetFive {
		t.Fatalf("post-start Coup preset update mutated room: got %q", updatedRoom.CoupPreset)
	}
	if !containsPreStartSettingsLockedMessage(w.Body.String()) {
		t.Fatalf("expected locked settings message, got: %s", w.Body.String())
	}
}

func TestPreStartSettingsFreezeRejectsAllSetupEndpointsAfterLobby(t *testing.T) {
	t.Run("treachery setup endpoints", func(t *testing.T) {
		tests := []struct {
			name        string
			path        string
			contentType string
			body        string
			assert      func(t *testing.T, room *game.Room)
		}{
			{
				name:        "role preset",
				path:        "/config/preset",
				contentType: "application/x-www-form-urlencoded",
				body:        "preset=custom",
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.PresetName != "standard" {
						t.Fatalf("preset mutated to %q", room.RoleConfig.PresetName)
					}
				},
			},
			{
				name: "player count",
				path: "/config/player-count/increment",
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.MaxPlayers != 5 {
						t.Fatalf("player count mutated to %d", room.RoleConfig.MaxPlayers)
					}
				},
			},
			{
				name: "role type count",
				path: "/config/role-type/Guardian/increment",
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.RoleTypes["Guardian"].Count != 2 {
						t.Fatalf("guardian count mutated to %d", room.RoleConfig.RoleTypes["Guardian"].Count)
					}
				},
			},
			{
				name:        "card toggle",
				path:        "/config/card-toggle",
				contentType: "application/json",
				body:        `{}`,
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.PresetName != "standard" {
						t.Fatalf("card toggle mutated preset to %q", room.RoleConfig.PresetName)
					}
				},
			},
			{
				name:        "leaderless variant",
				path:        "/config/leaderless",
				contentType: "application/json",
				body:        `{"allowLeaderless":true}`,
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.AllowLeaderlessGame {
						t.Fatal("leaderless variant mutated")
					}
				},
			},
			{
				name:        "hide distribution variant",
				path:        "/config/hide-distribution",
				contentType: "application/json",
				body:        `{"hideRoleDistribution":true}`,
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.HideRoleDistribution {
						t.Fatal("hide distribution variant mutated")
					}
				},
			},
			{
				name:        "fully random variant",
				path:        "/config/fully-random",
				contentType: "application/json",
				body:        `{"fullyRandomRoles":true}`,
				assert: func(t *testing.T, room *game.Room) {
					if room.RoleConfig.FullyRandomRoles {
						t.Fatal("fully random variant mutated")
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h, room, operator := newTreacheryRoomForPreStartFreezeTest(t, game.StateCountdown)
				w := postPreStartSettingForTest(t, h, room, operator, tt.path, tt.contentType, tt.body)

				if w.Code != http.StatusConflict {
					t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
				}
				if !containsPreStartSettingsLockedMessage(w.Body.String()) {
					t.Fatalf("expected locked settings message, got: %s", w.Body.String())
				}
				updatedRoom, _ := h.store.GetRoom(room.Code)
				tt.assert(t, updatedRoom)
			})
		}
	})

	t.Run("coup setup endpoints", func(t *testing.T) {
		tests := []struct {
			name        string
			path        string
			contentType string
			body        string
			assert      func(t *testing.T, room *game.Room)
		}{
			{
				name:        "role preset",
				path:        "/config/coup-preset",
				contentType: "application/x-www-form-urlencoded",
				body:        "preset=" + string(game.CoupPresetSix),
				assert: func(t *testing.T, room *game.Room) {
					if room.CoupPreset != game.CoupPresetFive {
						t.Fatalf("Coup preset mutated to %q", room.CoupPreset)
					}
				},
			},
			{
				name: "player count",
				path: "/config/coup-player-count/increment",
				assert: func(t *testing.T, room *game.Room) {
					if room.CoupPreset != game.CoupPresetFive {
						t.Fatalf("Coup player count mutated preset to %q", room.CoupPreset)
					}
				},
			},
			{
				name:        "role counts",
				path:        "/config/coup-role-counts",
				contentType: "application/x-www-form-urlencoded",
				body:        "king=1&blueKnight=1&blackKnight=2&redKnight=1&greenKnight=0&wastelandKnight=0",
				assert: func(t *testing.T, room *game.Room) {
					if room.CoupRoleCounts[game.RoleGreenKnight] != 1 {
						t.Fatalf("Coup role counts mutated to %v", room.CoupRoleCounts)
					}
				},
			},
			{
				name:        "information policy",
				path:        "/config/coup-info",
				contentType: "application/x-www-form-urlencoded",
				body:        "kingToBlue=king-knows-no-blue&redToBlack=red-knows-no-black&blackToRed=all-black-know-red&blackNetwork=black-network-all",
				assert: func(t *testing.T, room *game.Room) {
					if room.CoupInfoPolicy.KingToBlue == game.CoupKingKnowsNoBlue {
						t.Fatalf("Coup information policy mutated to %+v", room.CoupInfoPolicy)
					}
				},
			},
			{
				name:        "Royal Guard variant",
				path:        "/config/coup-royal-guard",
				contentType: "application/x-www-form-urlencoded",
				body:        "blockerLimit=1",
				assert: func(t *testing.T, room *game.Room) {
					if room.CoupRoyalGuardBlockerLimit != 0 {
						t.Fatalf("Royal Guard setting mutated to %d", room.CoupRoyalGuardBlockerLimit)
					}
				},
			},
			{
				name:        "Inquisition variant",
				path:        "/config/coup-inquisition",
				contentType: "application/x-www-form-urlencoded",
				body:        "resultPolicy=private",
				assert: func(t *testing.T, room *game.Room) {
					if room.CoupInquisitionResultPolicy == game.CoupInquisitionResultPrivate {
						t.Fatalf("Inquisition setting mutated to %q", room.CoupInquisitionResultPolicy)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h, room, operator := newCoupRoomForPreStartFreezeTest(t, game.StatePlaying)
				w := postPreStartSettingForTest(t, h, room, operator, tt.path, tt.contentType, tt.body)

				if w.Code != http.StatusConflict {
					t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
				}
				if !containsPreStartSettingsLockedMessage(w.Body.String()) {
					t.Fatalf("expected locked settings message, got: %s", w.Body.String())
				}
				updatedRoom, _ := h.store.GetRoom(room.Code)
				tt.assert(t, updatedRoom)
			})
		}
	})
}

func TestPreStartSettingsFreezeRejectsDebugModeSetupMutationAfterLobby(t *testing.T) {
	h, room, operator := newCoupRoomForPreStartFreezeTest(t, game.StatePlaying)
	h.config.Server.DebugModeEnabled = true

	w := postPreStartSettingForTest(t, h, room, operator, "/config/coup-info", "application/x-www-form-urlencoded", "kingToBlue=king-knows-no-blue")

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
	if !containsPreStartSettingsLockedMessage(w.Body.String()) {
		t.Fatalf("expected locked settings message, got: %s", w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.CoupInfoPolicy.KingToBlue == game.CoupKingKnowsNoBlue {
		t.Fatalf("debug mode setup mutation changed information policy to %+v", updatedRoom.CoupInfoPolicy)
	}
}

func TestPreStartSettingsFreezeRejectsLegacyRoleOptionsAfterLobby(t *testing.T) {
	h := newTestHandler()

	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeTreachery
	room.State = game.StatePlaying
	roleConfig, err := h.roleConfigService.CreateFromPreset("standard", 5)
	if err != nil {
		t.Fatalf("failed to create role config: %v", err)
	}
	room.RoleConfig = roleConfig
	host := game.NewPlayer("host", "Non-playing Host", "session-host")
	host.IsHost = true
	room.AddPlayer(host)
	markRoomOperatorForTest(room, host)
	h.store.UpdateRoom(room)

	w := postPreStartSettingForTest(t, h, room, host, "/options", "application/x-www-form-urlencoded", "card_id=31&key=use_all_cards&value=true")

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
	if !containsPreStartSettingsLockedMessage(w.Body.String()) {
		t.Fatalf("expected locked settings message, got: %s", w.Body.String())
	}
	updatedRoom, _ := h.store.GetRoom(room.Code)
	if updatedRoom.RoleOptionsManager != nil && updatedRoom.RoleOptionsManager.HasOptions(31) {
		t.Fatal("post-start role option update created card options")
	}
}

func newTreacheryRoomForPreStartFreezeTest(t *testing.T, state game.GameState) (*Handler, *game.Room, *game.Player) {
	t.Helper()

	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeTreachery
	room.State = state
	roleConfig, err := h.roleConfigService.CreateFromPreset("standard", 5)
	if err != nil {
		t.Fatalf("failed to create role config: %v", err)
	}
	room.RoleConfig = roleConfig
	operator := game.NewPlayer("operator", "Playing Operator", "session-operator")
	room.AddPlayer(operator)
	markRoomOperatorForTest(room, operator)
	h.store.UpdateRoom(room)

	return h, room, operator
}

func newCoupRoomForPreStartFreezeTest(t *testing.T, state game.GameState) (*Handler, *game.Room, *game.Player) {
	t.Helper()

	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = game.RulesModeCoup
	room.State = state
	room.CoupPreset = game.CoupPresetFive
	counts, ok := game.CoupRoleCountsForPreset(game.CoupPresetFive)
	if !ok {
		t.Fatal("missing five-player Coup preset")
	}
	room.CoupRoleCounts = counts
	room.CoupInfoPolicy = game.NormalizeCoupInformationPolicy(game.CoupInformationPolicy{})
	room.CoupRoyalGuardBlockerLimit = 0
	room.CoupInquisitionResultPolicy = game.CoupInquisitionResultPublic
	operator := game.NewPlayer("operator", "Playing Operator", "session-operator")
	room.AddPlayer(operator)
	markRoomOperatorForTest(room, operator)
	h.store.UpdateRoom(room)

	return h, room, operator
}

func postPreStartSettingForTest(t *testing.T, h *Handler, room *game.Room, operator *game.Player, path string, contentType string, body string) *httptest.ResponseRecorder {
	t.Helper()

	router := SetupRouter(h, h.config, &RouterOptions{
		DisableRateLimiting:  true,
		DisableRequestLogger: true,
	})

	req := httptest.NewRequest("POST", "/room/"+room.Code+path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	addPlayerSessionCookiesForTest(req, room, operator)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	return w
}

func TestPreStartSettingsFreezeOmitsSetupControlsFromPostStartDOM(t *testing.T) {
	for _, mode := range []game.RulesMode{game.RulesModeTreachery, game.RulesModeCoup} {
		t.Run(string(mode), func(t *testing.T) {
			actors := []struct {
				name      string
				actorKind string
				path      func(room *game.Room) string
			}{
				{
					name:      "playing Room Operator explicit dashboard",
					actorKind: "playing-operator",
					path: func(room *game.Room) string {
						return "/room/" + room.Code + "/operator"
					},
				},
				{
					name:      "non-playing Host dashboard",
					actorKind: "host",
					path: func(room *game.Room) string {
						return "/game/" + room.Code
					},
				},
				{
					name:      "normal Player View",
					actorKind: "player",
					path: func(room *game.Room) string {
						return "/game/" + room.Code
					},
				},
			}

			for _, state := range []game.GameState{game.StateCountdown, game.StatePlaying} {
				t.Run(string(state), func(t *testing.T) {
					for _, tt := range actors {
						t.Run(tt.name, func(t *testing.T) {
							h, room, actor := newPostStartRoomForPreStartDOMTest(t, mode, state, tt.actorKind)
							router := SetupRouter(h, h.config, &RouterOptions{
								DisableRateLimiting:  true,
								DisableRequestLogger: true,
							})

							req := httptest.NewRequest("GET", tt.path(room), nil)
							addPlayerSessionCookiesForTest(req, room, actor)
							w := httptest.NewRecorder()

							router.ServeHTTP(w, req)

							if w.Code != http.StatusOK {
								t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
							}
							assertNoPreStartSetupDOM(t, w.Body.String())
						})
					}
				})
			}
		})
	}
}

func newPostStartRoomForPreStartDOMTest(t *testing.T, mode game.RulesMode, state game.GameState, actorKind string) (*Handler, *game.Room, *game.Player) {
	t.Helper()

	h := newTestHandler()
	room, _ := h.store.CreateRoom()
	room.RulesMode = mode
	room.State = state
	if mode == game.RulesModeCoup {
		room.CoupPreset = game.CoupPresetFive
		counts, ok := game.CoupRoleCountsForPreset(game.CoupPresetFive)
		if !ok {
			t.Fatal("missing five-player Coup preset")
		}
		room.CoupRoleCounts = counts
	} else {
		roleConfig, err := h.roleConfigService.CreateFromPreset("standard", 5)
		if err != nil {
			t.Fatalf("failed to create role config: %v", err)
		}
		room.RoleConfig = roleConfig
	}

	operator := game.NewPlayer("operator", "Playing Operator", "session-operator")
	host := game.NewPlayer("host", "Non-playing Host", "session-host")
	host.IsHost = true
	player := game.NewPlayer("player", "Normal Player", "session-player")
	for _, p := range []*game.Player{operator, host, player} {
		if mode == game.RulesModeCoup {
			p.Role = mockHandlerCoupCard(1005, "Green Knight")
		} else {
			p.Role = mockGuardianCard()
		}
		room.AddPlayer(p)
	}
	room.OperatorSessionID = operator.SessionID
	if actorKind == "host" {
		room.OperatorSessionID = host.SessionID
	}
	h.store.UpdateRoom(room)

	switch actorKind {
	case "playing-operator":
		return h, room, operator
	case "host":
		return h, room, host
	case "player":
		return h, room, player
	default:
		t.Fatalf("unknown actor kind %q", actorKind)
		return nil, nil, nil
	}
}

func assertNoPreStartSetupDOM(t *testing.T, body string) {
	t.Helper()

	forbidden := []string{
		`Role Count Configuration`,
		`Unsafe Role Count Override`,
		`Allow Leaderless Games`,
		`Hide Role Distribution`,
		`Fully Random Roles`,
		`id="role-config"`,
		`id="host-dashboard-coup-setup"`,
		`id="operator-start-game"`,
		`/config/player-count`,
		`/config/role-type/`,
		`/config/preset`,
		`/config/coup-`,
	}
	for _, marker := range forbidden {
		if strings.Contains(body, marker) {
			t.Fatalf("post-start DOM rendered pre-start setup marker %q in %q", marker, body)
		}
	}
}
