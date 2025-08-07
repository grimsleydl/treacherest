package game

// CardTypes represents the type breakdown of a card
type CardTypes struct {
	Supertype string `json:"supertype"`
	Subtype   string `json:"subtype"`
}

// Card represents a MTG Treachery card
type Card struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	NameAnchor string    `json:"name_anchor"`
	URI        string    `json:"uri"`
	Cost       string    `json:"cost"`
	CMC        int       `json:"cmc"`
	Color      string    `json:"color"`
	Type       string    `json:"type"`
	Types      CardTypes `json:"types"`
	Rarity     string    `json:"rarity"`
	Text       string    `json:"text"`
	Flavor     string    `json:"flavor"`
	Artist     string    `json:"artist"`
	Rulings    []string  `json:"rulings"`
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
