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

		if !IsCardSupported(card.ID) {
			continue
		}

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
	leaders := cs.GetCardsForRoleType(RoleLeader)
	if len(leaders) == 0 {
		return nil
	}
	return leaders[rand.Intn(len(leaders))]
}

// GetRandomGuardian returns a random Guardian card
func (cs *CardService) GetRandomGuardian() *Card {
	guardians := cs.GetCardsForRoleType(RoleGuardian)
	if len(guardians) == 0 {
		return nil
	}
	return guardians[rand.Intn(len(guardians))]
}

// GetRandomAssassin returns a random Assassin card
func (cs *CardService) GetRandomAssassin() *Card {
	assassins := cs.GetCardsForRoleType(RoleAssassin)
	if len(assassins) == 0 {
		return nil
	}
	return assassins[rand.Intn(len(assassins))]
}

// GetRandomTraitor returns a random Traitor card
func (cs *CardService) GetRandomTraitor() *Card {
	traitors := cs.GetCardsForRoleType(RoleTraitor)
	if len(traitors) == 0 {
		return nil
	}
	return traitors[rand.Intn(len(traitors))]
}

// GetCardsForRoleType returns supported cards in a dealable role category.
func (cs *CardService) GetCardsForRoleType(cardType RoleType) []*Card {
	var cards []*Card

	switch cardType {
	case RoleLeader:
		cards = cs.Leaders
	case RoleGuardian:
		cards = cs.Guardians
	case RoleAssassin:
		cards = cs.Assassins
	case RoleTraitor:
		cards = cs.Traitors
	default:
		return nil
	}

	return filterSupportedCards(cards)
}

// IsCardConfigurable reports whether a named card can be enabled for a role type.
func (cs *CardService) IsCardConfigurable(cardType RoleType, cardName string) bool {
	for _, card := range cs.GetCardsForRoleType(cardType) {
		if card.Name == cardName {
			return true
		}
	}
	return false
}

// GetAllCards returns all cards from the card service
func (cs *CardService) GetAllCards() []*Card {
	cards := make([]*Card, len(cs.allCards))
	for i := range cs.allCards {
		cards[i] = &cs.allCards[i]
	}
	return cards
}

// GetRandomCards returns a specified number of random cards from a category
// ensuring no duplicates
func (cs *CardService) GetRandomCards(cardType RoleType, count int) []*Card {
	pool := cs.GetCardsForRoleType(cardType)

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
