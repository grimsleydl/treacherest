package game

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
)

// CardService manages the loaded cards and provides methods to access them
type CardService struct {
	Leaders   []*Card
	Guardians []*Card
	Assassins []*Card
	Traitors  []*Card
	allCards  []Card
}

// NewCardService creates a new CardService by loading cards from the JSON file
func NewCardService() (*CardService, error) {
	// Try multiple possible paths for the JSON file
	possiblePaths := []string{
		"docs/external/treachery-cards.json",
		"../../docs/external/treachery-cards.json",
		filepath.Join(os.Getenv("PRJ_ROOT"), "docs/external/treachery-cards.json"),
		"/workspace/docs/external/treachery-cards.json",
	}

	var jsonData []byte
	var err error
	var successPath string

	for _, path := range possiblePaths {
		jsonData, err = os.ReadFile(path)
		if err == nil {
			successPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read treachery-cards.json from any known location: %w", err)
	}

	var collection CardCollection
	if err := json.Unmarshal(jsonData, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse treachery-cards.json from %s: %w", successPath, err)
	}

	service := &CardService{
		Leaders:   make([]*Card, 0),
		Guardians: make([]*Card, 0),
		Assassins: make([]*Card, 0),
		Traitors:  make([]*Card, 0),
		allCards:  collection.Cards,
	}

	// Find the base path for images
	var basePath string
	imagePaths := []string{
		"static/images/cards",
		"../../static/images/cards",
		filepath.Join(os.Getenv("PRJ_ROOT"), "nix/app/static/images/cards"),
		"/workspace/nix/app/static/images/cards",
	}
	
	for _, path := range imagePaths {
		if _, err := os.Stat(filepath.Join(path, "1.jpg")); err == nil {
			basePath = path
			break
		}
	}
	
	if basePath == "" {
		return nil, fmt.Errorf("failed to find card images directory")
	}

	// Categorize cards by subtype and load images
	for i := range collection.Cards {
		card := &collection.Cards[i]
		
		// Load and encode the image
		imagePath := filepath.Join(basePath, fmt.Sprintf("%d.jpg", card.ID))
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read image for card %d (%s): %w", card.ID, card.Name, err)
		}
		
		// Detect MIME type
		mimeType := http.DetectContentType(imageData)
		
		// Create base64 data URI
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		card.Base64Image = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
		
		// Keep image path for backward compatibility
		card.ImagePath = fmt.Sprintf("/static/images/cards/%d.jpg", card.ID)
		
		switch card.Types.Subtype {
		case "Leader":
			service.Leaders = append(service.Leaders, card)
		case "Guardian":
			service.Guardians = append(service.Guardians, card)
		case "Assassin":
			service.Assassins = append(service.Assassins, card)
		case "Traitor":
			service.Traitors = append(service.Traitors, card)
		}
	}

	return service, nil
}

// GetRandomLeader returns a random Leader card
func (cs *CardService) GetRandomLeader() *Card {
	if len(cs.Leaders) == 0 {
		return nil
	}
	return cs.Leaders[rand.Intn(len(cs.Leaders))]
}

// GetRandomGuardian returns a random Guardian card
func (cs *CardService) GetRandomGuardian() *Card {
	if len(cs.Guardians) == 0 {
		return nil
	}
	return cs.Guardians[rand.Intn(len(cs.Guardians))]
}

// GetRandomAssassin returns a random Assassin card
func (cs *CardService) GetRandomAssassin() *Card {
	if len(cs.Assassins) == 0 {
		return nil
	}
	return cs.Assassins[rand.Intn(len(cs.Assassins))]
}

// GetRandomTraitor returns a random Traitor card
func (cs *CardService) GetRandomTraitor() *Card {
	if len(cs.Traitors) == 0 {
		return nil
	}
	return cs.Traitors[rand.Intn(len(cs.Traitors))]
}

// GetRandomCards returns a specified number of random cards from a category
// ensuring no duplicates
func (cs *CardService) GetRandomCards(cardType RoleType, count int) []*Card {
	var pool []*Card

	switch cardType {
	case RoleLeader:
		pool = cs.Leaders
	case RoleGuardian:
		pool = cs.Guardians
	case RoleAssassin:
		pool = cs.Assassins
	case RoleTraitor:
		pool = cs.Traitors
	default:
		return nil
	}

	if count > len(pool) {
		count = len(pool)
	}

	// Create a copy of the pool to avoid modifying the original
	poolCopy := make([]*Card, len(pool))
	copy(poolCopy, pool)

	// Fisher-Yates shuffle
	for i := len(poolCopy) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		poolCopy[i], poolCopy[j] = poolCopy[j], poolCopy[i]
	}

	return poolCopy[:count]
}
