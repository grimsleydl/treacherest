package ability

import (
	"testing"
)

// TestTriggerTypes tests trigger type constants
func TestTriggerTypes(t *testing.T) {
	tests := []struct {
		name     string
		trigger  TriggerType
		expected string
	}{
		{"on unveil", TriggerOnUnveil, "on_unveil"},
		{"as unveil", TriggerAsUnveil, "as_unveil"},
		{"at phase", TriggerAtPhaseStep, "at_phase_step"},
		{"whenever", TriggerWhenever, "whenever"},
		{"activated", TriggerActivated, "activated"},
		{"static", TriggerStatic, "static"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.trigger) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.trigger))
			}
		})
	}
}

// TestEffectTypes tests effect type constants
func TestEffectTypes(t *testing.T) {
	effects := []EffectType{
		EffectDamage,
		EffectExile,
		EffectDestroy,
		EffectCounter,
		EffectDraw,
		EffectLifeGainLoss,
		EffectCreateToken,
		EffectCopy,
		EffectControlChange,
		EffectStateChange,
		EffectReveal,
		EffectSearch,
		EffectReanimate,
		EffectTapUntap,
		EffectTransform,
	}

	// Ensure each effect type is unique
	seen := make(map[EffectType]bool)
	for _, effect := range effects {
		if seen[effect] {
			t.Errorf("Duplicate effect type: %s", effect)
		}
		seen[effect] = true
	}

	if len(effects) != 15 {
		t.Errorf("Expected 15 effect types, got %d", len(effects))
	}
}

// TestManaCost tests mana cost structure
func TestManaCost(t *testing.T) {
	tests := []struct {
		name    string
		cost    *ManaCost
		hasX    bool
		hasY    bool
		generic int
	}{
		{
			name:    "simple generic",
			cost:    &ManaCost{Generic: 3, X: false, Y: false},
			hasX:    false,
			hasY:    false,
			generic: 3,
		},
		{
			name:    "with X",
			cost:    &ManaCost{Generic: 0, X: true, Y: false},
			hasX:    true,
			hasY:    false,
			generic: 0,
		},
		{
			name:    "with X and Y",
			cost:    &ManaCost{Generic: 0, X: true, Y: true},
			hasX:    true,
			hasY:    true,
			generic: 0,
		},
		{
			name:    "X plus generic",
			cost:    &ManaCost{Generic: 4, X: true, Y: false},
			hasX:    true,
			hasY:    false,
			generic: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cost.X != tt.hasX {
				t.Errorf("Expected X=%v, got X=%v", tt.hasX, tt.cost.X)
			}
			if tt.cost.Y != tt.hasY {
				t.Errorf("Expected Y=%v, got Y=%v", tt.hasY, tt.cost.Y)
			}
			if tt.cost.Generic != tt.generic {
				t.Errorf("Expected Generic=%d, got Generic=%d", tt.generic, tt.cost.Generic)
			}
		})
	}
}

// TestTargetSpec tests target specification
func TestTargetSpec(t *testing.T) {
	tests := []struct {
		name      string
		target    *TargetSpec
		wantType  TargetType
		wantCount CountSpec
	}{
		{
			name: "target player",
			target: &TargetSpec{
				Type:  TargetPlayer,
				Count: CountSpec{Min: 1, Max: 1, UsesX: false},
			},
			wantType:  TargetPlayer,
			wantCount: CountSpec{Min: 1, Max: 1, UsesX: false},
		},
		{
			name: "up to X players",
			target: &TargetSpec{
				Type:  TargetPlayer,
				Count: CountSpec{Min: 0, Max: -1, UsesX: true},
			},
			wantType:  TargetPlayer,
			wantCount: CountSpec{Min: 0, Max: -1, UsesX: true},
		},
		{
			name: "any number of permanents",
			target: &TargetSpec{
				Type:  TargetPermanent,
				Count: CountSpec{Min: 0, Max: -1, UsesX: false},
			},
			wantType:  TargetPermanent,
			wantCount: CountSpec{Min: 0, Max: -1, UsesX: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.target.Type != tt.wantType {
				t.Errorf("Expected Type=%v, got Type=%v", tt.wantType, tt.target.Type)
			}
			if tt.target.Count != tt.wantCount {
				t.Errorf("Expected Count=%+v, got Count=%+v", tt.wantCount, tt.target.Count)
			}
		})
	}
}

// TestFilter tests filter structure
func TestFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		want   Filter
	}{
		{
			name:   "non-Leader filter",
			filter: Filter{Type: FilterRole, Value: "Leader", Negate: true},
			want:   Filter{Type: FilterRole, Value: "Leader", Negate: true},
		},
		{
			name:   "nonland filter",
			filter: Filter{Type: FilterPermanentType, Value: "land", Negate: true},
			want:   Filter{Type: FilterPermanentType, Value: "land", Negate: true},
		},
		{
			name:   "opponent filter",
			filter: Filter{Type: FilterController, Value: "opponent", Negate: false},
			want:   Filter{Type: FilterController, Value: "opponent", Negate: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filter != tt.want {
				t.Errorf("Expected %+v, got %+v", tt.want, tt.filter)
			}
		})
	}
}

// TestDuration tests duration types
func TestDuration(t *testing.T) {
	tests := []struct {
		name         string
		duration     *Duration
		wantType     DurationType
		wantEndPhase string
		wantCond     string
	}{
		{
			name:         "one shot",
			duration:     &Duration{Type: DurationOneShot},
			wantType:     DurationOneShot,
			wantEndPhase: "",
			wantCond:     "",
		},
		{
			name:         "until end of turn",
			duration:     &Duration{Type: DurationUntilEOT, EndPhase: "end_of_turn"},
			wantType:     DurationUntilEOT,
			wantEndPhase: "end_of_turn",
			wantCond:     "",
		},
		{
			name:         "until face down",
			duration:     &Duration{Type: DurationUntilCondition, Condition: "face_down"},
			wantType:     DurationUntilCondition,
			wantEndPhase: "",
			wantCond:     "face_down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.duration.Type != tt.wantType {
				t.Errorf("Expected Type=%v, got Type=%v", tt.wantType, tt.duration.Type)
			}
			if tt.duration.EndPhase != tt.wantEndPhase {
				t.Errorf("Expected EndPhase=%s, got EndPhase=%s", tt.wantEndPhase, tt.duration.EndPhase)
			}
			if tt.duration.Condition != tt.wantCond {
				t.Errorf("Expected Condition=%s, got Condition=%s", tt.wantCond, tt.duration.Condition)
			}
		})
	}
}

// TestCondition tests condition types
func TestCondition(t *testing.T) {
	conditions := []ConditionType{
		ConditionYouMay,
		ConditionIfYouDo,
		ConditionIfXOrMore,
		ConditionForEach,
		ConditionIfYourTurn,
	}

	// Ensure each condition type is unique
	seen := make(map[ConditionType]bool)
	for _, cond := range conditions {
		if seen[cond] {
			t.Errorf("Duplicate condition type: %s", cond)
		}
		seen[cond] = true
	}

	if len(conditions) != 5 {
		t.Errorf("Expected 5 condition types, got %d", len(conditions))
	}
}

// TestTransformParams tests transform effect parameters
func TestTransformParams(t *testing.T) {
	params := &TransformParams{
		SourceZone:     ZoneOutsideGame,
		Filter:         []Filter{{Type: FilterRole, Value: "Leader", Negate: true}},
		RevealCount:    3,
		RequiresChoice: true,
		KeepTypes:      []string{"Traitor"},
		CopyAbilities:  true,
	}

	if params.EffectType() != EffectTransform {
		t.Errorf("Expected EffectTransform, got %v", params.EffectType())
	}
	if params.SourceZone != ZoneOutsideGame {
		t.Errorf("Expected ZoneOutsideGame, got %v", params.SourceZone)
	}
	if len(params.Filter) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(params.Filter))
	}
	if params.RevealCount != 3 {
		t.Errorf("Expected RevealCount=3, got %d", params.RevealCount)
	}
	if !params.RequiresChoice {
		t.Error("Expected RequiresChoice=true")
	}
	if len(params.KeepTypes) != 1 || params.KeepTypes[0] != "Traitor" {
		t.Errorf("Expected KeepTypes=[Traitor], got %v", params.KeepTypes)
	}
	if !params.CopyAbilities {
		t.Error("Expected CopyAbilities=true")
	}
}

// TestAbility tests complete ability structure
func TestAbility(t *testing.T) {
	ability := &Ability{
		CardID: 31, // The Wearer of Masks
		Trigger: Trigger{
			Type:   TriggerAsUnveil,
			Timing: "as",
		},
		Cost: &Cost{
			Mana: &ManaCost{X: true, Generic: 0},
		},
		Target: &TargetSpec{
			Type:  TargetCardInZone,
			Count: CountSpec{Min: 1, Max: 1, UsesX: false},
		},
		Effects: []Effect{
			{
				Type: EffectTransform,
				Params: &TransformParams{
					SourceZone:     ZoneOutsideGame,
					RevealCount:    3,
					RequiresChoice: true,
					KeepTypes:      []string{"Traitor"},
				},
			},
		},
		Duration: &Duration{
			Type:      DurationUntilCondition,
			Condition: "face_down",
		},
		Conditions: []Condition{
			{Type: ConditionYouMay},
		},
	}

	if ability.CardID != 31 {
		t.Errorf("Expected CardID=31, got %d", ability.CardID)
	}
	if ability.Trigger.Type != TriggerAsUnveil {
		t.Errorf("Expected TriggerAsUnveil, got %v", ability.Trigger.Type)
	}
	if ability.Cost.Mana.X != true {
		t.Error("Expected X cost")
	}
	if len(ability.Effects) != 1 {
		t.Errorf("Expected 1 effect, got %d", len(ability.Effects))
	}
	if ability.Effects[0].Type != EffectTransform {
		t.Errorf("Expected EffectTransform, got %v", ability.Effects[0].Type)
	}
	if ability.Duration.Type != DurationUntilCondition {
		t.Errorf("Expected DurationUntilCondition, got %v", ability.Duration.Type)
	}
	if len(ability.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(ability.Conditions))
	}
}
