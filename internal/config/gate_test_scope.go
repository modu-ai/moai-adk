package config

import "strings"

// gate_test_scope.go — the accepted values of gate.pre_commit.test_scope.mode
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-011).

const (
	// GateTestScopeAffected tests the packages that own a changed file plus
	// their direct dependents. The default.
	GateTestScopeAffected = "affected"
	// GateTestScopeFull tests every package in the module — the scope the
	// pre-commit context had before the affected mode existed.
	GateTestScopeFull = "full"

	// GateTestScopeModeKey is the dotted gate.yaml path of the setting, as it
	// appears in rejection messages.
	GateTestScopeModeKey = "gate.pre_commit.test_scope.mode"
)

// ValidateGateTestScopeMode reports whether mode is one of the accepted
// values. An unknown value is rejected with an error naming the key and the
// accepted values; there is no silent fallback to either scope. The empty
// string is the zero value of a Config that never set the key, so it passes
// and every consumer resolves it to GateTestScopeAffected (the default).
func ValidateGateTestScopeMode(mode string) []ValidationError {
	switch mode {
	case "", GateTestScopeAffected, GateTestScopeFull:
		return nil
	}
	return []ValidationError{{
		Field:   GateTestScopeModeKey,
		Message: "must be one of: " + strings.Join([]string{GateTestScopeAffected, GateTestScopeFull}, ", "),
		Value:   mode,
		Wrapped: ErrInvalidConfig,
	}}
}
