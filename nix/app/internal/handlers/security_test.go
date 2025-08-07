package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestURLParameterSecurityFix verifies that the URL parameter vulnerability is fixed
func TestURLParameterSecurityFix(t *testing.T) {
	h := newTestHandler()
	testRouter := setupTestRouter(h)

	// Create a room
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=Alice"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.CreateRoom(w, r)

	location := w.Header().Get("Location")
	roomCode := strings.TrimPrefix(location, "/room/")

	// Get cookies
	cookies := w.Result().Cookies()
	var sessionCookie, playerCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
		} else if c.Name == "player_"+roomCode {
			playerCookie = c
		}
	}

	t.Run("URLParameterDoesNotChangeName", func(t *testing.T) {
		// Try to access room with ?name= parameter (the vulnerability)
		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name=HACKER", nil)
		req.AddCookie(sessionCookie)
		req.AddCookie(playerCookie)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Should still show the original player name, not "HACKER"
		body := w.Body.String()
		if strings.Contains(body, "HACKER") {
			t.Error("URL parameter vulnerability not fixed - name parameter was processed")
		}
		if !strings.Contains(body, "Alice") {
			t.Error("Original player name not preserved")
		}

		// Verify player name in store wasn't changed
		room, _ := h.store.GetRoom(roomCode)
		for _, player := range room.Players {
			if player.Name == "HACKER" {
				t.Error("Player name was changed via URL parameter - vulnerability exists!")
			}
		}
	})

	t.Run("JoinRequiresPOST", func(t *testing.T) {
		// Create a new room to test joining
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=Bob"))
		r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		h.CreateRoom(w2, r2)

		location2 := w2.Header().Get("Location")
		roomCode2 := strings.TrimPrefix(location2, "/room/")

		// Try to join with GET and name parameter (old vulnerable way)
		req := httptest.NewRequest("GET", "/room/"+roomCode2+"?name=Charlie", nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Should show join form, not add player
		body := w.Body.String()
		if !strings.Contains(body, "Enter your name") || !strings.Contains(body, "Join Game") {
			t.Error("Join form not shown - GET request might be processing name")
		}

		// Verify player was NOT added
		room, _ := h.store.GetRoom(roomCode2)
		if len(room.Players) != 1 { // Should only have Bob
			t.Errorf("Expected 1 player, got %d - GET request might have added player", len(room.Players))
		}
		for _, player := range room.Players {
			if player.Name == "Charlie" {
				t.Error("Player added via GET request - vulnerability exists!")
			}
		}
	})

	t.Run("JoinOnlyViaPOST", func(t *testing.T) {
		// Create a new room
		w3 := httptest.NewRecorder()
		r3 := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=Dave"))
		r3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		h.CreateRoom(w3, r3)

		location3 := w3.Header().Get("Location")
		roomCode3 := strings.TrimPrefix(location3, "/room/")

		// Join properly via POST
		formData := "room_code=" + roomCode3 + "&player_name=Eve"
		req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Should redirect after successful join
		if w.Code != http.StatusSeeOther {
			t.Errorf("Expected redirect status 303, got %d", w.Code)
		}

		// Verify player was added
		room, _ := h.store.GetRoom(roomCode3)
		found := false
		for _, player := range room.Players {
			if player.Name == "Eve" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Player not added via POST - proper join flow broken")
		}
	})

	t.Run("NameValidationEnforced", func(t *testing.T) {
		// Test that name validation is enforced in POST endpoint
		testCases := []struct {
			name        string
			playerName  string
			shouldFail  bool
			description string
		}{
			{"", "EmptyName", true, "empty name should fail"},
			{"A", "A", false, "single character should pass"},
			{"ValidName123", "ValidName123", false, "alphanumeric should pass"},
			{"Name With Spaces", "Name With Spaces", false, "spaces should pass"},
			{"Name@Hacker!", "Name@Hacker!", true, "special characters should fail"},
			{"<script>alert('xss')</script>", "<script>alert('xss')</script>", true, "XSS attempt should fail"},
			{"Name!@#$%", "Name!@#$%", true, "multiple special chars should fail"},
			{"123456789012345678901", "123456789012345678901", true, "name too long should fail"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Create a new room for each test
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/room/create", strings.NewReader("playerName=TestHost"))
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				h.CreateRoom(w, r)

				location := w.Header().Get("Location")
				testRoomCode := strings.TrimPrefix(location, "/room/")

				// Try to join with the test name
				formData := "room_code=" + testRoomCode + "&player_name=" + tc.name
				req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.AddCookie(sessionCookie)

				w2 := httptest.NewRecorder()
				testRouter.ServeHTTP(w2, req)

				if tc.shouldFail {
					if w2.Code == http.StatusSeeOther {
						t.Errorf("Name %q should have been rejected but was accepted", tc.name)
					}
				} else {
					if w2.Code != http.StatusSeeOther {
						t.Errorf("Name %q should have been accepted but was rejected with status %d", tc.name, w2.Code)
						t.Logf("Response: %s", w2.Body.String())
					}
				}
			})
		}
	})
}