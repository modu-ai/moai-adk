package gar

// NegativeCases collects non-error returns that the over-broad `return $ERR`
// rule incorrectly matched before SPEC-GATE-ASTGREP-REPAIR-001 M1.
// After refinement, NONE of these should match go-error-not-wrapped.
func NegativeInt() int              { return 0 }
func NegativeBoolTrue() bool        { return true }
func NegativeBoolFalse() bool       { return false }
func NegativeNilError() error       { return nil }
func NegativeEmptyString() string   { return "" }
func NegativeLiteralString() string { return "literal" }
func NegativeInt42() int            { return 42 }
