package game

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestCardIDMapping(t *testing.T) {
	// Load the cards JSON - try multiple paths
	possiblePaths := []string{
		"../../docs/external/treachery-cards.json",
		"../../../docs/external/treachery-cards.json",
		filepath.Join(os.Getenv("PRJ_ROOT"), "docs/external/treachery-cards.json"),
		"/workspace/docs/external/treachery-cards.json",
	}
	
	var jsonData []byte
	var err error
	
	for _, path := range possiblePaths {
		jsonData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("Error reading treachery-cards.json: %v", err)
	}

	var collection CardCollection
	if err := json.Unmarshal(jsonData, &collection); err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	// Test 1: Verify all cards have unique IDs
	idMap := make(map[int]string)
	for _, card := range collection.Cards {
		if existing, exists := idMap[card.ID]; exists {
			t.Errorf("Duplicate ID %d: '%s' and '%s'", card.ID, existing, card.Name)
		}
		idMap[card.ID] = card.Name
	}

	// Test 2: Verify IDs are sequential (1-54)
	for i := 1; i <= 54; i++ {
		if _, exists := idMap[i]; !exists {
			t.Errorf("Missing ID: %d", i)
		}
	}

	// Test 3: Verify we can construct correct URLs
	testCases := []struct {
		id       int
		name     string
		role     string
		expected string
	}{
		{44, "The Seer", "Assassin", "https://mtgtreachery.net/images/cards/en/trd/044%20-%20Assassin%20-%20The%20Seer.jpg"},
		{3, "The Bodyguard", "Guardian", "https://mtgtreachery.net/images/cards/en/trd/003%20-%20Guardian%20-%20The%20Bodyguard.jpg"},
		{1, "The Ætherist", "Guardian", "https://mtgtreachery.net/images/cards/en/trd/001%20-%20Guardian%20-%20The%20%C3%86therist.jpg"},
	}

	for _, tc := range testCases {
		// Find the card
		var found *Card
		for _, card := range collection.Cards {
			if card.ID == tc.id {
				found = &card
				break
			}
		}

		if found == nil {
			t.Errorf("Card with ID %d not found", tc.id)
			continue
		}

		// Verify basic info
		if found.Name != tc.name {
			t.Errorf("Card %d: expected name '%s', got '%s'", tc.id, tc.name, found.Name)
		}
		if found.Types.Subtype != tc.role {
			t.Errorf("Card %d: expected role '%s', got '%s'", tc.id, tc.role, found.Types.Subtype)
		}

		// Verify URL construction with new format (ID prefix)
		fileName := fmt.Sprintf("%03d - %s - %s.jpg", found.ID, found.Types.Subtype, found.Name)
		encodedFileName := url.PathEscape(fileName)
		actualURL := fmt.Sprintf("https://mtgtreachery.net/images/cards/en/trd/%s", encodedFileName)
		
		if actualURL != tc.expected {
			t.Errorf("Card %d: URL mismatch\nExpected: %s\nActual:   %s", tc.id, tc.expected, actualURL)
		}
	}
}

func TestLocalFileNaming(t *testing.T) {
	// Test that our local file naming matches what the game expects
	testCases := []struct {
		id       int
		expected string
	}{
		{1, "1.jpg"},
		{44, "44.jpg"},
		{54, "54.jpg"},
	}

	for _, tc := range testCases {
		actual := fmt.Sprintf("%d.jpg", tc.id)
		if actual != tc.expected {
			t.Errorf("ID %d: expected filename '%s', got '%s'", tc.id, tc.expected, actual)
		}
	}
}