package ability

import "testing"

func TestUndercoverCardsShareAdvisoryOnlyUnveilRequirement(t *testing.T) {
	supplier := GetUnveilRequirements(17)
	warlock := GetUnveilRequirements(18)

	if supplier != warlock {
		t.Fatal("Supplier and Warlock should share one Undercover requirement entry")
	}
	if supplier.AdvisoryType != UndercoverUnveilAdvisory {
		t.Fatalf("Undercover advisory type = %q, want %q", supplier.AdvisoryType, UndercoverUnveilAdvisory)
	}
	if supplier.InputType != NoInput {
		t.Fatalf("Undercover input type = %q, want advisory-only %q", supplier.InputType, NoInput)
	}
	if RequiresInput(17) || RequiresInput(18) {
		t.Fatal("Undercover requirements must never block unveiling on app-observed state")
	}
}
