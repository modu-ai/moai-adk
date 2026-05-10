package config

import (
	"encoding/json"
	"time"
)

// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-005 REQ-002, REQ-003 — typed-state contract; every config field that needs to answer "where from?" wraps via Value[T].
// @MX:REASON: Generic wrapper is the type-safety contract; consumers reach via cfg.<field>.V.
//
// Provenance tracks where a configuration value came from.
// Every config value carries provenance to answer "which file set this?"
// without needing to grep through files.
type Provenance struct {
	// Source is the tier that provided this value (e.g., SrcPolicy, SrcUser).
	Source Source

	// Origin is the absolute file path where this value was defined.
	// For builtin defaults, this is "internal/config/defaults.go".
	Origin string

	// Loaded is the timestamp when this value was read from disk.
	// For builtin defaults, this is the process start time.
	Loaded time.Time

	// SchemaVersion is the schema version from the file that provided this value.
	// Zero means no schema version was specified.
	SchemaVersion int

	// OverriddenBy contains paths of lower tiers that had values for this key
	// but were overridden by a higher-priority tier.
	// Only populated when there are actual overrides.
	OverriddenBy []string
}

// Value wraps a configuration value with its provenance.
// The generic type T allows any configuration type to carry provenance.
type Value[T any] struct {
	// V is the actual configuration value.
	V T

	// P is the provenance information for this value.
	P Provenance
}

// Unwrap returns the underlying value.
// This is a convenience method for accessing V directly.
func (v Value[T]) Unwrap() T {
	return v.V
}

// Origin returns the file path where this value came from.
// This is a convenience method for accessing P.Origin.
func (v Value[T]) Origin() string {
	return v.P.Origin
}

// NewValue creates a new Value with the given value and provenance.
func NewValue[T any](v T, p Provenance) Value[T] {
	return Value[T]{
		V: v,
		P: p,
	}
}

// IsBuiltin returns true if this value came from the builtin defaults tier.
func (v Value[T]) IsBuiltin() bool {
	return v.P.Source == SrcBuiltin
}

// IsDefault returns true if this value is a builtin default (human-readable alias).
// This is used in `moai doctor config dump` output.
func (v Value[T]) IsDefault() bool {
	return v.IsBuiltin()
}

// provenanceJSON is the JSON-serializable representation of Provenance.
// It uses Source.String() for the human-readable source name instead of the numeric int,
// ensuring stable and readable JSON output.
type provenanceJSON struct {
	Source        string   `json:"source"`
	Origin        string   `json:"origin"`
	Loaded        string   `json:"loaded"`
	SchemaVersion int      `json:"schema_version,omitempty"`
	OverriddenBy  []string `json:"overridden_by,omitempty"`
}

// MarshalJSON implements json.Marshaler for Provenance.
// Source is serialized using Source.String() (e.g., "policy", "user") instead of the
// raw int value, producing stable and human-readable JSON for `moai doctor config dump`.
//
// REQ-V3R2-RT-005-006, AC-02: every key has {V, P: {Source, Origin, Loaded}} in JSON.
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M3 GREEN — Source.String() in MarshalJSON prevents int leakage.
func (p Provenance) MarshalJSON() ([]byte, error) {
	pj := provenanceJSON{
		Source:        p.Source.String(),
		Origin:        p.Origin,
		Loaded:        p.Loaded.Format(time.RFC3339),
		SchemaVersion: p.SchemaVersion,
		OverriddenBy:  p.OverriddenBy,
	}
	return json.Marshal(pj)
}
