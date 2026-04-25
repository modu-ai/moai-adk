package constitution_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestZoneEnumValuesExactlyTwo는 Zone 타입이 정확히 2개 값만 가짐을 검증한다.
// AC-CON-001-004 매핑.
func TestZoneEnumValuesExactlyTwo(t *testing.T) {
	t.Parallel()

	// ZoneFrozen과 ZoneEvolvable이 정의되어 있어야 한다.
	if constitution.ZoneFrozen == constitution.ZoneEvolvable {
		t.Fatal("ZoneFrozen과 ZoneEvolvable이 같은 값을 가진다")
	}
}

// TestZoneFrozenIsZero는 ZoneFrozen이 iota=0임을 검증한다.
func TestZoneFrozenIsZero(t *testing.T) {
	t.Parallel()

	if constitution.ZoneFrozen != 0 {
		t.Errorf("ZoneFrozen = %d, 0이어야 한다", constitution.ZoneFrozen)
	}
}

// TestZoneEvolvableIsOne은 ZoneEvolvable이 iota=1임을 검증한다.
func TestZoneEvolvableIsOne(t *testing.T) {
	t.Parallel()

	if constitution.ZoneEvolvable != 1 {
		t.Errorf("ZoneEvolvable = %d, 1이어야 한다", constitution.ZoneEvolvable)
	}
}

// TestZoneString은 Zone.String() 메서드의 출력을 검증한다.
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

// TestParseZoneValidInputs는 유효한 입력에 대한 ParseZone을 검증한다.
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
				t.Fatalf("ParseZone(%q) 오류: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseZone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseZoneInvalidInputs는 알 수 없는 값에 대한 ParseZone 오류 반환을 검증한다.
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
				t.Errorf("ParseZone(%q) 오류를 반환해야 하지만 nil을 반환했다", input)
			}
		})
	}
}
