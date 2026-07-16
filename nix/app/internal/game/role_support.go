package game

// unsupportedCardIDs is the interim exclusion list for role mechanics that the
// app cannot represent yet. Remove IDs from this set as support lands.
var unsupportedCardIDs = map[int]struct{}{}

// IsCardSupported reports whether a card may be offered or dealt.
func IsCardSupported(cardID int) bool {
	_, unsupported := unsupportedCardIDs[cardID]
	return !unsupported
}

func filterSupportedCards(cards []*Card) []*Card {
	filtered := make([]*Card, 0, len(cards))
	for _, card := range cards {
		if card != nil && IsCardSupported(card.ID) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}
