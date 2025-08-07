package handlers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"treacherest/internal/store"
)

func TestNew(t *testing.T) {
	memStore := store.NewMemoryStore()
	handler := New(memStore)
	
	if handler == nil {
		t.Fatal("New returned nil handler")
	}
	
	if handler.store != memStore {
		t.Error("handler store is not the provided store")
	}
	
	if handler.eventBus == nil {
		t.Error("handler eventBus is nil")
	}
}

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	
	if eb == nil {
		t.Fatal("NewEventBus returned nil")
	}
	
	if eb.subscribers == nil {
		t.Fatal("subscribers map not initialized")
	}
	
	if len(eb.subscribers) != 0 {
		t.Errorf("expected empty subscribers map, got %d", len(eb.subscribers))
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	eb := NewEventBus()
	
	t.Run("creates subscription channel", func(t *testing.T) {
		ch := eb.Subscribe("room1")
		
		if ch == nil {
			t.Fatal("Subscribe returned nil channel")
		}
		
		// Verify channel is buffered
		select {
		case ch <- Event{Type: "test"}:
			// Should not block
		default:
			t.Error("channel appears to be unbuffered")
		}
	})
	
	t.Run("multiple subscriptions to same room", func(t *testing.T) {
		ch1 := eb.Subscribe("room2")
		ch2 := eb.Subscribe("room2")
		ch3 := eb.Subscribe("room2")
		
		if ch1 == ch2 || ch2 == ch3 || ch1 == ch3 {
			t.Error("Subscribe returned same channel for different subscriptions")
		}
		
		// Verify all channels are stored
		eb.mu.RLock()
		subs := eb.subscribers["room2"]
		eb.mu.RUnlock()
		
		if len(subs) != 3 {
			t.Errorf("expected 3 subscribers for room2, got %d", len(subs))
		}
	})
	
	t.Run("subscriptions to different rooms", func(t *testing.T) {
		eb := NewEventBus() // Fresh event bus
		
		ch1 := eb.Subscribe("room1")
		ch2 := eb.Subscribe("room2")
		ch3 := eb.Subscribe("room3")
		
		// Verify separate storage
		eb.mu.RLock()
		defer eb.mu.RUnlock()
		
		if len(eb.subscribers["room1"]) != 1 {
			t.Errorf("expected 1 subscriber for room1, got %d", len(eb.subscribers["room1"]))
		}
		if len(eb.subscribers["room2"]) != 1 {
			t.Errorf("expected 1 subscriber for room2, got %d", len(eb.subscribers["room2"]))
		}
		if len(eb.subscribers["room3"]) != 1 {
			t.Errorf("expected 1 subscriber for room3, got %d", len(eb.subscribers["room3"]))
		}
		
		// Verify channels are different
		if ch1 == ch2 || ch2 == ch3 || ch1 == ch3 {
			t.Error("Subscribe returned same channel for different rooms")
		}
	})
}

func TestEventBus_Unsubscribe(t *testing.T) {
	t.Run("unsubscribe existing channel", func(t *testing.T) {
		eb := NewEventBus()
		
		ch1 := eb.Subscribe("room1")
		ch2 := eb.Subscribe("room1")
		ch3 := eb.Subscribe("room1")
		
		// Clean up to avoid unused variable warning
		defer close(ch1)
		defer close(ch3)
		
		// Unsubscribe ch2
		eb.Unsubscribe("room1", ch2)
		
		// Verify ch2 is closed
		select {
		case _, ok := <-ch2:
			if ok {
				t.Error("channel should be closed")
			}
		default:
			t.Error("channel should be closed and readable")
		}
		
		// Verify only 2 subscribers remain
		eb.mu.RLock()
		subs := eb.subscribers["room1"]
		eb.mu.RUnlock()
		
		if len(subs) != 2 {
			t.Errorf("expected 2 subscribers after unsubscribe, got %d", len(subs))
		}
		
		// Verify correct channels remain
		for _, sub := range subs {
			if sub == ch2 {
				t.Error("unsubscribed channel still in subscribers")
			}
		}
	})
	
	t.Run("unsubscribe non-existent channel", func(t *testing.T) {
		eb := NewEventBus()
		ch := make(chan Event)
		
		// Should not panic
		eb.Unsubscribe("room1", ch)
	})
	
	t.Run("unsubscribe from non-existent room", func(t *testing.T) {
		eb := NewEventBus()
		ch := make(chan Event)
		
		// Should not panic
		eb.Unsubscribe("nonexistent", ch)
	})
	
	t.Run("unsubscribe last subscriber", func(t *testing.T) {
		eb := NewEventBus()
		ch := eb.Subscribe("room1")
		
		eb.Unsubscribe("room1", ch)
		
		// Verify room still exists but with empty slice
		eb.mu.RLock()
		subs, exists := eb.subscribers["room1"]
		eb.mu.RUnlock()
		
		if !exists {
			t.Error("room key was deleted instead of having empty slice")
		}
		
		if len(subs) != 0 {
			t.Errorf("expected 0 subscribers, got %d", len(subs))
		}
	})
}

func TestEventBus_Publish(t *testing.T) {
	t.Run("publish to subscribers", func(t *testing.T) {
		eb := NewEventBus()
		
		ch1 := eb.Subscribe("room1")
		ch2 := eb.Subscribe("room1")
		ch3 := eb.Subscribe("room2")
		
		event := Event{
			Type:     "test",
			RoomCode: "room1",
			Data:     "test data",
		}
		
		eb.Publish(event)
		
		// Verify ch1 and ch2 receive the event
		select {
		case e := <-ch1:
			if e.Type != event.Type || e.RoomCode != event.RoomCode {
				t.Error("received incorrect event on ch1")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("ch1 did not receive event")
		}
		
		select {
		case e := <-ch2:
			if e.Type != event.Type || e.RoomCode != event.RoomCode {
				t.Error("received incorrect event on ch2")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("ch2 did not receive event")
		}
		
		// Verify ch3 does not receive the event
		select {
		case <-ch3:
			t.Error("ch3 received event for different room")
		case <-time.After(100 * time.Millisecond):
			// Expected
		}
	})
	
	t.Run("publish to full channel", func(t *testing.T) {
		eb := NewEventBus()
		ch := eb.Subscribe("room1")
		
		// Fill the channel
		for i := 0; i < 10; i++ {
			ch <- Event{Type: "fill"}
		}
		
		// Publish should not block even though channel is full
		done := make(chan bool)
		go func() {
			eb.Publish(Event{Type: "test", RoomCode: "room1"})
			done <- true
		}()
		
		select {
		case <-done:
			// Good, didn't block
		case <-time.After(100 * time.Millisecond):
			t.Error("Publish blocked on full channel")
		}
	})
	
	t.Run("publish to no subscribers", func(t *testing.T) {
		eb := NewEventBus()
		
		// Should not panic
		eb.Publish(Event{Type: "test", RoomCode: "room1"})
	})
}

func TestEventBus_ConcurrentOperations(t *testing.T) {
	eb := NewEventBus()
	var wg sync.WaitGroup
	
	// Multiple goroutines subscribing
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ch := eb.Subscribe("room1")
				// Simulate some work
				time.Sleep(time.Microsecond)
				eb.Unsubscribe("room1", ch)
			}
		}(i)
	}
	
	// Multiple goroutines publishing
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				eb.Publish(Event{
					Type:     "test",
					RoomCode: "room1",
					Data:     id,
				})
			}
		}(i)
	}
	
	wg.Wait()
}

func TestGetOrCreateSession(t *testing.T) {
	t.Run("creates new session when no cookie exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		
		sessionID := getOrCreateSession(w, req)
		
		if sessionID == "" {
			t.Error("expected non-empty session ID")
		}
		
		// Verify cookie was set
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected 1 cookie, got %d", len(cookies))
		}
		
		cookie := cookies[0]
		if cookie.Name != "session" {
			t.Errorf("expected cookie name 'session', got %s", cookie.Name)
		}
		if cookie.Value != sessionID {
			t.Errorf("cookie value %s doesn't match returned session ID %s", cookie.Value, sessionID)
		}
		if !cookie.HttpOnly {
			t.Error("cookie should be HttpOnly")
		}
		if cookie.SameSite != http.SameSiteLaxMode {
			t.Error("cookie should have SameSite=Lax")
		}
		if cookie.MaxAge != 86400*7 {
			t.Errorf("expected MaxAge 7 days, got %d", cookie.MaxAge)
		}
	})
	
	t.Run("returns existing session from cookie", func(t *testing.T) {
		existingSession := "existing-session-id"
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: existingSession,
		})
		w := httptest.NewRecorder()
		
		sessionID := getOrCreateSession(w, req)
		
		if sessionID != existingSession {
			t.Errorf("expected %s, got %s", existingSession, sessionID)
		}
		
		// Verify no new cookie was set
		cookies := w.Result().Cookies()
		if len(cookies) != 0 {
			t.Errorf("expected no new cookies, got %d", len(cookies))
		}
	})
	
	t.Run("generates unique session IDs", func(t *testing.T) {
		sessions := make(map[string]bool)
		
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			sessionID := getOrCreateSession(w, req)
			
			if sessions[sessionID] {
				t.Errorf("duplicate session ID generated: %s", sessionID)
			}
			sessions[sessionID] = true
		}
	})
}

func TestGeneratePlayerID(t *testing.T) {
	t.Run("generates non-empty ID", func(t *testing.T) {
		id := generatePlayerID()
		if id == "" {
			t.Error("expected non-empty player ID")
		}
	})
	
	t.Run("generates hex string", func(t *testing.T) {
		id := generatePlayerID()
		// Hex string of 8 bytes should be 16 characters
		if len(id) != 16 {
			t.Errorf("expected 16 character hex string, got %d", len(id))
		}
		
		// Verify it's valid hex
		for _, c := range id {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("invalid hex character: %c", c)
			}
		}
	})
	
	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		
		for i := 0; i < 1000; i++ {
			id := generatePlayerID()
			if ids[id] {
				t.Errorf("duplicate player ID generated: %s", id)
			}
			ids[id] = true
		}
	})
}