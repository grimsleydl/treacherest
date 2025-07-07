package game

import "fmt"

// CardTypes represents the type breakdown of a card
type CardTypes struct {
	Supertype string `json:"supertype"`
	Subtype   string `json:"subtype"`
}

// Card represents a MTG Treachery card
type Card struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	NameAnchor  string    `json:"name_anchor"`
	URI         string    `json:"uri"`
	Cost        string    `json:"cost"`
	CMC         int       `json:"cmc"`
	Color       string    `json:"color"`
	Type        string    `json:"type"`
	Types       CardTypes `json:"types"`
	Rarity      string    `json:"rarity"`
	Text        string    `json:"text"`
	Flavor      string    `json:"flavor"`
	Artist      string    `json:"artist"`
	Rulings     []string  `json:"rulings"`
	ImagePath   string    `json:"-"` // Local image path, not from JSON
	Base64Image string    `json:"-"` // Base64-encoded image data URI
}

// CardCollection represents the full JSON structure
type CardCollection struct {
	GameVariant string  `json:"game_variant"`
	APIAuthor   string  `json:"api_author"`
	APIVersion  float64 `json:"api_version"`
	SetName     string  `json:"set_name"`
	SetCode     string  `json:"set_code"`
	SetLang     string  `json:"set_lang"`
	CardsCount  int     `json:"cards_count"`
	Cards       []Card  `json:"cards"`
}

// GetRoleType returns the RoleType equivalent for a card's subtype
func (c *Card) GetRoleType() RoleType {
	switch c.Types.Subtype {
	case "Leader":
		return RoleLeader
	case "Guardian":
		return RoleGuardian
	case "Assassin":
		return RoleAssassin
	case "Traitor":
		return RoleTraitor
	default:
		return ""
	}
}

// GetWinCondition returns a simplified win condition based on card type
func (c *Card) GetWinCondition() string {
	switch c.Types.Subtype {
	case "Leader":
		return "Survive and be the last player standing"
	case "Guardian":
		return "Win or lose with the Leader"
	case "Assassin":
		return "Win if the Leader is eliminated"
	case "Traitor":
		return "Be the last player standing"
	default:
		return ""
	}
}

// GetLeaderlessWinCondition returns win conditions for leaderless games
func (c *Card) GetLeaderlessWinCondition() string {
	switch c.Types.Subtype {
	case "Leader":
		return "Not applicable in leaderless games"
	case "Guardian":
		return "Win if majority of non-Traitor players survive"
	case "Assassin":
		return "Win by eliminating your secret target"
	case "Traitor":
		return "Be the last player standing"
	default:
		return ""
	}
}

// IsLeaderDependent returns true if the card's ability references the Leader
func (c *Card) IsLeaderDependent() bool {
	// List of cards that specifically reference the Leader
	leaderDependentCards := map[string]bool{
		"The Golem":         true,
		"The Great Martyr":  true,
		"The Oracle":        true,
		"The Quellmaster":   true,
		"The Metamorph":     true,
		"The Puppet Master": true,
	}

	return leaderDependentCards[c.Name]
}

// GetImagePath returns the local path to the card image
func (c *Card) GetImagePath() string {
	if c.ImagePath != "" {
		return c.ImagePath
	}
	// Default path based on card ID
	return fmt.Sprintf("/static/images/cards/%d.jpg", c.ID)
}

// GetPlaceholderPath returns the path to the placeholder image
func (c *Card) GetPlaceholderPath() string {
	return "/static/images/cards/placeholder.svg"
}

// GetImageBase64 returns the base64-encoded image data URI
func (c *Card) GetImageBase64() string {
	return c.Base64Image
}
