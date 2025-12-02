package ability

import (
	"fmt"
	"regexp"
	"strings"
)

// CardLike is an interface that represents card properties needed for ability detection
// This avoids circular imports with the game package
type CardLike interface {
	GetID() int
	GetText() string
}

// AbilityRegistry manages all card abilities
type AbilityRegistry struct {
	abilities map[int]*Ability
}

// NewAbilityRegistry creates a new registry and auto-registers abilities
func NewAbilityRegistry(cards []CardLike) *AbilityRegistry {
	registry := &AbilityRegistry{
		abilities: make(map[int]*Ability),
	}

	// Auto-register pattern-based abilities
	for _, card := range cards {
		if ability := DetectAbilityPattern(card); ability != nil {
			registry.abilities[card.GetID()] = ability
		}
	}

	return registry
}

// Register manually registers an ability for a card
func (r *AbilityRegistry) Register(cardID int, ability *Ability) {
	r.abilities[cardID] = ability
}

// Get retrieves an ability by card ID
func (r *AbilityRegistry) Get(cardID int) *Ability {
	return r.abilities[cardID]
}

// Has checks if an ability exists for a card
func (r *AbilityRegistry) Has(cardID int) bool {
	_, exists := r.abilities[cardID]
	return exists
}

// Count returns the number of registered abilities
func (r *AbilityRegistry) Count() int {
	return len(r.abilities)
}

// ListAll returns all registered abilities
func (r *AbilityRegistry) ListAll() []*Ability {
	abilities := make([]*Ability, 0, len(r.abilities))
	for _, ability := range r.abilities {
		abilities = append(abilities, ability)
	}
	return abilities
}

// DetectAbilityPattern attempts to automatically detect an ability from card text
// Returns nil if the card doesn't fit known patterns
func DetectAbilityPattern(card CardLike) *Ability {
	// Empty text = no ability
	if card.GetText() == "" {
		return nil
	}

	// Detect trigger type
	triggerType := detectTriggerType(card)

	// Build basic ability structure
	ability := &Ability{
		CardID: card.GetID(),
		Trigger: Trigger{
			Type: triggerType,
		},
	}

	// Parse unveil cost if present
	if cost := parseUnveilCost(card); cost != "" {
		ability.Cost = parseCost(cost)
	}

	// Detect effect types
	effectTypes := detectEffectTypes(card.GetText())
	ability.Effects = make([]Effect, len(effectTypes))
	for i, effectType := range effectTypes {
		ability.Effects[i] = Effect{
			Type: effectType,
		}
	}

	// Detect conditions
	ability.Conditions = detectConditions(card.GetText())

	// Detect duration
	ability.Duration = detectDuration(card.GetText())

	return ability
}

// detectTriggerType determines the trigger type from card text
func detectTriggerType(card CardLike) TriggerType {
	text := card.GetText()

	// Check for "As X is unveiled" (replacement effect)
	if strings.Contains(text, "As ") && strings.Contains(text, " is unveiled") {
		return TriggerAsUnveil
	}

	// Check for "When X is unveiled"
	if strings.Contains(text, "When ") && strings.Contains(text, " is unveiled") {
		return TriggerOnUnveil
	}

	// Check for "At the beginning of"
	if strings.Contains(text, "At the beginning of") {
		return TriggerAtPhaseStep
	}

	// Check for "Whenever"
	if strings.Contains(text, "Whenever ") {
		return TriggerWhenever
	}

	// Check for activated ability pattern: {Cost}:
	activatedPattern := regexp.MustCompile(`\{[^}]+\}:`)
	if activatedPattern.MatchString(text) {
		return TriggerActivated
	}

	// Default to static ability
	return TriggerStatic
}

// parseUnveilCost extracts the unveil cost from card text
func parseUnveilCost(card CardLike) string {
	text := card.GetText()

	// Pattern 1: "Unveil {X}" or "Unveil {3}" etc.
	unveilPattern := regexp.MustCompile(`Unveil\s+(\{[^}]+\}(?:\{[^}]+\})*)`)
	if matches := unveilPattern.FindStringSubmatch(text); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 2: "Unveil—" followed by cost description (up to period or pipe)
	unveilDashPattern := regexp.MustCompile(`Unveil—([^.|]+)`)
	if matches := unveilDashPattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// parseCost converts a cost string into a Cost struct
func parseCost(costStr string) *Cost {
	cost := &Cost{}

	// Check for mana cost patterns
	manaPattern := regexp.MustCompile(`\{([^}]+)\}`)
	manaMatches := manaPattern.FindAllStringSubmatch(costStr, -1)

	if len(manaMatches) > 0 {
		cost.Mana = &ManaCost{}
		for _, match := range manaMatches {
			symbol := match[1]
			if symbol == "X" {
				cost.Mana.X = true
			} else if symbol == "Y" {
				cost.Mana.Y = true
			} else {
				// Try to parse as number
				var generic int
				if _, err := fmt.Sscanf(symbol, "%d", &generic); err == nil {
					cost.Mana.Generic += generic
				}
			}
		}
	}

	// Check for life payment
	lifePattern := regexp.MustCompile(`Pay (\d+) life`)
	if matches := lifePattern.FindStringSubmatch(costStr); len(matches) > 1 {
		var life int
		fmt.Sscanf(matches[1], "%d", &life)
		cost.Life = life
	}

	// Check for discard
	if strings.Contains(costStr, "Discard a card") {
		cost.Discard = 1
	}
	discardPattern := regexp.MustCompile(`Discard (\d+) card`)
	if matches := discardPattern.FindStringSubmatch(costStr); len(matches) > 1 {
		var discard int
		fmt.Sscanf(matches[1], "%d", &discard)
		cost.Discard = discard
	}

	return cost
}

// detectEffectTypes identifies effect types from card text
func detectEffectTypes(text string) []EffectType {
	var effects []EffectType
	textLower := strings.ToLower(text)

	// Order matters - check more specific patterns first

	// Transform/Copy effects
	if strings.Contains(textLower, "becomes a copy") || strings.Contains(textLower, "become a copy") {
		effects = append(effects, EffectTransform)
	}

	// Damage effects - check more specific pattern first to avoid duplicates
	if (strings.Contains(textLower, "deals") || strings.Contains(textLower, "deal")) && strings.Contains(textLower, "damage") {
		effects = append(effects, EffectDamage)
	}

	// Draw effects
	if strings.Contains(textLower, "draw") && (strings.Contains(textLower, "card") || strings.Contains(textLower, "cards")) {
		effects = append(effects, EffectDraw)
	}

	// Exile effects
	if strings.Contains(textLower, "exile") {
		effects = append(effects, EffectExile)
	}

	// Destroy effects
	if strings.Contains(textLower, "destroy") {
		effects = append(effects, EffectDestroy)
	}

	// Counter effects
	if strings.Contains(textLower, "counter target") {
		effects = append(effects, EffectCounter)
	}

	// Life gain/loss
	if strings.Contains(textLower, "gain") && strings.Contains(textLower, "life") {
		effects = append(effects, EffectLifeGainLoss)
	}
	if strings.Contains(textLower, "lose") && strings.Contains(textLower, "life") {
		effects = append(effects, EffectLifeGainLoss)
	}

	// Token creation
	if strings.Contains(textLower, "create") && strings.Contains(textLower, "token") {
		effects = append(effects, EffectCreateToken)
	}

	// Control change
	if strings.Contains(textLower, "gain control") {
		effects = append(effects, EffectControlChange)
	}

	// State change (face up/down)
	if strings.Contains(textLower, "turn") && (strings.Contains(textLower, "face down") || strings.Contains(textLower, "face up")) {
		effects = append(effects, EffectStateChange)
	}

	// Reveal effects
	if strings.Contains(textLower, "reveal") {
		effects = append(effects, EffectReveal)
	}

	// Search effects
	if strings.Contains(textLower, "search") && strings.Contains(textLower, "library") {
		effects = append(effects, EffectSearch)
	}

	// Reanimate effects
	if strings.Contains(textLower, "return") && strings.Contains(textLower, "from") && strings.Contains(textLower, "graveyard") {
		effects = append(effects, EffectReanimate)
	}
	if strings.Contains(textLower, "put") && strings.Contains(textLower, "graveyard") && strings.Contains(textLower, "battlefield") {
		effects = append(effects, EffectReanimate)
	}

	// Tap/Untap effects
	if strings.Contains(textLower, "tap") || strings.Contains(textLower, "untap") {
		effects = append(effects, EffectTapUntap)
	}

	return effects
}

// detectConditions identifies conditional execution patterns
func detectConditions(text string) []Condition {
	var conditions []Condition
	textLower := strings.ToLower(text)

	if strings.Contains(textLower, "you may") {
		conditions = append(conditions, Condition{Type: ConditionYouMay})
	}

	if strings.Contains(textLower, "if you do") {
		conditions = append(conditions, Condition{Type: ConditionIfYouDo})
	}

	if strings.Contains(text, "if it's your turn") {
		conditions = append(conditions, Condition{Type: ConditionIfYourTurn})
	}

	if strings.Contains(textLower, "for each") {
		conditions = append(conditions, Condition{Type: ConditionForEach})
	}

	return conditions
}

// detectDuration determines how long an effect lasts
func detectDuration(text string) *Duration {
	textLower := strings.ToLower(text)

	// Check for "until end of turn"
	if strings.Contains(textLower, "until end of turn") {
		return &Duration{
			Type:     DurationUntilEOT,
			EndPhase: "end_of_turn",
		}
	}

	// Check for "until [condition]"
	if strings.Contains(textLower, "until") {
		// Check for face down condition
		if strings.Contains(textLower, "face down") {
			return &Duration{
				Type:      DurationUntilCondition,
				Condition: "face_down",
			}
		}
		// Check for "until you lose the game"
		if strings.Contains(textLower, "until you lose the game") {
			return &Duration{
				Type:      DurationUntilCondition,
				Condition: "lose_game",
			}
		}
	}

	// Check for "this turn" (one-shot within turn)
	if strings.Contains(textLower, "this turn") {
		return &Duration{
			Type: DurationUntilEOT,
		}
	}

	// Check for permanent effects (continuous abilities)
	if strings.Contains(textLower, "have") || strings.Contains(textLower, "has") {
		// "Players have", "Leader players have", etc.
		return &Duration{
			Type: DurationPermanent,
		}
	}

	// Default to one-shot
	return &Duration{
		Type: DurationOneShot,
	}
}
