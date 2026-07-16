package ability

// UnveilRequirementType defines the type of input needed before unveiling
type UnveilRequirementType string

// UnveilAdvisoryType identifies an advisory-only unveil condition.
type UnveilAdvisoryType string

const (
	// NoInput - card unveils immediately with no player input
	NoInput UnveilRequirementType = "no_input"
	// NumericInput - player must provide a numeric value (X)
	NumericInput UnveilRequirementType = "numeric_input"
	// ChoiceInput - player must make a selection from options
	ChoiceInput UnveilRequirementType = "choice_input"

	// NoUnveilAdvisory means the unveil flow has no advisory-only condition.
	NoUnveilAdvisory UnveilAdvisoryType = ""
	// UndercoverUnveilAdvisory reminds the owner about the table-adjudicated
	// Undercover condition without preventing them from unveiling.
	UndercoverUnveilAdvisory UnveilAdvisoryType = "undercover"
)

// UnveilRequirements defines what a card needs when being unveiled
type UnveilRequirements struct {
	// CardID is the card this requirement applies to.
	CardID int

	// CardIDs are used when multiple cards share the same requirement entry.
	CardIDs []int

	// RequiresLeaderConfirmation - Leader must confirm physical card reveal
	// before the player sees ability options (prevents peeking)
	RequiresLeaderConfirmation bool

	// InputType specifies what input is needed before ability triggers
	InputType UnveilRequirementType

	// InputLabel is the user-friendly label for the input (e.g., "Mana to Spend (X)")
	InputLabel string

	// InputDescription provides context for the input
	InputDescription string

	// AdvisoryType identifies a non-blocking advisory shown in the unveil flow.
	AdvisoryType UnveilAdvisoryType

	// MinValue for numeric inputs
	MinValue int

	// MaxValue for numeric inputs (-1 for no limit)
	MaxValue int

	// DefaultValue for numeric inputs
	DefaultValue int

	// SetsFaceUp - whether unveiling this card sets it face up (usually true)
	SetsFaceUp bool
}

// unveilRequirementEntries defines each distinct unveil requirement once. Cards
// with the same mechanic share one entry and only add another ID to CardIDs.
var unveilRequirementEntries = []*UnveilRequirements{
	// The Supplier (ID 17) and The Warlock (ID 18)
	// "Undercover" permits unveiling after another identity has been unveiled or
	// another player attacked the Leader. The app can observe only the first half,
	// so this requirement is advisory-only and remains NoInput.
	{
		CardIDs:      []int{17, 18},
		InputType:    NoInput,
		AdvisoryType: UndercoverUnveilAdvisory,
		SetsFaceUp:   true,
	},

	// The Metamorph (ID 25)
	// "Identity — Traitor
	// Unveil—Pay 8 life, Discard a card.
	// When The Metamorph is unveiled, until end of turn, as an opponent loses the game,
	// you may remove The Metamorph from the game. If you do, gain control of that player's
	// identity card and turn it face down if it isn't a Leader."
	{
		CardID:                     25,
		RequiresLeaderConfirmation: true,
		InputType:                  NoInput,
		InputLabel:                 "",
		InputDescription:           "Pay 8 life and discard a card to unveil The Metamorph. Until end of turn, when an opponent loses, you may steal their identity.",
		SetsFaceUp:                 true,
	},

	// The Puppet Master (ID 27)
	// "Unveil {6}: When The Puppet Master is unveiled, redistribute control of any number
	// of other identity cards. Then turn face down each of those cards that isn't a Leader.
	// You may look at face-down identity cards you don't control any time.
	// At the beginning of your end step, draw two cards."
	{
		CardID:                     27,
		RequiresLeaderConfirmation: true,
		InputType:                  ChoiceInput,
		InputLabel:                 "Redistribute Identity Cards",
		InputDescription:           "Pay {6} to unveil. Choose which players to include in the redistribution and how to reassign their identity cards.",
		SetsFaceUp:                 true,
	},

	// The Wearer of Masks (ID 31)
	// "Unveil {X}: As The Wearer of Masks is unveiled, reveal up to X non-Leader
	// identity cards at random from outside the game..."
	{
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
}

// unveilRequirements indexes the shared entries by every card ID they cover.
var unveilRequirements = indexUnveilRequirements(unveilRequirementEntries)

func indexUnveilRequirements(entries []*UnveilRequirements) map[int]*UnveilRequirements {
	indexed := make(map[int]*UnveilRequirements)
	for _, requirement := range entries {
		if requirement.CardID != 0 {
			indexed[requirement.CardID] = requirement
		}
		for _, cardID := range requirement.CardIDs {
			indexed[cardID] = requirement
		}
	}
	return indexed
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
