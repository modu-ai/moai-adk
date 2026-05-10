package config

import (
	"errors"
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

// TestMergeAll_PolicyOverrideRejected verifies that strict_mode in the policy tier
// causes a lower-tier override attempt to return PolicyOverrideRejected error.
//
// AC-V3R2-RT-005-07: Given policy tier sets strict_mode: true and a policy-designated key,
// When SrcProject tries to override the key, Then PolicyOverrideRejected error is returned.
//
// REQ-V3R2-RT-005-022, AC-07
func TestMergeAll_PolicyOverrideRejected(t *testing.T) {
	// Arrange
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcPolicy: {
			"policy.strict_mode":            true,
			"permission.network_allowlist":   []string{"host1"},
		},
		SrcProject: {
			"permission.network_allowlist": []string{"host1", "host2"}, // attempted override
		},
	}
	origins := map[Source]string{
		SrcPolicy:  "/etc/moai/settings.json",
		SrcProject: ".moai/config/config.yaml",
	}

	// Act
	_, err := MergeAll(tiers, origins, now)

	// Assert: PolicyOverrideRejected error expected
	if err == nil {
		t.Fatal("MergeAll() with strict_mode=true and lower-tier override should return error")
	}

	var policyErr *PolicyOverrideRejected
	if !errors.As(err, &policyErr) {
		t.Errorf("MergeAll() error type = %T, want *PolicyOverrideRejected", err)
	} else {
		if policyErr.Key == "" {
			t.Error("PolicyOverrideRejected.Key is empty, want the offending key name")
		}
		if policyErr.PolicySource == "" {
			t.Error("PolicyOverrideRejected.PolicySource is empty, want the policy file path")
		}
	}
}

// TestMergeAll_OverriddenByPopulated verifies that when multiple tiers have non-zero values,
// the lower-tier origins are collected in Provenance.OverriddenBy.
//
// AC-V3R2-RT-005-01: Given policy=true, user=false, project=true (all non-zero),
// When Load() merges, Then winner=policy, OverriddenBy=[user_path, project_path].
//
// REQ-V3R2-RT-005-012, AC-01
func TestMergeAll_OverriddenByPopulated(t *testing.T) {
	// Arrange: all 3 tiers have the same key with non-zero values
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcPolicy:  {"test.key": true},
		SrcUser:    {"test.key": false},  // false is zero for bool → should be excluded
		SrcProject: {"test.key": "project_value"},
	}
	origins := map[Source]string{
		SrcPolicy:  "/etc/moai/settings.json",
		SrcUser:    "~/.moai/settings.json",
		SrcProject: ".moai/config/config.yaml",
	}

	// Act
	merged, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("MergeAll() unexpected error: %v", err)
	}

	// Assert
	val, ok := merged.Get("test.key")
	if !ok {
		t.Fatal("MergeAll() key 'test.key' not found in result")
	}

	// Policy should win
	if val.P.Source != SrcPolicy {
		t.Errorf("winning source = %v, want SrcPolicy", val.P.Source)
	}

	// OverriddenBy should contain project (user=false is zero, excluded)
	// At minimum, project path should be in OverriddenBy
	foundProject := false
	for _, path := range val.P.OverriddenBy {
		if path == ".moai/config/config.yaml" {
			foundProject = true
		}
	}
	if !foundProject {
		t.Errorf("OverriddenBy %v does not contain project path '.moai/config/config.yaml'", val.P.OverriddenBy)
	}

	// User (false = zero) should NOT be in OverriddenBy
	for _, path := range val.P.OverriddenBy {
		if path == "~/.moai/settings.json" {
			t.Errorf("OverriddenBy contains user path %q, but user had zero value (false)", path)
		}
	}
}

// TestMergeAll_ZeroValuesExcluded verifies that zero values are excluded from OverriddenBy.
//
// AC-V3R2-RT-005-01 edge case: policy=true, user="" (zero), project=true.
// OverriddenBy should contain only [project_path], user excluded (zero value).
//
// REQ-V3R2-RT-005-005, REQ-V3R2-RT-005-012, AC-01
func TestMergeAll_ZeroValuesExcluded(t *testing.T) {
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcPolicy:  {"perm.mode": "strict"},
		SrcUser:    {"perm.mode": ""},        // empty string = zero
		SrcProject: {"perm.mode": "lenient"},
	}
	origins := map[Source]string{
		SrcPolicy:  "/etc/moai/settings.json",
		SrcUser:    "~/.moai/settings.json",
		SrcProject: ".moai/config/config.yaml",
	}

	merged, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("MergeAll() unexpected error: %v", err)
	}

	val, ok := merged.Get("perm.mode")
	if !ok {
		t.Fatal("MergeAll() key not found")
	}

	// Verify no zero-value tier appears in OverriddenBy
	for _, path := range val.P.OverriddenBy {
		if path == "~/.moai/settings.json" {
			t.Errorf("zero-value user tier should not appear in OverriddenBy, got: %v", val.P.OverriddenBy)
		}
	}
}

// TestMergeAll_ByteStableJSON verifies that identical tier inputs produce byte-identical merged output.
//
// REQ-V3R2-RT-005 §7 Constraints (Determinism), AC-02 byte-stability
func TestMergeAll_ByteStableJSON(t *testing.T) {
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcUser:    {"quality.coverage_threshold": 85, "llm.mode": "claude"},
		SrcBuiltin: {"quality.coverage_threshold": 80},
	}
	origins := map[Source]string{
		SrcUser:    "~/.moai/settings.json",
		SrcBuiltin: "internal/config/defaults.go",
	}

	// Act: produce merged settings twice
	merged1, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("first MergeAll() error: %v", err)
	}
	merged2, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("second MergeAll() error: %v", err)
	}

	// Assert: JSON output must be byte-identical
	json1, err := merged1.Dump("json")
	if err != nil {
		t.Fatalf("Dump() first call error: %v", err)
	}
	json2, err := merged2.Dump("json")
	if err != nil {
		t.Fatalf("Dump() second call error: %v", err)
	}

	if json1 != json2 {
		t.Errorf("Dump() is not byte-stable:\nfirst: %s\nsecond: %s", json1, json2)
	}
}

// TestMergeAll_StrictModeFalseAllowsOverride verifies that when strict_mode is false (or unset),
// policy-designated keys can be overridden by lower tiers (policy still wins by priority, no error).
//
// AC-V3R2-RT-005-07 edge case: strict_mode=false allows normal priority-based override without error.
//
// REQ-V3R2-RT-005-022, AC-07
func TestMergeAll_StrictModeFalseAllowsOverride(t *testing.T) {
	now := time.Now()

	tiers := map[Source]map[string]any{
		SrcPolicy: {
			"policy.strict_mode":          false, // strict mode disabled
			"permission.network_allowlist": []string{"host1"},
		},
		SrcProject: {
			"permission.network_allowlist": []string{"host1", "host2"},
		},
	}
	origins := map[Source]string{
		SrcPolicy:  "/etc/moai/settings.json",
		SrcProject: ".moai/config/config.yaml",
	}

	// Act: should NOT return PolicyOverrideRejected when strict_mode=false
	merged, err := MergeAll(tiers, origins, now)
	if err != nil {
		t.Fatalf("MergeAll() with strict_mode=false should not return error, got: %v", err)
	}

	// Policy still wins by priority
	val, ok := merged.Get("permission.network_allowlist")
	if !ok {
		t.Fatal("key 'permission.network_allowlist' not found in merged result")
	}
	if val.P.Source != SrcPolicy {
		t.Errorf("source = %v, want SrcPolicy (policy always wins by priority)", val.P.Source)
	}

	// Project path should appear in OverriddenBy (not rejected, just overridden)
	foundProject := false
	for _, path := range val.P.OverriddenBy {
		if path == ".moai/config/config.yaml" {
			foundProject = true
		}
	}
	if !foundProject {
		t.Errorf("OverriddenBy %v should contain project path when strict_mode=false", val.P.OverriddenBy)
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
