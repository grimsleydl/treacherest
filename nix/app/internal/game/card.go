package game

import (
	"fmt"
	"strings"
)

var coupGreenBlueHuntWinConditionBullets = []string{
	"You serve neither crown.",
	"A crown is legitimate only after the hidden guard bleeds.",
	"You are hunting Blue Knights.",
	"Your Hunt is satisfied when at least one Blue Knight dies before King Fall.",
	"Blue dying with the King does not count.",
	"If Inquisition succeeds, you may share a King-side victory even without a Blue death.",
	"You may share a Red-side victory only if your Hunt was satisfied before King Fall.",
	"Broad Amnesty can let successful Inquisition before King Fall satisfy that Red-side lock.",
	"You do not share Black or Wasteland victories.",
}

var coupGreenBlueHuntPublicWinConditionBullets = []string{
	"Green serves neither crown.",
	"Green hunts Blue Knights.",
	"The default Hunt is satisfied when at least one Blue Knight dies before King Fall.",
	"Blue dying with the King does not count.",
	"If Inquisition succeeds, Green may share a King-side victory even without a Blue death.",
	"Green may share a Red-side victory only if Green Hunt was satisfied before King Fall.",
	"Broad Amnesty can let successful Inquisition before King Fall satisfy that Red-side lock.",
	"Green does not share Black or Wasteland victories.",
}

var CoupGreenBlueHuntWinCondition = strings.Join(coupGreenBlueHuntWinConditionBullets, " ")

// CoupStrictGreenWinCondition is retained for compatibility with older saved
// room/card data. New Coup role text should use CoupGreenBlueHuntWinCondition.
var CoupStrictGreenWinCondition = CoupGreenBlueHuntWinCondition

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
	case "King":
		return RoleKing
	case "Blue Knight":
		return RoleBlueKnight
	case "Black Knight":
		return RoleBlackKnight
	case "Red Knight":
		return RoleRedKnight
	case "Green Knight":
		return RoleGreenKnight
	case "Wasteland Knight":
		return RoleWasteland
	default:
		return ""
	}
}

// GetWinCondition returns a simplified win condition based on card type
func (c *Card) GetWinCondition() string {
	switch c.Types.Subtype {
	case "Leader":
		return "The Leader and their Guardians win if they are the last players standing."
	case "Guardian":
		return "The Guardians help the Leader, they win or lose with them."
	case "Assassin":
		return "The Assassins win if the Leader is eliminated."
	case "Traitor":
		return "The Traitor wins if they are the last player standing."
	case "King":
		return "Win if alive when Black, Red, and Wasteland threats are eliminated."
	case "Blue Knight":
		return "Win with the King. Lose when the King loses."
	case "Black Knight":
		return "Win if the King is dead, at least one Black Knight survives, and Red is dead."
	case "Red Knight":
		return "Win if the King is dead, Red survives, and all Black Knights are dead."
	case "Green Knight":
		return CoupGreenBlueHuntWinCondition
	case "Wasteland Knight":
		return "Win alone when every other player is eliminated."
	default:
		return ""
	}
}

func (c *Card) GetWinConditionBullets() []string {
	if c.GetRoleType() != RoleGreenKnight {
		return nil
	}
	bullets := make([]string, len(coupGreenBlueHuntWinConditionBullets))
	copy(bullets, coupGreenBlueHuntWinConditionBullets)
	return bullets
}

func (c *Card) GetPublicWinCondition() string {
	if c.GetRoleType() == RoleGreenKnight {
		return strings.Join(coupGreenBlueHuntPublicWinConditionBullets, " ")
	}
	return c.GetWinCondition()
}

func (c *Card) GetPublicWinConditionBullets() []string {
	if c.GetRoleType() != RoleGreenKnight {
		return nil
	}
	bullets := make([]string, len(coupGreenBlueHuntPublicWinConditionBullets))
	copy(bullets, coupGreenBlueHuntPublicWinConditionBullets)
	return bullets
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

// GetID returns the card ID (implements ability.CardLike interface)
func (c *Card) GetID() int {
	return c.ID
}

// GetText returns the card text (implements ability.CardLike interface)
func (c *Card) GetText() string {
	return c.Text
}
