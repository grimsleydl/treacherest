package game

import "testing"

func TestNormalizeCoupGreenHuntRequirement(t *testing.T) {
	tests := []struct {
		name string
		in   CoupGreenHuntRequirement
		want CoupGreenHuntRequirement
	}{
		{name: "empty defaults to one blue", in: "", want: CoupGreenHuntOneBlue},
		{name: "one blue remains one blue", in: CoupGreenHuntOneBlue, want: CoupGreenHuntOneBlue},
		{name: "all blues remains all blues", in: CoupGreenHuntAllBlues, want: CoupGreenHuntAllBlues},
		{name: "unknown defaults to one blue", in: "blue-exposed", want: CoupGreenHuntOneBlue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCoupGreenHuntRequirement(tt.in); got != tt.want {
				t.Fatalf("NormalizeCoupGreenHuntRequirement(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeCoupInquisitionAmnesty(t *testing.T) {
	tests := []struct {
		name string
		in   CoupInquisitionAmnesty
		want CoupInquisitionAmnesty
	}{
		{name: "empty defaults to king side only", in: "", want: CoupInquisitionAmnestyKingSideOnly},
		{name: "king side only remains king side only", in: CoupInquisitionAmnestyKingSideOnly, want: CoupInquisitionAmnestyKingSideOnly},
		{name: "broad remains broad", in: CoupInquisitionAmnestyBroad, want: CoupInquisitionAmnestyBroad},
		{name: "unknown defaults to king side only", in: "mandate", want: CoupInquisitionAmnestyKingSideOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCoupInquisitionAmnesty(tt.in); got != tt.want {
				t.Fatalf("NormalizeCoupInquisitionAmnesty(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
