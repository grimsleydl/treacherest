package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var errDownloadFailed = errors.New("one or more card images failed to download")

// Card represents the minimal structure needed for downloading
type Card struct {
	ID    int       `json:"id"`
	Name  string    `json:"name"`
	Types CardTypes `json:"types"`
}

// CardTypes represents the type breakdown of a card
type CardTypes struct {
	Subtype string `json:"subtype"`
}

// CardCollection represents the JSON structure
type CardCollection struct {
	Cards []Card `json:"cards"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("MTG Treachery Card Image Downloader")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Println("This script will download card images from mtgtreachery.net")
	fmt.Println("It includes a 1-second delay between requests to be polite to their server.")
	fmt.Println()
	fmt.Print("Press Enter to continue or Ctrl+C to cancel...")
	fmt.Scanln()

	// Load the cards JSON from the current directory
	jsonData, err := os.ReadFile("treachery-cards.json")
	if err != nil {
		return fmt.Errorf("read treachery-cards.json: %w", err)
	}

	var collection CardCollection
	if err := json.Unmarshal(jsonData, &collection); err != nil {
		return fmt.Errorf("parse treachery-cards.json: %w", err)
	}

	// Create output directory in the current working directory.
	// The Nix build script will move this to the correct final location.
	outputDir := "cards"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	absPath, _ := filepath.Abs(outputDir)
	fmt.Printf("Images will be saved to: %s\n\n", absPath)

	// Download each card image
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	failed := 0
	for _, card := range collection.Cards {
		// Construct the correct URL pattern: https://mtgtreachery.net/images/cards/en/trd/{ID} - {Role} - {Card Name}.jpg
		role := card.Types.Subtype
		fileName := fmt.Sprintf("%03d - %s - %s.jpg", card.ID, role, card.Name)
		encodedFileName := url.PathEscape(fileName)
		imageURL := fmt.Sprintf("https://mtgtreachery.net/images/cards/en/trd/%s", encodedFileName)

		outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.jpg", card.ID))

		// Check if file already exists
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("Skipping %s (already exists)\n", card.Name)
			continue
		}

		fmt.Printf("Downloading %s...\n", card.Name)
		fmt.Printf("  URL: %s\n", imageURL)

		// Create request with proper user agent
		req, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			fmt.Printf("Error creating request for %s: %v\n", card.Name, err)
			failed++
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")

		// Download the image
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", card.Name, err)
			failed++
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error downloading %s: HTTP %d\n", card.Name, resp.StatusCode)
			resp.Body.Close()
			failed++
			continue
		}

		// Create the file
		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating file for %s: %v\n", card.Name, err)
			resp.Body.Close()
			failed++
			continue
		}

		// Write the image data
		if _, err = io.Copy(file, resp.Body); err != nil {
			fmt.Printf("Error writing %s: %v\n", card.Name, err)
			file.Close()
			resp.Body.Close()
			failed++
			continue
		}
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file for %s: %v\n", card.Name, err)
			resp.Body.Close()
			failed++
			continue
		}
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response for %s: %v\n", card.Name, err)
			failed++
			continue
		}

		fmt.Printf("Downloaded %s successfully\n", card.Name)

		// Delay to be polite to the server (1 second between requests)
		time.Sleep(1 * time.Second)
	}

	// Create attribution file
	attribution := `Card Image Attribution
=====================

The card images in this directory are from MTG Treachery (https://mtgtreachery.net).

As stated on their website:
- Pictures used for the custom Identity cards are owned by their illustrators
- MTG Treachery does not own any artworks depicted here
- The artworks were taken from various websites and properly credited to the artists at the bottom of each card
- Using these artworks should be seen as advertisement for the respective artists and their amazing work

The literal and graphical information presented about Magic: The Gathering, including card images, 
the mana symbols, and Oracle text, is copyright Wizards of the Coast, LLC, a subsidiary of Hasbro, Inc.

This project is not produced by, endorsed by, supported by, or affiliated with Wizards of the Coast.
`

	if err := os.WriteFile(filepath.Join(outputDir, "ATTRIBUTION.txt"), []byte(attribution), 0644); err != nil {
		return fmt.Errorf("write attribution file: %w", err)
	}

	if failed > 0 {
		return fmt.Errorf("%w: %d failed", errDownloadFailed, failed)
	}

	fmt.Println("\nDownload complete!")
	return nil
}
