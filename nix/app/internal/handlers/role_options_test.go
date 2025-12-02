package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"treacherest/internal/config"
	"treacherest/internal/game"
	"treacherest/internal/store"

	"github.com/go-chi/chi/v5"
)

// TestGetRoleOptions tests the GET role options endpoint
func TestGetRoleOptions(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room with options
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	room.AddPlayer(player1)

	// Set some options
	room.RoleOptionsManager.GetOrCreateOptions(31).SetOption("use_all_cards", true)
	room.RoleOptionsManager.GetOrCreateOptions(31).SetOption("max_reveal", 5)
	memStore.UpdateRoom(room)

	t.Run("Get existing options", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/options?card_id=31", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetRoleOptions(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "31") {
			t.Errorf("Expected response to contain card_id 31, got %s", body)
		}
	})

	t.Run("Get non-existent options", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/options?card_id=999", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetRoleOptions(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body != "{}" {
			t.Errorf("Expected empty JSON object, got %s", body)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/INVALID/options?card_id=31", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_INVALID",
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetRoleOptions(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("No player cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/options?card_id=31", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetRoleOptions(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Invalid card ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/room/"+room.Code+"/options?card_id=invalid", nil)
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetRoleOptions(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestSetRoleOption tests the POST role options endpoint
func TestSetRoleOption(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	player1.IsHost = true // Make player1 the host
	player2 := game.NewPlayer("player2", "Bob", "session2")
	room.AddPlayer(player1)
	room.AddPlayer(player2)
	memStore.UpdateRoom(room)

	t.Run("Set boolean option", func(t *testing.T) {
		form := url.Values{}
		form.Add("card_id", "31")
		form.Add("key", "use_all_cards")
		form.Add("value", "true")

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/options", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SetRoleOption(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify option was set
		updatedRoom, _ := memStore.GetRoom(room.Code)
		opts := updatedRoom.RoleOptionsManager.GetOrCreateOptions(31)
		val, _ := opts.GetBoolOption("use_all_cards")
		if !val {
			t.Error("Expected use_all_cards to be true")
		}
	})

	t.Run("Set integer option", func(t *testing.T) {
		form := url.Values{}
		form.Add("card_id", "31")
		form.Add("key", "max_reveal")
		form.Add("value", "5")

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/options", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SetRoleOption(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify option was set
		updatedRoom, _ := memStore.GetRoom(room.Code)
		opts := updatedRoom.RoleOptionsManager.GetOrCreateOptions(31)
		val, _ := opts.GetIntOption("max_reveal")
		if val != 5 {
			t.Errorf("Expected max_reveal to be 5, got %d", val)
		}
	})

	t.Run("Set string option", func(t *testing.T) {
		form := url.Values{}
		form.Add("card_id", "25")
		form.Add("key", "filter_mode")
		form.Add("value", "traitors_only")

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/options", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SetRoleOption(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify option was set
		updatedRoom, _ := memStore.GetRoom(room.Code)
		opts := updatedRoom.RoleOptionsManager.GetOrCreateOptions(25)
		val, _ := opts.GetStringOption("filter_mode")
		if val != "traitors_only" {
			t.Errorf("Expected filter_mode to be 'traitors_only', got %s", val)
		}
	})

	t.Run("Non-host cannot set options", func(t *testing.T) {
		form := url.Values{}
		form.Add("card_id", "31")
		form.Add("key", "test")
		form.Add("value", "value")

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/options", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player2", // Non-host
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SetRoleOption(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) {
		form := url.Values{}
		form.Add("key", "test") // Missing card_id

		req := httptest.NewRequest("POST", "/room/"+room.Code+"/options", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "player_" + room.Code,
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", room.Code)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SetRoleOption(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Room not found", func(t *testing.T) {
		form := url.Values{}
		form.Add("card_id", "31")
		form.Add("key", "test")
		form.Add("value", "value")

		req := httptest.NewRequest("POST", "/room/INVALID/options", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "player_INVALID",
			Value: "player1",
		})

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "INVALID")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.SetRoleOption(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// TestRoleOptionsEventPublishing tests that events are published correctly
func TestRoleOptionsEventPublishing(t *testing.T) {
	cfg := config.DefaultConfig()
	memStore := store.NewMemoryStore(cfg)
	eventBus := NewEventBus()
	handler := &Handler{
		store:    memStore,
		config:   cfg,
		eventBus: eventBus,
	}

	// Create room
	room, _ := memStore.CreateRoom()
	player1 := game.NewPlayer("player1", "Alice", "session1")
	player1.IsHost = true
	room.AddPlayer(player1)
	memStore.UpdateRoom(room)

	// Subscribe to events
	eventChan := eventBus.Subscribe(room.Code)

	// Set an option
	form := url.Values{}
	form.Add("card_id", "31")
	form.Add("key", "use_all_cards")
	form.Add("value", "true")

	req := httptest.NewRequest("POST", "/room/"+room.Code+"/options", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "player_" + room.Code,
		Value: "player1",
	})

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", room.Code)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SetRoleOption(w, req)

	// Wait for event
	select {
	case event := <-eventChan:
		if event.Type != "role_options_changed" {
			t.Errorf("Expected event type 'role_options_changed', got %s", event.Type)
		}
		if event.RoomCode != room.Code {
			t.Errorf("Expected room code %s, got %s", room.Code, event.RoomCode)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for role_options_changed event")
	}
}
