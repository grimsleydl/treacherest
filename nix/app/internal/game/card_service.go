package game

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

// CardService manages the loaded cards and provides methods to access them
type CardService struct {
	Leaders   []*Card
	Guardians []*Card
	Assassins []*Card
	Traitors  []*Card
	allCards  []Card
}

// NewCardService creates a new CardService by loading cards from embedded data
func NewCardService(jsonData []byte, imagesFS embed.FS) (*CardService, error) {
	// Parse the embedded JSON data
	var collection CardCollection
	if err := json.Unmarshal(jsonData, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse embedded treachery-cards.json: %w", err)
	}

	service := &CardService{
		Leaders:   make([]*Card, 0),
		Guardians: make([]*Card, 0),
		Assassins: make([]*Card, 0),
		Traitors:  make([]*Card, 0),
		allCards:  collection.Cards,
	}

	// Categorize cards by subtype and load images
	for i := range collection.Cards {
		card := &collection.Cards[i]
		
		// Load and encode the image from embedded filesystem
		imagePath := fmt.Sprintf("static/images/cards/%d.jpg", card.ID)
		imageData, err := imagesFS.ReadFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded image for card %d (%s): %w", card.ID, card.Name, err)
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
