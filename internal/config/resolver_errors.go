package config

import "fmt"

// @MX:NOTE: [AUTO] Resolver-specific error types for 8-tier configuration loading and validation
//
// ConfigTypeError is returned when a value cannot be parsed against its typed schema.
// This maps to REQ-V3R2-RT-005-013.
type ConfigTypeError struct {
	File          string
	Key           string
	ExpectedType  string
	ActualValue   any
}

func (e *ConfigTypeError) Error() string {
	return fmt.Sprintf("config type error in %s: key %s expects %s, got %v",
		e.File, e.Key, e.ExpectedType, e.ActualValue)
}

// ConfigAmbiguous is returned when two sibling files in the same tier define
// the same key with conflicting values.
// This maps to REQ-V3R2-RT-005-041.
type ConfigAmbiguous struct {
	Key   string
	File1 string
	File2 string
}

func (e *ConfigAmbiguous) Error() string {
	return fmt.Sprintf("config ambiguous: key %s defined in both %s and %s with conflicting values",
		e.Key, e.File1, e.File2)
}

// PolicyOverrideRejected is returned when a lower tier attempts to override
// a policy-designated key while policy.strict_mode is true.
// This maps to REQ-V3R2-RT-005-022.
type PolicyOverrideRejected struct {
	Key            string
	PolicySource   string
	AttemptedSource string
}

func (e *PolicyOverrideRejected) Error() string {
	return fmt.Sprintf("policy override rejected: key %s from %s cannot override policy setting from %s (strict mode enabled)",
		e.Key, e.AttemptedSource, e.PolicySource)
}

// ConfigSchemaMismatch is returned when a value's Go type changes between
// schema versions without a migration.
// This maps to REQ-V3R2-RT-005-042.
type ConfigSchemaMismatch struct {
	Field             string
	OldType           string
	NewType           string
	MigrationVersion  string
}

func (e *ConfigSchemaMismatch) Error() string {
	return fmt.Sprintf("config schema mismatch: field %s changed from %s to %s; migration %s required but not found",
		e.Field, e.OldType, e.NewType, e.MigrationVersion)
}

// TierReadError is returned when a tier file exists but cannot be opened.
// This maps to REQ-V3R2-RT-005-040.
type TierReadError struct {
	Source Source
	Path   string
	Err    error
}

func (e *TierReadError) Error() string {
	return fmt.Sprintf("tier read error: cannot read %s from %s: %v", e.Source, e.Path, e.Err)
}

// Unwrap returns the underlying error for use with errors.Is/As.
func (e *TierReadError) Unwrap() error {
	return e.Err
}
