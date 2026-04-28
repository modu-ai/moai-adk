package config

import (
	"testing"
	"time"
)

func TestValueUnwrap(t *testing.T) {
	provenance := Provenance{
		Source: SrcUser,
		Origin: "/path/to/config.yaml",
		Loaded: time.Now(),
	}

	val := NewValue(42, provenance)

	if got := val.Unwrap(); got != 42 {
		t.Errorf("Unwrap() = %v, want 42", got)
	}
}

func TestValueOrigin(t *testing.T) {
	expectedOrigin := "/path/to/config.yaml"
	provenance := Provenance{
		Source: SrcUser,
		Origin: expectedOrigin,
		Loaded: time.Now(),
	}

	val := NewValue(42, provenance)

	if got := val.Origin(); got != expectedOrigin {
		t.Errorf("Origin() = %v, want %v", got, expectedOrigin)
	}
}

func TestValueIsBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		source   Source
		expected bool
	}{
		{"builtin", SrcBuiltin, true},
		{"user", SrcUser, false},
		{"policy", SrcPolicy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := Value[int]{
				V: 42,
				P: Provenance{Source: tt.source},
			}

			if got := val.IsBuiltin(); got != tt.expected {
				t.Errorf("IsBuiltin() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValueIsDefault(t *testing.T) {
	tests := []struct {
		name     string
		source   Source
		expected bool
	}{
		{"builtin", SrcBuiltin, true},
		{"user", SrcUser, false},
		{"project", SrcProject, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := Value[int]{
				V: 42,
				P: Provenance{Source: tt.source},
			}

			if got := val.IsDefault(); got != tt.expected {
				t.Errorf("IsDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewValue(t *testing.T) {
	provenance := Provenance{
		Source: SrcProject,
		Origin: "/path/to/config.yaml",
		Loaded: time.Now(),
	}

	val := NewValue("test value", provenance)

	if val.V != "test value" {
		t.Errorf("NewValue() V = %v, want 'test value'", val.V)
	}
	if val.P.Source != provenance.Source || val.P.Origin != provenance.Origin {
		t.Error("NewValue() P does not match provided provenance")
	}
}

func TestProvenanceFields(t *testing.T) {
	now := time.Now()
	provenance := Provenance{
		Source:        SrcPolicy,
		Origin:        "/etc/moai/settings.json",
		Loaded:        now,
		SchemaVersion: 3,
		OverriddenBy:  []string{"/path/to/user/settings.json", "/path/to/project/config.yaml"},
	}

	if provenance.Source != SrcPolicy {
		t.Errorf("Source = %v, want SrcPolicy", provenance.Source)
	}
	if provenance.Origin != "/etc/moai/settings.json" {
		t.Errorf("Origin = %v, want '/etc/moai/settings.json'", provenance.Origin)
	}
	if !provenance.Loaded.Equal(now) {
		t.Error("Loaded time does not match")
	}
	if provenance.SchemaVersion != 3 {
		t.Errorf("SchemaVersion = %d, want 3", provenance.SchemaVersion)
	}
	if len(provenance.OverriddenBy) != 2 {
		t.Errorf("OverriddenBy length = %d, want 2", len(provenance.OverriddenBy))
	}
}
