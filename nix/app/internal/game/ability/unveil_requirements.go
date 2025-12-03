package ability

// UnveilRequirementType defines the type of input needed before unveiling
type UnveilRequirementType string

const (
	// NoInput - card unveils immediately with no player input
	NoInput UnveilRequirementType = "no_input"
	// NumericInput - player must provide a numeric value (X)
	NumericInput UnveilRequirementType = "numeric_input"
	// ChoiceInput - player must make a selection from options
	ChoiceInput UnveilRequirementType = "choice_input"
)

// UnveilRequirements defines what a card needs when being unveiled
type UnveilRequirements struct {
	// CardID is the card this requirement applies to
	CardID int

	// RequiresLeaderConfirmation - Leader must confirm physical card reveal
	// before the player sees ability options (prevents peeking)
	RequiresLeaderConfirmation bool

	// InputType specifies what input is needed before ability triggers
	InputType UnveilRequirementType

	// InputLabel is the user-friendly label for the input (e.g., "Mana to Spend (X)")
	InputLabel string

	// InputDescription provides context for the input
	InputDescription string

	// MinValue for numeric inputs
	MinValue int

	// MaxValue for numeric inputs (-1 for no limit)
	MaxValue int

	// DefaultValue for numeric inputs
	DefaultValue int

	// SetsFaceUp - whether unveiling this card sets it face up (usually true)
	SetsFaceUp bool
}

// unveilRequirements is the registry of all card unveil requirements
var unveilRequirements = map[int]*UnveilRequirements{
	// The Wearer of Masks (ID 31)
	// "Unveil {X}: As The Wearer of Masks is unveiled, reveal up to X non-Leader
	// identity cards at random from outside the game..."
	31: {
		CardID:                     31,
		RequiresLeaderConfirmation: true,
		InputType:                  NumericInput,
		InputLabel:                 "Mana to Spend (X)",
		InputDescription:           "Choose how much mana to spend. You will reveal up to X non-Leader identity cards to choose from.",
		MinValue:                   0,
		MaxValue:                   -1, // No limit
		DefaultValue:               3,
		SetsFaceUp:                 true,
	},

	// Default behavior for most cards will be handled by GetUnveilRequirements
	// returning a default struct
}

// GetUnveilRequirements returns the unveil requirements for a card
// Returns default requirements (no input, sets face up) if not explicitly defined
func GetUnveilRequirements(cardID int) *UnveilRequirements {
	if req, exists := unveilRequirements[cardID]; exists {
		return req
	}

	// Default requirements for cards without explicit definition
	return &UnveilRequirements{
		CardID:                     cardID,
		RequiresLeaderConfirmation: false,
		InputType:                  NoInput,
		SetsFaceUp:                 true,
	}
}

// HasUnveilRequirements checks if a card has any special unveil requirements
func HasUnveilRequirements(cardID int) bool {
	_, exists := unveilRequirements[cardID]
	return exists
}

// RequiresInput checks if a card needs player input before unveiling
func RequiresInput(cardID int) bool {
	req := GetUnveilRequirements(cardID)
	return req.InputType != NoInput
}

// RequiresConfirmation checks if a card needs leader confirmation before proceeding
func RequiresConfirmation(cardID int) bool {
	req := GetUnveilRequirements(cardID)
	return req.RequiresLeaderConfirmation
}
