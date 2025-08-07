package handlers

import (
	"testing"
	"treacherest/internal/store"
)

func TestHandler_Store(t *testing.T) {
	memStore := store.NewMemoryStore()
	handler := New(memStore)

	// Test that Store() returns the same store
	if handler.Store() != memStore {
		t.Error("Store() did not return the expected store instance")
	}

	// Test that we can use the store through the handler
	room, err := handler.Store().CreateRoom()
	if err != nil {
		t.Fatalf("failed to create room through handler store: %v", err)
	}

	// Verify room was created
	retrievedRoom, err := handler.Store().GetRoom(room.Code)
	if err != nil {
		t.Fatalf("failed to retrieve room: %v", err)
	}

	if retrievedRoom.Code != room.Code {
		t.Error("retrieved room has different code")
	}
}
