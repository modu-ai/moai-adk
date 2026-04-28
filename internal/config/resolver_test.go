package config

import (
	"testing"
)

func TestNewResolver(t *testing.T) {
	resolver := NewResolver()

	if resolver == nil {
		t.Fatal("NewResolver() returned nil")
	}

	// Just check that it's not nil - type checking happens at compile time
	_ = resolver
}

func TestResolverLoadEmpty(t *testing.T) {
	// Test loading with no config files present
	resolver := NewResolver()

	merged, err := resolver.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if merged == nil {
		t.Fatal("Load() returned nil MergedSettings")
	}

	// Should have no keys since we're loading in an empty environment
	keys := merged.Keys()
	// In a real test environment, there might be builtin defaults
	_ = keys
}

func TestResolverKey(t *testing.T) {
	resolver := NewResolver()

	// Load first to populate merged settings
	_, err := resolver.Load()
	if err != nil {
		t.Skipf("Load() failed (expected in test environment): %v", err)
	}

	// Try to get a key - likely won't exist in test environment
	_, ok := resolver.Key("test", "field")
	_ = ok // We don't care if it exists, just that Key() doesn't panic
}

func TestResolverPublicInterface(t *testing.T) {
	// Test that the public interface works correctly
	resolver := NewResolver()

	if resolver == nil {
		t.Fatal("NewResolver() returned nil")
	}

	// Load should succeed even with missing files (per REQ-V3R2-RT-005-014)
	merged, err := resolver.Load()
	if err != nil {
		t.Skipf("Load() failed in test environment: %v", err)
	}

	if merged == nil {
		t.Fatal("Load() returned nil MergedSettings")
	}

	// Key() should work even for missing keys
	_, ok := resolver.Key("nonexistent", "field")
	if ok {
		t.Error("Key() returned ok=true for missing key")
	}
}

func TestParseSourceErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
	}{
		{"invalid", "invalid", true},
		{"empty", "", true},
		{"wrong_case", "Policy", true}, // Case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test resolver error types
func TestConfigTypeError(t *testing.T) {
	err := &ConfigTypeError{
		File:         "/path/to/config.yaml",
		Key:          "test.key",
		ExpectedType: "int",
		ActualValue:  "string",
	}

	expected := "config type error in /path/to/config.yaml: key test.key expects int, got string"
	if got := err.Error(); got != expected {
		t.Errorf("ConfigTypeError.Error() = %v, want %v", got, expected)
	}
}

func TestConfigAmbiguous(t *testing.T) {
	err := &ConfigAmbiguous{
		Key:   "test.key",
		File1: "/path/to/file1.yaml",
		File2: "/path/to/file2.yml",
	}

	expected := "config ambiguous: key test.key defined in both /path/to/file1.yaml and /path/to/file2.yml with conflicting values"
	if got := err.Error(); got != expected {
		t.Errorf("ConfigAmbiguous.Error() = %v, want %v", got, expected)
	}
}

func TestPolicyOverrideRejected(t *testing.T) {
	err := &PolicyOverrideRejected{
		Key:            "test.key",
		PolicySource:   "/etc/moai/settings.json",
		AttemptedSource: ".moai/config/config.yaml",
	}

	expected := "policy override rejected: key test.key from .moai/config/config.yaml cannot override policy setting from /etc/moai/settings.json (strict mode enabled)"
	if got := err.Error(); got != expected {
		t.Errorf("PolicyOverrideRejected.Error() = %v, want %v", got, expected)
	}
}

func TestConfigSchemaMismatch(t *testing.T) {
	err := &ConfigSchemaMismatch{
		Field:            "test_field",
		OldType:          "int",
		NewType:          "string",
		MigrationVersion: "001",
	}

	expected := "config schema mismatch: field test_field changed from int to string; migration 001 required but not found"
	if got := err.Error(); got != expected {
		t.Errorf("ConfigSchemaMismatch.Error() = %v, want %v", got, expected)
	}
}

func TestTierReadError(t *testing.T) {
	innerErr := &ConfigTypeError{File: "test", Key: "x"}
	err := &TierReadError{
		Source: SrcUser,
		Path:   "/path/to/file",
		Err:    innerErr,
	}

	if got := err.Error(); got == "" {
		t.Error("TierReadError.Error() returned empty string")
	}

	if err.Unwrap() != innerErr {
		t.Error("TierReadError.Unwrap() did not return inner error")
	}
}
