package ability

// TriggerType represents when an ability triggers
// Based on pattern analysis: 80.6% use "when unveiled", 3.2% use "as unveiled"
type TriggerType string

const (
	TriggerOnUnveil    TriggerType = "on_unveil"     // 50 cards - "When X is unveiled"
	TriggerAsUnveil    TriggerType = "as_unveil"     // 2 cards - "As X is unveiled" (replacement)
	TriggerAtPhaseStep TriggerType = "at_phase_step" // 14 cards - "At the beginning of..."
	TriggerWhenever    TriggerType = "whenever"      // 5 cards - "Whenever [event]"
	TriggerActivated   TriggerType = "activated"     // 10 cards - "{Cost}: Effect"
	TriggerStatic      TriggerType = "static"        // 5 cards - Continuous effects
)

// Trigger defines when and how an ability activates
type Trigger struct {
	Type      TriggerType
	Timing    string // "when" vs "as" for replacement effects
	Event     string // For "whenever" triggers (e.g., "player loses game")
	PhaseStep string // For phase triggers (e.g., "upkeep", "end_step")
}

// EffectType represents categories of ability effects
// Based on pattern analysis: 15 distinct effect types identified
type EffectType string

const (
	EffectDamage        EffectType = "damage"         // 6 cards
	EffectExile         EffectType = "exile"          // 13 cards
	EffectDestroy       EffectType = "destroy"        // 4 cards
	EffectCounter       EffectType = "counter"        // 6 cards (spell/ability)
	EffectDraw          EffectType = "draw"           // 10 cards
	EffectLifeGainLoss  EffectType = "life_gain_loss" // 18 cards
	EffectCreateToken   EffectType = "create_token"   // 17 cards
	EffectCopy          EffectType = "copy"           // 5 cards
	EffectControlChange EffectType = "control_change" // 6 cards
	EffectStateChange   EffectType = "state_change"   // 5 cards (face up/down)
	EffectReveal        EffectType = "reveal"         // 7 cards
	EffectSearch        EffectType = "search"         // 3 cards
	EffectReanimate     EffectType = "reanimate"      // 4 cards
	EffectTapUntap      EffectType = "tap_untap"      // 7 cards
	EffectTransform     EffectType = "transform"      // 2 cards (Wearer, Mirror)
)

// ManaCost represents mana payment requirements
// Supports {X}, {Y}, and generic mana with optional cost reduction
type ManaCost struct {
	Generic   int
	X         bool
	Y         bool
	Reduction *CostReduction // For cards like The Quellmaster
}

// CostReduction represents conditional cost reduction
type CostReduction struct {
	Amount    int
	Condition string // e.g., "per creature"
}

// Cost represents all possible cost types
// Analysis shows: 35 fixed mana, 10 with {X}, 3 life payment, 3 discard, etc.
type Cost struct {
	Mana      *ManaCost
	Life      int
	Discard   int
	Sacrifice *SacrificeSpec
	Exile     *ExileSpec
}

// SacrificeSpec defines what must be sacrificed
type SacrificeSpec struct {
	Count   int
	Filters []Filter
}

// ExileSpec defines what must be exiled
type ExileSpec struct {
	Count   int
	Filters []Filter
	Zone    Zone
}

// TargetType represents what can be targeted
type TargetType string

const (
	TargetPlayer      TargetType = "player"
	TargetPermanent   TargetType = "permanent"
	TargetSpell       TargetType = "spell"
	TargetCardInZone  TargetType = "card_in_zone"
	TargetAbility     TargetType = "ability"
)

// TargetSpec defines targeting requirements
// Analysis: 25 cards target players, 12 use "up to X", 18 target permanents
type TargetSpec struct {
	Type     TargetType
	Count    CountSpec
	Filters  []Filter
	Zone     Zone // For card targets
	Optional bool
}

// CountSpec defines how many targets
type CountSpec struct {
	Min   int  // 0 for "up to"
	Max   int  // -1 for "any number"
	UsesX bool // For "up to X" patterns
}

// FilterType categorizes target restrictions
type FilterType string

const (
	FilterRole           FilterType = "role"            // "Leader", "Traitor", etc.
	FilterPermanentType  FilterType = "permanent_type"  // "creature", "land", etc.
	FilterController     FilterType = "controller"      // "opponent", "you"
	FilterSubtype        FilterType = "subtype"
	FilterCardType       FilterType = "card_type"
)

// Filter represents a targeting restriction
// Commonly used: "non-Leader" (5x), "nonland" (12x), "nontoken" (4x)
type Filter struct {
	Type   FilterType
	Value  string
	Negate bool // For "non-" filters
}

// Zone represents game zones
type Zone string

const (
	ZoneBattlefield  Zone = "battlefield"
	ZoneGraveyard    Zone = "graveyard"
	ZoneHand         Zone = "hand"
	ZoneLibrary      Zone = "library"
	ZoneExile        Zone = "exile"
	ZoneCommand      Zone = "command"     // Identity cards
	ZoneStack        Zone = "stack"
	ZoneOutsideGame  Zone = "outside_game" // For Wearer of Masks
)

// Effect represents a single effect in an ability
type Effect struct {
	Type      EffectType
	Params    EffectParams
	Reflexive *ReflexiveTrigger // "When you do X, Y happens"
}

// EffectParams is an interface for effect-specific parameters
type EffectParams interface {
	EffectType() EffectType
}

// TransformParams defines transformation effect parameters
// Used by The Wearer of Masks and similar cards
type TransformParams struct {
	SourceZone     Zone
	Filter         []Filter
	RevealCount    int
	RequiresChoice bool
	KeepTypes      []string // e.g., ["Traitor"]
	CopyAbilities  bool
}

func (p *TransformParams) EffectType() EffectType {
	return EffectTransform
}

// DamageParams defines damage effect parameters
type DamageParams struct {
	Amount     int
	Targets    []string // Player IDs or permanent IDs
	Divided    bool     // "divided as you choose"
}

func (p *DamageParams) EffectType() EffectType {
	return EffectDamage
}

// ReflexiveTrigger represents triggers that happen based on previous effects
// Pattern: "When one or more cards are milled this way, ..."
type ReflexiveTrigger struct {
	Condition string
	Effect    *Effect
}

// DurationType defines how long effects last
// Analysis: 35 one-shot, 12 "until end of turn", 5 "until condition"
type DurationType string

const (
	DurationOneShot        DurationType = "one_shot"
	DurationUntilEOT       DurationType = "until_eot"
	DurationUntilCondition DurationType = "until_condition"
	DurationPermanent      DurationType = "permanent"
)

// Duration defines how long an effect lasts
type Duration struct {
	Type      DurationType
	EndPhase  string // "end_of_turn", "next_upkeep"
	Condition string // "face_down", "lose_game"
}

// ConditionType represents conditional execution patterns
// Analysis: 22 "you may", 8 "if you do", 15 "up to X"
type ConditionType string

const (
	ConditionYouMay     ConditionType = "you_may"
	ConditionIfYouDo    ConditionType = "if_you_do"
	ConditionIfXOrMore  ConditionType = "if_x_or_more"
	ConditionForEach    ConditionType = "for_each"
	ConditionIfYourTurn ConditionType = "if_your_turn"
)

// Condition represents conditional execution requirements
type Condition struct {
	Type     ConditionType
	Variable string // Variable to check (e.g., "X" for "if X is 5 or more")
}

// Ability represents a complete ability definition
// Combines all components: trigger, cost, target, effects, duration, conditions
type Ability struct {
	CardID         int
	Trigger        Trigger
	Cost           *Cost
	Target         *TargetSpec
	Effects        []Effect
	Duration       *Duration
	Conditions     []Condition
	CustomResolver AbilityResolver // For complex cards requiring custom logic
}

// AbilityResolver interface for cards that don't fit pattern library
// Analysis shows ~10 cards (16%) require fully custom implementations
type AbilityResolver interface {
	// Called when ability triggers
	OnTrigger(ctx *AbilityContext) (*PendingAbility, error)

	// Called when player makes a choice
	OnChoice(ctx *AbilityContext, choice interface{}) error

	// Called to check if ability can activate
	CanActivate(ctx *AbilityContext) bool

	// For continuous effects
	ApplyEffect(ctx *AbilityContext) error
	RemoveEffect(ctx *AbilityContext) error
}

// AbilityContext provides game state for ability resolution
type AbilityContext struct {
	RoomCode  string
	PlayerID  string
	CardID    int
	TempData  map[string]interface{} // Mutable state for multi-step abilities
}

// PendingAbility represents an ability awaiting player choice
type PendingAbility struct {
	ID             string
	PlayerID       string
	CardID         int
	AbilityType    string
	Data           map[string]interface{}
	ModalDismissed bool
	ModalState     map[string]interface{}
}
