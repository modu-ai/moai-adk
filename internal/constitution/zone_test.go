package constitution_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestZoneEnumValuesExactlyTwo verifies that the Zone type has exactly 2 values.
// AC-CON-001-004 mapping.
func TestZoneEnumValuesExactlyTwo(t *testing.T) {
	t.Parallel()

	// ZoneFrozen and ZoneEvolvable must be defined.
	if constitution.ZoneFrozen == constitution.ZoneEvolvable {
		t.Fatal("ZoneFrozen and ZoneEvolvable have the same value")
	}
}

// TestZoneFrozenIsZero verifies that ZoneFrozen is iota=0.
func TestZoneFrozenIsZero(t *testing.T) {
	t.Parallel()

	if constitution.ZoneFrozen != 0 {
		t.Errorf("ZoneFrozen = %d, must be 0", constitution.ZoneFrozen)
	}
}

// TestZoneEvolvableIsOne verifies that ZoneEvolvable is iota=1.
func TestZoneEvolvableIsOne(t *testing.T) {
	t.Parallel()

	if constitution.ZoneEvolvable != 1 {
		t.Errorf("ZoneEvolvable = %d, must be 1", constitution.ZoneEvolvable)
	}
}

// TestZoneString verifies the output of the Zone.String() method.
func TestZoneString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		zone constitution.Zone
		want string
	}{
		{constitution.ZoneFrozen, "Frozen"},
		{constitution.ZoneEvolvable, "Evolvable"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := tt.zone.String()
			if got != tt.want {
				t.Errorf("Zone(%d).String() = %q, want %q", tt.zone, got, tt.want)
			}
		})
	}
}

// TestParseZoneValidInputs verifies ParseZone for valid inputs.
func TestParseZoneValidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  constitution.Zone
	}{
		{"Frozen", constitution.ZoneFrozen},
		{"frozen", constitution.ZoneFrozen},
		{"FROZEN", constitution.ZoneFrozen},
		{"Evolvable", constitution.ZoneEvolvable},
		{"evolvable", constitution.ZoneEvolvable},
		{"EVOLVABLE", constitution.ZoneEvolvable},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := constitution.ParseZone(tt.input)
			if err != nil {
				t.Fatalf("ParseZone(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseZone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseZoneInvalidInputs verifies that ParseZone returns an error for unknown values.
func TestParseZoneInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"unknown",
		"Tentative",
		"frozen_with_extra",
	}

	for _, input := range tests {
		t.Run("invalid_"+input, func(t *testing.T) {
			t.Parallel()
			_, err := constitution.ParseZone(input)
			if err == nil {
				t.Errorf("ParseZone(%q) must return an error but returned nil", input)
			}
		})
	}
}
