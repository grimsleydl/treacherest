package game

// RulesMode identifies which ruleset a room is using.
type RulesMode string

const (
	RulesModeTreachery RulesMode = "treachery"
	RulesModeCoup      RulesMode = "coup"
)

// ParseRulesMode converts form/input values into supported rules modes.
func ParseRulesMode(value string) (RulesMode, bool) {
	switch value {
	case "", string(RulesModeTreachery):
		return RulesModeTreachery, true
	case string(RulesModeCoup):
		return RulesModeCoup, true
	default:
		return "", false
	}
}
