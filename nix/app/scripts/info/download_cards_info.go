package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// CardInfo represents the minimal structure needed for info display
type CardInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

// CardInfoCollection represents the JSON structure
type CardInfoCollection struct {
	Cards []CardInfo `json:"cards"`
}

func main() {
	fmt.Println("MTG Treachery Card Image Downloader")
	fmt.Println("===================================")
	fmt.Println()

	// Load the cards JSON - try multiple paths
	possiblePaths := []string{
		"../../docs/external/treachery-cards.json",
		"../../../docs/external/treachery-cards.json",
		os.Getenv("PRJ_ROOT") + "/docs/external/treachery-cards.json",
	}

	var jsonData []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		jsonData, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		fmt.Printf("Error reading treachery-cards.json: %v\n", err)
		return
	}

	var collection CardInfoCollection
	if err := json.Unmarshal(jsonData, &collection); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	fmt.Printf("Found %d cards in the collection (loaded from %s)\n\n", len(collection.Cards), foundPath)

	fmt.Println("NOTE: The card image URLs need to be determined from mtgtreachery.net")
	fmt.Println("The expected pattern (https://mtgtreachery.net/cards/TRD/{id}.jpg) returns 404")
	fmt.Println()
	fmt.Println("To find the correct URLs:")
	fmt.Println("1. Visit https://mtgtreachery.net/rules/oracle/")
	fmt.Println("2. Click on a card to view it")
	fmt.Println("3. Inspect the page to find the actual image URL")
	fmt.Println("4. Update the download_cards.go script with the correct URL pattern")
	fmt.Println()

	// Show some example cards that would be downloaded
	fmt.Println("Sample cards that would be downloaded:")
	for i, card := range collection.Cards {
		if i >= 5 {
			break
		}
		fmt.Printf("- Card #%d: %s (by %s)\n", card.ID, card.Name, card.Artist)
	}
	fmt.Printf("... and %d more\n", len(collection.Cards)-5)
	fmt.Println()

	fmt.Println("Once the correct URL pattern is found, run:")
	fmt.Println("  download-cards")
	fmt.Println()
	fmt.Println("For now, the game will use placeholder images.")
}
