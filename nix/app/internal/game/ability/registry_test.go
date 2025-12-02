package ability

import (
	"testing"
)

// testCard is a simple implementation of CardLike for testing
type testCard struct {
	id   int
	text string
}

func (tc *testCard) GetID() int {
	return tc.id
}

func (tc *testCard) GetText() string {
	return tc.text
}

// TestNewAbilityRegistry tests registry creation
func TestNewAbilityRegistry(t *testing.T) {
	cards := []CardLike{
		&testCard{id: 1, text: "When Test Card 1 is unveiled, draw a card."},
		&testCard{id: 2, text: "When Test Card 2 is unveiled, deal 3 damage."},
	}

	registry := NewAbilityRegistry(cards)

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	if len(registry.abilities) == 0 {
		t.Error("Expected registry to have abilities registered")
	}
}

// TestRegisterAbility tests manual ability registration
func TestRegisterAbility(t *testing.T) {
	registry := &AbilityRegistry{
		abilities: make(map[int]*Ability),
	}

	ability := &Ability{
		CardID: 31,
		Trigger: Trigger{
			Type: TriggerAsUnveil,
		},
	}

	registry.Register(31, ability)

	retrieved := registry.Get(31)
	if retrieved == nil {
		t.Fatal("Expected to retrieve registered ability")
	}

	if retrieved.CardID != 31 {
		t.Errorf("Expected CardID=31, got %d", retrieved.CardID)
	}
}

// TestGetAbility tests ability lookup
func TestGetAbility(t *testing.T) {
	registry := &AbilityRegistry{
		abilities: make(map[int]*Ability),
	}

	ability := &Ability{CardID: 5}
	registry.abilities[5] = ability

	retrieved := registry.Get(5)
	if retrieved == nil {
		t.Fatal("Expected to retrieve ability")
	}

	if retrieved.CardID != 5 {
		t.Errorf("Expected CardID=5, got %d", retrieved.CardID)
	}

	// Test non-existent card
	notFound := registry.Get(999)
	if notFound != nil {
		t.Error("Expected nil for non-existent ability")
	}
}

// TestHasAbility tests ability existence check
func TestHasAbility(t *testing.T) {
	registry := &AbilityRegistry{
		abilities: make(map[int]*Ability),
	}

	registry.abilities[10] = &Ability{CardID: 10}

	if !registry.Has(10) {
		t.Error("Expected Has(10) to return true")
	}

	if registry.Has(999) {
		t.Error("Expected Has(999) to return false")
	}
}

// TestDetectTriggerType tests trigger type detection from card text
func TestDetectTriggerType(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected TriggerType
	}{
		{
			name:     "when unveiled",
			text:     "When The Bodyguard is unveiled, prevent all damage.",
			expected: TriggerOnUnveil,
		},
		{
			name:     "as unveiled replacement",
			text:     "As The Wearer of Masks is unveiled, reveal cards.",
			expected: TriggerAsUnveil,
		},
		{
			name:     "at beginning phase",
			text:     "At the beginning of your upkeep, draw a card.",
			expected: TriggerAtPhaseStep,
		},
		{
			name:     "whenever trigger",
			text:     "Whenever a creature dies, gain 1 life.",
			expected: TriggerWhenever,
		},
		{
			name:     "activated ability",
			text:     "{3}: Destroy target permanent.",
			expected: TriggerActivated,
		},
		{
			name:     "no trigger static",
			text:     "Leader players have hexproof.",
			expected: TriggerStatic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &testCard{text: tt.text}
			triggerType := detectTriggerType(card)
			if triggerType != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, triggerType)
			}
		})
	}
}

// TestParseUnveilCost tests unveil cost parsing
func TestParseUnveilCost(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "simple X cost",
			text:     "Unveil {X}",
			expected: "{X}",
		},
		{
			name:     "fixed cost",
			text:     "Unveil {3}",
			expected: "{3}",
		},
		{
			name:     "X plus generic",
			text:     "Unveil {X}{4}",
			expected: "{X}{4}",
		},
		{
			name:     "life payment",
			text:     "Unveil—Pay 8 life, Discard a card.",
			expected: "Pay 8 life, Discard a card",
		},
		{
			name:     "no unveil cost",
			text:     "When this is unveiled, draw a card.",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &testCard{text: tt.text}
			cost := parseUnveilCost(card)
			if cost != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, cost)
			}
		})
	}
}

// TestDetectEffectType tests effect type detection
func TestDetectEffectType(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []EffectType
	}{
		{
			name:     "damage effect",
			text:     "deals 6 damage to target player",
			expected: []EffectType{EffectDamage},
		},
		{
			name:     "draw effect",
			text:     "draw two cards",
			expected: []EffectType{EffectDraw},
		},
		{
			name:     "exile effect",
			text:     "exile target creature",
			expected: []EffectType{EffectExile},
		},
		{
			name:     "transform effect",
			text:     "becomes a copy of that card",
			expected: []EffectType{EffectTransform},
		},
		{
			name:     "multiple effects",
			text:     "draw a card, then deal 2 damage",
			expected: []EffectType{EffectDamage, EffectDraw}, // Detection order matches implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := detectEffectTypes(tt.text)
			if len(effects) != len(tt.expected) {
				t.Errorf("Expected %d effects, got %d", len(tt.expected), len(effects))
			}
			for i, expected := range tt.expected {
				if i < len(effects) && effects[i] != expected {
					t.Errorf("Expected effect %d to be %s, got %s", i, expected, effects[i])
				}
			}
		})
	}
}

// TestPatternDetection tests complete pattern detection
func TestPatternDetection(t *testing.T) {
	tests := []struct {
		name          string
		card          CardLike
		expectAbility bool
	}{
		{
			name: "simple unveil ability",
			card: &testCard{
				id:   1,
				text: "When Simple Card is unveiled, draw a card.",
			},
			expectAbility: true,
		},
		{
			name: "complex custom ability",
			card: &testCard{
				id:   31,
				text: "As The Wearer of Masks is unveiled, reveal up to X non-Leader identity cards...",
			},
			expectAbility: true,
		},
		{
			name: "no ability text",
			card: &testCard{
				id:   100,
				text: "",
			},
			expectAbility: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ability := DetectAbilityPattern(tt.card)
			hasAbility := ability != nil
			if hasAbility != tt.expectAbility {
				t.Errorf("Expected ability=%v, got ability=%v", tt.expectAbility, hasAbility)
			}
		})
	}
}

// TestRegistryCount tests ability count
func TestRegistryCount(t *testing.T) {
	registry := &AbilityRegistry{
		abilities: make(map[int]*Ability),
	}

	registry.abilities[1] = &Ability{CardID: 1}
	registry.abilities[2] = &Ability{CardID: 2}
	registry.abilities[3] = &Ability{CardID: 3}

	count := registry.Count()
	if count != 3 {
		t.Errorf("Expected count=3, got %d", count)
	}
}

// TestRegistryListAll tests listing all abilities
func TestRegistryListAll(t *testing.T) {
	registry := &AbilityRegistry{
		abilities: make(map[int]*Ability),
	}

	registry.abilities[1] = &Ability{CardID: 1}
	registry.abilities[2] = &Ability{CardID: 2}

	all := registry.ListAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 abilities, got %d", len(all))
	}

	// Check that all abilities are present
	found := make(map[int]bool)
	for _, ability := range all {
		found[ability.CardID] = true
	}

	if !found[1] || !found[2] {
		t.Error("Expected both abilities to be in list")
	}
}
