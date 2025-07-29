package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"treacherest/internal/game"
)

func TestGenerateRandomName(t *testing.T) {
	// Test multiple generations to ensure randomness
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := generateRandomName()
		
		// Check length
		if len(name) != 5 {
			t.Errorf("Expected name length 5, got %d for name '%s'", len(name), name)
		}
		
		// Check all lowercase
		for _, ch := range name {
			if ch < 'a' || ch > 'z' {
				t.Errorf("Expected all lowercase letters, got '%c' in name '%s'", ch, name)
			}
		}
		
		names[name] = true
	}
	
	// Should have generated at least 90 unique names out of 100 (allowing for some collisions)
	if len(names) < 90 {
		t.Errorf("Expected at least 90 unique names out of 100 generations, got %d", len(names))
	}
}

func TestJoinRoomPostWithRandomName(t *testing.T) {
	h := newTestHandler()
	
	// Create a room
	room, _ := h.store.CreateRoom()
	
	// Join with empty name
	formData := "room_code=" + room.Code + "&player_name="
	req := httptest.NewRequest("POST", "/join-room", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	
	h.JoinRoomPost(w, req)
	
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}
	
	// Verify player was added with generated name
	room, _ = h.store.GetRoom(room.Code)
	if len(room.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(room.Players))
	}
	
	// Get the first player from the map
	var player *game.Player
	for _, p := range room.Players {
		player = p
		break
	}
	if len(player.Name) != 5 {
		t.Errorf("expected generated name to be 5 characters, got %d", len(player.Name))
	}
	
	// Check all characters are lowercase letters
	for _, ch := range player.Name {
		if ch < 'a' || ch > 'z' {
			t.Errorf("expected all lowercase letters, got '%c' in name '%s'", ch, player.Name)
		}
	}
}