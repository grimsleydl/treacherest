package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Create output directory
	outputDir := "../static/images/cards"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Download a few sample cards for testing
	// Using the actual URL pattern: https://mtgtreachery.net/images/cards/en/trd/{Role} - {Card Name}.jpg
	sampleCards := []struct {
		ID   int
		Name string
		Role string
	}{
		{44, "The Blood Empress", "Leader"},
		{3, "The Bodyguard", "Guardian"},
		{28, "The Ambitious Queen", "Assassin"},
		{17, "The Banisher", "Traitor"},
	}
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for _, card := range sampleCards {
		// Construct the correct URL
		fileName := fmt.Sprintf("%s - %s.jpg", card.Role, card.Name)
		encodedFileName := url.PathEscape(fileName)
		imageURL := fmt.Sprintf("https://mtgtreachery.net/images/cards/en/trd/%s", encodedFileName)
		
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.jpg", card.ID))

		// Check if file already exists
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("Skipping %s (already exists)\n", card.Name)
			continue
		}

		fmt.Printf("Downloading %s...\n", card.Name)

		// Create request with proper user agent
		req, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			fmt.Printf("Error creating request for %s: %v\n", card.Name, err)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
		
		// Download the image
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", card.Name, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error downloading %s: HTTP %d\n", card.Name, resp.StatusCode)
			continue
		}

		// Create the file
		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating file for %s: %v\n", card.Name, err)
			continue
		}
		defer file.Close()

		// Write the image data
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", card.Name, err)
			continue
		}

		fmt.Printf("Downloaded %s successfully\n", card.Name)

		// Small delay to be polite to the server
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\nSample download complete!")
}