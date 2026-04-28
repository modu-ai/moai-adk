package config

import (
	"testing"
	"time"
)

func TestNewMergedSettings(t *testing.T) {
	ms := NewMergedSettings()

	if ms == nil {
		t.Fatal("NewMergedSettings() returned nil")
	}

	if ms.values == nil {
		t.Error("NewMergedSettings() values map is nil")
	}
}

func TestMergedSettingsGetSet(t *testing.T) {
	ms := NewMergedSettings()
	key := "test.field"
	expected := Value[any]{
		V: 42,
		P: Provenance{
			Source: SrcUser,
			Origin: "/path/to/config.yaml",
			Loaded: time.Now(),
		},
	}

	ms.Set(key, expected)

	got, ok := ms.Get(key)
	if !ok {
		t.Fatal("Get() returned ok=false for existing key")
	}

	if got.V != expected.V {
		t.Errorf("Get() V = %v, want %v", got.V, expected.V)
	}
	if got.P.Source != expected.P.Source {
		t.Errorf("Get() P.Source = %v, want %v", got.P.Source, expected.P.Source)
	}
}

func TestMergedSettingsGetMissing(t *testing.T) {
	ms := NewMergedSettings()

	_, ok := ms.Get("nonexistent")
	if ok {
		t.Error("Get() returned ok=true for missing key")
	}
}

func TestMergedSettingsKeys(t *testing.T) {
	ms := NewMergedSettings()
	keys := []string{"key1", "key2", "key3"}

	for _, key := range keys {
		ms.Set(key, Value[any]{V: key, P: Provenance{Source: SrcBuiltin}})
	}

	got := ms.Keys()
	if len(got) != len(keys) {
		t.Errorf("Keys() returned %d keys, want %d", len(got), len(keys))
	}
}

func TestMergedSettingsAll(t *testing.T) {
	ms := NewMergedSettings()
	expected := map[string]Value[any]{
		"key1": {V: 1, P: Provenance{Source: SrcBuiltin}},
		"key2": {V: 2, P: Provenance{Source: SrcUser}},
	}

	for k, v := range expected {
		ms.Set(k, v)
	}

	got := ms.All()
	if len(got) != len(expected) {
		t.Errorf("All() returned %d items, want %d", len(got), len(expected))
	}
}

func TestMergeAllPriorityOrder(t *testing.T) {
	// Test that higher priority sources override lower priority ones
	// This maps to REQ-V3R2-RT-005-005
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcUser:    {"test.key": "user_value"},
		SrcProject: {"test.key": "project_value"},
		SrcBuiltin: {"test.key": "builtin_value"},
	}

	origins := map[Source]string{
		SrcUser:    "~/.moai/settings.json",
		SrcProject: ".moai/config/config.yaml",
		SrcBuiltin: "internal/config/defaults.go",
	}

	merged, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("MergeAll() error = %v", err)
	}

	val, ok := merged.Get("test.key")
	if !ok {
		t.Fatal("MergeAll() did not set expected key")
	}

	// SrcUser (priority 1) should win over SrcProject (2) and SrcBuiltin (7)
	if val.P.Source != SrcUser {
		t.Errorf("MergeAll() source = %v, want SrcUser", val.P.Source)
	}
	if val.V != "user_value" {
		t.Errorf("MergeAll() value = %v, want 'user_value'", val.V)
	}

	// Check that overridden tiers are tracked
	if len(val.P.OverriddenBy) != 2 {
		t.Errorf("MergeAll() OverriddenBy length = %d, want 2", len(val.P.OverriddenBy))
	}
}

func TestMergeAllZeroValues(t *testing.T) {
	// Test that zero values are skipped
	// This maps to the isZero function behavior
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcUser:    {"test.key": ""},
		SrcProject: {"test.key": "project_value"},
	}

	origins := map[Source]string{
		SrcUser:    "~/.moai/settings.json",
		SrcProject: ".moai/config/config.yaml",
	}

	merged, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("MergeAll() error = %v", err)
	}

	val, ok := merged.Get("test.key")
	if !ok {
		t.Fatal("MergeAll() did not set expected key")
	}

	// Empty string should be skipped, project value should win
	if val.P.Source != SrcProject {
		t.Errorf("MergeAll() source = %v, want SrcProject", val.P.Source)
	}
}

func TestMergeAllEmptyTiers(t *testing.T) {
	// Test with no tiers
	merged, err := MergeAll(map[Source]map[string]any{}, map[Source]string{}, time.Now())
	if err != nil {
		t.Fatalf("MergeAll() error = %v", err)
	}

	if merged == nil {
		t.Fatal("MergeAll() returned nil")
	}

	keys := merged.Keys()
	if len(keys) != 0 {
		t.Errorf("MergeAll() returned %d keys, want 0", len(keys))
	}
}

func TestDiff(t *testing.T) {
	// Test diff between two MergedSettings
	now := time.Now()

	msA := NewMergedSettings()
	msA.Set("same", Value[any]{V: 1, P: Provenance{Source: SrcUser, Loaded: now}})
	msA.Set("only_a", Value[any]{V: 2, P: Provenance{Source: SrcUser, Loaded: now}})
	msA.Set("different", Value[any]{V: 3, P: Provenance{Source: SrcUser, Loaded: now}})

	msB := NewMergedSettings()
	msB.Set("same", Value[any]{V: 1, P: Provenance{Source: SrcProject, Loaded: now}})
	msB.Set("only_b", Value[any]{V: 4, P: Provenance{Source: SrcProject, Loaded: now}})
	msB.Set("different", Value[any]{V: 5, P: Provenance{Source: SrcProject, Loaded: now}})

	diff := Diff(msA, msB)

	// Should have 3 differences: only_a, only_b, different
	if len(diff) != 3 {
		t.Errorf("Diff() returned %d differences, want 3", len(diff))
	}

	// Check that "same" is not in the diff (values are equal)
	if _, ok := diff["same"]; ok {
		t.Error("Diff() included 'same' key which has equal values")
	}

	// Check "different" key
	if diffA, ok := diff["different"]; ok {
		if diffA.A.V != 3 || diffA.B.V != 5 {
			t.Errorf("Diff() 'different' values = (%v, %v), want (3, 5)", diffA.A.V, diffA.B.V)
		}
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"nil", nil, true},
		{"zero_int", 0, true},
		{"zero_float", 0.0, true},
		{"zero_bool", false, true},
		{"empty_string", "", true},
		{"non_zero_int", 1, false},
		{"non_zero_float", 1.0, false},
		{"true", true, false},
		{"non_empty_string", "test", false},
		{"empty_slice", []int{}, true},
		{"nil_slice", ([]int)(nil), true},
		{"empty_map", map[string]int{}, true},
		{"nil_map", (map[string]int)(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isZero(tt.value); got != tt.expected {
				t.Errorf("isZero(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}
